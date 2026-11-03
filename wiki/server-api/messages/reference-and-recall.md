# 引用与撤回消息

被动回复来源、展示引用索引和撤回消息 ID 各自用途不同。不要把三者都存成一个没有语义的 `id`。

## 先区分标识

| 标识 | 来源 | 用途 |
| --- | --- | --- |
| 当前接收消息 `d.id` | `m.ID` | 当前消息的被动回复来源，由 Context 自动填写 |
| 当前消息 `msg_idx` | `m.MessageScene.MessageIndex()` | 展示引用当前消息 |
| 被引用消息 `ref_msg_idx` | `m.MessageScene.ReferencedIndex()` | 读取平台提供的引用上下文 |
| 机器人发送结果 `id` | `result.ID` | 撤回刚发送的消息 |
| 机器人展示索引 `ref_idx` | `result.ExtInfo.RefIndex` | 后续展示引用该机器人消息 |
| 交互信封 `id` | `ctx.Payload.ID` | 交互消息回复的 `event_id`，不同于交互 ACK 使用的 `d.id` |

## 引用当前接收消息

以下处理函数在没有展示索引时退回普通回复，不猜测或伪造索引。

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

`MessageScene` 为 nil 时其 `MessageIndex` 方法安全返回空串。`Reference` 只设置展示引用，不改变被动回复所用的 `msg_id`。

## 发送后撤回机器人回复

以下单聊或群聊处理函数演示获取发送 ID，然后立即撤回这次回复。实际业务通常保存 ID，在明确的业务动作中撤回。

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

`Context.Recall` 使用当前会话；后台业务可使用 `Client.RecallMessage(ctx, target, messageID)` 指定单聊或群聊目标。它不会默认撤回当前接收消息，也不会从 REFIDX 推算消息 ID。

## 权限和引用可靠性

单聊只能撤回机器人消息；群聊撤回成员消息需要机器人管理员权限。两分钟窗口及最终资格由 QQ 检查。频道撤回使用[频道专用旧入口](channel/messages.md)，不在 Context.Recall 的支持范围内。

平台可能在多级引用中下发不一致的 `msg_idx/ref_msg_idx`、`TMP_` 索引或折叠后的文本摘要。SDK 原样保留，不承诺全局稳定 ID 或完整引用链恢复。因此业务桥接、消息关联和历史索引不能只依赖这些展示索引。

## 另请参阅

[读取富文本](rich-content.md) · [官方单聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages_message_id.delete.html) · [官方群聊撤回](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages_message_id.delete.html) · [官方群聊引用字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
