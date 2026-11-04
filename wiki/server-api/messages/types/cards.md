# 发送图文卡片

用 `card.type=tuwen` 在 QQ 单聊或群聊发送链接预览。

## 从业务参数构造卡片

在消息处理函数中调用下面的函数。地址由业务服务提供。构造时不拉取网页或图片，调用 `ctx.Reply` 才发送。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyArticle(ctx *types.Context, articleURL, pictureURL string) error {
	return ctx.Reply(message.Card(message.ImageTextCard{
		Title:       "本周更新",
		Description: "查看新功能和使用说明。",
		URL:         articleURL,
		PictureURL:  pictureURL,
	}))
}
```

## 参数限制

构造器按下表在本地检查字段。

| 字段 | 本地校验 |
| --- | --- |
| `Title` | 1..50 个字符 |
| `Description` | 1..200 个字符 |
| `URL` | HTTPS，含主机名，最多 1024 个字符 |
| `PictureURL` | HTTPS，含主机名，最多 1024 个字符 |

## 注意事项

- 图文卡片、[Markdown 图片](markdown.md)、接收的 ARKData 和[旧 ARK 模板](ark.md)各用不同消息结构。
- `Card` 仅支持 QQ 单聊和群聊。向频道发送会返回本地错误。
- 平台仍会审核内容、域名和终端展示。示例需传入真实可访问的业务地址。
- 用 `message.New(message.CARD).WithField("card", ...)` 可扩展其他卡片字段。这个入口不表示 SDK 已实现或实测所有卡片类型。见[扩展消息字段](../../../advanced/raw-api.md)。

## 另请参阅

[Markdown](markdown.md) · [ARK 模板](ark.md) · [官方单聊 card 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) · [官方群聊 card 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
