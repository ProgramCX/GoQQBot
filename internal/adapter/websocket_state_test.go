package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/shared/types"
	"github.com/gorilla/websocket"
)

func TestDispatchWatermarkAtomicCapacity(t *testing.T) {
	p := &dispatchWatermark{}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := p.tryAdd(nil, 7); ok {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 7 {
		t.Fatalf("accepted %d deliveries, want 7", accepted.Load())
	}
}

func TestSessionReconnectIsolatesLateCompletions(t *testing.T) {
	s := &session{}
	s.invalidate()
	s.id = "old-session"
	first, second := int64(5), int64(9)
	done, _ := s.watermark.tryAdd(&first, 2)
	done()
	late, _ := s.watermark.tryAdd(&second, 2)
	if got := s.beginConnection(); got == nil || *got != 5 {
		t.Fatalf("checkpoint = %v", got)
	}
	late()
	if got := s.watermark.checkpoint(); got == nil || *got != 5 {
		t.Fatal("old completion advanced new connection")
	}
	s.authRetry = true
	s.invalidate()
	if s.id != "" || s.watermark.checkpoint() != nil || !s.authRetry {
		t.Fatal("invalidation retained session or allowed another token refresh")
	}
}

func TestIdentifyRejectionWaitState(t *testing.T) {
	c := &gatewayConnection{adapter: &WebSocketAdapter{}, session: &session{}, phase: stateIdentifying}
	c.session.invalidate()
	c.adapter.config.DialTimeout = time.Second
	c.login = time.NewTimer(time.Hour)
	defer c.login.Stop()
	if err := c.invalidSession(json.RawMessage(`false`)); err != nil {
		t.Fatal(err)
	}
	if c.phase != stateAwaitingRejectionClose {
		t.Fatal("Identify rejection did not enter close-wait state")
	}
	for _, p := range []types.Payload{
		{Op: types.InvalidSessionOp, D: json.RawMessage(`false`)},
		{Op: types.EventOp, T: "READY", D: json.RawMessage(`{"session_id":"late"}`)},
		{Op: types.HeartbeatOp},
	} {
		if err := c.handleFrame(p); err != nil {
			t.Fatal(err)
		}
	}
	if c.phase != stateAwaitingRejectionClose || c.session.id != "" {
		t.Fatal("late frame revived rejected login")
	}
	if err := c.heartbeatTick(); err != nil {
		t.Fatal(err)
	}
	var terminal *terminalError
	if !errors.As(c.loginTimeout(), &terminal) {
		t.Fatal("rejection without close reason must terminate")
	}
}

type refreshGateway struct {
	testGateway
	refreshes int
}

func (g *refreshGateway) InvalidateToken(string) { g.refreshes++ }

func TestGatewayCloseClassification(t *testing.T) {
	for _, code := range []int{closeInvalidShard, closeShardingRequired, closeInvalidVersion, closeInvalidIntents, closeUnauthorizedIntent, closeBotOffline, closeBotBanned, closeSessionIPForbidden, closeInvalidSession, closeInvalidSequence} {
		c := &gatewayConnection{adapter: &WebSocketAdapter{}, session: &session{id: "s", watermark: &dispatchWatermark{}}, phase: stateReady}
		err := c.readFailure(&websocket.CloseError{Code: code, Text: "private-close-text"})
		var terminal *terminalError
		wantTerminal := code != closeInvalidSession && code != closeInvalidSequence
		if errors.As(err, &terminal) != wantTerminal || strings.Contains(err.Error(), "private-close-text") {
			t.Fatalf("close %d: %v", code, err)
		}
		if !wantTerminal && c.session.id != "" {
			t.Fatalf("close %d retained invalid session", code)
		}
	}
	g := &refreshGateway{}
	c := &gatewayConnection{adapter: &WebSocketAdapter{client: g}, session: &session{}, phase: stateAwaitingRejectionClose}
	var terminal *terminalError
	if err := c.readFailure(&websocket.CloseError{Code: closeInvalidToken}); errors.As(err, &terminal) {
		t.Fatal("first token rejection should refresh")
	}
	if err := c.readFailure(&websocket.CloseError{Code: closeInvalidToken}); !errors.As(err, &terminal) || g.refreshes != 1 {
		t.Fatal("token rejection retried more than once")
	}
}

func TestGatewayWritesPreserveCause(t *testing.T) {
	w, outcomes := testAdapter(t, func(conn *websocket.Conn) error {
		if err := send(conn, 10, 0, "", map[string]int{"heartbeat_interval": 1000}); err != nil {
			return err
		}
		var p types.Payload
		_ = conn.ReadJSON(&p)
		return nil
	}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, token, err := w.dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.readHello(conn); err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	c := &gatewayConnection{adapter: w, conn: conn, token: token, session: &session{}}
	for _, err := range []error{c.authenticate(nil), c.sendHeartbeat()} {
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("write lost closed-connection cause: %v", err)
		}
	}
	checkServer(t, outcomes, 1)
}

func TestIdentifyRejectionWithoutCloseReasonTimesOut(t *testing.T) {
	w, outcomes := testAdapter(t, func(conn *websocket.Conn) error {
		if _, err := helloAndAuth(conn, 1000); err != nil {
			return err
		}
		if err := send(conn, 9, 0, "", false); err != nil {
			return err
		}
		var p types.Payload
		if err := conn.ReadJSON(&p); err == nil {
			return errors.New("rejected login still sent a heartbeat")
		}
		return nil
	}, nil)
	w.config.DialTimeout = 200 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := w.Run(ctx)
	if err == nil || !strings.Contains(err.Error(), "server supplied no close reason") {
		t.Fatalf("unexpected rejection timeout: %v", err)
	}
	checkServer(t, outcomes, 1)
}

func TestGatewayHeartbeatACKAndServerRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var complete func()
	w, outcomes := testAdapter(t, func(conn *websocket.Conn) error {
		if _, err := helloAndAuth(conn, 50); err != nil {
			return err
		}
		if err := send(conn, 0, 1, "READY", map[string]string{"session_id": "s"}); err != nil {
			return err
		}
		if err := send(conn, 0, 8, "C2C_MESSAGE_CREATE", map[string]string{"content": "pending"}); err != nil {
			return err
		}
		for i := 0; i < 3; i++ {
			var heartbeat types.Payload
			if err := conn.ReadJSON(&heartbeat); err != nil {
				return err
			}
			if heartbeat.Op != types.HeartbeatOp || string(heartbeat.D) != "8" {
				return errors.New("heartbeat did not use latest received sequence")
			}
			if err := conn.WriteJSON(map[string]any{"op": types.HeartbeatACKOp}); err != nil {
				return err
			}
			if i == 0 {
				if err := conn.WriteJSON(map[string]any{"op": types.HeartbeatOp}); err != nil {
					return err
				}
			}
		}
		return send(conn, 7, 0, "", nil)
	}, func(_ *types.Payload, done func()) error { complete = done; return nil })
	s := &session{}
	ready, err := w.connect(ctx, s)
	if !ready || err == nil || err.Error() != "gateway requested reconnect" {
		t.Fatalf("connection ended unexpectedly: ready=%v err=%v", ready, err)
	}
	if got := s.watermark.checkpoint(); got == nil || *got != 1 {
		t.Fatal("heartbeat advanced past unfinished handler")
	}
	complete()
	if got := s.watermark.checkpoint(); got == nil || *got != 8 {
		t.Fatal("completion did not advance watermark")
	}
	checkServer(t, outcomes, 1)
}
