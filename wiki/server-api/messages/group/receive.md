# 接收群聊消息

注册 `OnMessage(eventType, handler)` 处理群消息正文和附件。命令处理优先；未知命令及只包含附件的消息仍能进入此入口。

## 将消息交给业务服务

以下注册函数在运行前调用。业务服务获得独立解码结果，可以继续读取作者、卡片及引用条目。

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

`GROUP_AND_C2C_EVENT` 用于常规群 @与单聊消息；群全量能力需要在平台开通并选择相应订阅。群消息事件与群成员管理事件是不同订阅，不要用注册一个处理函数代替平台配置。

当前 `IntentMask` 未支持官方的 `GROUP_MEMBER_EVENT` 名称；它不是群全量消息开关。成员增减、入群审批等也没有统一类型化接收入口，不能把本页普通消息示例用于宣称这些能力已完成。

处理函数可并发执行，若业务要求“同一群逐条处理”，应在业务服务中协调。WebSocket 重放可能重复投递，写入业务需幂等。

## 另请参阅

[发送群聊消息](send.md) · [读取富文本](../rich-content.md) · [官方群 @事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html) · [官方群全量事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_message_create.html)
