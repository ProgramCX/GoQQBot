# 消息类型

`pkg/message` 的构造器返回不可变的 `Outgoing`。构造时不联网；`Build(scene)` 在本地校验，`Context.Reply` 或 `Client.SendReply` 才实际发送。

## 按内容选择构造器

| 内容 | 构造入口 | 场景或限制 |
| --- | --- | --- |
| [纯文本](text.md) | `message.Text` | 群聊、单聊、频道及频道私信 |
| [Markdown](markdown.md) | `message.Markdown`、`MarkdownTemplate` | 频道需相应权限 |
| [图文卡片](cards.md) | `message.Card` | QQ 单聊和群聊 |
| [ARK 模板](ark.md) | `message.Ark` | 旧协议，受模板和场景权限限制 |
| 普通富媒体 | `message.Media(fileInfo)` | 单聊和群聊，需先完成对应会话上传 |
| 输入状态 | `message.Typing(seconds)` | 仅单聊，秒数为 1..60 |

## 先在本地校验

以下函数不会创建客户端或发送消息，可用于确认业务构造是否满足 SDK 本地限制。

```go
package guide

import (
	"encoding/json"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func BuildMessagePreview() (map[string]json.RawMessage, error) {
	return message.Markdown("## 订单完成\n请查看详情。").Build(message.SceneGroup)
}
```

校验通过不代表平台一定接受；权限、审核、客户端展示、文件有效期和配额仍由 QQ 决定。

## 扩展字段

`New(kind).WithField(name, value)` 可承载新增消息类型和字段；字段在设置时立即序列化为快照。已知非法组合仍会返回错误，`msg_type`、`msg_id`、`event_id`、`msg_seq` 和 `is_wakeup` 不能通过 WithField 覆盖。

本地消息错误可用 `errors.Is(err, message.ErrInvalidMessage)` 判断。零值 Outgoing 无效；请使用构造器。

## 另请参阅

[消息收发](../README.md) · [Markdown](markdown.md) · [官方消息类型](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/type/overview.html)
