# 发送单聊消息

用 `*types.Context` 回复当前用户。后台业务没有当前事件时，用 `*api.Client` 指定目标。

## 回复与输入状态

将下面的处理函数注册到单聊命令。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyPrivate(ctx *types.Context) error {
	if err := ctx.Reply(message.Typing(5)); err != nil {
		return err
	}
	return ctx.Reply(message.Markdown("## 查询完成\n这里放业务结果。"))
}
```

用 `ctx.ReplyText("查询完成")` 回复普通文字。构造消息不会联网，调用 `Reply` 才会发送。

## 后台业务主动发送

下面的函数接收业务系统提供的客户端和已知的 `user_openid`。调用方传入可取消的上下文。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func NotifyPrivate(ctx context.Context, client *api.Client, userOpenID string) (*message.ReplyResult, error) {
	return client.SendMessage(ctx, message.Target{
		Scene: message.SceneC2C,
		ID:    userOpenID,
	}, message.Text("你的任务已完成。"))
}
```

使用 `client.SendWakeup(ctx, userOpenID, outgoing)` 显式发送单聊唤醒消息。

## 序号与失败处理

用同一个 `Context` 回复多次时，SDK 会分配递增的 `msg_seq`。并发回复也会使用不同序号。需要自行控制重试时，可以通过 `.Sequence(n)` 指定序号。

`ctx.SendReply(ctx.RequestContext(), outgoing)` 返回发送结果。撤回使用 `ID`，展示引用使用 `ExtInfo.RefIndex`。详见[引用与撤回](../reference-and-recall.md)。

## 注意事项

- `Typing(seconds)` 仅用于单聊。取值为 1 至 60 秒，依据 SDK 的参数校验。每次调用只发一次请求，不会在后台刷新。
- QQ 检查主动消息（proactive message）的权限和配额。唤醒消息也无法绕过主动发送限制。
- 请求失败仍会消耗序号。SDK 不自动重试 POST。
- 混用自行固定序号的旧发送函数会破坏序号管理。
- 把展示引用索引当作撤回 ID，无法定位要撤回的消息。

## 另请参阅

[文本消息](../types/text.md) · [Markdown](../types/markdown.md) · [引用与撤回](../reference-and-recall.md) · [官方发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
