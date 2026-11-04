# 处理错误与 panic

处理函数返回错误后，SDK 继续接收事件。通过 `WithErrorHandler` 更新告警、指标或业务归档，无需解析日志文本。

## 观察事件失败

以下函数创建带观察函数的 Bot。示例只记录可控的元数据。

```go
package guide

import (
	"errors"
	"log/slog"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func NewObservedBot(path string, logger *slog.Logger) (*goqqrobot.Bot, error) {
	return goqqrobot.NewBot(path, nil,
		goqqrobot.WithLogger(logger),
		goqqrobot.WithErrorHandler(func(_ *types.Context, err error) {
			var failure *types.HandlerError
			if errors.As(err, &failure) {
				logger.Error("处理失败", "event", failure.EventType, "stage", failure.Stage)
			}
			var remote *api.APIError
			if errors.As(err, &remote) {
				logger.Error("平台调用失败", "status", remote.StatusCode,
					"code", remote.Code, "trace", remote.TraceID)
			}
			var panicErr *types.PanicError
			if errors.As(err, &panicErr) {
				logger.Error("处理函数发生 panic")
			}
		}),
	)
}
```

观察函数在工作协程中同步执行。SDK 调用观察函数时不持锁。示例同时启用 SDK 日志和自定义观察日志。实际应用可在观察函数中只更新指标，避免重复打印故障。

## 错误类型

根据错误类型或 `errors.Is` 的结果判断失败原因。

| 类型或判断 | 用途 |
| --- | --- |
| `*types.HandlerError` | 提供 `EventType`、`Stage` 和原始根因，通过 Unwrap 保留错误链 |
| `*types.PanicError` | 已恢复的处理函数或观察函数 panic |
| `*api.APIError` | HTTP 状态、平台错误码及 TraceID |
| `errors.Is(err, message.ErrInvalidMessage)` | 本地消息构造或场景校验失败 |
| `errors.Is(err, context.Canceled)` | 操作被取消，返回值可能还含其他错误 |

处理阶段包括 `decode`、`handler`、`task`、`logger` 和 `error_handler`。观察函数自身 panic 时，SDK 会隔离故障，不再递归调用该函数。

## 退出时仍要检查错误

实时观察故障后，仍需检查 `Run` 和 `Shutdown` 的返回值。退出流程见[运行与关闭](../getting-started/lifecycle.md)。

- Bot 保留首个事件失败和首个观察函数失败。未设置日志器或观察函数时也会保留。
- `Run` 或 `Shutdown` 结束时，Bot 将保留的失败与关闭错误合并返回。
- Bot 记录首个故障后继续接收事件，不会无限积累错误对象。

## 平台错误

业务分支可用 `errors.As` 取出 API 错误，再按 `Code` 和 `StatusCode` 处理。`APIError.Error()` 不包含平台原始响应正文。人工排查时可显式读取 `DiagnosticMessage()`。

媒体错误可分开排查。

- QQ 从 URL 下载失败。
- QQ 上传处理超时。
- 本地请求超时。
- 业务队列过载。

向 QQ 支持反馈时，可提供平台 TraceID。

## 注意事项

- 示例函数需传入非 nil 的日志器。
- 观察函数可能并发调用，应及时返回。Context 和错误对象保持只读。
- 整体格式化 Context、原始 Payload、`PanicError.Value` 或 `Stack` 可能暴露敏感信息。记录 TraceID 时也应排除任意 URL、凭证和完整载荷。
- 合并错误匹配 `context.Canceled` 时，仍可能含有 `HandlerError`。
- `DiagnosticMessage()` 限制长度，并尽力脱敏。诊断文本不适合作为稳定的业务错误码。
- 消息发送或素材完成时，接口可能返回了结果，但 `err` 也不是 nil。HTTP 可能已成功，失败发生在回包解析阶段。先核对结果的 `Raw` 和业务状态，再判断是否发送成功。
- SDK 不自动重试消息 POST 或失败的业务处理。

## 另请参阅

[写入日志文件](write-log.md) · [API 调用指南](../getting-started/api-calls.md) · [官方错误码及 TraceID](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
