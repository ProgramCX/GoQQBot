# 发送及撤回频道消息

普通文本和 Markdown 优先使用 `ctx.Reply`。需要发送 Embed 或使用完整频道协议结构时，旧的 `api/channel` 包仍提供入口，但应先了解其返回模型限制。

## 使用统一回复入口

将下面函数注册到频道消息事件；Markdown 在频道场景是否可用由平台权限决定。

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

统一发送结果使用 `message.ReplyResult`，其中 `Timestamp` 是 `time.Time`。发送成功但响应解析失败时会同时返回结果及错误，`Raw` 保存平台回包。

## 使用旧频道协议入口

以下函数演示 Embed 的结构和旧 API 用法，`channelID`、`sourceID` 由调用方明确提供。它是现有接口用法说明，不表示其响应模型已完成修复或真实联调。

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

**已知限制：**这些旧方法仍将回包解析到 `message.SendMessageResponse`，其 `Timestamp` 是 `int64`。平台返回 RFC3339 字符串时间戳时会解码失败，消息却可能已经发送成功。旧方法此时没有统一入口那样的带 Raw 成功结果，不应看到解码错误就自动重发。

## 撤回明确的频道消息

频道撤回当前使用旧入口。以下函数只撤回调用方明确传入的消息，不猜测或自动撤回接收消息。

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

最后的布尔参数为 `hideTip`，是否有权隐藏提示仍由平台检查。`Context.Recall` / `Client.RecallMessage` 目前只支持 QQ 单聊和群聊，不能代替上述频道入口。

## 另请参阅

[频道概述](README.md) · [频道表态](reactions.md) · [官方频道发送](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道撤回](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/recall.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
