# 日志与错误处理

SDK 默认不输出日志、不创建日志文件、不修改 `slog.Default()`。应用通过选项注入自己的日志器，并决定输出位置、格式、级别和文件生命周期。

| 需求 | 阅读 |
| --- | --- |
| 将日志写入文件或同时写入控制台 | [写入日志文件](write-log.md) |
| 结构化 JSON、动态日志级别 | [写入日志文件：格式与级别](write-log.md#格式与级别) |
| 观察处理函数错误及 panic | [错误处理](errors.md) |
| 检查平台错误码及 TraceID | [错误处理：平台错误](errors.md#平台错误) |

`WithLogger` 只作用于当前 Bot，不配置其他机器人实例或独立 `api.Client` 的日志。`WithErrorHandler` 是处理失败的程序化入口，和日志输出相互独立。

默认 SDK 故障日志只记录事件类型、处理阶段、是否 panic 以及 API 状态、错误码、TraceID 等元数据。业务文本、原始载荷、panic 值及堆栈可能包含敏感信息，应由应用显式决定是否记录。

## 另请参阅

- [快速开始](../getting-started/quick-start.md)
- [运行与关闭](../getting-started/lifecycle.md)
- [官方 API 错误及全链路追踪](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
