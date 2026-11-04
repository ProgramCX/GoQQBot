# 收发频道消息

子频道消息和频道私信可以共用 `ctx.Reply`。

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

SDK 已有频道发送、私信发送、撤回和表态的旧封装。

## 注意事项

- 订阅仍需匹配平台开放的权限。换成 `MESSAGE_CREATE` 不会开通全量消息权限。
- 频道消息和私信与 QQ 群聊、单聊使用不同 ID 和权限。普通频道服务器 ID 无法替代私信会话的 `guild_id`。
- SDK 尚未完整封装[频道管理](../../channel-management/README.md)，包括成员、身份组、公告、帖子和日程。
- 创建频道私信会话的 `POST /users/@me/dms` 尚未封装。

## 另请参阅

[频道发送与撤回](messages.md) · [频道表情表态](reactions.md) · [官方发送子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
