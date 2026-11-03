# 快捷菜单与指令面板

当前 SDK 可以接收快捷菜单互动，并按 `feature_id` 分发。**尚未提供菜单查询更新、指令面板创建及增删改查的领域封装。** 接收到回调，不等于 SDK 已替你创建平台资源。

## 接收已配置的菜单操作

先在官方支持的方式中配置菜单，再在 `bot.Run` 前注册相同的功能 ID：

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

`OnMenu` 仅匹配类型 12 的 `feature_id`。需要统一处理多个功能时，注册 `OnInteraction(interaction.InteractionTypeMenu, handler)`；具体 ID 处理函数更优先。

## 命令注册与平台资源

`OnCommand("help", eventType, handler)` 注册本进程的消息路由，不会创建 QQ 菜单或指令面板。命令按钮让用户发送消息，菜单操作则可能产生类型 12 互动，处理方式不能混用。

未封装的管理接口可以通过 `Client.Request` 显式调用，但调用方需要自行根据官方接口定义请求及响应模型、权限和错误处理；这不算已具备完整类型化 SDK。

## 另请参阅

- [按钮回调、ACK 与路由优先级](button-callbacks.md)
- [SDK 接口覆盖清单](../../../../docs/api-coverage.md)
- [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html)
- [QQ 官方 API 导航：机器人、全局菜单与指令面板](https://bot.q.qq.com/wiki/develop/api-v2/)
