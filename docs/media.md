# 群聊、单聊消息与富媒体

按 2026-11-03 的 QQ 官方文档核对。接收结构、发送结构、上传操作分别建模；它们不是同一种 JSON，不能把接收到的 ARK 或合并转发结构直接回传为发送体。

## 接收与解耦

`ctx.Message()` 返回独立的 `*message.Received` 快照，不要求 Client，也不下载附件。

| 官方结构 | SDK 读取入口 | 说明 |
| --- | --- | --- |
| MessageAttachment | `m.Attachments` / `message.MessageAttachment` | 类型、URL、名称、尺寸、大小、WAV URL、ASR 参考文本 |
| ARKData | `m.ARK` | Prompt、Type、Name、Fields；字段值保持任意 JSON 类型 |
| MsgElement | `m.MsgElements` | MessageIndex、Author、MessageType、Content、Attachments、ARK、递归 MsgElements、Raw |
| MessageScene | `m.MessageScene` / `message.MessageScene` | Source、Ext；Lookup、MessageIndex、ReferencedIndex |

旧 `ArkData`、`Elements` 原始 JSON 字段保留兼容；新增 `ARK`、`MsgElements` 是独立解码的类型化视图。修改一个不会改变另一个。`ReceivedAttachment`、`ReceivedScene` 原名保留，官方结构名以别名提供，不迫使已有项目迁移。

```go
bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
    m, err := ctx.Message()
    if err != nil { return err }
    return service.Handle(ctx.RequestContext(), m)
})
```

这里的 `service` 是调用方自己的业务对象，只接收 `context.Context` 和消息 DTO，不需要持有路由器、Bot、HTTP Client 或 SDK Context。群聊可将同一个 handler 注册到 `GROUP_AT_MESSAGE_CREATE` / `GROUP_MESSAGE_CREATE`。群全量消息仍需要平台权限和订阅。

```go
var tags []string
found, err := m.ARK.Field("tags", &tags) // 数组、对象、数字不会被强制转换成字符串。
index := m.MessageScene.MessageIndex() // msg_idx，用于显示引用。
token, exists := m.MessageScene.Lookup("auth_token") // nil scene 安全；值内的 = 会保留。
```

每次 Message 调用重新解码；需要多次处理时，在业务处理函数内保留一次结果即可。没有隐藏全局缓存或跨事件共享状态。`Raw`、`MessageScene.Ext`、下载链接和上传预签名 URL 可能包含敏感信息，不要整体记录。

## 发送、引用和撤回

```go
// 富文本与按钮复用已有的不可变 builder。
err := ctx.Reply(message.Markdown("## 结果\n处理完成").Buttons(
    button.Command("再试一次", "/retry").AutoSend(),
))

// 图文卡片：构造时无网络，Reply 时检查已知字段限制。
err = ctx.Reply(message.Card(message.ImageTextCard{
    Title: "结果", Description: "点击查看详情",
    URL: "https://example.com/result", PictureURL: "https://example.com/image.png",
}))

// 引用收到的消息，使用平台给出的 REFIDX。
m, err := ctx.Message()
if err != nil { return err }
index := m.MessageScene.MessageIndex()
if index != "" { return ctx.Reply(message.Text("引用回复").Reference(index)) }
```

本地验证已知卡片的标题 1–50 字符、描述 1–200 字符、HTTPS 链接不超过 1024 字符。未来卡片可用 `message.New(message.CARD).WithField("card", message.MessageCard{Type: ..., Content: ...})`，未知类型保留给平台判断。`message.New` / `WithField` 仍支持未来消息类型和新字段。

```go
sent, err := ctx.SendReply(ctx.RequestContext(), message.Text("可撤回的机器人回复"))
if err != nil { return err }
return ctx.Recall(sent.ID)
```

三个标识不能互换：

| 标识 | 来源 | 用途 |
| --- | --- | --- |
| 收到消息的 ID | `m.ID` | 被动回复；群管理员在平台允许时撤回该成员消息 |
| 已发送消息的 ID | `sent.ID` | 撤回机器人发出的消息 |
| 显示引用索引 | `m.MessageScene.MessageIndex()` / `sent.ExtInfo.RefIndex` | `Outgoing.Reference`，不能作为撤回 ID |

`ctx.Recall(id)` / `RecallContext(ctx, id)` 不会默认删除当前用户消息。群管理员如确实要撤回收到的成员消息，显式调用 `ctx.Recall(m.ID)`。群内普通机器人只能撤回自身消息；单聊只能撤回机器人发给该用户的消息。两类撤回都由平台执行两分钟限制和权限检查，SDK 不依赖本地时间推断。空响应算成功；权限、过期等错误通过 `errors.As(err, &apiError)` 获取 `*api.APIError`。

旧 group/c2c 的 `UploadMedia`、`RecallMessage` 已复用共同引擎。旧发送结构中的字符串 `MessageReference` 在 JSON 序列化时转换为 `{ "message_id": "REFIDX_..." }`；也可设置新的 `Reference *message.MessageReference`，该字段优先。优先采用 `ctx.Reply`，统一场景校验与消息序号。

没有接收 Context 的定时任务可直接使用 `client.SendReply(ctx, message.ReplyTarget{...}, outgoing)`；主动发送显式使用 `client.SendMessage(ctx, message.Target{...}, outgoing)`，C2C 互动召回用 `client.SendWakeup`。主动发送权限和配额由平台决定。所有发送不会自动重试；返回非 nil 结果并伴随解码错误表示 HTTP 已成功，不能直接重发。

## 上传后再发送

```go
uploaded, err := ctx.UploadMedia(ctx.RequestContext(), message.RichMediaRequest{
    FileType: message.FileTypeImage,
    URL: "https://example.com/image.png",
})
if err != nil { return err }
return ctx.Reply(message.Media(uploaded.FileInfo))
```

图片、视频、语音、文件分别用 `FileTypeImage`、`FileTypeVideo`、`FileTypeVoice`、`FileTypeFile`。默认 `SrvSendMsg=false`，只上传；显式 true 会让 QQ 直接发送并占用主动消息次数。上传不自动消耗被动回复序号，之后的 `ctx.Reply` 正常分配序号。

本地文件推荐分片上传：

```go
file, err := os.Open("image.png")
if err != nil { return err }
defer file.Close()
info, err := file.Stat()
if err != nil { return err }
op, cancel := context.WithTimeout(ctx.RequestContext(), 2*time.Minute)
defer cancel()
uploaded, err := ctx.UploadFile(op, api.FileUpload{
    Reader: file, Size: info.Size(), Name: info.Name(), Type: message.FileTypeImage,
})
if err != nil { return err }
return ctx.ReplyContext(op, message.Media(uploaded.FileInfo))
```

`api.FileUpload` 接收 `io.ReaderAt`，读取时不修改文件偏移，SDK 不关闭它。调用期间文件内容必须稳定。文件哈希使用 64 KiB 缓冲区，不将整个文件读入内存；块大小和最后一个块的长度来自服务端计划。上传顺序为：

1. 计算整文件 MD5、SHA1，以及前 **10002432 字节** MD5。
2. `PrepareUpload`：POST `/v2/{users|groups}/{id}/upload_prepare`。
3. 检查整个分片计划，按序号上传；`UploadPart` 用无 QQ Authorization、无 Cookie 的 HTTPS PUT 请求传输原始字节，拒绝重定向。
4. 每个 PUT 成功后 `FinishUploadPart`：POST `upload_part_finish`，包含上传 ID、服务端返回的原始分片序号、实际大小和块 MD5。文档定义序号从 0 开始；SDK 同时兼容完整连续的 `1..N` 计划。文件偏移使用排序后的分片位置计算，完成通知不擅自减一。重复、缺失或越界编号仍会在传输前报错。
5. `UploadMedia` 携带 `UploadID` 和文件名调用 `files` 合并，返回 `file_info`。

高层 UploadFile 使用单个 worker，同步返回、首错即停，没有自动重试，也没有隐藏 goroutine。服务端 UploadConfig 的并发和重试建议保留在 PrepareUpload 返回值中；需要并发、断点续传或自定义重试时，组合上述低层 API，并由调用方保存上传计划与已确认分片。高层失败不返回续传进度。提供的 ReaderAt 本身需要保证读操作能返回，取消 context 无法强制中断调用方自定义的阻塞 reader。

`client.UploadFile(ctx, target, file)` 与 `ctx.UploadFile` 共用实现。Target 只有 Scene 和 ID，不依赖接收事件；有助于将文件上传放入独立业务服务。

URL、UploadID 和 FileData 必须三选一。官方把 `file_data` 定义为字符串和 5M 限制，未说明字符串编码；SDK 作为低层字段原样转交，不猜测编码或宣称已验证二进制直传。建议使用无编码歧义的 UploadFile 分片流程。

文件类型 4 已支持。超过软限制时平台可能降级为文件，超过硬限制由平台报错；SDK 不擅自转换文件格式。上传得到的 `file_info` 应在相应场景、会话和 TTL 内使用，不跨群聊/单聊复用。TTL=0 表示可长期使用，其他值单位为秒；SDK 不缓存，也不自动刷新。RawURL 是可选、临时的下载地址。

## C2C 输入状态和流式更新

```go
err := ctx.Reply(message.Typing(10)) // 1..60 秒；仅 C2C。
first, err := ctx.SendStreamChunk(ctx.RequestContext(), message.StreamChunk{
    InputState: 1, ContentType: "markdown", Content: "开始生成。",
})
if err != nil { return err }
_, err = ctx.SendStreamChunk(ctx.RequestContext(), message.StreamChunk{
    InputState: 10, Index: 1, ContentType: "markdown", Content: "生成完成。",
    StreamMessageID: first.ID,
})
return err
```

`InputMode` 默认 append，可选 replace；replace 必须保留已发送前缀，由业务管理该状态。InputState=1 生成中，10 结束。Index 从 0 递增，后续携带前一响应 ID。返回 `RemainingLength` 和完整 Raw。Context 与普通回复共享自动序号；独立 API 调用可用 `client.SendStreamChunk` 并显式管理序号。事件来源不变，SDK 不持久化流状态，不自动缓冲、排序或恢复。群聊调用输入状态或流式发送会返回明确的场景错误，调用方可按 Target.Scene 分支。

## 按钮验收示例

```powershell
go run ./examples/media -config config.yaml `
  -image-url "https://your-public-host/image.png" -image-file "E:\fixtures\image.png" `
  -video-file "E:\fixtures\video.mp4" -voice-file "E:\fixtures\voice.mp3" `
  -file-file "E:\fixtures\sample.pdf" `
  -card-url "https://your-public-host/page" -card-image "https://your-public-host/image.png"
```

单聊或群内发送 `/media`（群内按平台要求 @机器人），有 14 个按钮及对应命令处理器。四类素材各有 URL 和分片按钮，另有卡片、撤回机器人回复、引用、输入状态、流式消息和接收结构按钮。每类素材可同时配置 `-<type>-url` 与 `-<type>-file`；未配置会明确说明缺失参数，不显示“测试成功”。普通非命令消息由有类型的 OnMessage 汇总接收结构，不输出凭证、消息正文或临时链接。可发送卡片、图片、语音、文件、引用、转发消息观察结果；详情由业务通过 Message 的结构读取。

模拟测试覆盖 14 个处理器在两类会话中的执行，以及上传路径/校验和/失败停止/授权隔离等协议行为。它不能证明 QQ 侧素材格式、权限、配额和客户端显示已通过验收；真实验收需运行示例逐项点击。

群内 `@机器人 /media` 无响应时，重启示例并追加 `-debug-events`，然后分别发送 `@机器人 /media-ping` 和 `@机器人 /media`。该开关只打印事件类型、处理路由、是否命令、正文长度和处理结果，不打印正文或鉴权数据。

- 出现 `media receive: event=GROUP_AT_MESSAGE_CREATE`：群事件已到达处理器；后续 `success=false` 和错误日志用于排查回复失败。
- `/media-ping` 有回复而 `/media` 失败：检查菜单回复的 API 错误，不能归因于事件订阅。
- 两个命令都没有 `media receive`，但私聊有：尚无群事件到达这些处理器的证据；核对是否实际选中机器人完成 @、平台的测试群/成员权限及是否有另一个机器人进程占用相同连接。此时不要仅靠放宽文本解析判断为路由缺陷。

上传 API 失败时，错误链会标明 `prepare media upload`（预上传）、`upload part`（存储 PUT）、`finish media upload part`（分片完成）或 `merge media upload`（合并）。示例的 `send uploaded ... media` 表示上传之后的消息发送失败。`errors.As` 仍可取得 APIError 的 HTTP 状态、平台错误码和 trace，不输出原始响应或预签名 URL。官方上传页面未列出的错误码（例如本次联调遇到的 `40093004`）应结合失败阶段和 trace 排查，不能直接认定为权限、文件格式或容量问题。

平台响应若提供 `message`、`msg` 或 `err_msg`，可用 `apiError.DiagnosticMessage()` 读取最多 512 字符的脱敏说明（截断时附加省略号）。已知凭证、URL、常见敏感字段和长的不透明字符串会被过滤；这是尽力脱敏，不应作为原始载荷日志入口。默认 `Error()` 不含该说明；示例仅在 `-debug-events` 时用引号转义输出 `media API diagnostic`，供确认未公开错误码的实际原因。平台未返回说明时，该值为空。

## 官方依据

- [单聊接收事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html)、[群聊接收事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)
- [单聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)、[群聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)、[C2C 流式发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_stream_messages.post.html)
- [单聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages_message_id.delete.html)、[群聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages_message_id.delete.html)
- [富媒体概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/rich-media.html)
- [单聊上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_files.post.html)、[单聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_prepare.post.html)、[单聊分片完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_part_finish.post.html)
- [群聊上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_files.post.html)、[群聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_prepare.post.html)、[群聊分片完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_part_finish.post.html)
