# 发送群聊消息

单聊和群聊可以共用消息构造器和处理函数。SDK 按事件选择群会话，并填写被动回复来源。

## 回复一个菜单

把以下函数注册到群聊命令。单聊命令也可复用该函数。

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

使用 `AutoSend()` 表达自动发送偏好。按钮的详细用法见[发送按钮](../interactions/buttons.md)。

## 后台业务发到明确的群

定时任务或业务服务可调用以下函数。

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

要回复已有事件，使用 `Client.SendReply` 和 `message.ReplyTarget`。调用方保存回复来源并管理序号。

## 群聊不能自动适配的能力

跨场景复用时，可用 `.Strict()` 在本地发现不支持的交互提示。

## 注意事项

- 单聊保留 `AutoSend()` 的自动发送提示。群聊退回输入框预填，不保证客户端直接发送。
- 使用 `.Strict()` 时，不支持的提示会返回本地错误。
- `groupOpenID` 取自当前机器人的有效群标识。
- 主动消息（proactive message）受平台权限和配额约束。缺少有效被动来源时，伪造的 `msg_id` 无法替代真实来源。
- 单聊输入状态和单聊流式消息不适用于群聊。
- SDK 只调整预填或自动发送等提示。频道身份组按钮的权限不会改为“所有人可用”。

## 另请参阅

[Markdown](../types/markdown.md) · [引用与撤回](../reference-and-recall.md) · [官方群聊发送字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
