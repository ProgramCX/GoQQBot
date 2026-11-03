# 注册命令与普通消息处理函数

为不同消息类型注册同一个处理函数，可以让群聊和单聊复用业务代码。回复目标交给上下文选择，不必自己拼群或用户的发送路径。

## 按类型注册

以下完整函数在 Bot 启动前调用：

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func Register(bot *goqqrobot.Bot) {
	hello := func(ctx *types.Context) error {
		name := "朋友"
		if len(ctx.Args) > 0 {
			name = ctx.Args[0]
		}
		return ctx.ReplyText("你好，" + name)
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, hello)
	bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, hello)
	bot.OnMessage(message.C2C_MESSAGE_CREATE, receive)
	bot.OnMessage(message.GROUP_AT_MESSAGE_CREATE, receive)
}

func receive(ctx *types.Context) error {
	msg, err := ctx.Message()
	if err != nil {
		return err
	}
	if len(msg.Images()) > 0 {
		return ctx.ReplyText("收到图片附件")
	}
	if ctx.Text() == "" {
		return nil
	}
	return ctx.ReplyText("收到：" + ctx.Text())
}
```

`OnCommand` 先传命令，再传事件类型；`OnMessage` 先传事件类型，再传处理函数。`ctx.Args` 在使用下标前检查长度。`ctx.Text()` 是移除首尾空白后的文本，`ctx.Message().Content` 保留原始正文。

## 路由优先级

普通消息按以下顺序选择一个处理函数，不是广播：

1. 当前事件类型匹配的已注册命令。
2. 当前类型的 `OnMessage`。
3. `OnAnyMessage`，仅限已知消息类型。
4. 旧构造器默认处理函数等兼容路径。

未知命令和仅带附件的消息可以进入 `OnMessage` / `OnAnyMessage`。同一个槽位后注册覆盖前注册，nil 处理函数被忽略。

群里艾特机器人对应 `GROUP_AT_MESSAGE_CREATE`；开通并实际收到的群全量消息对应 `GROUP_MESSAGE_CREATE`，注册后者不会自动把前者也注册。订阅与平台权限也必须匹配。

## 原始事件与交互

生命周期事件可通过 `OnEvent` 读取 `ctx.Payload.D`，按对应官方模型显式解码。当前消息型 `OnEvent` 保留旧版兼容语义，不应当作普通消息的统一兜底；需要正文和附件时使用 `OnMessage`。

交互使用独立路由：具体 `OnButton` / `OnMenu` 优先，其次按类型的 `OnInteraction`，最后 `OnEvent(INTERACTION_CREATE, ...)`。交互不会进入 `OnAnyMessage`，也不会自动 ACK。

处理函数可能并发执行，应该响应 `ctx.RequestContext()`；载荷及 Context 公开字段保持只读，不复制 Context。

## 另请参阅

- [消息富文本接收](../../server-api/messages/rich-content.md)
- [按钮回调](../../server-api/messages/interactions/button-callbacks.md)
- [官方事件结构](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
- [官方群聊艾特消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)
