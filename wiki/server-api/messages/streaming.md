# 流式发送单聊消息

流式消息可以在内容陆续生成时，分段更新同一条单聊回复。每次调用传入一个片段。

## 顺序发送两段内容

将下面的函数注册到单聊事件或命令。第一段从 `Index=0` 开始，后续片段携带上一次响应 ID。

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

发送模型生成的内容时，先等待当前片段发送完成，再把返回的最新 ID 用于发送下一段。

## 参数约束

根据更新方式和生成状态设置以下字段。

| 字段 | 说明 |
| --- | --- |
| `InputMode` | 省略或填 `append` 表示追加，`replace` 表示替换 |
| `InputState` | `1` 生成中，`10` 结束 |
| `Index` | 从零开始，后续片段由调用方递增 |
| `ContentType` | `text` 或 `markdown` |
| `StreamMessageID` | 第一段为空，后续填写前一次响应 ID |
| `Sequence` | Context 默认自动分配回复序号，也可显式管理 |
| `Wakeup` | 显式请求唤醒 |

## 错误恢复

后台业务可用 `Client.SendStreamChunk`。调用时显式提供 `ReplyTarget`，见[指定回复目标](../../getting-started/api-calls.md)。

## 注意事项

- SDK 逐段发送，不聚合生成结果，也不启动后台刷新。流状态、顺序及持久化由业务负责。
- 并发发送同一条流的片段可能破坏顺序。替换内容需保留已发送的前缀。
- `StreamResult.RemainingLength` 是QQ 返回的剩余字符数，决定还能追加多少内容。
- 群聊、频道和频道私信不支持该入口。
- 普通片段需提供回复来源。只有显式使用 `Wakeup` 才可无来源唤醒。平台仍检查资格和配额。
- 平台接收后仍可能遇到请求取消、连接中断或响应解析失败。SDK 不重试，也不自动从旧 ID 恢复。
- 返回了结果但 `err` 也不是 nil 时，先检查保留的 `Raw`。直接重发可能重复片段。

## 另请参阅

[发送单聊消息](c2c/send.md) · [引用与撤回](reference-and-recall.md) · [官方单聊流式接口](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_stream_messages.post.html)
