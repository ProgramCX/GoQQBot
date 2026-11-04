# 调用未封装的官方接口

SDK 还没有专用方法时，后台服务可以用 `api.Client` 直接请求 QQ。这个客户端可以单独使用，不用先连 WebSocket 或启动 Webhook。`Get`、`Post`、`Put`、`Delete` 和 `Request` 用的是同一套令牌和错误处理逻辑。

## 明确路径、方法和响应模型

下面的函数查询[机器人详情](../server-api/robot/README.md)，把 QQ 返回的 JSON 交给你。需要哪些字段，就解析哪些字段。

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
| 返回正文 | 可在自己的程序里解析或保存 |

把 ID 拼进路径之前，用 `url.PathEscape` 转义。游标等查询参数用 `url.Values.Encode` 编码。



## 扩展消息字段

消息里要带上尚未封装的字段时，可以用 `message.New` 和 `WithField` 自己填。下面以已支持的[图文卡片](../server-api/messages/types/cards.md)为例。

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

调用 `WithField` 时，SDK 会将字段值编码为 JSON 并保存。后续修改原值不会影响消息。编码失败时，`Build` 或发送方法会返回错误。

## 注意事项

- `path` 只接收相对路径。直接拼接不可信游标可能改变查询参数。
- `Client.Request` 会读完整个响应正文。调用时选择响应大小可预期的官方 JSON 接口。下载任意文件可能占用大量内存。
- 完整记录敏感事件或签名 URL 会泄露敏感数据。
- 翻页、补字段和重试都由你的程序决定。如果只是在解析响应时出错，QQ 可能已经处理了 POST。重发前先查这次操作是否成功。
- 已知消息类型仍校验字段和场景。权限检查同样适用。
- `msg_id`、`event_id`、`msg_seq`、`is_wakeup` 由对应的发送方法填写，通过 `WithField` 改不了。

## 另请参阅

[官方 API 调用指南](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html) · [官方机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) · [HTTP 客户端配置](../getting-started/api-calls.md) · [错误处理](../logging/errors.md)
