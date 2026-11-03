# 消息交互覆盖清单与用法

核对日期：2026-11-03。范围是官方“消息交互”栏目下的概述、按钮互动（互动事件、互动事件响应）及文本交互，
同时对照群聊/单聊消息发送接口的 keyboard 字段。平台控制台菜单管理、指令面板管理属于其他接口，
本清单不将接收菜单回调等同于已经实现菜单管理 API。

之前的 SDK 有按钮发送和基础 ACK，但未完整覆盖该栏目。本次新增如下：

| 官方能力 | SDK 入口/覆盖情况 |
| --- | --- |
| 模板、自定义键盘、行列、ID、展示文字/样式 | `KeyboardTemplate`、`Keyboard`、`Buttons`、`Row`、`Columns`；原始结构继续可用 |
| 跳转、指令、回调按钮 | `button.Link`、`Command`、`Callback` |
| 权限、回调互斥分组、确认对话框 | `Users`、`Roles`、`AdminOnly`、`All`、`Group`、`Confirm`、`Confirmation` |
| 自动发送、引用回复、选图、频道选择器 | `AutoSend`、`Reply`、`PickImage`、`ChooseChannel`；不支持的场景默认回退为预填，`Strict` 可禁止回退 |
| 旧 click_limit、低版本提示 | `ClickLimit`（已标记弃用）、`UnsupportedTips`、原始 Action 字段 |
| 互动事件所有已列举类型及字段 | `InteractionCreateEvent`、`interaction.Decode`、`ctx.Interaction()`；新增反馈、会话、故事、模型、授权字段 |
| 按钮/菜单/类型路由 | `bot.OnButton`、`OnMenu`、`OnInteraction`，保留 `OnEvent(INTERACTION_CREATE, ...)` 总兜底 |
| 互动响应 code=0..5 | `ctx.AckInteraction`、`AckInteractionContext`；底层 `Client.AcknowledgeInteraction` |
| @用户、@全部成员 | `text.AtUser`、`AtEveryone`，使用最新 qqbot 标签 |
| 点击直接发送、预填参数命令 | `text.Enter`、`Input`、`Label`、`Reference`，自动 urlencode |
| 跳转子频道、系统表情 | `text.Channel`、`Emoji` |

原始事件保存在 `ctx.Payload.D`；`ctx.Interaction()` 返回的 `Raw` 也是独立完整快照。
未知互动类型可以通过 `OnInteraction(新类型, ...)` 或总兜底接收，不要求更新枚举后才能处理。
未列明的新增事件字段保留在 Raw 中，不在 SDK 内猜测其含义。

## 完整示例

```sh
go run ./examples/interactions -config config.yaml
```

WebSocket 配置需要订阅：

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
      - INTERACTION
```

Webhook 模式则在开放平台订阅相应互动事件。SDK 不自动修改平台权限和订阅配置。
群聊 @机器人发送 `/interact`，或单聊发送 `/interact`，按目录中的分类逐项点击。
示例仅在收到用户命令或点击后回复，不自动发送主动消息。

## 按钮回调与分发

```go
bot.OnButton("refresh", func(ctx *types.Context) error {
    event, err := ctx.Interaction()
    if err != nil {
        return err
    }
    // event.Data.Resolved.ButtonData 是业务数据；权限与业务校验应在应用里完成。
    _ = event.Data.Resolved.ButtonData
    if err := ctx.AckInteraction(interaction.InteractionCallbackCodeSuccess); err != nil {
        return err
    }
    return ctx.ReplyMarkdown("刷新完成")
})
```

发送对应按钮：

```go
return ctx.ReplyMarkdown("## 数据面板",
    button.Callback("刷新", "refresh-data").ID("refresh").Primary(),
)
```

`OnButton` 仅匹配 type=11 的 button_id；`OnMenu` 仅匹配 type=12 的 feature_id。
优先级为：具体按钮/菜单 ID → 互动类型 → `OnEvent(INTERACTION_CREATE)`。
每个事件最多选择一个 handler，不会同时调用多个 handler 导致重复 ACK。
如果平台仅提供 button_data 而没有 button_id，可用 `OnInteraction(InteractionTypeButton, ...)` 自行分发。

指令按钮/文本指令产生的普通消息继续走 `OnCommand`；回调按钮不会被解析成斜杠命令。
SDK 不自动返回成功 ACK：调用方需要在确认业务权限后选择成功、无权限、失败等响应码。

## 互动类型与 ACK

SDK 当前列举：11 消息按钮、12 快捷菜单、13 消息反馈、14 清空会话、15 进入/退出故事集、
16 切换模型、18 用户授权、19 群授权、20 群授权状态变更。
`event.Kind()` 优先读取外层 type，在旧事件缺少该值时回退到 data.type。

只有 type=11 和 12 必须 ACK，可用 `event.NeedsAcknowledgement()` 判断。
其他类型无需 ACK，但官方允许显式调用；例如 type=14 且 code=0 可触发清空会话提示。
SDK 不负责清空应用自己的数据库/聊天历史，相关业务处理仍由应用完成。

响应码保持官方原值：0 成功、1 失败、2 频繁、3 重复、4 无权限、5 仅管理员。
ACK 使用 `PUT /interactions/{d.id}`，参数不带 `INTERACTION_CREATE:` 前缀；
发送被动消息使用的 envelope.id/event_id 是另一个字段，不能相互替代。

官方互动概述要求指令回调类场景在 3 秒内回应。应先 ACK，再做耗时工作；可主动限制请求时长：

```go
ackCtx, cancel := context.WithTimeout(ctx.RequestContext(), 2*time.Second)
defer cancel()
return ctx.AckInteractionContext(ackCtx, interaction.InteractionCallbackCodeSuccess)
```

同一 Context 只允许一次 ACK 尝试，包含并发调用；再次调用返回 `ErrInteractionResponseAttempted`。
调用前 context 已取消时不占用机会。首次 HTTP 请求失败后，平台是否已经接收可能不确定，
SDK 不会悄悄重试；需要重试时调用方可明确使用底层 `Client.AcknowledgeInteraction` 并处理平台结果。
重复投递产生不同 Context，因此这不是跨事件、跨进程去重；应用仍需按互动 ID 做业务幂等。

事件依旧通过有界线程池处理；若队列等待或业务处理过长，可能错过平台时限。
SDK 没有承诺客户端版本、平台权限或端到端耗时一定满足要求。

## 文本交互

```go
return ctx.Reply(message.Markdown("## 欢迎\n").Append(
    text.AtUser(memberOpenID),
    text.Plain("，请选择：\n"),
    text.Input("/bind ").Label("绑定游戏账号").Reference(),
))
```

从片段开始构建也可以使用 `message.MarkdownParts(...)` 或 `message.TextParts(...)`。
`text.Plain` 和初始正文按原样保留 Markdown/文本，不对用户传入的原始标签二次解析或改写。
对不可信 Markdown 的内容转义由应用负责；SDK 会安全转义自己构造的标签属性。

| 片段 | 官方格式/支持范围 |
| --- | --- |
| `AtUser(id)` | `<qqbot-at-user id="..." />`；群聊仅 Markdown，id 使用 member_openid；频道使用 user_openid |
| `AtEveryone()` | `<qqbot-at-everyone />`；群聊文本/Markdown（机器人需为管理员，每群每天一次），频道需相应权限 |
| `Enter(command)` | `<qqbot-cmd-enter text="..." />`；仅 Markdown，C2C 保留，群聊/频道默认回退为 Input，Strict 时返回错误 |
| `Input(command)` | `<qqbot-cmd-input text="..." show="..." reference="true" />`；仅 Markdown，Label/Reference 为可选配置 |
| `Channel(id)` | `<#channel_id>`；仅频道，目标须属于当前频道 |
| `Emoji(id)` | `<emoji:id>`；仅频道，系统 type=1 表情，type=2 按普通字符串填写 |

Input/Enter 的 text 和 Input 的 show 按编码前最多 100 个字符校验。
输入原文即可，SDK 编码一次；空格编码为 `%20`，字面 `+` 编码为 `%2B`，避免不同客户端把空格解释为加号。
`/bind ` 的末尾空格也保留。不要先自行 urlencode 再传给构造器。

@用户/全体、跳转频道和表情没有安全且等价的跨场景替代：用于不支持场景时返回明确错误，
不会偷偷改变提及对象。平台管理员权限、频道成员关系及每日次数由平台校验，SDK 不在本地伪造判断。

## 验证与边界

单元测试使用官方字段表和本地 HTTP 替身，验证所有已列举互动类型、0..5 ACK 码、
同一 Context 的并发/重复 ACK、取消与网络失败、分发优先级、文本编码/转义与场景回退。
本次未使用真实账号发送消息或点击互动；真实客户端展示与平台权限仍需应用侧联调。

## 官方依据

- [消息交互概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/overview.html)
- [互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html)
- [互动事件响应](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/interactions_interaction_id.put.html)
- [文本交互](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/text-chain.html)
- [群聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
- [单聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)

逐项按钮、处理器、平台资源与人工验收清单见 [交互验收指南](interactions-acceptance.md)。
