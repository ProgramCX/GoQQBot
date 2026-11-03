# 频道消息

频道中的子频道消息和频道私信可以使用统一上下文回复；它们与 QQ 群聊、QQ 单聊的 ID 和权限不同。

## 接收并回复频道 @消息

以下函数在运行前注册。WebSocket 接收频道 @消息时选择 `PUBLIC_GUILD_MESSAGES`；频道私信选择 `DIRECT_MESSAGE`。订阅仍需匹配平台实际开放的权限。

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

`MESSAGE_CREATE` 是另一种频道消息事件，接收范围受权限限制；不要通过更换事件名假定已经取得全量消息权限。

## ID 的区别

| 场景 | `message.Scene` | 发送目标 |
| --- | --- | --- |
| 子频道消息 | `SceneChannel` | `channel_id` |
| 频道私信 | `SceneDirectMessage` | 私信会话的 `guild_id` |
| QQ 单聊 | `SceneC2C` | `user_openid` |

统一 `ctx.Reply` 根据当前事件选择目标。不要把频道私信的 `guild_id` 替换为普通频道服务器 ID。

## 当前覆盖范围

已有发送、频道私信发送、撤回和表态旧封装；没有完整的频道成员、身份组、公告、帖子、日程等管理封装。创建频道私信会话的 `POST /users/@me/dms` 也尚未封装。

## 另请参阅

[频道发送与撤回](messages.md) · [频道表情表态](reactions.md) · [官方发送子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) · [官方频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html)
