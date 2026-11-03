package goqqrobot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type diagnosticTransport func(*http.Request) (*http.Response, error)

func (f diagnosticTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Exercise the real Run/pool/router/shutdown path with the network held at token refresh.
// 将网络阻塞在令牌刷新阶段，验证真实 Run/线程池/路由/关闭路径。
// Synthetic tasks enter at the same queue boundary as a decoded transport event.
// 合成任务从与已解码传输事件相同的队列边界进入。
func runningDiagnosticBot(t *testing.T, options ...Option) (*Bot, <-chan error) {
	t.Helper()
	ready := make(chan struct{})
	var once sync.Once
	client := api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: diagnosticTransport(func(r *http.Request) (*http.Response, error) {
		once.Do(func() { close(ready) })
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}))
	options = append(options, WithAPIClient(client))
	cfg := config.Config{Bot: config.Bot{AppId: "test", AppSecret: "secret"}, Connection: config.Connection{Mode: "websocket"}, Runtime: config.Runtime{Pool: config.ServerPool{Workers: 1, QueueSize: 8}}}
	bot, err := NewWithConfig(cfg, nil, options...)
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- bot.Run(context.Background()) }()
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not reach transport")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = bot.Shutdown(ctx)
	})
	return bot, result
}

func submitDiagnostic(t *testing.T, bot *Bot, text string, done func()) {
	t.Helper()
	data, _ := json.Marshal(map[string]string{"content": text})
	if err := bot.pool.Submit(&types.Task{Ctx: context.Background(), Client: bot.apiClient, Payload: &types.Payload{T: message.C2C_MESSAGE_CREATE, D: data}, Done: done}); err != nil {
		t.Fatal(err)
	}
}

func stopDiagnostic(t *testing.T, bot *Bot, result <-chan error) (error, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownErr := bot.Shutdown(ctx)
	select {
	case runErr := <-result:
		return runErr, shutdownErr
	case <-ctx.Done():
		t.Fatal("Run did not finish")
	}
	return nil, nil
}

func TestHandlerPanicIsObservableAndWorkerContinues(t *testing.T) {
	reported := make(chan error, 2)
	var successful atomic.Int32
	var completed atomic.Int32
	bot, result := runningDiagnosticBot(t, WithErrorHandler(func(ctx *types.Context, err error) {
		if ctx.Command != "panic" || len(ctx.Args) != 1 || ctx.Args[0] != "argument" {
			t.Error("panic observer lost original handler context")
		}
		reported <- err
	}))
	bot.OnCommand("panic", message.C2C_MESSAGE_CREATE, func(*types.Context) error { panic("private-panic-content") })
	bot.OnCommand("ok", message.C2C_MESSAGE_CREATE, func(*types.Context) error { successful.Add(1); return nil })
	submitDiagnostic(t, bot, "/panic argument", func() { completed.Add(1) })
	submitDiagnostic(t, bot, "/ok", func() { completed.Add(1) })
	runErr, shutdownErr := stopDiagnostic(t, bot, result)
	for _, err := range []error{<-reported, runErr, shutdownErr} {
		var failure *types.HandlerError
		var panicErr *types.PanicError
		if !errors.As(err, &failure) || !errors.As(err, &panicErr) || panicErr.Value != "private-panic-content" || len(panicErr.Stack) == 0 {
			t.Fatalf("panic was lost: %v", err)
		}
		if strings.Contains(err.Error(), "private-panic-content") {
			t.Fatal("panic payload formatted by default")
		}
	}
	if successful.Load() != 1 || completed.Load() != 2 {
		t.Fatal("panic lost a worker or task completion")
	}
}

func TestErrorWithoutObserverIsReturnedOnShutdown(t *testing.T) {
	want := errors.New("private-user-error")
	bot, result := runningDiagnosticBot(t)
	bot.OnCommand("fail", message.C2C_MESSAGE_CREATE, func(*types.Context) error { return want })
	submitDiagnostic(t, bot, "/fail", nil)
	runErr, shutdownErr := stopDiagnostic(t, bot, result)
	if !errors.Is(runErr, want) || !errors.Is(shutdownErr, want) {
		t.Fatal("error disappeared without observer/logger")
	}
	if strings.Contains(runErr.Error(), want.Error()) {
		t.Fatal("arbitrary callback data leaked through summary")
	}
}

func TestObserverPanicIsContainedAndReturnedWithoutRecursion(t *testing.T) {
	var calls atomic.Int32
	bot, result := runningDiagnosticBot(t, WithErrorHandler(func(*types.Context, error) { calls.Add(1); panic("observer-secret") }))
	bot.OnCommand("fail", message.C2C_MESSAGE_CREATE, func(*types.Context) error { return errors.New("handler error") })
	for range 3 {
		submitDiagnostic(t, bot, "/fail", nil)
	}
	runErr, _ := stopDiagnostic(t, bot, result)
	var panicErr *types.PanicError
	if calls.Load() != 3 || !errors.As(runErr, &panicErr) || panicErr.Value != "observer-secret" {
		t.Fatalf("observer failure was lost or recursed: %v", runErr)
	}
	if strings.Contains(runErr.Error(), "observer-secret") {
		t.Fatal("observer panic leaked")
	}
}

func TestLoggersArePerBotFilterableAndNeverLogPayloads(t *testing.T) {
	var first, second bytes.Buffer
	var level slog.LevelVar
	level.Set(slog.LevelError)
	log := slog.New(slog.NewTextHandler(&first, &slog.HandlerOptions{Level: &level}))
	bot, result := runningDiagnosticBot(t, WithLogger(log))
	other, err := NewWithConfig(config.Config{Bot: config.Bot{AppId: "test", AppSecret: "secret"}}, nil, WithLogger(slog.New(slog.NewTextHandler(&second, nil))))
	if err != nil {
		t.Fatal(err)
	}
	other.OnCommand("other", message.C2C_MESSAGE_CREATE, func(*types.Context) error { return nil })
	bot.OnCommand("fail", message.C2C_MESSAGE_CREATE, func(*types.Context) error { return errors.New("credential-like-private-data") })
	submitDiagnostic(t, bot, "/fail secret-argument", nil)
	_, _ = stopDiagnostic(t, bot, result)
	if !strings.Contains(first.String(), "event processing failed") || second.Len() != 0 {
		t.Fatal("logger isolation or level filtering failed")
	}
	for _, secret := range []string{"credential-like-private-data", "secret-argument", "registered command"} {
		if strings.Contains(first.String(), secret) {
			t.Fatal("filtered or sensitive data logged")
		}
	}
	first.Reset()
	level.Set(slog.LevelDebug)
	bot.OnCommand("visible-registration", message.C2C_MESSAGE_CREATE, func(*types.Context) error { return nil })
	if !strings.Contains(first.String(), "registered command") {
		t.Fatal("slog.LevelVar change was ignored")
	}
}

func TestSDKImportDoesNotCreateFiles(t *testing.T) {
	if os.Getenv("GOQQBOT_LOG_IMPORT_HELPER") == "1" {
		return
	}
	dir := t.TempDir()
	command := exec.Command(os.Args[0], "-test.run=^TestSDKImportDoesNotCreateFiles$")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOQQBOT_LOG_IMPORT_HELPER=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("import helper: %v %s", err, output)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatal("SDK import created files")
	}
}
