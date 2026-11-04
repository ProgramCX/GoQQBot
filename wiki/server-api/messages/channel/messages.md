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
- 统一入口发送成功但解析回包失败时，会返回结果，`err` 也不是 nil。`Raw` 保留平台回包。
- 旧方法把回包解码为 `message.SendMessageResponse`，其中 `Timestamp` 是 `int64`。收到 RFC3339 字符串时间戳会解码失败，消息仍可能已发出。
- 旧方法解码失败时没有带 `Raw` 的成功结果。此时重发可能产生重复消息。旧入口示例不表示响应模型已修复或经过真实联调。
- 隐藏撤回提示的权限由平台检查。[统一撤回入口](../reference-and-recall.md)中的 `Context.Recall` 和 `Client.RecallMessage` 仅支持 QQ 单聊和群聊。

## 另请参阅

[频道概述](README.md) · [频道表态](reactions.md) · [官方频道发送](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道撤回](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/recall.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
