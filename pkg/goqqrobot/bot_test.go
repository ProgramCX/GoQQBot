package goqqrobot_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func configuration() config.Config {
	return config.Config{Bot: config.Bot{AppId: "test", AppSecret: "example-secret"}, Connection: config.Connection{Mode: "websocket"}}
}

func TestRunCancellationConcurrentShutdownAndRestart(t *testing.T) {
	requested := make(chan struct{})
	client := api.NewClient("test", "example-secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		close(requested)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}))
	bot, err := goqqrobot.NewWithConfig(configuration(), nil, goqqrobot.WithAPIClient(client))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- bot.Start() }()
	select {
	case <-requested:
	case <-time.After(5 * time.Second):
		t.Fatal("bot did not start")
	}
	if err := bot.Start(); !errors.Is(err, goqqrobot.ErrAlreadyStarted) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdowns := make(chan error, 4)
	for i := 0; i < 4; i++ {
		go func() { shutdowns <- bot.Shutdown(ctx) }()
	}
	for i := 0; i < 4; i++ {
		if err := <-shutdowns; err != nil {
			t.Fatal(err)
		}
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if err := bot.Start(); !errors.Is(err, goqqrobot.ErrClosed) {
		t.Fatal(err)
	}
}

func TestStartupFailureReturnsWithoutSignal(t *testing.T) {
	client := api.NewClient("test", "example-secret", api.WithHTTPClient(&http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":100016}`)), Header: make(http.Header)}, nil
	})}))
	bot, err := goqqrobot.NewWithConfig(configuration(), nil, goqqrobot.WithAPIClient(client))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- bot.StartWithGracefulShutdown() }()
	select {
	case err := <-result:
		var failure *api.APIError
		if !errors.As(err, &failure) || failure.Code != 100016 {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("startup error was hidden while waiting for a signal")
	}
	if err := bot.Stop(); err != nil {
		t.Fatal(err)
	}
}
func TestShutdownBeforeRunAndInputOwnership(t *testing.T) {
	cfg := configuration()
	cfg.Connection.WebSocket.Intents = []string{"GROUP_AND_C2C_EVENT"}
	bot, err := goqqrobot.NewWithConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Connection.WebSocket.Intents[0] = "INVALID"
	if err := bot.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := bot.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := bot.Start(); !errors.Is(err, goqqrobot.ErrClosed) {
		t.Fatal(err)
	}
}
