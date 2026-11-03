# 调用 OpenAPI

在事件处理函数中，优先使用 `ctx.Reply`、`ctx.UploadFile` 等方法；定时任务、后台服务和已有业务系统可直接持有 `*api.Client`，不必启动 Bot 或构造假的接收事件。

## 使用明确的回复目标

下面是可放入业务包的完整函数。客户端由应用注入，`userOpenID` 和 `sourceMessageID` 从真实的单聊事件获取，不能使用普通 QQ 号码或引用索引：

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

`ReplyTarget.MessageID` 与 `EventID` 必须二选一，分别表示消息 `d.id` 和可回复事件的外层 `id`。`Client.SendReply` 不负责跨调用自动分配序号，多次回复由调用方设置 `.Sequence(n)`；事件上下文的新回复方法则共用自动计数器。

主动发送使用 `Client.SendMessage(ctx, message.Target{...}, outgoing)`；单聊唤醒使用 `SendWakeup`。主动发送能力、额度、接收方状态由平台检查，不能通过省略回复来源绕过限制。

## 注入 HTTP 客户端

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

默认客户端超时为 30 秒。注入的客户端同时用于令牌获取和 API 请求；客户端及其传输资源仍由调用方管理。构造后不要并发修改其配置。

## 调用尚无专用方法的端点

`Request` 是通用 JSON 请求入口，**不代表 SDK 已完成该领域的类型化封装**。下面展示调用方式，实际路径、请求体、权限及返回模型需对照对应官方接口：

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

`path` 传相对 OpenAPI 路径，例如 `/users/@me`，不是完整 URL。查询参数使用 `url.Values.Encode()` 编码；JSON body 不传预先编码的字符串。`Get`、`Post`、`Put`、`Delete` 是同一请求实现的便利方法。

SDK 不自动重试消息发送和上传操作。超时可能意味着平台已经执行成功；同时返回非 nil 结果和错误时先检查对应结果的 `Raw`，不要盲目重发。

## 另请参阅

- [获取访问凭证](credentials.md)
- [错误处理与 TraceID](../logging/errors.md)
- [官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
