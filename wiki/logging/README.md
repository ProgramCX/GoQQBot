# 日志与错误处理

给 Bot 注入日志器，就能控制输出位置、格式和级别。应用负责管理日志文件的生命周期。

| 需求 | 阅读 |
| --- | --- |
| 将日志写入文件或同时写入控制台 | [写入日志文件](write-log.md) |
| 结构化 JSON、动态日志级别 | [写入日志文件：格式与级别](write-log.md#格式与级别) |
| 观察处理函数错误及 panic | [错误处理](errors.md) |
| 检查平台错误码及 TraceID | [错误处理：平台错误](errors.md#平台错误) |

`WithErrorHandler` 让应用用代码处理失败，与日志输出相互独立。

## 注意事项

- SDK 默认不输出日志，不创建日志文件，也不修改 `slog.Default()`。
- `WithLogger` 只影响当前 Bot。其他机器人实例和独立的 `api.Client` 不受影响。
- 启用 SDK 日志后，故障日志默认只记录元数据。字段包括事件类型、处理阶段、是否 panic、API 状态、错误码和 TraceID。
- 业务文本、原始载荷、panic 值和堆栈可能含有敏感信息。应用自行判断是否记录。

## 另请参阅

[快速开始](../getting-started/quick-start.md) · [运行与关闭](../getting-started/lifecycle.md) · [官方 API 错误及全链路追踪](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
