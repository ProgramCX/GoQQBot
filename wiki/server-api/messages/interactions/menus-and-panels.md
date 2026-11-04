# 接收快捷菜单与指令面板事件

SDK 可以接收快捷菜单互动，并按 `feature_id` 分发。

## 接收已配置的菜单操作

先按官方方式配置菜单。在 `bot.Run` 前注册相同的功能 ID。

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

`OnCommand("help", eventType, handler)` 注册本进程的 [消息路由](../../../getting-started/events/routing.md)。命令按钮让用户发送消息。菜单操作可能产生类型 12 互动，应走互动处理入口。

用 [`Client.Request`](../../../advanced/raw-api.md) 显式调用未封装的管理接口。按官方接口定义请求和响应模型，并处理权限与错误。

## 注意事项

- SDK 尚未封装菜单查询与更新。指令面板（command panel）的增删改查也尚未封装。
- 收到菜单回调只表示互动已送达。`OnCommand` 和回调注册均不会创建 QQ 菜单或指令面板。
- `Client.Request` 提供原始请求能力。调用方自建模型与类型化 SDK 封装是两回事。

## 另请参阅

[按钮回调、ACK 与路由优先级](button-callbacks.md) · [SDK 接口覆盖清单](../../../../docs/api-coverage.md) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html) · [QQ 官方 API 导航：机器人、全局菜单与指令面板](https://bot.q.qq.com/wiki/develop/api-v2/)
