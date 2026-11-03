# 常见问题

## 群里 @机器人后没有回复

先确认平台已开放群聊权限并订阅对应事件。实际 @机器人，再发送 `/hello` 或示例支持的命令。普通 `@机器人` 文本不等于 QQ 的真实艾特。

`OnCommand` 注册顺序为“命令、事件类型、处理函数”；`OnMessage` 为“事件类型、处理函数”。注册普通群 @命令应使用 `GROUP_AT_MESSAGE_CREATE`。只有平台获批并实际下发全量事件时，才使用 `GROUP_MESSAGE_CREATE`。

SDK 默认不输出日志；按[日志指南](../logging/README.md)显式注入日志器。使用 media 示例时加 `-debug-events`，观察是否收到事件、是否选中路由、失败发生在上传还是发送。不要为排查问题直接打印凭证和完整载荷。

## 收到“没有默认消息处理函数”或非命令未处理

注册 `OnMessage` 或 `OnAnyMessage` 接收普通消息。注册 `/hello` 不会自动处理图片、普通文字或未知命令。处理优先级和旧版默认 handler 的区别见[事件路由](../getting-started/events/routing.md)。

## 菜单提示缺少 `-image-file`

这是示例启动参数，不是在 QQ 中发送的命令参数。停止示例，带 `-image-file` 重新启动，然后点击本地图片上传按钮。URL 上传按钮另需 `-image-url`。完整命令见[人工验收](../examples/README.md)。

## URL 上传超时，本地文件上传却成功

URL 模式由 QQ 服务器拉取地址，受源站可达性、网络、响应方式和文件处理影响。先查看 `APIError` 的状态、错误码及显式诊断原因，再决定是否改用 SDK 的本地分片上传。增加本机 HTTP 超时不一定能改变平台拉取超时。

例如已观察到的 `40034003` 表示富媒体文件上传超时、`40093007` 表示富媒体文件下载失败；应结合[官方错误码](https://bot.q.qq.com/wiki/develop/api-v2/openapi/error/error.html)和当次平台原因判断，不将全部失败归为同一问题。

SDK 不会在 URL 失败时自动下载并上传文件。业务如需该策略，应显式决定下载目标、大小上限、超时及重试边界。

## 按钮布局超限或群聊不能自动发送

键盘最多五行，每行最多五个按钮；默认两列最多十个。十四个按钮需设置 `Columns(3)`，或者由业务分页。SDK 不自动丢弃按钮或增加列数。

`AutoSend()` 在支持的单聊中保留自动发送；群聊默认降级为预填。显式选择 `.Strict()` 后，不支持的提示会返回错误。权限不会随场景降级而被放宽，详见[按钮指南](../server-api/messages/interactions/buttons.md)。

## 图片在电脑显示，在手机只看到替代文字

检查图片来源及消息类型，区分普通富媒体 `file_info` 与 Markdown 素材引用；不要互换。为 Markdown 素材提供正确显示尺寸，并在删除素材前分别验收客户端。当前 SDK 提供尺寸读取及 `WithDimensions`，但不能保证所有手机版本的显示一致。

详见[素材管理](../server-api/messages/material-management.md)。不要用“HTTP 发送成功”直接推断手机端已经显示图片。

## `msg_id`、`event_id`、`REFIDX`、交互 ID 怎么选

| 标识 | 使用场景 |
| --- | --- |
| 当前消息 `d.id` | 普通消息被动回复的 `msg_id` |
| 外层事件 `id` | 允许事件回复的 `event_id` |
| 交互 `d.id` | `AckInteraction` 的确认目标 |
| `msg_idx` 或发送结果 `ext_info.ref_idx` | 展示引用的 `Reference` 参数 |
| 机器人发送结果 `ID` | 撤回刚发送的机器人消息 |

展示索引不等于撤回 ID，也不能保证稳定的消息溯源。见[引用与撤回](../server-api/messages/reference-and-recall.md)。

## 可以给所有 API 自动加重试吗

不应无条件重试。请求超时不代表 QQ 没有完成发送、上传或审批。SDK 消息和上传 POST 不自动重试；部分成功而解码失败时还可能同时返回结果和错误。

应用应区分查询、可安全重试的操作和结果不确定的写入，再管理幂等键、序号及查询恢复流程。错误分支见[错误处理](../logging/errors.md)。
