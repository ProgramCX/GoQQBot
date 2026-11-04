# 接收群聊消息

注册 `OnMessage(eventType, handler)` 处理群消息的正文和附件。

## 将消息交给业务服务

在机器人启动前调用下面的注册函数。业务服务收到独立的解码结果，可读取作者、卡片和引用条目。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterGroupReceiver(bot *goqqrobot.Bot, consume func(context.Context, *message.Received) error) {
	handler := func(ctx *types.Context) error {
		m, err := ctx.Message()
		if err != nil {
			return err
		}
		return consume(ctx.RequestContext(), m)
	}
	bot.OnMessage(message.GROUP_AT_MESSAGE_CREATE, handler)
	bot.OnMessage(message.GROUP_MESSAGE_CREATE, handler)
}
```

## 常用字段

使用 `ctx.Message()` 返回的消息读取以下字段。卡片和嵌套结构见[读取富文本](../rich-content.md)。

| 读取内容 | 字段或方法 |
| --- | --- |
| 群会话 | `m.GroupOpenID` / `m.ConversationID()` |
| 发送成员 | `m.Author.MemberOpenID` / `m.SenderID()` |
| 原始正文 | `m.Content` |
| 消息类型 | `m.MessageType` |
| 被提及成员 | `m.Mentions` |
| 当前及嵌套图片 | `m.Images()` |
| 引用上下文 | `m.MessageScene`、`m.MsgElements` |

## 订阅不等于注册

为常规群 @和单聊消息订阅 `GROUP_AND_C2C_EVENT`。配置方式见[事件订阅](../../../getting-started/events/README.md)。

## 注意事项

- 命令优先处理。未知命令和只有附件的消息仍可进入 `OnMessage`。
- 群全量能力需在平台开通并选择相应订阅。注册函数不会代替平台配置。
- 群消息与群成员管理事件各用不同订阅。
- `IntentMask` 尚不支持官方的 `GROUP_MEMBER_EVENT`。该名称也不是群全量消息开关。
- 成员增减和入群审批尚无统一的类型化接收入口。
- 处理函数可并发运行。同一群需要逐条处理时，由业务服务协调。
- WebSocket 重放可能重复投递。业务写入需实现幂等。

## 另请参阅

[发送群聊消息](send.md) · [读取富文本](../rich-content.md) · [官方群 @事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html) · [官方群全量事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_message_create.html)
