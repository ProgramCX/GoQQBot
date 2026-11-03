# 添加和读取频道表情表态

表态是在已有频道消息上添加一个表情，不是发送表情图片。现有 `api/channel` 提供添加、删除及查询表态用户的方法。

## 添加一个系统表情表态

以下函数接收平台支持的表情 ID；不要把图片 URL 当作表情 ID。

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

`EmojiTypeSystem` 为类型 1，`EmojiTypeCustom` 为类型 2；是否可用、ID 的合法范围及权限由平台决定。SDK 不提供表情 ID 到图片 URL 的统一转换。

## 查询第一页用户

以下代码只查询首页，返回类型中的 `Cookie` 和 `IsEnd` 用于平台分页。

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

**当前限制：**旧方法直接拼接非空 `cookie`，没有进行查询参数转义。包含 `+`、`&`、`=` 等字符的游标可能导致分页请求错误。因此本页仅展示空游标的首页，不能把该方法描述为已覆盖所有分页情况。自主管理分页时可通过底层 `Client.Request` 配合 `url.Values` 构造请求，并按官方接口验证，不要直接拼接游标。

## 另请参阅

[频道消息](messages.md) · [读取表情与图片](../rich-content.md) · [官方表情表态接口](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/emoji.html)
