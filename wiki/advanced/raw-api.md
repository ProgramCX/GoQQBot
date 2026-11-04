# 调用未封装的官方接口

独立服务可以直接用 `api.Client` 调用官方接口。应用无需启动 WebSocket 或 Webhook。`Get`、`Post`、`Put`、`Delete` 和 `Request` 共用访问令牌和 API 错误处理。

## 明确路径、方法和响应模型

下面的函数查询[机器人详情](../server-api/robot/README.md)。应用持有返回的原始 JSON，可按需解码字段。

```go
package guide

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func ReadBot(ctx context.Context, client *api.Client) (json.RawMessage, error) {
	data, err := client.Request(ctx, http.MethodGet, "/users/@me", nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
```

| 参数 | 约定 |
| --- | --- |
| `ctx` | 传入有效上下文，建议设置截止时间 |
| `method` | 使用 `http.MethodGet` 等标准方法常量。PATCH 通过 `Request` 调用。 |
| `path` | 官方 API 相对路径，如 `/users/@me` |
| `body` | SDK 将非 nil 的值编码为 JSON。GET 通常传 nil。 |
| 返回正文 | 由应用持有 |

路径参数用 `url.PathEscape` 编码。查询参数用 `url.Values.Encode` 编码。



## 扩展消息字段

使用 `message.New` 和 `WithField` 构造新增字段。下例借用已支持的[图文卡片](../server-api/messages/types/cards.md)，演示原始字段入口。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyRawCard(ctx *types.Context) error {
	return ctx.Reply(message.New(message.CARD).WithField("card", map[string]any{
		"type": "tuwen",
		"content": map[string]any{
			"title":       "公告",
			"description": "打开页面查看详情",
			"url":         "https://example.com/news",
			"pic_url":     "https://example.com/image.png",
		},
	}))
}
```

`WithField` 会立即序列化字段并保存副本。`Build` 或发送方法会返回构造错误。

## 注意事项

- `path` 只接收相对路径。直接拼接不可信游标可能改变查询参数。
- `Client.Request` 会读完整个响应正文。调用时选择响应大小可预期的官方 JSON 接口。下载任意文件可能占用大量内存。
- 完整记录敏感事件或签名 URL 会泄露敏感数据。
- 原始接口不管理分页，也不补字段或自动重试。POST 可能已成功，即使响应解码失败。先确认服务端结果，再决定如何恢复。
- 已知消息类型仍校验字段和场景。权限检查同样适用。
- `msg_id`、`event_id`、`msg_seq`、`is_wakeup` 由专用入口管理。`WithField` 无法覆盖这些字段。

## 另请参阅

[官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html) · [官方机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) · [HTTP 客户端配置](../getting-started/api-calls.md) · [错误处理](../logging/errors.md)
