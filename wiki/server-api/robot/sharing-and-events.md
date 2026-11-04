# 分享机器人链接并接收事件

官方分享接口可用 `callback_data` 区分邀请来源。应用通过 `Client.Post` 调用，并保存来源映射。

## 生成分享链接

下面的函数封装业务层的分享请求。

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

## 接收生命周期事件

通过 [`OnEvent` 路由](../../getting-started/events/routing.md)注册所需事件。

- `FRIEND_ADD` 和 `FRIEND_DEL`。
- `GROUP_ADD_ROBOT` 和 `GROUP_DEL_ROBOT`。

按对应官方事件文档解码 `ctx.Payload.D`。

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

WebSocket 使用对应 intent。Webhook 在平台配置订阅。

## 注意事项

- SDK 尚未封装分享接口。`GenerateInvite` 是业务层示例。
- 分享参数中的密钥或个人敏感信息可能随分享传播。平台回传数据也应按业务边界校验。
- SDK 当前仅提供事件名及原始信封。统一的类型化视图尚未覆盖全部字段。`OnEvent` 注册不会替应用完成平台订阅。
- `Context.Reply` 自动路由覆盖消息事件和场景明确的交互事件。生命周期事件不一定支持该入口。
- 官方允许使用外层事件 ID 回复时，可调用 [`Client.SendReply`](../../getting-started/api-calls.md)。显式填写场景、接收者和 `EventID`。事件类型名不能充当 ID。接收者也不能从无关字段推断。

## 另请参阅

[官方分享接口](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) · [事件订阅与路由](../../getting-started/events/routing.md) · [显式 API 调用](../../getting-started/api-calls.md)
