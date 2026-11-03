// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

// Package goqqrobot composes event transports, routing and API access.
// goqqrobot 包组合事件传输、路由和 API 访问。
// Register handlers before Run. A Bot runs once; Shutdown is safe concurrently and is idempotent.
// 请在 Run 前注册处理函数；Bot 只运行一次，Shutdown 可并发调用且幂等。
package goqqrobot

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ProgramCX/GoQQBot/internal/adapter"
	"github.com/ProgramCX/GoQQBot/internal/core/pool"
	"github.com/ProgramCX/GoQQBot/internal/core/router"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// Handler handles one command or event. Workers may invoke handlers concurrently.
// Handler 处理一个命令或事件，工作协程可能并发调用处理函数。
type Handler = router.Handler

// Lifecycle errors can be matched with errors.Is.
// 生命周期错误可使用 errors.Is 匹配。
var (
	ErrAlreadyStarted = errors.New("bot already started")
	ErrClosed         = errors.New("bot is closed")
)

// Bot owns its transport and workers; construction starts no goroutines or connections.
// Bot 持有其传输及工作协程，构造时不启动协程或连接。
type Bot struct {
	pool              *pool.Pool
	router            *router.CommandRouter
	config            config.Config
	apiClient         *api.Client
	validator         adapter.Validator
	mu                sync.Mutex
	started           bool
	closed            bool
	cancel            context.CancelFunc
	done              chan struct{}
	shutdownErr       error
	log               *slog.Logger
	onError           func(*types.Context, error)
	firstHandlerError error
	observerError     error
}

// Option customizes dependencies before construction completes.
// Option 在构造完成前定制依赖。
type Option func(*Bot)

// WithAPIClient supplies a caller-owned API client, for example with a custom HTTP transport.
// WithAPIClient 注入调用方持有的 API 客户端，例如使用自定义 HTTP 传输。
func WithAPIClient(client *api.Client) Option {
	return func(b *Bot) {
		if client != nil {
			b.apiClient = client
		}
	}
}

// NewWithConfig creates a dormant bot from an independent copy of cfg.
// NewWithConfig 从 cfg 的独立副本创建尚未启动的机器人。
// A nil defaultHandler leaves legacy non-command messages unhandled; OnMessage can be registered separately.
// defaultHandler 为 nil 时不处理旧版非命令消息，可另行注册 OnMessage。
// Zero optional settings use defaults.
// 可选配置为零时使用默认值。
func NewWithConfig(cfg config.Config, defaultHandler Handler, options ...Option) (*Bot, error) {
	cfg = cfg.WithDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	b := &Bot{config: cfg, apiClient: api.NewClient(cfg.Bot.AppId, cfg.Bot.AppSecret), validator: *adapter.NewValidator(cfg.Bot.AppSecret), done: make(chan struct{})}
	for _, option := range options {
		option(b)
	}
	b.log = logger.OrDiscard(b.log)
	b.router = router.New(defaultHandler, router.WithLogger(b.log), router.WithErrorHandler(b.reportFailure))
	b.pool = pool.New(cfg.Runtime.Pool.Workers, cfg.Runtime.Pool.QueueSize, b.router.Handle, func(task *types.Task, err *types.PanicError) {
		b.reportFailure(&types.Context{Context: task.Ctx, Payload: task.Payload, Client: task.Client}, err)
	})
	return b, nil
}

// New preserves the legacy constructor. Prefer NewWithConfig to avoid internal adapter types.
// New 保留旧版构造器；建议使用 NewWithConfig 避免依赖内部适配器类型。
func New(cfg config.Config, validator adapter.Validator, defaultHandler Handler) (*Bot, error) {
	b, err := NewWithConfig(cfg, defaultHandler)
	if err != nil {
		return nil, err
	}
	b.validator = validator
	return b, nil
}

// NewBot loads a YAML configuration and creates a dormant bot.
// NewBot 加载 YAML 配置并创建尚未启动的机器人。
func NewBot(configPath string, defaultHandler Handler, options ...Option) (*Bot, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewWithConfig(*cfg, defaultHandler, options...)
}

// OnCommand registers a command for an event type. Register before Run for deterministic routing.
// OnCommand 为事件类型注册命令；在 Run 前注册可保证路由确定性。
func (b *Bot) OnCommand(cmd string, eventID message.EventId, handler Handler) {
	b.router.OnCommand(cmd, eventID, handler)
}

// OnEvent registers an event handler. Register before Run for deterministic routing.
// OnEvent 注册事件处理函数；在 Run 前注册可保证路由确定性。
func (b *Bot) OnEvent(eventID message.EventId, handler Handler) { b.router.OnEvent(eventID, handler) }

// Run blocks until canceled, shut down, or a permanent transport error occurs.
// Run 阻塞运行，直到取消、关闭或发生不可恢复的传输错误。
// Accepted handlers may run concurrently and must honor their Context for prompt shutdown.
// 已接受的处理函数可能并发运行，必须响应 Context 以便及时关闭。
// A stopped Bot cannot be restarted. Caller cancellation is returned as ctx.Err().
// 已停止的 Bot 无法重启；调用方取消时返回 ctx.Err()。
// The first event/observer failure observed before shutdown is joined into the result; inspect with errors.As.
// 关闭前观察到的首个事件或观察函数故障会合并到返回值，可通过 errors.As 检查。
func (b *Bot) Run(ctx context.Context) (result error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrClosed
	}
	if b.started {
		b.mu.Unlock()
		return ErrAlreadyStarted
	}
	runCtx, cancel := context.WithCancel(ctx)
	b.started = true
	b.cancel = cancel
	b.mu.Unlock()
	b.pool.Start()
	defer func() {
		cancel()
		shutdownCtx, finish := context.WithTimeout(context.Background(), b.config.Runtime.ShutdownTimeout)
		defer finish()
		drainErr := b.pool.Shutdown(shutdownCtx)
		b.mu.Lock()
		b.shutdownErr = errors.Join(drainErr, b.firstHandlerError, b.observerError)
		result = errors.Join(result, b.shutdownErr)
		b.closed = true
		close(b.done)
		b.mu.Unlock()
	}()
	receive := func(payload *types.Payload, completed func()) error {
		return b.pool.Submit(&types.Task{Ctx: runCtx, Payload: payload, Done: completed, Client: b.apiClient})
	}
	var transport interface{ Run(context.Context) error }
	if b.config.Connection.Mode == "websocket" {
		transport = adapter.NewWebSocket(b.config.Connection.WebSocket, b.apiClient, receive, b.config.Runtime.Pool.Workers+b.config.Runtime.Pool.QueueSize)
	} else {
		transport = adapter.New(b.config, b.validator, receive, b.log)
	}
	err := transport.Run(runCtx)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

// Start runs until Stop/Shutdown or a permanent transport failure.
// Start 运行到 Stop/Shutdown 或不可恢复的传输故障发生。
func (b *Bot) Start() error { return b.Run(context.Background()) }

// StartWithGracefulShutdown runs until an interrupt, termination signal, or transport failure.
// StartWithGracefulShutdown 运行到中断、终止信号或传输故障发生。
func (b *Bot) StartWithGracefulShutdown() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := b.Run(ctx)
	if ctx.Err() != nil && errors.Is(err, context.Canceled) {
		b.mu.Lock()
		defer b.mu.Unlock()
		return b.shutdownErr
	}
	return err
}

// Shutdown cancels transport and handler contexts and waits up to ctx's deadline.
// Shutdown 取消传输及处理函数上下文，并等待至 ctx 的截止时间。
// Calling it before Run permanently closes the bot. Repeated calls are safe.
// Run 前调用会永久关闭机器人；重复调用是安全的。
func (b *Bot) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	if !b.started && !b.closed {
		b.closed = true
		close(b.done)
	}
	cancel := b.cancel
	b.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	select {
	case <-b.done:
		b.mu.Lock()
		err := b.shutdownErr
		b.mu.Unlock()
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop shuts down using runtime.shutdownTimeout. Use Shutdown to supply your own deadline.
// Stop 使用 runtime.shutdownTimeout 关闭；自定义截止时间请使用 Shutdown。
func (b *Bot) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.Runtime.ShutdownTimeout)
	defer cancel()
	return b.Shutdown(ctx)
}

// OnInteraction registers a handler for an interaction kind. More specific OnButton/OnMenu handlers win.
// OnInteraction 注册一种交互的处理函数，更具体的 OnButton/OnMenu 优先。
// Register before Run. Nil handlers are ignored. No ACK is sent automatically.
// 请在 Run 前注册；忽略 nil，不自动发送 ACK。
func (b *Bot) OnInteraction(kind interaction.InteractionType, handler Handler) {
	b.router.OnInteraction(kind, handler)
}

// OnButton registers a type-11 callback by button ID. Empty IDs and nil handlers are ignored.
// OnButton 按按钮 ID 注册类型 11 回调；忽略空 ID 和 nil 处理函数。
// Buttons without an ID can be handled with OnInteraction or OnEvent using button_data.
// 无 ID 按钮可通过 OnInteraction 或 OnEvent 使用 button_data 处理。
func (b *Bot) OnButton(id string, handler Handler) { b.router.OnButton(id, handler) }

// OnMenu registers a type-12 callback by feature_id; menu configuration remains managed by the platform.
// OnMenu 按 feature_id 注册类型 12 回调，菜单配置仍由平台管理。
// Empty IDs and nil handlers are ignored. Register before Run.
// 忽略空 ID 和 nil 处理函数；请在 Run 前注册。
func (b *Bot) OnMenu(featureID string, handler Handler) { b.router.OnMenu(featureID, handler) }

// OnMessage registers unconsumed messages for one event type, such as C2C_MESSAGE_CREATE or GROUP_AT_MESSAGE_CREATE.
// OnMessage 注册指定事件类型中尚未消费的消息，例如单聊或群聊艾特消息。
// Commands take precedence; unknown commands and attachment-only messages are included.
// 命令优先；包括未知命令和仅包含附件的消息。
// It overrides OnAnyMessage for that event type. Later registration replaces the slot; nil is ignored.
// 覆盖该事件类型的 OnAnyMessage；后注册覆盖原处理函数，忽略 nil。
// Register before Run. Interaction callbacks use OnButton/OnInteraction instead.
// 请在 Run 前注册；交互回调应使用 OnButton/OnInteraction。
func (b *Bot) OnMessage(eventID message.EventId, handler Handler) {
	b.router.OnMessage(eventID, handler)
}

// OnAnyMessage registers a fallback across all known message event types, after commands and scoped OnMessage handlers.
// OnAnyMessage 在命令和指定类型 OnMessage 之后，为所有已知消息类型提供兜底。
// It takes precedence over the legacy constructor defaultHandler. Later registration replaces it; nil is ignored.
// 优先于旧版构造器的 defaultHandler；后注册覆盖原处理函数，忽略 nil。
func (b *Bot) OnAnyMessage(handler Handler) { b.router.OnAnyMessage(handler) }
