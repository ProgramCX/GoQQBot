# 回复纯文本

直接回复文字时使用 `ctx.ReplyText`。先构造模板、设置引用或显式序号时，使用 `message.Text` 与 `ctx.Reply`。

## 共用纯文本处理函数

以下函数可注册到单聊、群聊或频道消息事件。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func ReplyText(ctx *types.Context) error {
	name := "访客"
	if len(ctx.Args) > 0 {
		name = ctx.Args[0]
	}
	return ctx.ReplyText("你好，" + name)
}
```

路由负责解析命令参数。`ctx.Text()` 返回规范化的输入文字。需要保留空白时，读取 `ctx.Message()` 的 `Content`。

## 复用消息构造结果

以下函数只构造消息，不发送。返回值可交给 `ctx.Reply`，也可在不同 goroutine 中安全复用。

```go
package guide

import "github.com/ProgramCX/GoQQBot/pkg/message"

func TextResult() message.Outgoing {
	return message.Text("查询已完成，请稍后查看结果。")
}
```

链式方法返回独立值，不修改共享模板。

## 注意事项

- 读取命令参数前先检查长度。空参数列表无法读取首项。
- 空正文或全为空白的正文不符合纯文本校验。纯文本与 Markdown、卡片等正文对象二选一。
- 艾特和可点击命令受场景及格式限制。客户端不一定会解释普通文本里的任意标签。见[文本交互](../interactions/text.md)。

## 另请参阅

[Markdown](markdown.md) · [群聊发送](../group/send.md) · [官方单聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) · [官方群聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
