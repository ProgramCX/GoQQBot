# 引用与撤回消息

回复来源、引用索引和撤回所需的消息 ID 用途不同，应分别保存。下表说明何时使用哪一种标识。

## 先区分标识

按目标选择对应的标识。

| 标识 | 来源 | 用途 |
| --- | --- | --- |
| 当前接收消息 `d.id` | `m.ID` | 当前消息的被动回复来源，由 Context 自动填写 |
| 当前消息 `msg_idx` | `m.MessageScene.MessageIndex()` | 展示引用当前消息 |
| 被引用消息 `ref_msg_idx` | `m.MessageScene.ReferencedIndex()` | 读取平台提供的引用上下文 |
| 机器人发送结果 `id` | `result.ID` | 撤回刚发送的消息 |
| 机器人展示索引 `ref_idx` | `result.ExtInfo.RefIndex` | 后续展示引用该机器人消息 |
| 交互信封 `id` | `ctx.Payload.ID` | 交互消息回复的 `event_id`，不同于[交互 ACK](interactions/button-callbacks.md)使用的 `d.id` |

## 引用当前接收消息

有展示索引时设置引用。没有索引时退回普通回复，避免猜测或伪造索引。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyWithReference(ctx *types.Context) error {
	m, err := ctx.Message()
	if err != nil {
		return err
	}
	reply := message.Text("已收到这条消息。")
	if index := m.MessageScene.MessageIndex(); index != "" {
		reply = reply.Reference(index)
	}
	return ctx.Reply(reply)
}
```

`Reference` 只设置展示引用，不改变被动回复使用的 `msg_id`。

## 发送后撤回机器人回复

下面的处理函数发送回复后，使用返回的消息 ID 撤回该回复。实际应用可先保存 ID，在业务明确要求撤回时再调用接口。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func SendAndRecall(ctx *types.Context) error {
	result, err := ctx.SendReply(ctx.RequestContext(), message.Text("这是一条撤回演示消息。"))
	if err != nil {
		return err
	}
	return ctx.Recall(result.ID)
}
```

`Context.Recall` 使用当前会话。后台业务可用 `Client.RecallMessage(ctx, target, messageID)` 指定单聊或群聊目标。

## 权限和引用可靠性

按发送场景选撤回入口。频道撤回需要隐藏提示时，使用[频道专用入口](channel/messages.md)。

## 注意事项

- 通过 `errors.Is` 判断下表中的错误类别。通过 `errors.As` 读取 `*api.APIError`，可取得原始错误码、HTTP 状态及 TraceID。通用错误或未识别的错误码也会保留在 `*api.APIError` 中。
- QQ 返回查询或撤回失败时，先确认消息是否仍在，再决定是否重试。SDK 不会自动重试撤回。

| 平台错误码 | SDK 错误 | 含义 |
| --- | --- | --- |
| `40012010` | `api.ErrAPIAccessDenied` | 应用没有接口访问权限 |
| `40062003`、`306004` | `api.ErrOperationDenied` | 机器人没有操作权限 |
| `40064004` | `api.ErrRecallWindowExpired` | 超过撤回时限 |
| `40061001`、`306001` | `api.ErrInvalidAPIParameters` | 请求参数无效 |
| `40061002`、`306002` | `api.ErrInvalidMessageID` | 消息 ID 无效 |
| `306003` | `api.ErrRecallMessageLookup` | QQ 未能获取消息 |
| `306006` | `api.ErrRecallChannelLookup` | QQ 未能获取子频道 |
| `50065001`、`306005` | `api.ErrRecallFailed` | QQ 报告撤回失败 |

- `MessageScene` 为 nil 时，`MessageIndex` 返回空串。
- `Context.Recall` 不默认撤回接收消息，也不从 REFIDX 推算消息 ID。
- 单聊只允许撤回机器人消息。群聊撤回成员消息需要机器人管理员权限。QQ 按平台规则检查两分钟窗口和最终资格。
- `Context.Recall` 支持子频道和频道私信。频道私信只能撤回机器人自己的消息。
- 多级引用可能收到不一致的 `msg_idx/ref_msg_idx`、`TMP_` 索引或折叠后的文本摘要。SDK 原样保留。
- SDK 不保证全局稳定 ID 或恢复完整引用链。展示索引不足以支撑业务桥接、消息关联和历史索引。

## 另请参阅

[读取富文本](rich-content.md) · [官方单聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages_message_id.delete.html) · [官方群聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages_message_id.delete.html) · [官方群聊引用字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
