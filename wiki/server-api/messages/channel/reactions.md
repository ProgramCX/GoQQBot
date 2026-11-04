# 添加和读取频道表情表态

表情表态（reaction）把表情附在已有频道消息上。使用 `api/channel` 添加、删除或查询表态用户。

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
- 旧方法直接拼接非空 `cookie`，未转义查询参数。`+`、`&`、`=` 等字符可能导致分页请求错误。空游标的首页示例无法覆盖这些分页情况。
- 自行分页时，可用[底层请求](../../../advanced/raw-api.md)中的 `Client.Request`。用 `url.Values` 编码参数，并按官方接口验证。

## 另请参阅

[频道消息](messages.md) · [读取表情与图片](../rich-content.md) · [官方表情表态接口](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/emoji.html)
