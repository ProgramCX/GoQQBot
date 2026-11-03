// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package router

import (
	"encoding/json"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	typesShared "github.com/ProgramCX/GoQQBot/shared/types"
)

type Handler func(ctx *typesShared.Context) error
type CommandRouter struct {
	mu             sync.RWMutex
	handlers       map[RoutePair]Handler
	defaultHandler Handler
	interactions   map[interaction.InteractionType]Handler
	buttons        map[string]Handler
	menus          map[string]Handler
	messages       map[message.EventId]Handler
	messageHandler Handler
	log            *slog.Logger
	onError        func(*typesShared.Context, error)
}

type RoutePair struct {
	cmd     string
	eventId message.EventId
}

func New(defaultHandler Handler, options ...Option) *CommandRouter {
	r := &CommandRouter{
		handlers:       make(map[RoutePair]Handler),
		interactions:   make(map[interaction.InteractionType]Handler),
		buttons:        make(map[string]Handler),
		menus:          make(map[string]Handler),
		messages:       make(map[message.EventId]Handler),
		defaultHandler: defaultHandler,
		log:            logger.OrDiscard(nil),
	}
	for _, option := range options {
		option(r)
	}
	return r
}

// Option configures immutable dispatch dependencies before the router is used.
// Option 在路由器使用前配置不可变的派发依赖。
type Option func(*CommandRouter)

// WithLogger uses a caller-owned slog logger; nil keeps dispatch logging disabled.
// WithLogger 使用调用方持有的 slog 日志器；nil 表示禁用派发日志。
func WithLogger(log *slog.Logger) Option {
	return func(c *CommandRouter) { c.log = logger.OrDiscard(log) }
}

// WithErrorHandler reports dispatch failures outside registry locks. The observer must not panic.
// WithErrorHandler 在注册表锁之外报告派发失败；观察函数不得 panic。
func WithErrorHandler(handler func(*typesShared.Context, error)) Option {
	return func(c *CommandRouter) { c.onError = handler }
}

func (c *CommandRouter) fail(ctx *typesShared.Context, stage string, err error) {
	failure := &typesShared.HandlerError{EventType: ctx.Payload.T, Stage: stage, Err: err}
	if c.onError != nil {
		c.onError(ctx, failure)
		return
	}
	c.log.ErrorContext(ctx.RequestContext(), "event processing failed", "event", ctx.Payload.T, "stage", stage)
}

func (c *CommandRouter) invoke(ctx *typesShared.Context, handler Handler) {
	err := func() (err error) {
		defer func() {
			if value := recover(); value != nil {
				err = &typesShared.PanicError{Value: value, Stack: debug.Stack()}
			}
		}()
		return handler(ctx)
	}()
	if err != nil {
		c.fail(ctx, "handler", err)
	}
}

func (c *CommandRouter) OnCommand(cmd string, eventId message.EventId, handler Handler) {

	cmd = strings.TrimPrefix(cmd, "/")
	cr := RoutePair{
		cmd:     cmd,
		eventId: eventId,
	}

	if handler == nil {
		c.log.Warn("ignored nil command handler", "event", eventId)
	}

	if handler == nil {
		return
	}
	c.mu.Lock()
	c.handlers[cr] = handler
	c.mu.Unlock()
	c.log.Debug("registered command", "command", cmd, "event", eventId)
}

func (c *CommandRouter) OnEvent(eventId message.EventId, handler Handler) {
	cr := RoutePair{
		cmd:     "",
		eventId: eventId,
	}
	if handler == nil {
		c.log.Warn("ignored nil event handler", "event", eventId)
	}
	if handler == nil {
		return
	}
	c.mu.Lock()
	c.handlers[cr] = handler
	c.mu.Unlock()
	c.log.Debug("registered event", "event", eventId)
}

// OnMessage registers a fallback for one message event type, overriding OnAnyMessage for that type.
// OnMessage 注册指定消息事件类型的兜底处理函数，并覆盖该类型的 OnAnyMessage。
// Registered commands still take precedence. Nil handlers are ignored; later registration replaces the slot.
// 已注册命令仍优先；忽略 nil，重复注册覆盖原处理函数。
func (c *CommandRouter) OnMessage(eventID message.EventId, handler Handler) {
	if handler == nil {
		return
	}
	c.mu.Lock()
	c.messages[eventID] = handler
	c.mu.Unlock()
}

// OnAnyMessage registers the fallback for all known message event types, excluding interactions/lifecycle events.
// OnAnyMessage 为所有已知消息事件类型注册兜底，不包括交互和生命周期事件。
// Unknown commands and attachment-only messages are included. Nil handlers are ignored.
// 包括未知命令和仅带附件的消息；忽略 nil 处理函数。
func (c *CommandRouter) OnAnyMessage(handler Handler) {
	if handler == nil {
		return
	}
	c.mu.Lock()
	c.messageHandler = handler
	c.mu.Unlock()
}

func (c *CommandRouter) Handle(task *typesShared.Task) {
	if task == nil || task.Payload == nil {
		return
	}
	if task.Payload.T == message.INTERACTION_CREATE {
		c.handleEvent(task)
		return
	}
	var data *struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(task.Payload.D, &data); err != nil {
		c.fail(&typesShared.Context{Context: task.Ctx, Payload: task.Payload, Client: task.Client}, "decode", err)
		return
	}
	if data == nil {
		return
	}
	content := strings.TrimSpace(data.Content)
	ctx := &typesShared.Context{Context: task.Ctx, Payload: task.Payload, RawMessage: content, Args: []string{}, Client: task.Client}
	slash := strings.HasPrefix(content, "/")
	if slash {
		ctx.Command, ctx.Args = parseCommand(content)
	}
	c.mu.RLock()
	var commandHandler Handler
	if slash && ctx.Command != "" {
		commandHandler = c.handlers[RoutePair{cmd: ctx.Command, eventId: task.Payload.T}]
	}
	messageHandler, scoped := c.messages[task.Payload.T]
	if !scoped && message.IsMessageEvent(task.Payload.T) {
		messageHandler = c.messageHandler
	}
	// Preserve the legacy blank-content OnEvent path when no new message handler is selected.
	// 未选中新的消息处理函数时，保留旧版空正文 OnEvent 路径。
	eventHandler := c.handlers[RoutePair{eventId: task.Payload.T}]
	c.mu.RUnlock()
	invoke := func(handler Handler) {
		c.invoke(ctx, handler)
	}
	if commandHandler != nil {
		invoke(commandHandler)
		return
	}
	if messageHandler != nil {
		invoke(messageHandler)
		return
	}
	// Constructor default handlers retain their previous nonempty, non-command behavior.
	// 构造器默认处理函数保留原有的非空且非命令消息行为。
	if !slash && content != "" {
		if c.defaultHandler != nil {
			invoke(c.defaultHandler)
		}
		return
	}
	if content == "" && eventHandler != nil {
		invoke(eventHandler)
		return
	}
	if slash {
		c.log.Debug("unknown command", "event", task.Payload.T)
	}
}

func parseCommand(content string) (string, []string) {
	content = strings.TrimPrefix(content, "/")

	var args []string
	var current strings.Builder
	inQuote := false

	for _, ch := range content {
		switch ch {
		case '"':
			inQuote = !inQuote
		case ' ', '\t':
			if !inQuote && current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			} else if inQuote {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if len(args) == 0 {
		return "", nil
	}

	return args[0], args[1:]
}
