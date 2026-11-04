# 收发频道消息

子频道消息和频道私信可以共用 `ctx.Reply`。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - PUBLIC_GUILD_MESSAGES
      - DIRECT_MESSAGE
```

[全部 Intent 与下属事件](../../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`AT_MESSAGE_CREATE`（子频道 @机器人）和 `DIRECT_MESSAGE_CREATE`（频道私信），按实际场景勾选。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

## 接收并回复频道 @消息

在机器人运行前注册以下函数。通过 WebSocket 订阅频道 @消息时选择 `PUBLIC_GUILD_MESSAGES`。频道私信选择 `DIRECT_MESSAGE`。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterChannel(bot *goqqrobot.Bot) {
	bot.OnMessage(message.AT_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("已收到频道消息。")
	})
	bot.OnMessage(message.DIRECT_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("已收到频道私信。")
	})
}
```

## ID 的区别

`ctx.Reply` 根据当前事件选择目标。各场景的标识见下表。

| 场景 | `message.Scene` | 发送目标 |
| --- | --- | --- |
| 子频道消息 | `SceneChannel` | `channel_id` |
| 频道私信 | `SceneDirectMessage` | 私信会话的 `guild_id` |
| QQ 单聊 | `SceneC2C` | `user_openid` |

## 当前覆盖范围

`ctx.ChannelMessage()` 可读取频道专用字段，包括成员、消息序号和引用。审核及表态事件通过 `OnEvent` 注册，分别用 `ctx.ChannelMessageAudit()` 和 `ctx.ChannelReaction()` 解析。表态事件可用 `ctx.Reply` 回复当前子频道，SDK 使用外层事件 ID 作为回复来源。

## 注意事项

- 订阅仍需匹配平台开放的权限。换成 `MESSAGE_CREATE` 不会开通全量消息权限。
- 频道消息和私信与 QQ 群聊、单聊使用不同 ID 和权限。频道本身的 ID 无法替代频道私信会话的 `guild_id`。
- 发送频道私信前，可用 `CreateDirectSession` 与同一频道内的成员建立会话。频道资料、权限和内容操作见[频道管理](../../channel-management/README.md)。

## 另请参阅

[频道发送与撤回](messages.md) · [频道表情表态](reactions.md) · [官方发送子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
