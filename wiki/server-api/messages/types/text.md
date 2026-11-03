# 回复纯文本

简单文字使用 `ctx.ReplyText`。需要先构造模板、设置引用或显式序号时，使用 `message.Text` 与 `ctx.Reply`。

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

命令参数由路由解析，访问前检查长度。`ctx.Text()` 返回规范化的输入文字；需要保留空白时读取 `ctx.Message()` 的 `Content`。

## 复用消息构造结果

以下函数只是构造消息，不发送。返回值可交给 `ctx.Reply`，也可以在不同 goroutine 中安全复用。

```go
package guide

import "github.com/ProgramCX/GoQQBot/pkg/message"

func TextResult() message.Outgoing {
	return message.Text("查询已完成，请稍后查看结果。")
}
```

链式方法返回独立值，不会修改共享模板。纯文本不能为空或全部为空白；不要同时添加 Markdown、卡片或其他互斥的正文对象。协议中的艾特和可点击命令还受场景及格式限制，不能假定普通文本中的任意标签都会被客户端解释。

## 另请参阅

[Markdown](markdown.md) · [群聊发送](../group/send.md) · [官方单聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) · [官方群聊发送](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
