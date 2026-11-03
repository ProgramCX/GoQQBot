# 机器人

本模块对应官方“机器人”分类。当前 SDK 已实现访问凭证和网关发现，但尚未提供机器人资料、频道列表、分享链接和全局菜单管理的专用方法。

| 官方能力 | HTTP 接口 | 使用方式 |
| --- | --- | --- |
| [获取机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) | `GET /users/@me` | 用 `Client.Get`，自行声明所需响应字段 |
| [获取机器人频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me_guilds.get.html) | `GET /users/@me/guilds` | 用原始 API，按官方规定管理分页 |
| [生成分享链接](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) | `POST /v2/generate_url_link` | 用原始 API，显式传入业务参数 |
| 自定义菜单和指令面板 | `/v2/menu`、`/v2/panels` | [菜单和面板说明](../messages/interactions/menus-and-panels.md) |

## 查询机器人资料

下面的 `Profile` 是应用自定义响应模型，不是 SDK 新增的公开类型。函数只读取业务需要的字段，新增平台字段不会妨碍当前解码。

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

`api.Client` 负责认证和平台错误解析；应用负责解释资料字段、缓存及刷新频率。不要把资料查询获得的机器人 ID 与群成员 OpenID、单聊用户 OpenID 混用。

## 另请参阅

- [业务分享参数和事件](sharing-and-events.md)
- [访问凭证](../../getting-started/credentials.md)
- [原始 API 扩展](../../advanced/raw-api.md)
