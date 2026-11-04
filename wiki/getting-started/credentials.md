# 获取访问凭证

在 QQ 开放平台创建机器人，即可获得 `AppID` 和 `AppSecret`。SDK 用这两项凭证获取 Access Token，并为请求添加认证头。

## 使用独立客户端验证认证

完整程序从环境变量读取凭证，并验证能否获取令牌。程序不会打印令牌。

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

`api.NewClient` 创建客户端时不联网。首次需要调用 API 时，SDK 才会获取令牌。

- 同一客户端的并发请求共用刷新流程。发送消息不必逐条获取令牌。
- SDK 根据服务端给出的有效期提前让缓存过期。令牌刷新不需要额外的定时协程。
- 复用客户端可复用令牌缓存。不同客户端实例各自缓存令牌。
- SDK 会添加 `Authorization: QQBot ...` 认证头。

`AccessToken(ctx)` 支持取消。需要截止时间时，用 `AccessToken(ctx)` 替代旧的 `EnsureToken()`。

## 常见认证失败

获取凭证时，即使 HTTP 状态是 200，业务失败也会返回 `*api.APIError`。

| 官方凭证接口错误码 | 含义 |
| --- | --- |
| `100007` | AppID 无效或机器人状态异常 |
| `100016` | AppID 或 Secret 不正确 |
| `10004` | 机器人不存在 |

## 注意事项

- 旧的 `bot.token` 已弃用，SDK 会忽略该字段。用它替代 AppSecret 会导致认证失败。
- 旧的 `EnsureToken()` 使用后台上下文，无法通过调用方的上下文取消。
- 同一错误码在不同接口可能有不同含义。按发生错误的端点查找说明。
- 取令牌成功后，网关仍会检查 IP 白名单、订阅权限和机器人状态。各 API 也会检查权限。登录排查见 [WebSocket](events/websocket.md)。

## 另请参阅

[配置文件](configuration.md) · [API 调用指南](api-calls.md) · [错误处理](../logging/errors.md) · [官方获取访问凭证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/access-token.html)
