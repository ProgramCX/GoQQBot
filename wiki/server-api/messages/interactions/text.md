# 文本交互

在 Markdown 正文中插入可点击命令，或在支持的场景中艾特用户。用 `text` 构造片段，避免手写标签及重复 URL 编码。

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

上例在群聊或频道中把 `Enter` 降级成预填指令，在单聊中保留点击直接发送。需要禁止降级时，对消息加 `.Strict()`。

## 支持范围

| 片段 | 使用场景及参数 |
| --- | --- |
| `Plain(value)` | 原样正文，保留 Markdown 和现成标签 |
| `AtUser(id)` | 群 Markdown 使用 `member_openid`；频道文本或 Markdown 使用 `user_openid` |
| `AtEveryone()` | 群或频道，需平台权限；群要求机器人为管理员且每群每天一次 |
| `Input(command)` | Markdown，点击后预填输入框 |
| `Input(...).Label(label)` | 自定义显示文字，仅 Input 可用 |
| `Input(...).Reference()` | 预填时带原消息引用，仅 Input 可用 |
| `Enter(command)` | Markdown，仅单聊支持直接发送，其他场景默认降级 |
| `Channel(id)` | 频道及频道私信，目标为当前频道服务器内的子频道 |
| `Emoji(id)` | 频道及频道私信中的 type=1 系统表情，不是图片 URL 解析器 |

普通文本组合使用 `message.TextParts(...)`；在已有正文后追加片段使用 `message.Markdown("正文").Append(...)`。

## 艾特当前群消息发送者

以下为完整群聊处理函数，注册给 `GROUP_AT_MESSAGE_CREATE` 或平台已开通的 `GROUP_MESSAGE_CREATE`。不要直接拿单聊 OpenID 作为群成员 ID。

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

不支持的艾特、子频道和表情场景会返回错误，没有等价行为时不自动改写含义。SDK 负责标签构造，平台负责管理员权限、成员关系和次数限制。

## 编码与输入限制

`Input` / `Enter` 的命令、`Label` 的文字在编码前分别最多一百个字符。请直接传原文，SDK 只编码一次：空格变为 `%20`，字面加号变为 `%2B`，末尾空格保留。

SDK 转义自己生成的属性值，但 `Plain` 及初始 Markdown 正文按原样保留。来自用户或外部系统的不可信 Markdown，应由应用按业务需要转义，不能假定所有正文已被清洗。

Unicode 表情可以作为普通文字发送。`text.Emoji` 是发送频道系统表情标签的入口，不能据此获取收到的 QQ 内置表情图片地址。

## 另请参阅

- [按钮交互](buttons.md)
- [富文本接收](../rich-content.md)
- [QQ 官方：文本交互](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/text-chain.html)
