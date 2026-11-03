package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/shared/types"
	"github.com/gorilla/websocket"
)

type testGateway string

func (g testGateway) Gateway(context.Context) (string, error)     { return string(g), nil }
func (g testGateway) AccessToken(context.Context) (string, error) { return "test-token", nil }

func testAdapter(t *testing.T, handler func(*websocket.Conn) error, receive Receive) (*WebSocketAdapter, <-chan error) {
	t.Helper()
	outcomes := make(chan error, 16)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(rw, r, nil)
		if err != nil {
			outcomes <- err
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		outcomes <- handler(conn)
	}))
	t.Cleanup(server.Close)
	cfg := (config.Config{}).WithDefaults().Connection.WebSocket
	cfg.Reconnect.InitialDelay = time.Millisecond
	cfg.Reconnect.MaxDelay = 5 * time.Millisecond
	return NewWebSocket(cfg, testGateway("ws"+strings.TrimPrefix(server.URL, "http")), receive, 16), outcomes
}

func send(c *websocket.Conn, op int, seq int64, kind string, data any) error {
	return c.WriteJSON(map[string]any{"op": op, "s": seq, "t": kind, "d": data})
}
func helloAndAuth(c *websocket.Conn, interval int) (types.Payload, error) {
	if err := send(c, 10, 0, "", map[string]int{"heartbeat_interval": interval}); err != nil {
		return types.Payload{}, err
	}
	var p types.Payload
	err := c.ReadJSON(&p)
	return p, err
}
func checkServer(t *testing.T, outcomes <-chan error, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		select {
		case err := <-outcomes:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("mock gateway did not finish")
		}
	}
}

func TestWebSocketIdentifyResumeAndInvalidSession(t *testing.T) {
	for _, invalidate := range []bool{false, true} {
		t.Run(fmt.Sprint(invalidate), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var connections atomic.Int32
			var delivered atomic.Int32
			w, outcomes := testAdapter(t, func(c *websocket.Conn) error {
				n := connections.Add(1)
				auth, err := helloAndAuth(c, 1000)
				if err != nil {
					return err
				}
				expected := types.IdentifyOp
				if n == 2 {
					expected = types.ResumeOp
				}
				if auth.Op != expected {
					return fmt.Errorf("connection %d op %d, want %d", n, auth.Op, expected)
				}
				var credentials struct {
					Token   string `json:"token"`
					Seq     int64  `json:"seq"`
					Intents uint32 `json:"intents"`
				}
				if err := json.Unmarshal(auth.D, &credentials); err != nil {
					return err
				}
				expectedToken := "QQBot test-token"
				if auth.Op == types.ResumeOp {
					expectedToken = "test-token"
				}
				if credentials.Token != expectedToken {
					return errors.New("incorrect gateway token format")
				}
				if n == 2 && credentials.Seq != 2 {
					return fmt.Errorf("resume seq %d, want 2", credentials.Seq)
				}
				if n == 2 && invalidate {
					return send(c, 9, 0, "", false)
				}
				if auth.Op == types.IdentifyOp {
					if credentials.Intents != 1<<25 {
						return errors.New("incorrect intents")
					}
					if err := send(c, 0, 1, "READY", map[string]string{"session_id": "session"}); err != nil {
						return err
					}
				} else if err := send(c, 0, 3, "RESUMED", ""); err != nil {
					return err
				}
				if err := send(c, 0, 2, "C2C_MESSAGE_CREATE", map[string]string{"content": "/hello"}); err != nil {
					return err
				}
				// Wait for the receiver before forcing reconnect, so its completed checkpoint is observable.
				// 强制重连前等待接收方完成，以便观察已完成的检查点。
				for {
					var heartbeat types.Payload
					if err := c.ReadJSON(&heartbeat); err != nil {
						if ctx.Err() != nil {
							return nil
						}
						return err
					}
					if delivered.Load() == 1 && n == 1 {
						return send(c, 7, 0, "", nil)
					}
				}
			}, func(p *types.Payload, complete func()) error {
				if string(p.D) == "" || p.T != "C2C_MESSAGE_CREATE" {
					return errors.New("bad event")
				}
				complete()
				if delivered.Add(1) == 2 {
					cancel()
				}
				return nil
			})
			err := w.Run(ctx)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Run: %v", err)
			}
			expected := 2
			if invalidate {
				expected = 3
			}
			checkServer(t, outcomes, expected)
			if int(connections.Load()) != expected || delivered.Load() != 2 {
				t.Fatal("unexpected reconnect or dispatch count")
			}
		})
	}
}

func TestWebSocketOverloadRetainsCheckpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rejected := errors.New("queue full")
	w, outcomes := testAdapter(t, func(c *websocket.Conn) error {
		if _, err := helloAndAuth(c, 1000); err != nil {
			return err
		}
		if err := send(c, 0, 1, "READY", map[string]string{"session_id": "s"}); err != nil {
			return err
		}
		return send(c, 0, 2, "C2C_MESSAGE_CREATE", map[string]string{"content": "/hello"})
	}, func(*types.Payload, func()) error { return rejected })
	state := &session{watermark: &dispatchWatermark{}}
	_, err := w.connect(ctx, state)
	if !errors.Is(err, rejected) {
		t.Fatalf("error %v", err)
	}
	if seq := state.watermark.checkpoint(); seq == nil || *seq != 1 {
		t.Fatalf("checkpoint advanced past rejected event: %v", seq)
	}
	checkServer(t, outcomes, 1)
}

func TestWebSocketHeartbeatAndMissingACK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w, outcomes := testAdapter(t, func(c *websocket.Conn) error {
		if _, err := helloAndAuth(c, 20); err != nil {
			return err
		}
		if err := send(c, 0, 1, "READY", map[string]string{"session_id": "s"}); err != nil {
			return err
		}
		var heartbeat types.Payload
		if err := c.ReadJSON(&heartbeat); err != nil {
			return err
		}
		if heartbeat.Op != types.HeartbeatOp || string(heartbeat.D) != "1" {
			return fmt.Errorf("bad heartbeat: %+v", heartbeat)
		}
		// No ACK: the adapter must close the socket at the next heartbeat boundary.
		// 不发送 ACK：适配器应在下一个心跳边界关闭套接字。
		err := c.ReadJSON(&heartbeat)
		if err == nil {
			return errors.New("expected connection close")
		}
		return nil
	}, func(*types.Payload, func()) error { return nil })
	_, err := w.connect(ctx, &session{watermark: &dispatchWatermark{}})
	if err == nil || !strings.Contains(err.Error(), "ACK timed out") {
		t.Fatalf("unexpected error: %v", err)
	}
	checkServer(t, outcomes, 1)
}

func TestWebSocketIdentifyRejectionIsTerminal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w, outcomes := testAdapter(t, func(c *websocket.Conn) error {
		if _, err := helloAndAuth(c, 1000); err != nil {
			return err
		}
		return send(c, 9, 0, "", false)
	}, func(*types.Payload, func()) error { return nil })
	if err := w.Run(ctx); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}
	checkServer(t, outcomes, 1)
}

func TestProgressWaitsForEarlierCompletion(t *testing.T) {
	p := &dispatchWatermark{}
	first, second := int64(5), int64(9)
	doneFirst, _ := p.tryAdd(&first, 2)
	doneSecond, _ := p.tryAdd(&second, 2)
	doneSecond()
	if p.checkpoint() != nil {
		t.Fatal("checkpoint skipped unfinished work")
	}
	doneFirst()
	if *p.checkpoint() != 9 {
		t.Fatal("checkpoint did not advance")
	}
	doneFirst()
	doneSecond() // completion is safe if called twice.
	// 重复调用完成回调也是安全的。
}
func TestIdentifyRetainsIPRejectionReason(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w, outcomes := testAdapter(t, func(c *websocket.Conn) error {
		if _, err := helloAndAuth(c, 1000); err != nil {
			return err
		}
		if err := send(c, 9, 0, "", false); err != nil {
			return err
		}
		return c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4924, "session ip forbidden"))
	}, func(*types.Payload, func()) error { return nil })
	err := w.Run(ctx)
	if err == nil || !strings.Contains(err.Error(), "4924") || !strings.Contains(err.Error(), "session ip forbidden") {
		t.Fatalf("lost platform rejection reason: %v", err)
	}
	if strings.Contains(err.Error(), "credentials and intents") {
		t.Fatal("misleading rejection reason")
	}
	checkServer(t, outcomes, 1)
}
