# 发送单聊消息

回复当前用户时，从 `*types.Context` 发送即可；没有当前事件的后台业务，则使用 `*api.Client` 和明确目标。

## 回复与输入状态

下面是一个可注册到单聊命令的处理函数。输入状态是一次平台请求，不会在后台持续刷新。

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

`Typing(seconds)` 的范围为 1..60，只能在单聊使用。普通文字直接 `ctx.ReplyText("查询完成")`；消息构造本身不会联网，`Reply` 才发送。

## 后台业务主动发送

以下函数接收业务系统提供的客户端和已知 `user_openid`，使用调用方传入的可取消上下文。

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

主动发送能否成功由平台权限及配额决定。需要明确发送单聊唤醒消息时可用 `client.SendWakeup(ctx, userOpenID, outgoing)`；这不是绕开主动消息限制的方法。

## 序号与失败处理

同一 Context 上的单聊回复自动预留递增 `msg_seq`，并发调用也不会重复编号；失败同样消耗序号。显式使用 `.Sequence(n)` 可管理受控重试，但 SDK 不自动重试 POST。不要与自行固定序号的旧发送函数混用。

`ctx.SendReply(ctx.RequestContext(), outgoing)` 返回发送结果；其 `ID` 用于撤回，`ExtInfo.RefIndex` 用于展示引用，两者不可互换。

## 另请参阅

[文本消息](../types/text.md) · [Markdown](../types/markdown.md) · [引用与撤回](../reference-and-recall.md) · [官方发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
