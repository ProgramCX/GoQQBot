# GoQQBot v0.3.0

GoQQBot v0.3.0 新增 WebSocket 接入、群聊和频道管理，以及机器人菜单与指令面板 API。消息功能扩展到图文卡片、本地富媒体分片上传和可复用的 Markdown 图片素材。[完整提交对比](https://github.com/ProgramCX/GoQQBot/compare/v0.2.3-dev...v0.3.0)

## 对应 QQ 平台近期更新

| QQ 平台更新 | GoQQBot v0.3.0 |
| --- | --- |
| 8 月 10 日新增群禁言、入群申请审批、自动审批策略和入群申请事件 | 提供类型化的查询、写入和策略管理方法。通过 `GROUP_MEMBER_EVENT` 订阅后，可用 `ctx.GroupJoinRequest()` 读取申请。 |
| 8 月 12 日新增自定义菜单与指令面板管理 | `robot.Client` 可查询、替换全局单聊菜单，也可创建、查询、修改、删除指令面板及更新关联对象。 |
| 9 月 3 日新增群成员查询、批量移除和黑名单管理 | `group.Client` 提供对应方法和分页类型，保留批量操作结果。 |
| 9 月 16 日修订分享链接协议 | `GenerateShareLink` 支持 `callback_data`，并解析响应中的分享链接。 |
| 10 月新增 Markdown 图片素材、群聊 @全体成员和图文卡片 | 支持上传和复用 `qqbot-media://` 素材引用。群聊可用 `text.AtEveryone()` 构造 @全体成员片段，单聊和群聊可用 `message.Card` 构造图文卡片。平台仍检查管理员身份和发送配额。 |

[QQ 官方变更记录](https://bot.q.qq.com/wiki/develop/api-v2/changelog.html)

## 接入与消息

- QQ 机器人运营与管理后台的「开发设置 → 事件订阅与回调」可选择 WebSocket 或 Webhook。WebSocket 在 `config.yaml` 填写 Intent，由 SDK 处理网关鉴权、心跳、重连和会话恢复。Webhook 在后台设置 HTTPS 回调地址，可选择全部接收或按需勾选具体事件。两种方式共用同一套业务处理函数。[连接方式对比](https://programcx.github.io/GoQQBot/getting-started/events/connection.html) · [Intent 与事件对照](https://programcx.github.io/GoQQBot/getting-started/events/intents.html)
- 普通消息通过 `OnMessage` 按事件类型接收。业务可通过 `ctx.Message()` 读取正文、附件、卡片、引用和转发，并用 `ctx.Reply` 回复当前会话。新增消息撤回、单聊输入状态与流式发送用法。
- 单聊和群聊富媒体支持 URL 上传与本地文件分片上传。Markdown 图片素材另有上传、查询、复用和删除接口。
- 频道消息支持创建私信会话、发送本地图片、撤回、表态分页和消息审核。旧发送入口可以解析 QQ 返回的 RFC3339 时间戳。完整的频道发送响应可通过 `channel.DecodeSentMessage` 读取。

## 管理接口

- `group.Client` 管理群资料、机器人群内状态、成员、禁言、黑名单和入群审批。
- `channel.Client` 管理频道、子频道、成员、身份组、权限、公告、精华消息、日程、论坛和音频。
- `robot.Client` 管理机器人资料、分享链接、全局菜单和指令面板。三个客户端均可复用现有的 `api.Client`，便于从后台任务或独立服务调用。

## 升级

```sh
go get github.com/ProgramCX/GoQQBot@v0.3.0
```

项目要求 Go 1.24.4 或更高版本。新增处理函数建议使用 `OnMessage` 和 `ctx.Reply`。撤回等操作可用 `errors.Is` 判断已知错误类别，并通过 `errors.As` 读取平台错误码和 TraceID。

## 测试与平台限制

Go 测试、竞态检查、`go vet`、71 个文档示例编译和 VitePress 构建均通过。QQ 实测覆盖入群审批的同意与拒绝、普通成员禁言与解除、频道私信图片发送及测试面板管理。修改禁言时长时，QQ 返回成功，但随后查询的到期时间与请求值存在差异。

群成员列表、成员详情、黑名单查询、频道成员列表和频道私信撤回在测试应用中返回 `40012010`。这些接口的成功响应仍需使用符合平台授权条件的机器人测试。全局菜单实测覆盖查询与空菜单写回，非空菜单的显示和点击需在 QQ 客户端验证。

[使用文档](https://programcx.github.io/GoQQBot/) · [实机测试记录](https://github.com/ProgramCX/GoQQBot/blob/v0.3.0/examples/management/RELEASE-CHECK.md)
