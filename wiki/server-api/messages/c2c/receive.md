# 接收单聊消息

对无需命令前缀的对话，使用带事件类型的 `OnMessage`。已经匹配的命令优先处理，不会再进入这个普通消息函数。

## 读取正文和图片

以下注册函数在机器人运行前调用。示例只把字段交给业务函数，不自动下载或转存用户附件。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterPrivateReceiver(bot *goqqrobot.Bot, consume func(context.Context, string, string, []message.MessageAttachment) error) {
	bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
		m, err := ctx.Message()
		if err != nil {
			return err
		}
		return consume(ctx.RequestContext(), m.SenderID(), m.Content, m.Images())
	})
}
```

## 选择字段

| 字段或方法 | 含义 |
| --- | --- |
| `ctx.Text()` | 路由移除首尾空白后的文本，命令保留 `/` |
| `m.Content` | 平台下发的原始正文，保留空白 |
| `m.ID` | 当前接收消息的 ID |
| `m.Author.UserOpenID` / `m.SenderID()` | 单聊发送者标识 |
| `m.Attachments` | 当前消息直属附件 |
| `m.Images()` | 当前消息及嵌套引用、转发中的图片附件 |
| `m.ARK` / `m.MsgElements` | 类型化卡片及递归引用、转发条目 |

`ctx.Message()` 不需要 API 客户端，也不执行 I/O。返回的快照由业务持有；更改快照不改变 SDK 载荷。

## 注意事项

无正文的图片也会进入 `OnMessage`，不要先用 `ctx.Text() == ""` 把它过滤掉。SDK 只暴露平台实际下发的内容；内置表情不保证提供图片 URL。不要记录整份 `Raw` 或 `MessageScene`，其中可能有授权数据。

## 另请参阅

[发送单聊消息](send.md) · [读取富文本](../rich-content.md) · [官方单聊消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html)
