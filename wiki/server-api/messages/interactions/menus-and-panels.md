# 接收快捷菜单与指令面板事件

用户点击快捷菜单后，SDK 可以根据 `feature_id` 调用对应的处理函数。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - INTERACTION
```

[全部 Intent 与下属事件](../../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`INTERACTION_CREATE`（互动事件）。命令按钮发送的普通消息仍按会话类型订阅。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

## 接收已配置的菜单操作

先通过 QQ 平台或机器人客户端配置菜单，再用菜单的功能 ID 注册处理函数。注册要在 `bot.Run` 前完成。

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

## 管理菜单与指令面板

用 `robot.NewClient(client)` 创建机器人客户端，复用现有的 `*api.Client`。

| 任务 | SDK 方法 | 官方接口 |
| --- | --- | --- |
| 查看全局菜单及版本 | `GetMenu` | [查询全局菜单](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_menu.get.html) |
| 替换全局菜单配置 | `UpdateMenu` | [修改全局菜单](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_menu.put.html) |
| 按场景查询面板列表 | `ListPanels` | [查询面板列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels.get.html) |
| 创建面板并取得 ID | `CreatePanel` | [创建面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels.post.html) |
| 获取面板配置与关联对象 | `GetPanel` | [查询面板详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.get.html) |
| 替换面板元素和备注 | `UpdatePanel` | [修改面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.put.html) |
| 删除指定面板 | `DeletePanel` | [删除面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.delete.html) |
| 增删指定用户或群的关联 | `UpdatePanelTarget` | [修改关联对象](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id_target.put.html) |

查询面板列表时填写 `scope`，需要下一页时传入上一次返回的 `next_cursor`。创建面板后，保存返回的 `panel_id`，供后续查询、修改和删除使用。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
)

func ListGroupPanels(ctx context.Context, client *api.Client, cursor string) (*robot.PanelPage, error) {
	return robot.NewClient(client).ListPanels(ctx, robot.PanelScope("group"), cursor, 20)
}
```

## 注意事项

- 全局菜单仅适用于 QQ 单聊。更新会替换完整配置，并影响所有单聊用户。修改前先读取并保存当前配置。
- 收到菜单回调只表示互动已送达。`OnCommand` 和回调注册均不会创建 QQ 菜单或指令面板。
- 指定对象的面板只适用于单聊和群聊。文字子频道和频道私信只支持全局面板。修改面板配置会保留已有的关联对象，增删关联时使用 `UpdatePanelTarget`。

## 另请参阅

[按钮回调、ACK 与路由优先级](button-callbacks.md) · [SDK 接口覆盖清单](../../../../docs/api-coverage.md) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html) · [QQ 官方 API 导航：机器人、全局菜单与指令面板](https://bot.q.qq.com/wiki/develop/api-v2/)
