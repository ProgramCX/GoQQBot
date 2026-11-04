# 接收快捷菜单与指令面板事件

用户点击快捷菜单后，SDK 可以根据 `feature_id` 调用对应的处理函数。

## 接收已配置的菜单操作

先在 QQ 平台配置菜单，再用菜单的功能 ID 注册处理函数。注册要在 `bot.Run` 前完成。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func registerMenu(bot *goqqrobot.Bot) {
	bot.OnMenu("help", func(ctx *types.Context) error {
		event, err := ctx.Interaction()
		if err != nil {
			return err
		}
		if event.Data.Resolved.FeatureID != "help" {
			return nil
		}
		if err := ctx.AckInteraction(interaction.InteractionCallbackCodeSuccess); err != nil {
			return err
		}
		return ctx.ReplyText("发送 /help 查看可用命令。")
	})
}
```

`OnMenu` 仅匹配平台类型 12 的 `feature_id`。用 `OnInteraction(interaction.InteractionTypeMenu, handler)` 统一处理多个功能。具体 ID 的处理函数优先，见 [回调路由](button-callbacks.md)。

## 命令注册与平台资源

`OnCommand("help", eventType, handler)` 注册本进程的 [消息路由](../../../getting-started/events/routing.md)。命令按钮让用户发送消息。如果点击菜单产生了类型 12 的互动事件，用 `OnMenu` 或 `OnInteraction` 接收。

用 [`Client.Request`](../../../advanced/raw-api.md) 调用 SDK 还没封装的管理接口。请求和响应的结构按官方文档定义，权限和错误由应用处理。

## 注意事项

- SDK 尚未封装菜单查询与更新。指令面板（command panel）的增删改查也尚未封装。
- 收到菜单回调只表示互动已送达。`OnCommand` 和回调注册均不会创建 QQ 菜单或指令面板。
- `Client.Request` 可以直接调用 QQ 的接口。请求和响应的结构由应用按官方文档定义。

## 另请参阅

[按钮回调、ACK 与路由优先级](button-callbacks.md) · [SDK 接口覆盖清单](../../../../docs/api-coverage.md) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html) · [QQ 官方 API 导航：机器人、全局菜单与指令面板](https://bot.q.qq.com/wiki/develop/api-v2/)
