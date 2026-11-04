# 常见问题

机器人没回复、文件传不上去，或消息在手机和电脑上显示不一样时，可以按下面的问题逐项排查。

## 群里 @机器人后没有回复

确认平台已开放群聊权限，并[订阅对应事件](../getting-started/events/README.md)。实际 @ 机器人，再发送 `/hello` 或示例支持的命令。

- `OnCommand` 依次接收命令、事件类型和处理函数。
- `OnMessage` 依次接收事件类型和处理函数。
- 普通群 @ 命令使用 `GROUP_AT_MESSAGE_CREATE`。

SDK 默认不输出日志。请先按[日志指南](../logging/README.md)配置日志。运行 `media` 示例时，还可以启用 `-debug-events`，根据下列日志判断问题发生在哪个阶段：

- 是否收到事件。
- 事件有没有交给对应的处理函数。
- 失败发生在上传还是发送。

## 收到“没有默认消息处理函数”或非命令未处理

只注册 `/hello` 时，图片、普通文字和其他命令不会交给它处理。要接收这些消息，可以注册 `OnMessage` 或 `OnAnyMessage`。同时注册多个处理函数时，先查[事件路由](../getting-started/events/routing.md)里的匹配顺序和旧版默认处理规则。

## 菜单提示缺少 `-image-file`

`-image-file` 是启动参数。添加该参数并重新启动示例，再点击本地图片上传按钮。测试 URL 上传时，使用启动参数 `-image-url`。完整命令见[人工验收](../examples/README.md)。

## URL 上传超时，本地文件上传却成功

URL 模式由 QQ 服务器拉取文件。先读 `APIError` 的状态、错误码及诊断原因。再决定是否改用[本地分片上传](../server-api/messages/rich-media/multipart-upload.md)。

已观察到以下错误码。结合[官方错误码](https://bot.q.qq.com/wiki/develop/api-v2/openapi/error/error.html)和当次平台诊断区分原因。

- `40034003` 表示富媒体文件上传超时。
- `40093007` 表示富媒体文件下载失败。



## 按钮布局超限或群聊不能自动发送

默认两列最多 10 个按钮（5 行 × 2 列）。放 14 个按钮可设 `Columns(3)`（5 行 × 3 列），也可分页。

`AutoSend()` 在支持的单聊中保留自动发送。群聊默认退回预填。更多用法见[按钮指南](../server-api/messages/interactions/buttons.md)。

## 图片在电脑显示，在手机只看到替代文字

先检查图片来源和消息类型。Markdown 素材使用正确显示尺寸。SDK 提供尺寸读取和 `WithDimensions`。删除素材前，分别验证各客户端。

操作步骤见[素材管理](../server-api/messages/material-management.md)。

## `msg_id`、`event_id`、`REFIDX`、交互 ID 怎么选

按操作选择来源、引用索引或发送结果 ID。

| 标识 | 使用场景 |
| --- | --- |
| 当前消息 `d.id` | 普通消息被动回复的 `msg_id` |
| 外层事件 `id` | 允许事件回复的 `event_id` |
| 交互 `d.id` | `AckInteraction` 的确认目标 |
| `msg_idx` 或发送结果 `ext_info.ref_idx` | 展示引用的 `Reference` 参数 |
| 机器人发送结果 `ID` | 撤回刚发送的机器人消息 |

字段对应关系见[引用与撤回](../server-api/messages/reference-and-recall.md)。

## 可以给所有 API 自动加重试吗

按查询、可安全重试的操作和结果未知的写入分别处理。

幂等键、序号和查询恢复流程由应用管理。错误分支见[错误处理](../logging/errors.md)。

## 注意事项

- 普通 `@机器人` 文本不会触发 QQ 的真实艾特。`GROUP_MESSAGE_CREATE` 仅用于已获批且实际下发的全量事件。
- 日志记录完整载荷或凭证会泄露敏感信息。
- QQ 拉取受源站可达性、网络、响应方式和文件处理影响。增加本机 HTTP 超时不一定能改变平台拉取超时。URL 上传失败后，SDK 不会自动下载再上传。业务若自行处理，需设定下载目标、大小上限、超时和重试边界。
- 键盘上限为 5 行，每行 5 个按钮，依据 QQ 的键盘行列限制。SDK 不丢弃按钮或自动增加列数。`.Strict()` 会对场景不支持的提示返回错误。退回预填不会改变按钮权限。
- `file_info` 与 Markdown 素材引用不可互换。HTTP 发送成功只说明请求成功。各手机版本的显示需要分别验证。
- 展示引用索引不能用于撤回。引用索引也不能保证稳定溯源。
- 超时后，QQ 可能已经完成发送、上传或审批。SDK 不自动重试消息或上传 POST。返回了结果但 err 也不是 nil 时，可能是请求成功而解码失败。直接重试有重复执行的风险。

## 另请参阅

[事件路由](../getting-started/events/routing.md) · [错误处理](../logging/errors.md)
