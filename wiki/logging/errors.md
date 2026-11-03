# 处理错误与 panic

处理函数返回错误时，SDK 继续接收事件。需要告警、指标或业务归档时使用 `WithErrorHandler`，不要依赖解析日志文本。

## 观察事件失败

以下完整函数创建带观察函数的 Bot，日志只写可控元数据：

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

此函数要求传入非 nil 日志器。示例同时启用了 SDK 日志与自定义观察日志，实际应用可只在观察函数中更新指标，避免同一故障重复打印。

观察函数在工作协程中同步执行，可能并发调用，必须及时返回。SDK 不持锁调用它。保持 Context 和错误对象只读；不要直接格式化整个 Context、原始 Payload、`PanicError.Value` 或 `Stack`。

## 错误类型

| 类型或判断 | 用途 |
| --- | --- |
| `*types.HandlerError` | `EventType`、`Stage` 和原始根因；通过 Unwrap 保留错误链 |
| `*types.PanicError` | 已恢复的处理函数或观察函数 panic |
| `*api.APIError` | HTTP 状态、平台错误码及 TraceID |
| `errors.Is(err, message.ErrInvalidMessage)` | 本地消息构造或场景校验失败 |
| `errors.Is(err, context.Canceled)` | 操作被取消；可能同时存在其他合并错误 |

处理阶段包含 `decode`、`handler`、`task`、`logger`、`error_handler`。SDK 隔离观察函数自己的 panic，不递归再次调用该观察函数。

## 退出时仍要检查错误

即使未设置日志器或错误观察函数，Bot 仍保留首个事件失败及首个观察函数失败，在 `Run` / `Shutdown` 结束时与关闭错误合并返回。运行过程中不会因为保存错误而停止接收，也不会无限积累所有错误对象。

因此，既要处理实时观察回调，也要检查生命周期方法返回值。匹配到 `context.Canceled` 不代表合并错误里没有 `HandlerError`。

## 平台错误

`APIError.Error()` 不包含平台原始响应正文。需要人工排查原因时，可显式读取 `DiagnosticMessage()`，它长度受限并做尽力脱敏；它仍不是稳定业务错误码，分支逻辑应使用 `Code`、`StatusCode` 和 `errors.As`。

对于消息发送、素材完成等操作，非 nil 结果和错误可能表示 HTTP 已成功但回包解析失败。检查对应结果的 `Raw` 与业务状态，不要自动把它当成“尚未发送”。SDK 不自动重试消息 POST 或失败的业务处理。

出现媒体错误时，区分 QQ 从 URL 下载失败、QQ 上传处理超时、本地请求超时及业务队列过载。平台 TraceID 可用于向 QQ 支持反馈；不要把任意 URL、凭证或完整载荷一起记录。

## 另请参阅

- [写入日志文件](write-log.md)
- [API 调用指南](../getting-started/api-calls.md)
- [官方错误码及 TraceID](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
