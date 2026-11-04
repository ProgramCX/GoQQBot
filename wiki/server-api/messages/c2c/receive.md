# 接收单聊消息

用带事件类型的 `OnMessage` 接收无需命令前缀的单聊对话。

## 读取正文和图片

在机器人启动前调用下面的注册函数。示例把正文和图片附件交给业务函数。

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

`ctx.Message()` 可直接解码消息，无需 API 客户端，也不执行 I/O。返回的快照归业务持有。修改快照不影响 SDK 载荷。

| 字段或方法 | 含义 |
| --- | --- |
| `ctx.Text()` | 路由移除首尾空白后的文本，命令保留 `/` |
| `m.Content` | 平台下发的原始正文，保留空白 |
| `m.ID` | 当前接收消息的 ID |
| `m.Author.UserOpenID` / `m.SenderID()` | 单聊发送者标识 |
| `m.Attachments` | 当前消息直属附件 |
| `m.Images()` | 当前消息及嵌套引用、转发中的图片附件 |
| `m.ARK` / `m.MsgElements` | 类型化卡片及递归引用、转发条目 |

## 注意事项

- 已匹配的命令优先处理，不会再进入普通消息函数。见[事件路由](../../../getting-started/events/routing.md)。
- 无正文的图片也会进入 `OnMessage`。用 `ctx.Text() == ""` 提前过滤会漏掉这类图片。
- 示例不会自动下载或转存用户附件。SDK 仅提供平台下发的内容，内置表情不一定有图片 URL。
- `Raw` 或 `MessageScene` 可能含授权数据。把整份内容写入日志会暴露这些数据。

## 另请参阅

[发送单聊消息](send.md) · [读取富文本](../rich-content.md) · [官方单聊消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html)
