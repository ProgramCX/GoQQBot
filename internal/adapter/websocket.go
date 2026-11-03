// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/shared/types"
	"github.com/gorilla/websocket"
)

type gatewayClient interface {
	Gateway(context.Context) (string, error)
	AccessToken(context.Context) (string, error)
}

// Receive submits an event without blocking and calls completed when processing finishes.
// Receive 非阻塞提交事件，并在处理结束时调用 completed。
// A non-nil return means the event was not accepted; completed must not then be called.
// 返回非 nil 表示事件未被接受，此时不得调用 completed。
type Receive func(payload *types.Payload, completed func()) error

// WebSocketAdapter maintains one gateway session. Run is called once by the owning bot.
// WebSocketAdapter 维护一个网关会话，由所属 Bot 调用一次 Run。
type WebSocketAdapter struct {
	config     config.WebSocket
	client     gatewayClient
	receive    Receive
	maxPending int
}

// NewWebSocket constructs a dormant adapter with an injected gateway client and event sink.
// NewWebSocket 使用注入的网关客户端和事件接收函数构造尚未启动的适配器。
func NewWebSocket(cfg config.WebSocket, client gatewayClient, receive Receive, maxPending int) *WebSocketAdapter {
	cfg.Intents = append([]string(nil), cfg.Intents...)
	return &WebSocketAdapter{config: cfg, client: client, receive: receive, maxPending: maxPending}
}

// Run reconnects transient failures until ctx is canceled. Authentication/configuration failures return.
// Run 在临时故障后持续重连，直到 ctx 取消；认证或配置错误直接返回。
// Session checkpoints are in memory only; handlers must tolerate at-least-once delivery.
// 会话检查点仅保存在内存中，处理函数必须容忍至少一次投递。
func (w *WebSocketAdapter) Run(ctx context.Context) error {
	state := &session{}
	state.invalidate()
	delay := w.config.Reconnect.InitialDelay
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ready, err := w.connect(ctx, state)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var terminal *terminalError
		if errors.As(err, &terminal) {
			return terminal.err
		}
		if isTerminalAPIError(err) {
			return err
		}
		if ready {
			delay = w.config.Reconnect.InitialDelay
		}
		// Jitter stays within the configured cap and prevents synchronized reconnect storms.
		// 随机抖动不超过配置上限，可避免多个实例同时重连。
		wait := delay/2 + time.Duration(rand.Int64N(max(int64(delay-delay/2), 1)))
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if delay < w.config.Reconnect.MaxDelay/2 {
			delay *= 2
		} else {
			delay = w.config.Reconnect.MaxDelay
		}
	}
}

// connect owns one socket; the connection loop owns all protocol state and writes.
// connect 管理一个连接的资源生命周期；连接循环管理全部协议状态及写入。
func (w *WebSocketAdapter) connect(ctx context.Context, state *session) (bool, error) {
	conn, token, err := w.dial(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	stopClose := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stopClose()

	interval, err := w.readHello(conn)
	if err != nil {
		return false, err
	}
	checkpoint := state.beginConnection()
	c := &gatewayConnection{
		adapter: w, conn: conn, session: state, token: token,
		received: checkpoint, phase: stateIdentifying,
	}
	if err := c.authenticate(checkpoint); err != nil {
		return false, err
	}
	err = c.run(ctx, interval)
	return c.phase == stateReady, err
}

func (w *WebSocketAdapter) dial(ctx context.Context) (*websocket.Conn, string, error) {
	endpoint, err := w.client.Gateway(ctx)
	if err != nil {
		return nil, "", err
	}
	token, err := w.client.AccessToken(ctx)
	if err != nil {
		return nil, "", err
	}
	dialer := websocket.Dialer{HandshakeTimeout: w.config.DialTimeout, Proxy: http.ProxyFromEnvironment}
	conn, resp, err := dialer.DialContext(ctx, endpoint, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			return nil, "", &terminalError{fmt.Errorf("gateway handshake rejected: HTTP %d", resp.StatusCode)}
		}
		// Dial errors can include URLs. Do not expose them in diagnostics.
		// 拨号错误可能包含 URL，不将其暴露在诊断信息中。
		return nil, "", errors.New("gateway connection failed")
	}
	conn.SetReadLimit(4 << 20)
	return conn, token, nil
}

func (w *WebSocketAdapter) readHello(conn *websocket.Conn) (time.Duration, error) {
	if err := conn.SetReadDeadline(time.Now().Add(w.config.DialTimeout)); err != nil {
		return 0, err
	}
	var hello types.Payload
	if err := conn.ReadJSON(&hello); err != nil {
		return 0, errors.New("read gateway Hello failed")
	}
	var greeting struct {
		Interval int64 `json:"heartbeat_interval"`
	}
	if hello.Op != types.HelloOp || json.Unmarshal(hello.D, &greeting) != nil || greeting.Interval < 1 || greeting.Interval > 300000 {
		return 0, &terminalError{errors.New("invalid gateway Hello")}
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return 0, err
	}
	return time.Duration(greeting.Interval) * time.Millisecond, nil
}
