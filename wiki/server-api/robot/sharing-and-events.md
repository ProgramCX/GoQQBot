# 分享链接与机器人事件

需要区分邀请来源时，可以通过官方分享接口携带 `callback_data`。SDK 尚未封装分享接口；应用通过 `Client.Post` 调用，并自行保存来源映射。

## 生成分享链接

```go
package guide

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func GenerateInvite(ctx context.Context, client *api.Client, source string) (string, error) {
	if utf8.RuneCountInString(source) > 32 {
		return "", fmt.Errorf("callback_data exceeds 32 characters")
	}
	raw, err := client.Post(ctx, "/v2/generate_url_link", struct {
		CallbackData string `json:"callback_data,omitempty"`
	}{CallbackData: source})
	if err != nil {
		return "", err
	}
	var result struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		// POST 可能已经成功，不要因为解码失败自动重发。
		return "", fmt.Errorf("decode invite response: %w", err)
	}
	if result.Data.URL == "" {
		return "", fmt.Errorf("invite response has no URL")
	}
	return result.Data.URL, nil
}
```

本函数是业务层示例，不是 SDK 的专用分享封装。分享参数不要携带密钥或个人敏感信息；平台回传数据也需要按业务边界校验。

## 接收生命周期事件

通过 `OnEvent` 注册 `FRIEND_ADD`、`FRIEND_DEL`、`GROUP_ADD_ROBOT`、`GROUP_DEL_ROBOT` 等事件。当前只有事件名和原始信封，不具备覆盖所有事件字段的统一类型化视图；请按对应官方事件文档解码 `ctx.Payload.D`。

```go
package guide

import (
	"context"
	"encoding/json"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterFriendEvents(bot *goqqrobot.Bot, consume func(context.Context, json.RawMessage) error) {
	bot.OnEvent(message.FRIEND_ADD, func(ctx *types.Context) error {
		// 复制后交给业务；不要记录整份授权载荷。
		data := append(json.RawMessage(nil), ctx.Payload.D...)
		return consume(ctx.RequestContext(), data)
	})
}
```

`OnEvent` 注册不代表平台已经订阅此事件。WebSocket 需要对应 intent，Webhook 需要在平台配置订阅。

当前 `Context.Reply` 的自动路由覆盖消息事件和场景明确的交互事件，不能直接套用到所有生命周期事件。若某个官方事件明确允许通过外层事件 ID 回复，应使用 `Client.SendReply` 显式填写场景、接收者和 `EventID`；不要填写事件类型名，也不要从不相关字段猜测接收者。

## 另请参阅

- [官方分享接口](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html)
- [事件订阅与路由](../../getting-started/events/routing.md)
- [显式 API 调用](../../getting-started/api-calls.md)
