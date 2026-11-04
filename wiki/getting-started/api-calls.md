# 调用 OpenAPI

在事件处理函数里，可以直接调用 `ctx.Reply` 或 `ctx.UploadFile`。定时任务和后台服务可以用 `*api.Client` 直接调用 OpenAPI，无需启动 Bot 或模拟收到的事件。

## 使用明确的回复目标

把下面的函数放进业务包，调用时传入客户端。其余两个参数从[单聊事件](../server-api/messages/c2c/receive.md)中取出，分别是 `userOpenID` 和 `sourceMessageID`。

```go
package guide

import (
	"context"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func ReplyToUser(parent context.Context, client *api.Client, userOpenID, sourceMessageID string) (*message.ReplyResult, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	return client.SendReply(ctx, message.ReplyTarget{
		Scene:     message.SceneC2C,
		ID:        userOpenID,
		MessageID: sourceMessageID,
	}, message.Text("业务已处理"))
}
```

`ReplyTarget.MessageID` 与 `EventID` 二选一。回复消息时填消息的 `d.id`。回复可回复事件时填事件外层的 `id`。[引用与撤回](../server-api/messages/reference-and-recall.md)说明了这些标识的区别。

主动消息（proactive message）使用 `Client.SendMessage(ctx, message.Target{...}, outgoing)`。在单聊里唤醒用户时，调用 `SendWakeup`。

## 注入 HTTP 客户端

应用已经有 HTTP 客户端时，可以通过 `api.WithHTTPClient` 交给 SDK 使用。

```go
package guide

import (
	"net/http"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func NewClient(appID, appSecret string) *api.Client {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	return api.NewClient(appID, appSecret, api.WithHTTPClient(httpClient))
}
```

SDK 自带的客户端默认 30 秒超时。如果传入自己的客户端，获取令牌和调用 API 都会用它。客户端和底层连接由你管理。

## 调用尚无专用方法的端点

SDK 还没有封装某个接口时，可以用 `Request` 发送 JSON 请求。路径、请求体、权限和返回值都按该接口的官方文档填写。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func RequestJSON(ctx context.Context, client *api.Client, method, path string, body any) ([]byte, error) {
	return client.Request(ctx, method, path, body)
}
```

`path` 使用相对 OpenAPI 路径，例如 `/users/@me`。查询参数用 `url.Values.Encode()` 编码。`Get`、`Post`、`Put`、`Delete` 共用同一请求实现。



## 注意事项

- 从真实事件中取出用户和消息标识。填普通 QQ 号码或引用索引会用错标识。
- 用 `Client.SendReply` 回复同一条消息多次时，用 `.Sequence(n)` 为每次回复编号。通过事件上下文的新版回复方法发送时，SDK 会统一编号。
- 主动发送时，QQ 会检查发送资格、额度和接收方状态。即使省略回复来源，QQ 也会做这些检查。
- 同一个客户端已经在被使用时，并发修改配置可能产生数据竞争。
- 用 `Request` 时，请求字段由你按接口文档填写。路径只填 OpenAPI 相对路径，传完整 URL 会偏离约定。请求体传 Go 值，传入已经编码的 JSON 字符串会再编码一次。
- SDK 不会自动重试发送和上传。即使请求超时，QQ 也可能已经处理成功。如果调用同时返回结果和错误，先检查结果中的 `Raw`。直接重发可能会重复发送或上传。

## 另请参阅

[获取访问凭证](credentials.md) · [错误处理与 TraceID](../logging/errors.md) · [官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html) · [调用未封装的官方接口](../advanced/raw-api.md)
