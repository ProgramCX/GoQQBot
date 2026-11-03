# 频道资料与子频道管理

SDK 尚未提供频道、子频道 CRUD 的类型化方法。下面演示如何在应用服务中组合原始 API，按业务所需字段解码频道资料。

## 查询频道资料

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

示例保留字符串时间戳。需要时间运算时根据平台实际格式解析，不把 RFC3339/ISO8601 字符串直接解码为整数。

## 子频道操作

| 操作 | 官方文档 | HTTP |
| --- | --- | --- |
| 列表 | [获取子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) | `GET /guilds/{guild_id}/channels` |
| 创建 | [创建子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.post.html) | `POST /guilds/{guild_id}/channels` |
| 详情 | [获取子频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.get.html) | `GET /channels/{channel_id}` |
| 修改 | [修改子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.patch.html) | `PATCH /channels/{channel_id}` |
| 删除 | [删除子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.delete.html) | `DELETE /channels/{channel_id}` |

`Client` 没有 `Patch` 便捷方法，但可以使用 `client.Request(ctx, http.MethodPatch, path, body)`。请求模型、权限校验、删除确认和操作审计由应用负责；不要向公共可修改的端点变量写入自定义服务器地址。

## 另请参阅

- [官方获取频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html)
- [自定义 HTTP 客户端](../../getting-started/api-calls.md)
- [成员和权限](members-and-permissions.md)
