# 添加和读取频道表情表态

表情表态（reaction）把表情附在已有频道消息上。用 `channel.NewClient(client)` 创建频道客户端，再调用 `AddReaction`、`RemoveReaction` 或 `ListReactionUsers`。

## 配置事件订阅

下列订阅用于接收用户表态通知。主动添加、删除和查询表态可直接调用接口。

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GUILD_MESSAGE_REACTIONS
```

[全部 Intent 与下属事件](../../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`MESSAGE_REACTION_ADD` 和 `MESSAGE_REACTION_REMOVE`。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../../getting-started/events/webhook.md)

## 添加一个系统表情表态

以下函数接收平台支持的表情 ID。

```go
package guide

import (
	channelAPI "github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReactToChannel(ctx *types.Context, channelID, messageID, emojiID string) error {
	return channelAPI.AddReaction(ctx, channelID, messageID, interaction.EmojiTypeSystem, emojiID)
}

func RemoveChannelReaction(ctx *types.Context, channelID, messageID, emojiID string) error {
	return channelAPI.RemoveReaction(ctx, channelID, messageID, interaction.EmojiTypeSystem, emojiID)
}
```

平台将 `EmojiTypeSystem` 定为类型 1，将 `EmojiTypeCustom` 定为类型 2。

## 查询第一页用户

下面的示例查询第一页。可用 `IsEnd` 判断是否还有下一页，并用返回的 `Cookie` 继续查询。

```go
package guide

import (
	channelAPI "github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func FirstReactionUsers(ctx *types.Context, channelID, messageID, emojiID string) (*interaction.GetReactionUsersResponse, error) {
	return channelAPI.GetReactionUsers(ctx, channelID, messageID, interaction.EmojiTypeSystem, emojiID, "", 20)
}
```

## 注意事项

- 表态不会发送表情图片。图片 URL 无法充当表情 ID。
- 平台决定表情是否可用、ID 的范围及权限。SDK 没有统一的表情 ID 到图片 URL 的转换。
- 查询下一页时，原样传入上一页返回的 `Cookie`。SDK 负责查询参数编码。官方只在第一页读取 `Limit`，默认 20，最多 50。特殊接口可通过[底层请求](../../../advanced/raw-api.md)调用。

## 另请参阅

[频道消息](messages.md) · [读取表情与图片](../rich-content.md) · [官方表情表态接口](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/emoji.html)
