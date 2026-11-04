# 处理按钮回调

点下回调按钮后，SDK 会按按钮 ID 找到处理函数。给按钮设为 `button.Callback(...).ID("refresh")`，再用 `OnButton("refresh", ...)` 注册处理函数。注册要在 `bot.Run` 之前完成。处理函数里用 `ctx.Interaction()` 读取按钮带来的数据。

```go
package guide

import (
	"context"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func registerRefresh(bot *goqqrobot.Bot, refresh func(context.Context, string) error) {
	bot.OnButton("refresh", func(ctx *types.Context) error {
		event, err := ctx.Interaction()
		if err != nil {
			return err
		}
		ackCtx, cancel := context.WithTimeout(ctx.RequestContext(), 2*time.Second)
		err = ctx.AckInteractionContext(ackCtx, interaction.InteractionCallbackCodeSuccess)
		cancel()
		if err != nil {
			return err
		}
		if err := refresh(ctx.RequestContext(), event.Data.Resolved.ButtonData); err != nil {
			return err
		}
		return ctx.ReplyText("刷新完成")
	})
}
```

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

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`INTERACTION_CREATE`（互动事件）。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

## 路由与读取字段

SDK 按以下优先级选择一个处理函数。

- `OnButton` 或 `OnMenu`。
- `OnInteraction(kind, handler)`。
- `OnEvent(message.INTERACTION_CREATE, handler)`。

没有按钮 ID 时，可按交互类型解析 `ButtonData`。

| 字段或方法 | 含义 |
| --- | --- |
| `event.ID` | 互动正文 `d.id`，用于 ACK |
| `ctx.Payload.ID` | 外层事件 ID，用于互动产生的被动消息（passive message）回复 |
| `event.Kind()` | 优先使用外层 `type`，缺失时退回 `data.type` |
| `event.Data.Resolved.ButtonID` | 按钮标识 |
| `event.Data.Resolved.ButtonData` | 按钮携带的业务数据 |
| `event.Scene`、会话和用户字段 | 场景及操作者信息，按平台实际下发读取 |
| `event.Raw` | 原始事件数据的副本 |

`ctx.Interaction()` 直接读取本地收到的事件。即使没有配置 API 客户端，也可以用这个方法。SDK 不认识的字段仍会留在原始数据里。

## 交互类型

根据平台事件类型选择 SDK 常量和读取字段。

| 值 | SDK 常量 | 数据入口 |
| --- | --- | --- |
| 11 | `InteractionTypeButton` | `Resolved.ButtonID`、`ButtonData` |
| 12 | `InteractionTypeMenu` | `Resolved.FeatureID` |
| 13 | `InteractionTypeFeedback` | `Resolved.FeedbackOpt`、`Checked` 等实际字段 |
| 14 | `InteractionTypeClearSession` | `Resolved.MessageScene` 等实际字段 |
| 15 | `InteractionTypeStory` | `Resolved.Action`、`MessageScene` |
| 16 | `InteractionTypeSwitchModel` | `Resolved.MessageScene` 等实际字段 |
| 18 | `InteractionTypeUserAuthorize` | `Resolved.AuthorizeData` |
| 19 | `InteractionTypeGroupAuthorize` | `Resolved.AuthorizeData` |
| 20 | `InteractionTypeGroupAuthorizeStatus` | `Resolved.AuthorizeData` |

未知类型可用显式数值注册 `OnInteraction`，也可交给事件兜底。SDK 不猜测未知字段的业务含义。

## 正确 ACK

用 `event.NeedsAcknowledgement()` 判断是否需要 ACK。平台要求类型 11 和 12 回复 ACK。其他类型可按需确认。

| 响应码 | SDK 常量 |
| --- | --- |
| 0 成功 | `InteractionCallbackCodeSuccess` |
| 1 失败 | `InteractionCallbackCodeFailed` |
| 2 频繁 | `InteractionCallbackCodeFrequent` |
| 3 重复 | `InteractionCallbackCodeDuplicate` |
| 4 无权限 | `InteractionCallbackCodeNoPermission` |
| 5 仅管理员 | `InteractionCallbackCodeAdminOnly` |

ACK 请求为 `PUT /interactions/{d.id}`。`ctx.Reply` 使用外层事件 ID 作为 `event_id`，见 [API 调用](../../../getting-started/api-calls.md)。

## 注意事项

- 示例假定刷新操作不需要额外授权。真实业务在发送成功 ACK 前需快速判断权限。
- 发出成功 ACK 后，后台操作仍可能失败。如果回调把操作交给后台协程，应用要负责管理这些协程。
- `event.Raw` 可能包含敏感信息。整体写入日志会暴露这些字段。
- 事件不一定下发结构体中的每个字段。`AuthorizeData` 是可选指针，读取前检查 nil。
- 类型 14 的成功码可能触发客户端清空会话提示。SDK 不会清空业务数据库。
- 发送消息无法代替 ACK。外层事件 ID 不适用于 ACK。
- 平台要求指令回调类场景在三秒内响应。线程池排队也计入耗时。SDK 不自动 ACK、不自动重试，也不保证端到端时限。
- 同一 `Context` 仅允许一次 ACK 尝试。并发调用及失败请求也计入次数。第二次返回 `types.ErrInteractionResponseAttempted`。调用前已取消的上下文不消耗次数。
- 确需重试时，业务显式调用 `ctx.Client.AcknowledgeInteraction(op, event.ID, code)`。首次请求可能已到达平台。
- 同一次互动重复送达时，每次投递都会创建新的 `Context`，ACK 次数也分别计算。业务可以根据互动 ID 或业务键识别重复请求，避免重复执行操作。

## 另请参阅

[发送回调按钮](buttons.md) · [快捷菜单](menus-and-panels.md) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html) · [QQ 官方：互动事件响应](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/interactions_interaction_id.put.html)
