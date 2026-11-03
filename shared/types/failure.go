package types

import (
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// HandlerError describes one failed dispatch without formatting its potentially sensitive cause.
// HandlerError 描述一次派发失败，不格式化可能敏感的根因。
// Use errors.As/Is through Unwrap to inspect API errors, cancellation or PanicError.
// 可沿 Unwrap 使用 errors.As/Is 检查 API 错误、取消或 PanicError。
type HandlerError struct {
	EventType message.EventId
	// Stage identifies decode, handler, task, logger or error_handler.
	// Stage 标识 decode、handler、task、logger 或 error_handler 阶段。
	Stage string
	Err   error
}

// Error omits message text, arguments and the original error value.
// Error 不包含消息文本、参数或原始错误值。
func (e *HandlerError) Error() string {
	return fmt.Sprintf("event processing failed: event=%s stage=%s", e.EventType, e.Stage)
}

// Unwrap preserves the original cause for programmatic handling.
// Unwrap 保留原始根因以便程序化处理。
func (e *HandlerError) Unwrap() error { return e.Err }

// PanicError captures a recovered callback panic. Value and Stack may contain sensitive application data;
// PanicError 捕获已恢复的回调 panic，Value 和 Stack 可能包含敏感业务数据；
// inspect them explicitly, keep them read-only and do not indiscriminately log the struct.
// 请显式检查并保持只读，不得随意记录整个结构体。
type PanicError struct {
	Value any
	Stack []byte
}

// Error deliberately does not format the arbitrary panic value.
// Error 有意不格式化任意 panic 值。
func (e *PanicError) Error() string { return "event callback panicked" }

// Unwrap preserves an error-valued panic without converting other panic values to text.
// Unwrap 保留 error 类型的 panic，不将其他 panic 值转换为文本。
func (e *PanicError) Unwrap() error { value, _ := e.Value.(error); return value }
