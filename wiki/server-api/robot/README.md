# 查询和配置机器人

查询机器人资料和管理平台资源可用官方原始 API。SDK 已提供[访问凭证](../../getting-started/credentials.md)和网关发现。

| 官方能力 | HTTP 接口 | 使用方式 |
| --- | --- | --- |
| [获取机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) | `GET /users/@me` | 用 `Client.Get`，自行声明所需响应字段 |
| [获取机器人频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me_guilds.get.html) | `GET /users/@me/guilds` | 用原始 API，按官方规定管理分页 |
| [生成分享链接](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) | `POST /v2/generate_url_link` | 用原始 API，显式传入业务参数 |
| 自定义菜单和指令面板（command panel） | `/v2/menu`、`/v2/panels` | [菜单和面板说明](../messages/interactions/menus-and-panels.md) |

## 查询机器人资料

应用可按需定义响应结构。`Profile` 只读取业务需要的字段，平台新增字段不影响现有解码。

```go
package guide

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

type Profile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Bot      bool   `json:"bot"`
}

func LoadProfile(ctx context.Context, client *api.Client) (*Profile, error) {
	raw, err := client.Get(ctx, "/users/@me")
	if err != nil {
		return nil, err
	}
	var profile Profile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return nil, fmt.Errorf("decode bot profile: %w", err)
	}
	return &profile, nil
}
```

`api.Client` 负责认证和解析平台错误。应用解释资料字段，并管理缓存和刷新频率。

## 注意事项

- SDK 尚无机器人资料和频道列表的专用方法。分享链接和全局菜单管理同样缺少专用方法。`Profile` 是应用自定义类型。
- 资料中的机器人 ID、群成员 OpenID 和单聊用户 OpenID 用于不同对象。按接口所需的标识调用。

## 另请参阅

[业务分享参数和事件](sharing-and-events.md) · [访问凭证](../../getting-started/credentials.md) · [原始 API 扩展](../../advanced/raw-api.md)
