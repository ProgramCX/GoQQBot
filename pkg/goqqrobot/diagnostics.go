package goqqrobot

import (
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// WithLogger injects a caller-owned slog logger for this bot only. Nil disables SDK logs (the default).
// WithLogger 仅为当前机器人注入调用方持有的 slog 日志器；nil 默认禁用 SDK 日志。
// The SDK never changes slog.Default, opens a log file or closes the caller's output.
// SDK 不修改 slog.Default、不打开日志文件，也不关闭调用方输出。
// Use slog.HandlerOptions.Level or slog.LevelVar for filtering and runtime level changes.
// 使用 slog.HandlerOptions.Level 或 slog.LevelVar 过滤日志或动态调整级别。
func WithLogger(log *slog.Logger) Option { return func(b *Bot) { b.log = log } }

// WithErrorHandler observes dispatch errors and recovered panics synchronously on worker goroutines.
// WithErrorHandler 在工作协程中同步观察派发错误及已恢复的 panic。
// The callback may run concurrently and must return promptly; no SDK lock is held while it runs.
// 回调可能并发执行，应及时返回；执行期间不持有 SDK 锁。
// Errors are *types.HandlerError; use errors.As for *types.PanicError or *api.APIError.
// 错误类型为 *types.HandlerError，可用 errors.As 提取 *types.PanicError 或 *api.APIError。
// Keep ctx and errors read-only. Payload, panic Value and Stack can contain sensitive application data.
// 请保持 ctx 和错误只读；Payload、panic Value 及 Stack 可能包含敏感业务数据。
// Processing continues; the first failure is also returned by Run/Shutdown at termination, even with no observer.
// 处理会继续；即使无观察函数，首个故障也会在结束时由 Run/Shutdown 返回。
// An observer panic is contained and retained as an additional failure, without recursively invoking the observer.
// 观察函数自身的 panic 会被隔离并另行保留，不递归调用观察函数。
func WithErrorHandler(handler func(*types.Context, error)) Option {
	return func(b *Bot) { b.onError = handler }
}

func (b *Bot) reportFailure(ctx *types.Context, err error) {
	failure, ok := err.(*types.HandlerError)
	if !ok {
		failure = &types.HandlerError{Stage: "task", Err: err}
		if ctx != nil && ctx.Payload != nil {
			failure.EventType = ctx.Payload.T
		}
	}
	b.mu.Lock()
	if b.firstHandlerError == nil {
		b.firstHandlerError = failure
	}
	b.mu.Unlock()

	// Error and panic values are arbitrary user data; default logs contain metadata only.
	// 错误及 panic 值是任意用户数据，默认日志仅包含元数据。
	b.observeSafely(ctx, "logger", func() {
		args := []any{"event", failure.EventType, "stage", failure.Stage}
		var panicErr *types.PanicError
		if errors.As(failure, &panicErr) {
			args = append(args, "panic", true)
		}
		var apiErr *api.APIError
		if errors.As(failure, &apiErr) {
			args = append(args, "status", apiErr.StatusCode, "code", apiErr.Code, "trace", apiErr.TraceID)
		}
		b.log.ErrorContext(ctx.RequestContext(), "event processing failed", args...)
	})
	if b.onError != nil {
		b.observeSafely(ctx, "error_handler", func() { b.onError(ctx, failure) })
	}
}

func (b *Bot) observeSafely(ctx *types.Context, stage string, callback func()) {
	defer func() {
		if value := recover(); value != nil {
			failure := &types.HandlerError{Stage: stage, Err: &types.PanicError{Value: value, Stack: debug.Stack()}}
			if ctx != nil && ctx.Payload != nil {
				failure.EventType = ctx.Payload.T
			}
			b.mu.Lock()
			if b.observerError == nil {
				b.observerError = failure
			}
			b.mu.Unlock()
		}
	}()
	callback()
}
