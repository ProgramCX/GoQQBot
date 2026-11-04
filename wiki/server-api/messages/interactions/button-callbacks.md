# 处理按钮回调

用同一个按钮 ID 关联按钮和处理函数。为 `button.Callback(...).ID("refresh")` 注册 `OnButton("refresh", ...)`。在 `bot.Run` 前调用下面的注册函数。业务数据由 `ctx.Interaction()` 读取。

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
| `event.Raw` | 独立原始快照 |

`ctx.Interaction()` 不联网，也不要求 API 客户端。SDK 保留未知字段的原始数据。

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
- 成功 ACK 确认互动，后续耗时业务仍可能失败。回调中的后台协程需要明确的管理方。
- `event.Raw` 可能包含敏感信息。整体写入日志会暴露这些字段。
- 事件不一定下发结构体中的每个字段。`AuthorizeData` 是可选指针，读取前检查 nil。
- 类型 14 的成功码可能触发客户端清空会话提示。SDK 不会清空业务数据库。
- 发送消息无法代替 ACK。外层事件 ID 不适用于 ACK。
- 平台要求指令回调类场景在三秒内响应。线程池排队也计入耗时。SDK 不自动 ACK、不自动重试，也不保证端到端时限。
- 同一 `Context` 仅允许一次 ACK 尝试。并发调用及失败请求也计入次数。第二次返回 `types.ErrInteractionResponseAttempted`。调用前已取消的上下文不消耗次数。
- 确需重试时，业务显式调用 `ctx.Client.AcknowledgeInteraction(op, event.ID, code)`。首次请求可能已到达平台。
- 重复投递会创建新的 `Context`。一次 ACK 限制仅针对当前 `Context`，跨事件仍需用互动 ID 或业务键做幂等。

## 另请参阅

[发送回调按钮](buttons.md) · [快捷菜单](menus-and-panels.md) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html) · [QQ 官方：互动事件响应](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/interactions_interaction_id.put.html)
