# 处理按钮回调

为 `button.Callback(...).ID("refresh")` 注册 `OnButton("refresh", ...)`。下面的注册函数在 `bot.Run` 之前调用，业务数据通过 `ctx.Interaction()` 读取。

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

示例假定刷新操作不需要额外授权；真实业务应在发送成功 ACK 前做快速授权判断。ACK 成功表示确认互动，不代表后续耗时业务已经完成。避免在回调中启动无人管理的后台协程。

## 路由与读取字段

路由优先级为 `OnButton` / `OnMenu` → `OnInteraction(kind, handler)` → `OnEvent(message.INTERACTION_CREATE, handler)`。每次只选一个处理函数；没有按钮 ID 时可以按交互类型自行解析 `ButtonData`。

| 字段或方法 | 含义 |
| --- | --- |
| `event.ID` | 互动正文 `d.id`，用于 ACK |
| `ctx.Payload.ID` | 外层事件 ID，用于互动产生的被动消息回复 |
| `event.Kind()` | 优先外层 type，缺失时回退到 data.type |
| `event.Data.Resolved.ButtonID` | 按钮标识 |
| `event.Data.Resolved.ButtonData` | 按钮携带的业务数据 |
| `event.Scene`、会话和用户字段 | 当前场景及操作者信息，按平台实际下发读取 |
| `event.Raw` | 独立原始快照，可能包含敏感信息，不应整体记录 |

`ctx.Interaction()` 不联网、不要求 API 客户端，并保留未知字段的原始数据。

## 交互类型

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

结构体字段存在不表示每个事件都必填。`AuthorizeData` 是可选指针，读取前检查 nil。未知类型可使用显式数值注册 `OnInteraction`，或交给事件兜底；SDK 不猜测未知字段的业务含义。

## 正确 ACK

类型 11 和 12 必须 ACK，可通过 `event.NeedsAcknowledgement()` 判断。其他类型可按需显式确认，例如类型 14 使用成功码可能触发客户端清空会话提示；SDK 不会清空你的业务数据库。

| 响应码 | SDK 常量 |
| --- | --- |
| 0 成功 | `InteractionCallbackCodeSuccess` |
| 1 失败 | `InteractionCallbackCodeFailed` |
| 2 频繁 | `InteractionCallbackCodeFrequent` |
| 3 重复 | `InteractionCallbackCodeDuplicate` |
| 4 无权限 | `InteractionCallbackCodeNoPermission` |
| 5 仅管理员 | `InteractionCallbackCodeAdminOnly` |

ACK 请求为 `PUT /interactions/{d.id}`；`ctx.Reply` 则使用外层事件 ID 作为 `event_id`。发送回复不能代替 ACK，也不能把外层 ID 传给 ACK。

指令回调类场景要求三秒内响应，线程池排队也会占用时间。SDK 不自动 ACK、不自动重试，也不保证端到端时限。

同一 `Context` 仅允许一次 ACK 尝试，包含并发调用及失败请求；第二次返回 `types.ErrInteractionResponseAttempted`。调用前已取消的上下文不消耗尝试次数。确需重试时，由业务明确使用 `ctx.Client.AcknowledgeInteraction(op, event.ID, code)`，并处理首次请求可能已到达平台的不确定性。

重复投递会创建新的 Context，仍需以互动 ID 或业务键做幂等；一次 ACK 限制不等于跨事件去重。

## 另请参阅

- [发送回调按钮](buttons.md)
- [快捷菜单](menus-and-panels.md)
- [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html)
- [QQ 官方：互动事件响应](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/interactions_interaction_id.put.html)
