# 发送群聊消息

使用统一消息构造器，单聊和群聊可以共用处理函数。SDK 根据当前事件选择群会话，并填写被动回复所需的消息来源。

## 回复一个菜单

把以下函数注册到群聊命令；也可以注册到单聊命令，不必复制发送逻辑。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyGroupMenu(ctx *types.Context) error {
	return ctx.Reply(message.Markdown("## 功能菜单\n请选择操作。").Buttons(
		button.Command("帮助", "/help").AutoSend(),
		button.Command("查询", "/query "),
	))
}
```

`AutoSend()` 表达支持时自动发送的偏好：单聊保留，群聊降级为输入框预填。它不意味着群聊客户端一定支持直接发送。使用 `.Strict()` 时，不支持的提示会返回本地错误。

## 后台业务发到明确的群

以下函数供定时任务或业务服务调用；`groupOpenID` 必须来自当前机器人的有效群标识。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func NotifyGroup(ctx context.Context, client *api.Client, groupOpenID string) (*message.ReplyResult, error) {
	return client.SendMessage(ctx, message.Target{
		Scene: message.SceneGroup,
		ID:    groupOpenID,
	}, message.Text("活动开始了。"))
}
```

主动发送受平台权限和配额约束；当前没有有效被动来源时，不要伪造 `msg_id`。明确回复某个已有事件可使用 `Client.SendReply` 与 `message.ReplyTarget`，调用方需保存来源并管理回复序号。

## 群聊不能自动适配的能力

单聊输入状态、单聊流式消息不能在群聊发送。频道身份组权限按钮也不能改成“所有人可用”以绕过限制。SDK 的降级仅针对预填或自动发送等交互提示，不放宽权限。

## 另请参阅

[Markdown](../types/markdown.md) · [引用与撤回](../reference-and-recall.md) · [官方群聊发送字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
