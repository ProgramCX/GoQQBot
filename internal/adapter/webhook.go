// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ProgramCX/GoQQBot/internal/foundation/config"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type WebhookAdapter struct {
	config    config.Config
	validator Validator
	server    *http.Server
	onReceive func(payload *types.Payload)
}

// New WebhookAdapter Constructor
func New(cfg config.Config, validator Validator, onReceive func(*types.Payload)) *WebhookAdapter {
	return &WebhookAdapter{
		config:    cfg,
		validator: validator,
		onReceive: onReceive,
	}
}

// Start webhook server
func (web *WebhookAdapter) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", web.handleFunc)
	web.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", web.config.Server.Host, web.config.Server.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	logger.INFO("Webhook server started at %s:%d", web.config.Server.Host, web.config.Server.Port)
	return web.server.ListenAndServe()
}

// Shutdown webhook server
func (web *WebhookAdapter) Shutdown(ctx context.Context) error {
	return web.server.Shutdown(ctx)
}

// handleWebhook handle webhook request
func (web *WebhookAdapter) handleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.ERROR("Error in Read Request Body: %v", err)
		return
	}
	defer r.Body.Close()

	var payload types.Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Handle Verification Request
	if payload.Op == types.VerifyOp {
		web.handleValidation(w, payload.D)
		return
	}

	// call back
	if web.onReceive != nil {
		web.onReceive(&payload)
	} else {
		logger.WARN("onReceive callback is nil, dropping payload")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		logger.ERROR("Error in responding to request server！:%v", err)
		return
	}
}

// handle webhook validation
// See: https://q.qq.com/qqbot/#/developer/webhook-setting
func (web *WebhookAdapter) handleValidation(w http.ResponseWriter, data []byte) {
	resp, err := web.validator.Validate(data)
	if err != nil {
		http.Error(w, "validation failed", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.ERROR("Parse Json Error: %v", err)
		return
	}
	logger.INFO("Validation Succeed！")
}
