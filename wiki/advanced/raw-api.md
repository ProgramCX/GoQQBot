# 调用未封装的官方接口

`api.Client` 可用于独立服务，不要求启动 WebSocket 或 Webhook。`Get`、`Post`、`Put`、`Delete` 和 `Request` 共用访问令牌和 API 错误处理。

## 明确路径、方法和响应模型

以下函数查询机器人详情，返回应用持有的原始 JSON。业务可以进一步解码自己需要的字段。

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
| `ctx` | 必须是有效上下文，建议带操作截止时间 |
| `method` | 使用 `http.MethodGet` 等标准方法常量；PATCH 使用 `Request` |
| `path` | 官方 API 相对路径，如 `/users/@me`；不要传完整 URL |
| `body` | 非 nil 时按 JSON 序列化；GET 通常传 nil |
| 返回正文 | 应用持有；敏感事件及签名 URL 不应整份写入日志 |

路径参数使用 `url.PathEscape`，查询参数使用 `url.Values.Encode`；不要拼接不可信游标。`Client.Request` 当前会完整读取响应正文，应用应只调用预期大小的官方 JSON 端点，不把它当成任意文件下载器。

原始接口不会自动管理业务分页、补字段或重试。遇到 POST 已成功而解码失败等不确定状态，应先确认服务端结果，再由业务决定恢复方式。

## 扩展消息字段

消息新增字段时，可通过 `message.New` 和 `WithField` 显式构造。下例仍使用已支持的图文卡片类型，只演示原始字段入口。

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

`WithField` 立即序列化快照，构造错误由 `Build` 或发送方法返回。已知消息类型仍执行已知字段及场景校验；不能用此入口绕过权限检查。路由字段 `msg_id`、`event_id`、`msg_seq`、`is_wakeup` 由专用入口管理，不能通过它覆盖。

## 另请参阅

- [官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
- [官方机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html)
- [HTTP 客户端配置](../getting-started/api-calls.md)
- [错误处理](../logging/errors.md)
