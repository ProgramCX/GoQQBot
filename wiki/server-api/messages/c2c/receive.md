# 接收单聊消息

用带事件类型的 `OnMessage` 接收无需命令前缀的单聊对话。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
```

[全部 Intent 与下属事件](../../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选「单聊事件」里的 **C2C消息事件**（`C2C_MESSAGE_CREATE`）。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

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

`ctx.Message()` 在本地解析已收到的消息，不需要 API 客户端，也不会向 QQ 发请求。返回的是一份独立副本，修改其中的字段不会影响 SDK 内部的数据。

| 字段或方法 | 含义 |
| --- | --- |
| `ctx.Text()` | 路由移除首尾空白后的文本，命令保留 `/` |
| `m.Content` | 平台下发的原始正文，保留空白 |
| `m.ID` | 当前接收消息的 ID |
| `m.Author.UserOpenID` / `m.SenderID()` | 单聊发送者标识 |
| `m.Attachments` | 当前消息本身的附件 |
| `m.Images()` | 当前消息及嵌套引用、转发中的图片附件 |
| `m.ARK` / `m.MsgElements` | 卡片数据，以及各层引用和转发条目 |

## 注意事项

- 已匹配的命令优先处理，不会再进入普通消息函数。见[事件路由](../../../getting-started/events/routing.md)。
- 无正文的图片也会进入 `OnMessage`。用 `ctx.Text() == ""` 提前过滤会漏掉这类图片。
- 示例不会自动下载或转存用户附件。SDK 仅提供平台下发的内容，内置表情不一定有图片 URL。
- `Raw` 或 `MessageScene` 可能含授权数据。把整份内容写入日志会暴露这些数据。

## 另请参阅

[发送单聊消息](send.md) · [读取富文本](../rich-content.md) · [官方单聊消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html)
