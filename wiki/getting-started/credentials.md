# 获取访问凭证

在 QQ 开放平台创建机器人，即可获得 `AppID` 和 `AppSecret`。SDK 用这两项凭证向 QQ 换取 Access Token。调用 API 时，SDK 会自动把令牌放进请求头。

## 使用独立客户端验证认证

将凭证设为环境变量，再运行下面的程序验证能否获取令牌。程序只报告认证结果，不输出令牌。

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func main() {
	if err := check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check() error {
	appID, secret := os.Getenv("QQ_BOT_APP_ID"), os.Getenv("QQ_BOT_APP_SECRET")
	if appID == "" || secret == "" {
		return fmt.Errorf("请设置 QQ_BOT_APP_ID 和 QQ_BOT_APP_SECRET")
	}
	client := api.NewClient(appID, secret)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := client.AccessToken(ctx); err != nil {
		return err
	}
	fmt.Println("认证成功")
	return nil
}
```

## 刷新与资源边界

`api.NewClient` 创建客户端时不联网。第一次调用需要认证的 API 时，SDK 才向 QQ 获取令牌。

- 同一个客户端同时发出多个请求时，只需刷新一次令牌。发送多条消息也可以复用令牌。
- SDK 会根据 QQ 返回的有效期提前更新缓存。刷新发生在需要令牌时，无需额外的定时协程。
- 一个客户端有一份令牌缓存。后续请求只要用同一个客户端，就能用上这份缓存。
- SDK 会添加 `Authorization: QQBot ...` 认证头。

获取令牌时如果要支持取消或超时，就把上下文传给 `AccessToken(ctx)`。

## 常见认证失败

获取令牌时，QQ 可能返回 HTTP 200，并在响应体里说明认证失败。遇到这种情况，SDK 会返回 `*api.APIError`。

| 官方凭证接口错误码 | 含义 |
| --- | --- |
| `100007` | AppID 无效或机器人状态异常 |
| `100016` | AppID 或 Secret 不正确 |
| `10004` | 机器人不存在 |

## 注意事项

- 旧的 `bot.token` 已弃用，SDK 会忽略该字段。用它替代 AppSecret 会导致认证失败。
- 旧的 `EnsureToken()` 使用后台上下文，无法通过调用方的上下文取消。
- 同一个错误码在不同接口里可能意思不同。排查时，查实际调用的接口。
- 取得令牌后，登录网关仍需通过 IP 白名单、订阅权限和机器人状态检查。调用各个 API 时，QQ 也会检查对应的权限。登录排查见 [WebSocket](events/websocket.md)。

## 另请参阅

[配置文件](configuration.md) · [API 调用指南](api-calls.md) · [错误处理](../logging/errors.md) · [官方获取访问凭证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/access-token.html)
