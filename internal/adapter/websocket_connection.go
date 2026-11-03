package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/shared/types"
	"github.com/gorilla/websocket"
)

type connState uint8

const (
	stateIdentifying connState = iota
	stateResuming
	stateAwaitingRejectionClose
	stateReady
)

type heartbeatState uint8

const (
	heartbeatIdle heartbeatState = iota
	heartbeatAwaitingACK
)

// gatewayConnection is owned by one connection loop. Only the reader goroutine
// gatewayConnection 由单个连接循环持有，只有读取协程
// reads the socket; all protocol transitions and writes occur on the loop goroutine.
// 读取套接字；所有协议状态转换及写入均在连接循环协程中执行。
type gatewayConnection struct {
	adapter   *WebSocketAdapter
	conn      *websocket.Conn
	session   *session
	token     string
	phase     connState
	heartbeat heartbeatState
	received  *int64
	login     *time.Timer
}

func (c *gatewayConnection) write(op types.OpType, data any) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.adapter.config.DialTimeout)); err != nil {
		return err
	}
	return c.conn.WriteJSON(struct {
		Op   types.OpType `json:"op"`
		Data any          `json:"d"`
	}{op, data})
}

func (c *gatewayConnection) authenticate(checkpoint *int64) error {
	var err error
	if c.session.id != "" && checkpoint != nil {
		c.phase = stateResuming
		// Resume uses the raw access token; Identify uses the QQBot prefix.
		// Resume 使用原始访问令牌；Identify 使用 QQBot 前缀。
		err = c.write(types.ResumeOp, map[string]any{"token": c.token, "session_id": c.session.id, "seq": *checkpoint})
	} else {
		intents, maskErr := config.IntentMask(c.adapter.config.Intents)
		if maskErr != nil {
			return &terminalError{maskErr}
		}
		err = c.write(types.IdentifyOp, map[string]any{"token": "QQBot " + c.token, "intents": intents, "shard": [2]int{c.adapter.config.Shard.ID, c.adapter.config.Shard.Total}})
	}
	if err != nil {
		return fmt.Errorf("gateway authentication write failed: %w", err)
	}
	return nil
}

type gatewayFrame struct {
	payload types.Payload
	err     error
}

func readGatewayFrames(ctx context.Context, conn *websocket.Conn, frames chan<- gatewayFrame, done chan<- struct{}) {
	defer close(done)
	for {
		var frame gatewayFrame
		frame.err = conn.ReadJSON(&frame.payload)
		select {
		case frames <- frame:
		case <-ctx.Done():
			return
		}
		if frame.err != nil {
			return
		}
	}
}

func (c *gatewayConnection) run(ctx context.Context, interval time.Duration) error {
	frames := make(chan gatewayFrame, 1)
	readerDone := make(chan struct{})
	readCtx, cancelRead := context.WithCancel(ctx)
	go readGatewayFrames(readCtx, c.conn, frames, readerDone)
	defer func() { cancelRead(); _ = c.conn.Close(); <-readerDone }()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	c.login = time.NewTimer(c.adapter.config.DialTimeout)
	defer c.login.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.login.C:
			return c.loginTimeout()
		case <-ticker.C:
			if err := c.heartbeatTick(); err != nil {
				return err
			}
		case frame := <-frames:
			if frame.err != nil {
				return c.readFailure(frame.err)
			}
			if err := c.handleFrame(frame.payload); err != nil {
				return err
			}
		}
	}
}

func (c *gatewayConnection) loginTimeout() error {
	if c.phase == stateAwaitingRejectionClose {
		return &terminalError{errors.New("gateway Identify rejected (op 9); server supplied no close reason")}
	}
	return errors.New("gateway login timed out")
}

func (c *gatewayConnection) sendHeartbeat() error {
	if err := c.write(types.HeartbeatOp, c.received); err != nil {
		return fmt.Errorf("gateway heartbeat write failed: %w", err)
	}
	c.heartbeat = heartbeatAwaitingACK
	return nil
}

func (c *gatewayConnection) heartbeatTick() error {
	if c.phase == stateAwaitingRejectionClose {
		return nil
	}
	if c.heartbeat == heartbeatAwaitingACK {
		return errors.New("gateway heartbeat ACK timed out")
	}
	return c.sendHeartbeat()
}

func (c *gatewayConnection) handleFrame(p types.Payload) error {
	// Once Identify is rejected, only the close reason or grace timeout can end it.
	// Identify 被拒绝后，只能由关闭原因或等待超时结束该状态。
	// In particular, a late READY or heartbeat request must not revive this login.
	// 特别是迟到的 READY 或心跳请求不得恢复这次登录。
	if c.phase == stateAwaitingRejectionClose {
		return nil
	}
	if p.S != nil {
		value := *p.S
		c.received = &value
	}
	switch p.Op {
	case types.HeartbeatACKOp:
		c.heartbeat = heartbeatIdle
	case types.HeartbeatOp:
		return c.sendHeartbeat()
	case types.ReconnectOp:
		return errors.New("gateway requested reconnect")
	case types.InvalidSessionOp:
		return c.invalidSession(p.D)
	case types.EventOp:
		return c.dispatch(p)
	}
	return nil
}

func (c *gatewayConnection) invalidSession(data json.RawMessage) error {
	var resumable bool
	if err := json.Unmarshal(data, &resumable); err != nil {
		return &terminalError{errors.New("invalid gateway session response")}
	}
	if !resumable {
		c.session.invalidate()
	}
	if c.phase == stateIdentifying {
		// QQ can reject Identify with op 9 before sending the specific close code.
		// QQ 可能先用 op 9 拒绝 Identify，再发送具体关闭码。
		c.phase = stateAwaitingRejectionClose
		c.login.Reset(min(2*time.Second, c.adapter.config.DialTimeout))
		return nil
	}
	return errors.New("gateway session invalidated")
}

func (c *gatewayConnection) dispatch(p types.Payload) error {
	switch p.T {
	case "READY":
		var data struct {
			SessionID string `json:"session_id"`
		}
		if json.Unmarshal(p.D, &data) != nil || data.SessionID == "" {
			return &terminalError{errors.New("invalid gateway READY")}
		}
		c.session.id = data.SessionID
	case "RESUMED":
	default:
		if c.session.id == "" {
			return &terminalError{errors.New("gateway dispatched event before READY")}
		}
		complete, ok := c.session.watermark.tryAdd(p.S, c.adapter.maxPending)
		if !ok {
			return errors.New("gateway processing window is full")
		}
		if err := c.adapter.receive(&p, complete); err != nil {
			return fmt.Errorf("submit gateway event: %w", err)
		}
		return nil
	}
	c.session.authRetry = false
	c.phase = stateReady
	c.login.Stop()
	// READY/RESUMED are protocol events, not outstanding user deliveries. They
	// READY/RESUMED 是协议事件，不属于待处理的业务投递；它们
	// still join the ordered watermark so replayed work cannot be skipped.
	// 仍参与按投递顺序推进的水位线，以免跳过需要重放的任务。
	complete, ok := c.session.watermark.tryAdd(p.S, c.adapter.maxPending)
	if !ok {
		return errors.New("gateway processing window is full")
	}
	complete()
	return nil
}
