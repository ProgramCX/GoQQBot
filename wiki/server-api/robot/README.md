# 查询和配置机器人

查询机器人资料或管理 QQ 平台资源，可以直接调用官方接口。SDK 已经提供[访问凭证](../../getting-started/credentials.md)和查询网关地址的功能。

| 官方能力 | HTTP 接口 | 使用方式 |
| --- | --- | --- |
| [获取机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) | `GET /users/@me` | 用 `Client.Get`，自行声明所需响应字段 |
| [获取机器人频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me_guilds.get.html) | `GET /users/@me/guilds` | 用原始 API，按官方规定管理分页 |
| [生成分享链接](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) | `POST /v2/generate_url_link` | 用原始 API，显式传入业务参数 |
| 自定义菜单和指令面板（command panel） | `/v2/menu`、`/v2/panels` | [菜单和面板说明](../messages/interactions/menus-and-panels.md) |

## 查询机器人资料

可以自己定义接收结果的结构体。下面的 `Profile` 只接收业务用到的字段。QQ 后续加字段时，已有代码仍能读取原来的字段。

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

`api.Client` 会处理认证，并把 QQ 返回的错误解析出来。资料字段怎么用、要不要缓存、多久刷新一次，由应用决定。

## 注意事项

- SDK 尚无机器人资料和频道列表的专用方法。分享链接和全局菜单管理同样缺少专用方法。`Profile` 是应用自定义类型。
- 资料中的机器人 ID、群成员 OpenID 和单聊用户 OpenID 用于不同对象。按接口所需的标识调用。

## 另请参阅

[业务分享参数和事件](sharing-and-events.md) · [访问凭证](../../getting-started/credentials.md) · [原始 API 扩展](../../advanced/raw-api.md)
