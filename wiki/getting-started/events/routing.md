# 注册命令与普通消息处理函数

群聊和单聊可以共用同一个处理函数。上下文会选择回复目标，无需拼接群或用户的发送路径。

## 按类型注册

在 Bot 启动前调用下面的注册函数。

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

- `OnCommand` 先传命令，再传事件类型。
- `OnMessage` 先传事件类型，再传处理函数。
- `ctx.Text()` 去掉首尾空白。`ctx.Message().Content` 保留原始正文。

## 路由优先级

收到消息后，SDK 按以下顺序查找处理函数，只调用第一个匹配的函数。

1. 当前事件类型匹配的已注册命令。
2. 当前类型的 `OnMessage`。
3. `OnAnyMessage`，仅限已知消息类型。
4. 旧构造器默认处理函数等兼容路径。

命令 -> 当前类型的 OnMessage -> OnAnyMessage -> 兼容路径

未知命令和仅含附件的消息可进入 `OnMessage` 或 `OnAnyMessage`。

## 原始事件与交互

用 `OnEvent` 注册生命周期事件，再从 `ctx.Payload.D` 读取数据。解析时使用该事件对应的官方结构。

交互使用独立路由，依次匹配以下处理函数。

1. 按钮的 `OnButton` 或菜单的 `OnMenu`。
2. 按类型注册的 `OnInteraction`。
3. `OnEvent(INTERACTION_CREATE, ...)`。

## 注意事项

- 访问 `ctx.Args` 的下标前，先检查长度。
- 同一注册位置以后注册的处理函数为准。nil 处理函数会被忽略。
- 群聊艾特对应 `GROUP_AT_MESSAGE_CREATE`。开通后实际收到的群全量消息对应 `GROUP_MESSAGE_CREATE`。
- 注册群全量消息时，不会同时注册艾特事件。[事件订阅](README.md)需与平台权限匹配。
- 消息型 `OnEvent` 保留旧版兼容行为。统一接收正文和附件时，使用 `OnMessage`。
- 交互不会进入 `OnAnyMessage`。SDK 不自动 ACK，确认方式见[按钮回调](../../server-api/messages/interactions/button-callbacks.md)。
- 处理函数可能并发执行。处理函数应响应 `ctx.RequestContext()`。
- 通过原指针使用 `Context`，避免复制整个结构。事件数据和 `Context` 的公开字段只供读取。

## 另请参阅

[消息富文本接收](../../server-api/messages/rich-content.md) · [按钮回调](../../server-api/messages/interactions/button-callbacks.md) · [官方事件结构](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html) · [官方群聊艾特消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)
