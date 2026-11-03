// Package logger supplies a no-op default without global output, files or configuration.
// logger 包提供无操作默认日志器，不输出、不创建文件、不修改配置。
package logger

import (
	"context"
	"log/slog"
)

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (h discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h discardHandler) WithGroup(string) slog.Handler           { return h }

var discard = slog.New(discardHandler{})

// OrDiscard keeps the caller-owned logger or returns a silent logger. It never modifies slog.Default.
// OrDiscard 保留调用方日志器或返回静默日志器，绝不修改 slog.Default。
func OrDiscard(log *slog.Logger) *slog.Logger {
	if log != nil {
		return log
	}
	return discard
}
