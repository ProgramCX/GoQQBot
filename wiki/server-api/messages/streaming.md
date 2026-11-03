# 流式发送单聊消息

流式消息允许业务分段更新同一条单聊响应。SDK 发送每个明确片段，不自动聚合生成结果、启动后台刷新或管理业务持久化。

## 顺序发送两段内容

下面函数仅注册到单聊事件或命令。第一段从 `Index=0` 开始，后续片段带上前一次响应的 ID。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyStream(ctx *types.Context) error {
	first, err := ctx.SendStreamChunk(ctx.RequestContext(), message.StreamChunk{
		Index:       0,
		InputMode:   "append",
		InputState:  1,
		ContentType: "text",
		Content:     "第一段结果。",
	})
	if err != nil {
		return err
	}
	_, err = ctx.SendStreamChunk(ctx.RequestContext(), message.StreamChunk{
		Index:           1,
		InputMode:       "append",
		InputState:      10,
		ContentType:     "text",
		Content:         "第二段结果，生成结束。",
		StreamMessageID: first.ID,
	})
	return err
}
```

实际接入模型生成时，在业务循环中逐次等待发送完成，并将最新结果 ID 传给下一段。不要并发发送同一条流的片段。

## 参数约束

| 字段 | 说明 |
| --- | --- |
| `InputMode` | 省略或 `append` 为追加，`replace` 为替换；替换应保留已发送前缀 |
| `InputState` | `1` 生成中，`10` 结束 |
| `Index` | 从零开始；后续片段由调用方递增 |
| `ContentType` | `text` 或 `markdown` |
| `StreamMessageID` | 第一段为空，后续填写前一次响应 ID |
| `Sequence` | Context 默认自动分配回复序号，也可显式管理 |
| `Wakeup` | 显式请求唤醒，仍受平台资格和配额检查 |

`StreamResult.RemainingLength` 是平台返回的剩余字符预算。不要假定任意长度的生成结果都可以继续追加。

## 错误恢复

群聊、频道和频道私信不支持该入口。`Client.SendStreamChunk` 是后台业务的底层入口，需要明确 `ReplyTarget`，普通片段必须有回复来源；仅显式 Wakeup 允许无来源唤醒。

请求取消、连接中断或回包解码失败可能发生在平台已经接收之后。SDK 不重试，也不自动从旧 ID 恢复。调用方负责流状态、顺序及持久化；非 nil 结果伴随错误时先检查保留的 Raw，避免盲目重复片段。

## 另请参阅

[发送单聊消息](c2c/send.md) · [引用与撤回](reference-and-recall.md) · [官方单聊流式接口](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_stream_messages.post.html)
