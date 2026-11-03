# 发送图文卡片

图文卡片使用当前单聊及群聊发送协议的 `card.type=tuwen`，适合链接预览。它与 Markdown 图片、接收到的 ARKData、旧 ARK 模板不是同一种消息。

## 从业务参数构造卡片

以下函数在处理函数中调用，地址由业务服务提供。构造时不拉取网页或图片，`ctx.Reply` 才执行发送。

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

| 字段 | 本地校验 |
| --- | --- |
| `Title` | 1..50 个字符 |
| `Description` | 1..200 个字符 |
| `URL` | HTTPS，含主机名，最多 1024 个字符 |
| `PictureURL` | HTTPS，含主机名，最多 1024 个字符 |

`Card` 仅适用于 QQ 单聊和群聊，频道发送将返回本地错误。构造器校验不代替平台内容审核、域名要求和终端展示检查；示例中的地址必须替换为真实可访问的业务地址。

未来其他卡片类型可以显式使用 `message.New(message.CARD).WithField("card", ...)` 扩展，但这只提供协议字段入口，不意味着 SDK 已实现或实测所有卡片类型。

## 另请参阅

[Markdown](markdown.md) · [ARK 模板](ark.md) · [官方单聊 card 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) · [官方群聊 card 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
