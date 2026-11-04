# 选择消息类型

用 `pkg/message` 构造不可变的 `Outgoing`。构造过程不联网。用 `Build(scene)` 在本地校验，再用 `Context.Reply` 或 `Client.SendReply` 发送。

## 按内容选择构造器

| 内容 | 构造入口 | 场景 |
| --- | --- | --- |
| [纯文本](text.md) | `message.Text` | 群聊、单聊、频道及频道私信 |
| [Markdown](markdown.md) | `message.Markdown`、`MarkdownTemplate` | 见注意事项 |
| [图文卡片](cards.md) | `message.Card` | QQ 单聊和群聊 |
| [ARK 模板](ark.md) | `message.Ark` | 旧协议 |
| [普通富媒体](../rich-media/README.md) | `message.Media(fileInfo)` | QQ 单聊和群聊 |
| 输入状态 | `message.Typing(seconds)` | 仅单聊 |

## 先在本地校验

下面的函数在本地校验业务构造的消息。函数不创建客户端，也不发送消息。

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

## 扩展字段

用 `New(kind).WithField(name, value)` 添加新消息类型和字段。设置字段时会立即序列化为快照。

用 `errors.Is(err, message.ErrInvalidMessage)` 判断本地消息错误。

## 注意事项

- 频道 Markdown 需要相应权限。旧 ARK 协议受模板和场景权限限制。
- 普通富媒体先在目标会话上传，再发送。输入状态的范围是 1 至 60 秒，由 SDK 校验。
- 本地校验通过后，QQ 仍会检查权限和配额。审核、客户端展示及文件有效期也由 QQ 决定。
- 已知非法字段组合会返回错误。`WithField` 无法覆盖 `msg_type`、`msg_id`、`event_id`、`msg_seq` 和 `is_wakeup`。
- 零值 `Outgoing` 无效。构造器会生成有效的消息结构。

## 另请参阅

[消息收发](../README.md) · [Markdown](markdown.md) · [官方消息类型](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/type/overview.html)
