# 查询和配置机器人

使用 `robot.NewClient(client)` 查询机器人资料、生成分享链接或管理菜单。频道列表通过 `channel.NewClient(client)` 查询。两个客户端都复用现有的 `*api.Client`。

| 官方能力 | HTTP 接口 | 使用方式 |
| --- | --- | --- |
| [获取机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) | `GET /users/@me` | `robot.Client.GetProfile` |
| [获取机器人频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me_guilds.get.html) | `GET /users/@me/guilds` | `channel.Client.ListGuilds` |
| [生成分享链接](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) | `POST /v2/generate_url_link` | `robot.Client.GenerateShareLink` |
| 自定义菜单和指令面板（command panel） | `/v2/menu`、`/v2/panels` | [菜单和面板说明](../messages/interactions/menus-and-panels.md) |

## 查询机器人资料

```go
package guide

import (
	"context"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
)

func ProfileName(ctx context.Context, client *api.Client) (string, error) {
	profile, err := robot.NewClient(client).GetProfile(ctx)
	if err != nil {
		return "", fmt.Errorf("获取机器人资料: %w", err)
	}
	return profile.Username, nil
}
```

SDK 会为请求添加访问凭证。机器人资料的缓存与刷新由业务决定。

## 注意事项

- 资料中的机器人 ID、群成员 OpenID 和单聊用户 OpenID 用于不同对象。按接口所需的标识调用。

## 另请参阅

[业务分享参数和事件](sharing-and-events.md) · [访问凭证](../../getting-started/credentials.md) · [原始 API 扩展](../../advanced/raw-api.md)
