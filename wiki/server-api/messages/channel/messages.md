# 发送及撤回频道消息

发送普通文本和 Markdown 时优先用 `ctx.Reply`。发送嵌入消息（Embed）或完整频道协议结构时，用旧的 `api/channel` 包。

## 使用统一回复入口

将下面的函数注册到频道消息事件。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyChannelMarkdown(ctx *types.Context) error {
	return ctx.Reply(message.Markdown("## 今日公告\n服务运行正常。"))
}
```

统一发送结果为 `message.ReplyResult`。`Timestamp` 的类型是 `time.Time`。

## 使用旧频道协议入口

以下函数演示 Embed 结构与旧 API。调用方显式提供 `channelID` 和 `sourceID`。

```go
package guide

import (
	channelAPI "github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelMessage "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func SendChannelEmbed(ctx *types.Context, channelID, sourceID string) (*message.SendMessageResponse, error) {
	return channelAPI.SendEmbed(ctx, channelID, &channelMessage.MessageEmbed{
		Title:  "服务状态",
		Prompt: "服务状态更新",
		Fields: []channelMessage.MessageEmbedField{{Name: "运行正常"}},
	}, sourceID)
}
```

## 撤回明确的频道消息

通过旧入口指定要撤回的消息 ID。以下函数不会推断或自动撤回接收消息。

```go
package guide

import (
	channelAPI "github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RecallChannel(ctx *types.Context, channelID, messageID string) error {
	return channelAPI.RecallChannelMessage(ctx, channelID, messageID, false)
}

func RecallChannelDM(ctx *types.Context, dmGuildID, messageID string) error {
	return channelAPI.RecallDM(ctx, dmGuildID, messageID, false)
}
```

最后的布尔参数是 `hideTip`，用于请求隐藏撤回提示。

## 注意事项

- 频道 Markdown 能否使用由平台权限决定。
- 如果统一入口已经发出消息，但未能解析 QQ 的响应，调用会同时返回结果和错误。响应原文保存在结果的 `Raw` 中。
- 旧方法将响应解析为 `message.SendMessageResponse`，其中 `Timestamp` 是 `int64`。收到 RFC3339 字符串时间戳会解码失败，消息仍可能已发出。
- 旧方法解析失败时，不会返回带 `Raw` 的发送结果。由于消息可能已经发出，直接重发可能造成重复。旧入口示例不表示响应模型已修复或经过真实联调。
- 隐藏撤回提示的权限由平台检查。[统一撤回入口](../reference-and-recall.md)中的 `Context.Recall` 和 `Client.RecallMessage` 仅支持 QQ 单聊和群聊。

## 另请参阅

[频道概述](README.md) · [频道表态](reactions.md) · [官方频道发送](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道撤回](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/recall.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
