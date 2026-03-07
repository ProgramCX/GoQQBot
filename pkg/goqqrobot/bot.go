// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package goqqrobot

import (
	"context"
	"errors"
	"net/http"

	"github.com/ProgramCX/GoQQBot/internal/adapter"
	"github.com/ProgramCX/GoQQBot/internal/core/pool"
	"github.com/ProgramCX/GoQQBot/internal/core/router"
	"github.com/ProgramCX/GoQQBot/internal/foundation"
	"github.com/ProgramCX/GoQQBot/internal/foundation/ascii"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// Handler is the handler function type for bot commands and events.
type Handler = router.Handler

type Bot struct {
	adapter *adapter.WebhookAdapter
	pool    *pool.Pool
	router  *router.CommandRouter
	config  config.Config
}

func New(cfg config.Config, validator adapter.Validator) (*Bot, error) {
	ascii.PrintLogo()
	b := &Bot{
		config: cfg,
	}

	b.router = router.New()

	b.pool = pool.New(10, 100, b.router.Handle)
	b.pool.Start()

	onReceive := func(payload *types.Payload) {
		task := &types.Task{
			Ctx:     context.Background(),
			Payload: payload,
			Done:    func() {},
		}
		b.pool.Submit(task)
	}

	b.adapter = adapter.New(cfg, validator, onReceive)

	return b, nil
}

func NewBot(configPath string) (*Bot, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	va := adapter.NewValidator(cfg.Bot.AppSecret)
	bot, err := New(*cfg, *va)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func (b *Bot) OnCommand(cmd string, eventID message.EventId, handler Handler) {

	b.router.OnCommand(cmd, eventID, handler)
}

func (b *Bot) OnEvent(eventID message.EventId, handler Handler) {
	b.router.OnEvent(eventID, handler)
}

func (b *Bot) Start() error {
	return b.adapter.Start()
}

func (b *Bot) StartWithGracefulShutdown() error {
	go func() {
		if err := b.adapter.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ERROR("Webhook server error: %v", err)
		}
	}()

	sig := <-foundation.Notify()
	logger.INFO("[%s] Signal Received. Shutting Down...", sig.String())
	return b.Stop()
}

func (b *Bot) Stop() error {
	err := b.adapter.Shutdown(context.Background())
	if err != nil {
		return err
	}
	b.pool.Stop()
	return nil
}
