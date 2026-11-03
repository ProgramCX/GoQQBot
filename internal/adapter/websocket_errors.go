package adapter

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/gorilla/websocket"
)

const (
	closeInvalidToken       = 4004
	closeInvalidSession     = 4006
	closeInvalidSequence    = 4007
	closeInvalidShard       = 4010
	closeShardingRequired   = 4011
	closeInvalidVersion     = 4012
	closeInvalidIntents     = 4013
	closeUnauthorizedIntent = 4014
	closeBotOffline         = 4914
	closeBotBanned          = 4915
	closeSessionIPForbidden = 4924

	// Keep API authentication codes distinct from WebSocket close codes.
	// API 认证错误码与 WebSocket 关闭码分别定义。
	apiAuthCode100007 = 100007
	apiAuthCode100016 = 100016
	apiAuthCode10004  = 10004
)

type terminalError struct{ err error }

func (e *terminalError) Error() string { return e.err.Error() }
func (e *terminalError) Unwrap() error { return e.err }

func isTerminalAPIError(err error) bool {
	var failure *api.APIError
	if !errors.As(err, &failure) {
		return false
	}
	if failure.StatusCode == http.StatusUnauthorized || failure.StatusCode == http.StatusForbidden {
		return true
	}
	switch failure.Code {
	case apiAuthCode100007, apiAuthCode100016, apiAuthCode10004:
		return true
	}
	return false
}

func (c *gatewayConnection) readFailure(err error) error {
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		if closeErr.Code == closeSessionIPForbidden || (c.phase == stateAwaitingRejectionClose && closeErr.Code != closeInvalidToken) {
			return &terminalError{gatewayRejection(closeErr.Code)}
		}
		switch closeErr.Code {
		case closeInvalidToken:
			if invalidator, ok := c.adapter.client.(interface{ InvalidateToken(string) }); ok && !c.session.authRetry {
				invalidator.InvalidateToken(c.token)
				c.session.authRetry = true
				return errors.New("gateway access token rejected; refreshing")
			}
			return &terminalError{errors.New("gateway authentication failed after token refresh")}
		case closeInvalidShard, closeShardingRequired, closeInvalidVersion, closeInvalidIntents, closeUnauthorizedIntent, closeBotOffline, closeBotBanned:
			return &terminalError{gatewayRejection(closeErr.Code)}
		case closeInvalidSession, closeInvalidSequence:
			c.session.invalidate()
		}
	}
	if c.phase == stateAwaitingRejectionClose {
		return &terminalError{errors.New("gateway Identify rejected (op 9); connection ended without a close reason")}
	}
	return errors.New("gateway disconnected")
}

func gatewayRejection(code int) error {
	// Do not echo arbitrary close text: external input may contain sensitive data.
	// 不回显任意关闭文本，因为外部输入可能包含敏感数据。
	reasons := map[int]string{
		closeInvalidShard: "invalid shard", closeShardingRequired: "shard configuration requires adjustment",
		closeInvalidVersion: "invalid API version", closeInvalidIntents: "invalid intents", closeUnauthorizedIntent: "intents not authorized",
		closeBotOffline: "bot is offline", closeBotBanned: "bot is banned",
		closeSessionIPForbidden: "session ip forbidden; check this machine's public egress IP against the platform IP allowlist",
	}
	reason := reasons[code]
	if reason == "" {
		reason = "server rejected Identify"
	}
	return fmt.Errorf("gateway rejected connection: close code %d (%s)", code, reason)
}
