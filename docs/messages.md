# 简单回复与 Markdown 按钮

群聊、QQ 单聊和频道消息可以共用同一段 handler，不需要自己判断接收者、拼 HTTP 路径或填回复 ID。
以下 API 是可组合的便利层；现有 `group`、`c2c`、`channel` 包和原始协议结构体继续可用。

```go
return ctx.ReplyText("你好")
```

```go
return ctx.ReplyMarkdown("## 操作成功",
    button.Command("帮助", "/help"),
    button.Command("绑定", "/bind "),
)
```

导入：

```go
import (
    "github.com/ProgramCX/GoQQBot/pkg/message"
    "github.com/ProgramCX/GoQQBot/pkg/message/button"
)
```

完整、可编译的示例位于 `examples/menu`，填写本地配置后可运行：

```sh
go run ./examples/menu -config config.yaml
```

群聊 @机器人发送 `/menu`，或在私聊发送 `/menu`。这个示例不会自动修改机器人平台配置或发送主动消息。

## 消息组合

```go
return ctx.Reply(
    message.Markdown(content).Columns(2).Buttons(
        button.Command("签到", "/checkin").ID("checkin").Primary(),
        button.Command("绑定", "/bind ").ID("bind").Visited("继续填写"),
        button.Link("官网", "https://example.com"),
    ),
)
```

默认每行两列，按钮为灰色线框、所有人可用、命令只预填。
`Primary()` 使用蓝色线框；`Style(...)` 可指定其他平台样式。
布局最多五行、每行最多五个按钮。可以用 `.Columns(3)` 自动排版，或显式逐行添加：

```go
return ctx.Reply(message.Markdown(content).
    Row(button.Command("帮助", "/help"), button.Command("菜单", "/menu")).
    Row(button.Link("官网", "https://example.com")))
```

自动布局按指定列数排版，不会自动增加列数。**默认两列最多容纳 10 个按钮**，14 个按钮需要显式设置 `.Columns(3)`，形成 5 行：

```go
return ctx.Reply(message.Markdown("选择操作").Columns(3).Buttons(buttons...))
```

| 每行列数 | 最多按钮数 |
| --- | --- |
| 1 | 5 |
| 2（默认） | 10 |
| 3 | 15 |
| 4 | 20 |
| 5 | 25 |

SDK 在 `Build(scene)` / `Reply` 发出消息请求前检查限制。布局溢出时，错误包含按钮数、列数、所需行数和调整建议；可通过 `errors.Is(err, message.ErrInvalidMessage)` 或 `errors.Is(err, button.ErrInvalidButton)` 判断。直接调用 `button.Keyboard` 时返回后者。显式 `Row` 和原始 `Keyboard` 同样校验行列限制；平台管理的模板键盘仅有模板 ID，SDK 无法检查其服务端布局。

超过 25 个按钮应由应用分页或拆成多个菜单。SDK 不会静默丢弃按钮、改变指定列数或自动发送多条消息，以免改变布局和消息次数。

`Buttons` 和 `Row` 是两种互斥布局模式，不混用，也不与原始 `Keyboard` / `KeyboardTemplate` 混用。
每种链式方法返回独立值；共享消息模板、按钮、构造结果可以并发复用，修改分支不会改变原模板。
普通方法均不联网。`Build(scene)` 或实际回复时返回构造错误，没有隐藏的 panic 或发送行为。

没有指定 ID 时，SDK 在每个键盘内生成不冲突的 ID；需要长期识别回调动作时应显式指定稳定 ID。
显示文字按字符校验，最多 10 个字符。命令数据按原样保留，不删除 `/bind ` 末尾的空格。

权限可通过 `.Users(userID...)`、`.AdminOnly()`、`.Roles(roleID...)` 或 `.All()` 设置。
角色权限仅用于频道场景。ID 必须使用当前机器人在相应场景拿到的标识。

`AutoSend()` 表达“支持时自动发送”的偏好：QQ 单聊保留 enter，群聊和频道自动回退为预填命令。
原始键盘里的 command action.enter/anchor 也遵循这一默认规则；仅在非单聊发送副本中移除这两个提示字段，
不会修改共享模板、命令数据、权限或未知扩展字段。选图器 anchor 不可用时同样回退为预填命令。
同一段 handler 因此可以共用于群聊、单聊和频道。

需要强制保证能力、发现不支持就返回错误时，显式使用 `.Strict()`：

```go
return ctx.Reply(message.Markdown(content).Strict().Buttons(
    button.Command("帮助", "/help").AutoSend(),
))
```

回退仅处理 enter/anchor 这类交互提示，不改变授权语义。指定频道角色的按钮在群聊中仍返回错误，
不会自动放宽为“所有人”。命令按钮的 `.Reply()` 设置 action.reply；使用 `.Confirm("确认操作？")` 可以添加二次确认。
服务端权限、客户端版本、内容审核和实际按钮展示仍由 QQ 平台决定。

## 自定义 context、结果与多次回复

```go
callCtx, cancel := context.WithTimeout(ctx.RequestContext(), 5*time.Second)
defer cancel()
return ctx.ReplyContext(callCtx, message.Markdown(content))
```

`Reply` / `ReplyMarkdown` 使用 handler context；`ReplyContext` 将指定的 context 传到取 token 和发送请求。
需要撤回 ID、引用索引或发送时间时：

```go
result, err := ctx.SendReply(callCtx, message.Markdown(content))
if err != nil {
    // result 非 nil 表示 HTTP 已成功，但回包解析失败；Raw 保留回包，不应盲目重发。
    return err
}
messageID := result.ID
referenceIndex := result.ExtInfo.RefIndex
_ = messageID
_ = referenceIndex
```

同一 Context 的 QQ 单聊/群聊消息回复自动分配 `msg_seq=1,2,...`，并发调用也不会重复分配。
发送失败同样消耗一个序号，不自动重试 POST，避免请求超时但平台已成功发送时产生重复消息。
需要控制同一次请求的重试标识时可显式 `.Sequence(n)`；它会将后续自动序号推进到至少 n+1。
显式重复使用相同序号是调用者的重试策略，平台可能按重复请求拒绝。

Context 只能使用指针传递，第一次使用后不可复制。公共字段和事件载荷在处理期间应保持只读。
一个 Context 的序号只管理新回复 API；不要在同一消息的回复中混用固定 msg_seq 的旧接口。
重复投递或进程重启产生的新 Context 不是跨进程去重机制，业务幂等仍由应用处理。

## 消息引用与图片校验

```go
return ctx.Reply(message.Markdown(content).Reference("REFIDX_source").VerifyImages())
```

`Reference` 构造 `message_reference` 对象，参数是展示引用的消息索引。
它与被动回复鉴权的 `msg_id` 不同：用户消息索引来自 `message_scene.ext` 的 `msg_idx`；
机器人发送消息的索引来自回包 `ext_info.ref_idx`。
`VerifyImages` 设置 `markdown.force_verify_image_resource`，图片转存失败时中断发送。

## 回调按钮

```go
return ctx.Reply(message.Markdown("## 刷新数据").Buttons(
    button.Callback("刷新", "refresh").ID("refresh").Group("refresh-choice"),
))
```

回调按钮不等于发送 `/命令`。需要订阅 `INTERACTION`；可以用 `OnButton` 按 ID 注册，也可以用下面的总兜底：

```go
bot.OnEvent(message.INTERACTION_CREATE, func(ctx *types.Context) error {
    // 仅为示例：实际业务应从 Payload.D 解码按钮 ID/data，再执行对应逻辑。
    if err := ctx.AckInteraction(interaction.InteractionCallbackCodeSuccess); err != nil {
        return err
    }
    return ctx.ReplyMarkdown("已收到刷新操作")
})
```

按官方互动概述，指令回调类场景要求 3 秒内响应，建议先 ACK，再做耗时工作。
ACK 调用 `PUT /interactions/{d.id}`；消息回复使用最外层事件 `id` 填 `event_id`。
两者是不同操作，发送 Markdown 不会代替 ACK。

## 高级字段与拓展

原始键盘可以直接迁移：

```go
return ctx.Reply(message.Markdown(content).Keyboard(existingKeyboard))
```

原始 Markdown 模板也可通过 `MarkdownTemplate(id, params...)` 使用；新项目建议使用自定义 Markdown。
高级按钮可由 `button.Raw(interaction.KeyboardButton{...})` 包装，保留现有的配置能力。
新增协议字段可以通过 `WithField` 传入：

```go
return ctx.Reply(message.New(message.TypeMessage(8)).WithField("card", map[string]any{
    "type": "tuwen",
    "content": map[string]any{
        "title": "服务器公告", "description": "查看最新活动",
        "url": "https://example.com/news", "pic_url": "https://example.com/image.png",
    },
}))
```

`New` 和 `WithField` 是显式扩展点，不要求为每个新增消息类型改动 Context 的发送签名。
字段在配置时立即序列化快照；平台新增的未知字段/按钮样式会保留，已知非法组合仍会校验。
未封装的键盘字段可用 `.WithField("keyboard", rawObject)` 保留，验证不会重写其未知字段。
`msg_type` 由 `New` 管理；`msg_id`、`event_id`、`msg_seq`、`is_wakeup` 不允许通过 WithField 注入，
避免覆盖回复上下文或把回复变成主动/召回消息。主动发送仍使用原有场景 API。

低层 API 提供 `Client.SendReply(ctx, message.ReplyTarget{...}, outgoing)`，用于已知接收者和来源的高级集成。
`ReplyTarget` 的 `MessageID` 和 `EventID` 必须二选一；直接调用时默认序号为 1，调用者管理多次回复序号。
Context 自动路由支持 QQ 群 @、群全量、单聊、频道 @/普通消息、频道私信，以及 scene 明确的互动事件。
其他事件返回 `ErrUnsupportedReply`；不会根据字段猜测并意外发送主动消息。

## GoCortex 迁移参考

原 `buildCommandKeyboard` 的拆行与多层结构可以替换为：

```go
buttons := make([]button.Button, 0, len(recommended))
for _, a := range recommended {
    b := button.Command(a.Label, a.Command).
        ID(a.ID).Visited(a.VisitedLabel).Style(a.Style).
        UnsupportedTips("当前客户端不支持按钮，请手动发送 "+a.DisplayCommand)
    if a.AutoSend {
        b = b.AutoSend()
    }
    buttons = append(buttons, b)
}
return ctx.Reply(message.Markdown(appendRecommendations(content, recommended)).Buttons(buttons...))
```

GoCortex 可以继续传递 `AutoSend` 偏好：群聊发送时自动预填，单聊发送时保留自动发送能力。
文本推荐、业务动作筛选仍由 GoCortex 决定。
原 `ReplyAssistantMarkdown(ctx, event, content)` 的发送实现可缩成：

```go
return event.ReplyContext(ctx, message.Markdown(content))
```

本次仅在 SDK 中提供新能力与迁移示例，未修改 GoCortex 业务逻辑。
GoCortex 通过本地 replace 使用 SDK 时，还需让其自身的依赖文件包含新增 WebSocket 依赖；
若提示缺少 gorilla/websocket 的 go.sum 项，可在 GoCortex 目录运行 `go mod tidy`。

## 官方协议依据

核对日期：2026-11-03。群聊/单聊最新发送接口与较旧的示例说明可能存在差异；
本 SDK 优先依据当前接口字段表，例如 enter 的单聊限制，不从示例里的 enter=true 推断群聊支持。

- [Markdown 消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/type/markdown.html)
- [发送群聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
- [发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
- [消息交互概述与 ACK](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/overview.html)
- [键盘行列及权限说明](https://bot.q.qq.com/wiki/develop/pythonsdk/model/inline_keyboard.html)
- [发送子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html)
- [频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)

按钮字段、事件类型、按 ID 路由、ACK 一次性语义与文本标签的完整覆盖说明见 [消息交互指南](interactions.md)。

普通文本、未知命令和附件接收入口见 [普通消息接收指南](receiving.md)。
