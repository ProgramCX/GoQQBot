# 普通消息接收

`OnMessage` 接收没有被已注册命令消费的消息，不需要在构造器参数里放默认 handler。

```go
bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
    return ctx.ReplyText("收到：" + ctx.Text())
})
```

`OnMessage(type, handler)` 明确指定事件类型，与 `OnEvent(type, handler)` 保持一致。
跨场景统一接收用 `OnAnyMessage(handler)`，覆盖 QQ 单聊、群 @消息、群全量消息、频道消息和频道私信的已知消息事件。
这是事件路由，不会绕过平台的订阅和权限：群聊通常仍需 @机器人；接收群全量消息需要平台开启对应能力。
互动回调、好友增删等不是普通消息，不会进入OnAnyMessage。

## 分场景注册与优先级

```go
bot.OnAnyMessage(commonHandler)
bot.OnMessage(message.GROUP_AT_MESSAGE_CREATE, groupHandler)
bot.OnMessage(message.GROUP_MESSAGE_CREATE, groupHandler)
bot.OnMessage(message.C2C_MESSAGE_CREATE, privateHandler)
```

每条消息只选择一个处理器：

1. 已注册且匹配的 `OnCommand` 优先，消息不会再次进入 OnMessage。
2. 有按事件注册的 OnMessage 时使用它，否则使用 OnAnyMessage。
3. 没有新接收器时，构造器的旧 defaultHandler 保留原行为：处理非空的非命令文本。
4. 未配置接收器的普通消息可以正常忽略，不再打印两条“没有默认处理器”的重复 Debug 日志。

OnMessage 包括未知 `/命令`，此时 `ctx.Command` / `ctx.Args` 保留解析结果，便于应用回复帮助。
空正文的图片、语音、文件等也会进入新 OnMessage；不要只根据 `ctx.Text() == ""` 就丢弃它们。
同一注册槽位后注册会替换旧处理器，nil 注册被忽略。不是广播/观察器，不会重复调用所有注册函数。
回调锁在执行用户代码前释放；推荐在 Run 前注册，避免运行中更换规则造成业务理解上的不确定性。

## 正文、发送者与附件

```go
bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
    received, err := ctx.Message()
    if err != nil {
        return err
    }
    if len(received.Attachments) > 0 {
        return ctx.ReplyText(fmt.Sprintf("收到 %d 个附件", len(received.Attachments)))
    }
    return ctx.ReplyText("收到：" + ctx.Text())
})
```

- `ctx.Text()` 返回 Router 去除首尾空白后的正文（原 RawMessage）；命令消息仍保留 `/`。
- `received.Content` 保留原始正文空白，`received.ID` 是消息 ID。
- `received.Scene()`、`SenderID()`、`ConversationID()` 提供统一读取方式，不混用群成员和私聊 OpenID。
- `received.Attachments` 包含 URL、类型、文件名、尺寸、大小，以及语音 WAV URL 和 ASR 参考文本。
- `received.MessageType`、`Mentions`、`MessageScene` 保留场景信息；`ARK` 和递归 `MsgElements` 提供类型化卡片、引用、转发内容，旧 `ArkData`、`Elements` 仍保留原始 JSON。
- `received.Raw` 保留完整事件体，包含未来字段。它和 message_scene.ext 可能含平台鉴权信息，不应直接整包打印日志。

Message 每次返回独立快照，修改读取结果不会改变 Context.Payload 或其他读取结果。
它不发网络请求，不下载附件，也不需要 API Client。

类型化内容读取、业务解耦、撤回、引用及富媒体上传见 [群聊、单聊消息与富媒体](media.md)。
自行构造 Context 时，Text 只读取 RawMessage；要读取 Payload 中的正文请调用 Message，避免隐藏的解码错误。
调用方在处理期间应保持 Context 公共字段和 Payload 只读。

## 图片、表情和富文本怎样读取

```go
bot.OnMessage(message.GROUP_AT_MESSAGE_CREATE, func(ctx *types.Context) error {
    m, err := ctx.Message()
    if err != nil { return err }
    images := m.Images()              // 包括引用、转发中的嵌套图片附件。
    attachments := m.AllAttachments() // 图片、语音、视频、文件等，含嵌套内容。
    // 将 m.Content、images、attachments、m.ARK、m.MsgElements 交给自己的业务服务。
    return service.HandleRichMessage(ctx.RequestContext(), m.Content, images, attachments, m.ARK, m.MsgElements)
})
```

| 用户发来的内容 | 读取位置 | 能力边界 |
| --- | --- | --- |
| Unicode 表情，如 😀 | `m.Content` | 这是正文字符，不自带可下载图片 URL |
| 正文中的平台表情标记 | `m.Content` | 按平台实际下发值保留，SDK 不猜测表情 ID 与图片地址的对应关系 |
| 图片、图片形式的表情包 | `m.Attachments` / `m.Images()` | 仅在 QQ 实际下发图片附件时能获取 URL、MIME、大小和尺寸；不保证客户端每一种表情包都下发图片 |
| 语音、视频、文件 | `m.Attachments` / `m.AllAttachments()` | 可读取平台给的 URL；语音还可能有 VoiceWAVURL、ASRReferText |
| 卡片、小程序、位置等结构化内容 | `m.ARK` | Fields 是任意 JSON，按字段实际类型用 Field 解码；卡片里的 preview 等不冒充附件 |
| 引用、合并转发、并行消息 | `m.MsgElements` | 递归读取作者、正文、附件、卡片和子元素，保留 MessageIndex |

`m.Images()` 只按下发的图片 MIME 类型筛选，不按后缀或 URL 猜测；返回独立的附件切片，不发网络请求、不上传素材。需要保持原始消息归属时，用 `MsgElements`，不要只用扁平化后的列表。接收附件与 [Markdown 图片素材管理](materials.md) 是不同能力，SDK 不会把收到的图片自动存进机器人素材库。

QQ 群聊通常需要 @机器人，普通表情消息是否下发仍取决于机器人权限和订阅；OnMessage 不会凭空收到平台没推送的内容。未知字段可以通过 `m.Raw` 和各 `MsgElement.Raw` 进一步检查，但它们可能含鉴权数据，不要整包记录日志。收到纯表情时不应依赖只识别命令的 OnCommand；已注册命令仍优先于 OnMessage。

## 未来消息类型

```go
bot.OnMessage(message.EventId("FUTURE_MESSAGE_EVENT"), handler)
```

明确按事件注册时，可接收将来新增的消息事件；全局默认只认当前已知消息事件，避免把任意事件当成用户消息。
`ctx.Message()` / `message.DecodeReceived` 可以解码这些明确分类的未来消息体，并保留 Raw；
未知事件的 Scene/ConversationID 不会被猜测，自动回复路由也仍需新增规则或显式 ReplyTarget。

## 可点击验收

启动 `examples/interactions`，发送 `/interact`，打开“普通消息”：

- 普通文字：按钮预填/发送普通文本，OnMessage 回复收到的正文。
- 附件消息：单聊选图或直接发送图片/语音/文件，OnMessage 回复附件数；空正文也能接收。
- 未知命令：发送未注册的命令，OnMessage 显示明确的兜底回复。

按钮回调继续由 OnButton/OnInteraction 处理，不会被消息兜底吞掉。

验收示例显式忽略 `received.Author.Bot` 的消息，避免频道全量消息回显形成循环；SDK 核心不强制过滤机器人消息。
## 官方字段参考

- [单聊消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html)
- [群 @消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)

接口修正：此前示例里的 `OnMessage(handler, types...)` 已改为明确的 `OnMessage(type, handler)`；多个类型分别注册，原无类型兜底改用 `OnAnyMessage(handler)`。
