# 获取访问凭证

在 QQ 开放平台创建机器人后取得 `AppID` 和 `AppSecret`。SDK 使用它们获取 Access Token，并在请求时添加认证头。旧 `bot.token` 已弃用并被忽略，不能代替 AppSecret。

## 使用独立客户端验证认证

下面的完整程序从环境变量读取凭证，测试能否获取令牌，不打印令牌：

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

`api.NewClient` 本身不联网。令牌在需要调用 API 时按需获取；同一客户端的并发请求共享刷新，不会为每次发送重复获取令牌。SDK 根据服务端有效期提前使缓存过期，不额外启动定时刷新协程。

业务无需手动拼接 `Authorization: QQBot ...`。复用客户端即可，不要为每条消息创建新客户端；不同实例拥有各自缓存。

`AccessToken(ctx)` 支持取消。旧 `EnsureToken()` 使用后台上下文，需要操作截止时间时优先使用前者。

## 常见认证失败

获取访问凭证端点可能以 HTTP 200 返回业务错误，SDK 仍会返回 `*api.APIError`。该端点的 `100007` 表示 AppID 无效或机器人状态异常，`100016` 表示 AppID 或 Secret 不正确，`10004` 表示机器人不存在。错误码解释应结合具体端点，不与其他接口的同数值错误混用。

WebSocket 的 IP 白名单、订阅权限和机器人状态也可能导致登录失败；取令牌成功不等于网关登录或所有 API 都有权限。

## 另请参阅

- [配置文件](configuration.md)
- [API 调用指南](api-calls.md)
- [错误处理](../logging/errors.md)
- [官方获取访问凭证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/access-token.html)
