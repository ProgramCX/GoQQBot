# 发送及撤回频道消息

发送普通文本和 Markdown 时优先用 `ctx.Reply`。发送嵌入消息（Embed）、图片 URL 或本地图片时，使用 `channel.NewClient(client)` 创建频道客户端。

## 配置事件订阅

下列订阅用于接收审核结果。发送与撤回接口由应用直接调用。

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - MESSAGE_AUDIT
```

[全部 Intent 与下属事件](../../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`MESSAGE_AUDIT_PASS` 和 `MESSAGE_AUDIT_REJECT`。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

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

以下兼容示例演示 Embed 结构。新代码可调用 `SendChannelMessage(ctx, channelID, body, image)`，最后一个参数为 `nil` 时发送 JSON。发送本地图片时，传入 `&channel.LocalImage{Filename: "image.png", Data: data}`。

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

`ctx.Recall(messageID)` 撤回当前会话中的指定消息。需要隐藏撤回提示时，使用频道客户端的 `RecallChannelMessage` 或 `RecallDirectMessage`，并传入 `RecallOptions{HideTip: true}`。以下为兼容入口示例。

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

## 创建并发送频道私信

用 `CreateDirectSession(ctx, guildID, recipientID)` 创建会话，再把返回的 `GuildID` 传给 `SendDirectMessage`。`recipientID` 是同一频道内的成员 ID。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

func SendDirect(ctx context.Context, client *channel.Client, guildID, memberID string) (*message.ReplyResult, error) {
	session, err := client.CreateDirectSession(ctx, guildID, memberID)
	if err != nil {
		return nil, err
	}
	return client.SendDirectMessage(ctx, session.GuildID, channelmsg.ChannelSendMessage{
		Content: "你好，这是一条频道私信。",
	}, nil)
}
```

## 注意事项

频道私信的本地图片发送已在 QQ 客户端测试通过。

::: warning 频道私信撤回需要私域权限
测试频道私信撤回时，QQ 返回 `40012010`（应用无接口访问权限）。当前测试应用无法验证撤回成功的情况。QQ 仅向私域机器人开放此接口。权限开通后，请将机器人从频道移除，再重新添加。
:::

- 需要读取发送响应中的作者、附件和频道字段时，调用 `channel.DecodeSentMessage(result)`。如果解析失败，先确认消息是否已发出，再决定是否重试。

- 返回平台错误码 `304023` 或 `304024` 时，可以通过 `errors.As` 读取 `*api.APIError` 的 `MessageAuditID`。订阅 `MESSAGE_AUDIT` 后，使用审核事件核对结果。

- 频道 Markdown 能否使用由平台权限决定。
- 如果统一入口已经发出消息，但未能解析 QQ 的响应，调用会同时返回结果和错误。响应原文保存在结果的 `Raw` 中。
- 兼容入口保留 `message.SendMessageResponse.Timestamp` 的 `int64` 类型，并将 QQ 返回的 RFC3339 时间转换为 Unix 秒。
- 旧方法解析失败时，不会返回带 `Raw` 的发送结果。由于消息可能已经发出，直接重发可能造成重复。旧入口示例不表示响应模型已修复或经过真实联调。
- 隐藏撤回提示的权限由平台检查。子频道、频道私信和 QQ 群聊、单聊均可使用[统一撤回入口](../reference-and-recall.md)。

## 另请参阅

[频道概述](README.md) · [频道表态](reactions.md) · [官方频道发送](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道撤回](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/recall.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
