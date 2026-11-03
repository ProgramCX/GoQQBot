// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type WebhookAdapter struct {
	config    config.Config
	validator Validator
	onReceive Receive
	log       *slog.Logger
}

// New WebhookAdapter Constructor
// 创建 WebhookAdapter 的构造函数。
func New(cfg config.Config, validator Validator, onReceive Receive, logs ...*slog.Logger) *WebhookAdapter {
	web := &WebhookAdapter{
		config:    cfg,
		validator: validator,
		onReceive: onReceive,
		log:       logger.OrDiscard(nil),
	}
	if len(logs) > 0 {
		web.log = logger.OrDiscard(logs[0])
	}
	return web
}

// Run serves webhook requests until cancellation or a listener error.
// Run 提供 Webhook 服务，直到上下文取消或监听器出错。
func (web *WebhookAdapter) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", web.config.Server.Host, web.config.Server.Port))
	if err != nil {
		return fmt.Errorf("listen webhook: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", web.handleFunc)
	server := &http.Server{Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), web.config.Runtime.ShutdownTimeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			_ = server.Close()
		}
		<-result
		if err != nil {
			return fmt.Errorf("shutdown webhook: %w", err)
		}
		return ctx.Err()
	}
}

// handleWebhook handle webhook request
// handleWebhook 处理 Webhook 请求。
func (web *WebhookAdapter) handleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<20))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload types.Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Handle Verification Request (op=13)
	// 处理验证请求（op=13）。
	if payload.Op == types.VerifyOp {
		web.handleValidation(w, payload.D)
		return
	}

	// For normal events, verify Ed25519 signature
	// 对普通事件验证 Ed25519 签名。
	if !web.validator.VerifyEventSignatureFromRequest(r, body) {
		web.log.WarnContext(r.Context(), "webhook signature verification failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if web.onReceive == nil || payload.Op != types.EventOp {
		http.Error(w, "Unsupported event", http.StatusBadRequest)
		return
	}
	if err := web.onReceive(&payload, func() {}); err != nil {
		http.Error(w, "Event queue unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"op":12}`)); err != nil {
		web.log.ErrorContext(r.Context(), "write webhook acknowledgement failed")
		return
	}
}

// handle webhook validation
// 处理 Webhook 验证。
// See: https://q.qq.com/qqbot/#/developer/webhook-setting
// 参见：https://q.qq.com/qqbot/#/developer/webhook-setting
func (web *WebhookAdapter) handleValidation(w http.ResponseWriter, data []byte) {

	resp, err := web.validator.Validate(data)
	if err != nil {
		web.log.Warn("webhook validation failed")
		http.Error(w, "validation failed", http.StatusForbidden)
		return
	}

	respJSON, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(respJSON); err != nil {
		web.log.Error("write webhook validation response failed")
		return
	}

	web.log.Debug("webhook validation succeeded")
}
