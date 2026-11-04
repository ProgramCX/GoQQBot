# 调用 OpenAPI

事件处理函数可直接用 `ctx.Reply`、`ctx.UploadFile` 等方法。定时任务、后台服务和已有业务系统可持有 `*api.Client`。这些服务无需启动 Bot 或构造接收事件。

## 使用明确的回复目标

下面的函数可放入业务包。应用注入客户端，并从[单聊事件](../server-api/messages/c2c/receive.md)获取 `userOpenID` 和 `sourceMessageID`。

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

`ReplyTarget.MessageID` 与 `EventID` 二选一。前者用消息的 `d.id`。后者用可回复事件的外层 `id`。[引用与撤回](../server-api/messages/reference-and-recall.md)说明了这些标识的区别。

主动消息（proactive message）使用 `Client.SendMessage(ctx, message.Target{...}, outgoing)`。单聊唤醒使用 `SendWakeup`。

## 注入 HTTP 客户端

通过 `api.WithHTTPClient` 注入应用自己的客户端。

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

SDK 默认将客户端超时设为 30 秒。注入的客户端用于获取令牌和调用 API。客户端及其传输资源由调用方管理。

## 调用尚无专用方法的端点

`Request` 是通用 JSON 请求入口。实际路径、请求体、权限和返回模型以对应官方接口为准。

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

- 用户和消息标识取自真实事件。普通 QQ 号码和引用索引无法替代这些标识。
- `Client.SendReply` 不为多次调用自动编号。多次回复由调用方设置 `.Sequence(n)`。事件上下文的新回复方法则共用自动计数器。
- 平台检查主动发送资格、额度和接收方状态。省略回复来源无法绕过检查。
- 并发修改已构造客户端的配置会带来竞态。
- 通用 `Request` 入口不等于领域类型化封装。传入完整 URL 会偏离相对路径约定。预先编码的 JSON 字符串会再次被编码。
- SDK 不自动重试消息发送和上传。请求超时时，平台可能已经执行成功。返回了结果但 err 也不是 nil 时，先检查结果的 `Raw`。直接重发可能造成重复。

## 另请参阅

[获取访问凭证](credentials.md) · [错误处理与 TraceID](../logging/errors.md) · [官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html) · [调用未封装的官方接口](../advanced/raw-api.md)
