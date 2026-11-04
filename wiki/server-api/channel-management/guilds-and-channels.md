# 频道资料与子频道管理

频道和子频道管理可通过[原始 API](../../advanced/raw-api.md)接入。SDK 尚未提供用于创建、删除、修改和查询的专用方法与数据结构。

## 查询频道资料

下例在应用服务中查询频道，并按业务需要解码字段。

```go
package guide

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
	JoinedAt    string `json:"joined_at"`
}

func ReadGuild(ctx context.Context, client *api.Client, guildID string) (*Guild, error) {
	if guildID == "" {
		return nil, fmt.Errorf("guild ID is required")
	}
	raw, err := client.Get(ctx, "/guilds/"+url.PathEscape(guildID))
	if err != nil {
		return nil, err
	}
	var guild Guild
	if err := json.Unmarshal(raw, &guild); err != nil {
		return nil, fmt.Errorf("decode guild: %w", err)
	}
	return &guild, nil
}
```

示例保留字符串时间戳。

## 子频道操作

按操作选择官方路径和 HTTP 方法。

| 操作 | 官方文档 | HTTP |
| --- | --- | --- |
| 列表 | [获取子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) | `GET /guilds/{guild_id}/channels` |
| 创建 | [创建子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.post.html) | `POST /guilds/{guild_id}/channels` |
| 详情 | [获取子频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.get.html) | `GET /channels/{channel_id}` |
| 修改 | [修改子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.patch.html) | `PATCH /channels/{channel_id}` |
| 删除 | [删除子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.delete.html) | `DELETE /channels/{channel_id}` |

`Client` 没有 `Patch` 便捷方法。用 `client.Request(ctx, http.MethodPatch, path, body)` 发起修改。

## 注意事项

- 做时间运算前，按平台格式解析时间戳。RFC3339 或 ISO8601 字符串无法直接解码成整数。
- 请求模型、权限校验、删除确认和操作审计由应用负责。
- 公开的端点变量会影响共用地址。指定自定义服务器时，应使用独立的客户端配置。

## 另请参阅

[官方获取频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html) · [自定义 HTTP 客户端](../../getting-started/api-calls.md) · [成员和权限](members-and-permissions.md)
