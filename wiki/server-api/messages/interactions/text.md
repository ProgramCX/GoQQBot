# 发送文本交互

用 `text` 可以在消息正文里放可点击的命令，也可以在支持的会话里艾特用户。SDK 会生成 QQ 需要的标签，并处理 URL 编码。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/text"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func bindHint(ctx *types.Context) error {
	return ctx.Reply(message.MarkdownParts(
		text.Plain("## 绑定账号\n"),
		text.Input("/bind ").Label("填写游戏账号").Reference(),
		text.Plain("\n也可以点击："),
		text.Enter("/help"),
	))
}
```

## 支持范围

普通消息用 `message.TextParts(...)` 拼接文字和可点击片段。如果已经有 Markdown 正文，用 `message.Markdown("正文").Append(...)` 把片段接在后面。下表列出每种片段适用的会话。

| 片段 | 使用场景及参数 |
| --- | --- |
| `Plain(value)` | 原样正文，保留 Markdown 和现成标签 |
| `AtUser(id)` | 群 Markdown 用 `member_openid`。频道文本或 Markdown 用 `user_openid` |
| `AtEveryone()` | 群或频道 |
| `Input(command)` | Markdown，点击后预填输入框 |
| `Input(...).Label(label)` | 自定义显示文字 |
| `Input(...).Reference()` | 预填时带原消息引用 |
| `Enter(command)` | Markdown，在单聊中点击直接发送 |
| `Channel(id)` | 频道及频道私信。目标为当前频道服务器内的子频道 |
| `Emoji(id)` | 频道及频道私信中的平台 type=1 系统表情 |

## 艾特当前群消息发送者

将下面的处理函数注册给 `GROUP_AT_MESSAGE_CREATE`，或平台已开通的 `GROUP_MESSAGE_CREATE`。接收方式见 [群聊消息](../group/receive.md)。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/text"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func welcomeGroupMember(ctx *types.Context) error {
	incoming, err := ctx.Message()
	if err != nil {
		return err
	}
	return ctx.Reply(message.MarkdownParts(
		text.AtUser(incoming.SenderID()),
		text.Plain("，欢迎使用机器人。"),
	))
}
```

## 编码与输入限制

将命令和显示文字原样传给构造器。SDK 只编码一次。

- 空格编码为 `%20`。
- 字面加号编码为 `%2B`。
- 命令末尾的空格会保留。

Unicode 表情可作为普通文字发送。

## 注意事项

- 单聊支持点击 `Enter` 直接发送。群聊或频道会退回预填指令。添加 `.Strict()` 可让不支持直接发送的场景返回错误。
- 群成员 ID 与单聊 OpenID 不通用。`Label` 和 `Reference` 仅适用于 `Input`。
- `AtEveryone()` 受平台权限约束。群内要求机器人为管理员，平台限制每群每天一次。
- 不支持的艾特、子频道和表情场景会返回错误。SDK 只在能保留原有用途时才会转换片段。平台检查管理员权限、成员关系和次数。
- 编码前，`Input` 或 `Enter` 的命令及 `Label` 的文字各限一百个字符。
- SDK 转义自己生成的属性值。`Plain` 和初始 Markdown 正文按原样保留。应用按业务需要转义用户或外部系统传入的不可信 Markdown。
- `text.Emoji` 发送频道系统表情标签。读取 QQ 内置表情的图片地址需要其他能力，见 [富文本接收](../rich-content.md)。

## 另请参阅

[按钮交互](buttons.md) · [富文本接收](../rich-content.md) · [QQ 官方：文本交互](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/text-chain.html)
