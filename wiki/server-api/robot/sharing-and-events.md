# 分享机器人链接并接收事件

需要区分邀请来源时，通过 `robot.Client.GenerateShareLink` 附带渠道标记。用户使用链接添加机器人后，QQ 会将该标记传回应用。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`FRIEND_ADD`、`FRIEND_DEL`、`GROUP_ADD_ROBOT`、`GROUP_DEL_ROBOT` 等所需事件。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 生成分享链接

```go
package guide

import (
	"context"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
)

func GenerateInvite(ctx context.Context, client *api.Client, source string) (string, error) {
	link, err := robot.NewClient(client).GenerateShareLink(ctx, source)
	if err != nil {
		return "", fmt.Errorf("生成分享链接: %w", err)
	}
	return link.URL, nil
}
```

`source` 是业务侧的渠道标记，对应协议的 `callback_data`。无需区分邀请来源时可留空，填写时最长 32 个字符。

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

## 注意事项

- 分享参数中的密钥或个人敏感信息可能随分享传播。平台回传数据也应按业务边界校验。
- SDK 提供事件名和收到的原始事件数据，目前还没有为所有字段定义结构体。注册 `OnEvent` 后，还要按接入方式订阅对应的事件。
- `Context.Reply` 自动路由覆盖消息事件和场景明确的交互事件。生命周期事件不一定支持该入口。
- 官方允许使用外层事件 ID 回复时，可调用 [`Client.SendReply`](../../getting-started/api-calls.md)。显式填写场景、接收者和 `EventID`。事件类型名不能充当 ID。接收者也不能从无关字段推断。

## 另请参阅

[官方分享接口](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) · [事件订阅与路由](../../getting-started/events/routing.md) · [显式 API 调用](../../getting-started/api-calls.md)
