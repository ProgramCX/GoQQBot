# 发送 Markdown

Markdown 适合菜单、查询结果和图文说明。QQ 单聊和群聊的自定义 Markdown 无需申请模板；频道场景目前仍需平台开通。客户端支持的排版范围以 QQ 文档为准。

## 回复正文与按钮

下面函数可注册到单聊或群聊命令。构造器默认每行两列，最多五行。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func MarkdownMenu(ctx *types.Context) error {
	return ctx.Reply(message.Markdown("## 服务中心\n请选择操作。").Buttons(
		button.Command("帮助", "/help"),
		button.Link("官方网站", "https://example.com"),
	))
}
```

普通命令按钮默认只预填。`AutoSend()` 在单聊保留自动发送提示，在群聊降级为预填；`.Strict()` 则拒绝不支持的场景能力。授权限制不会被放宽。

## 使用已配置的模板

以下函数从业务配置接收模板 ID，不在代码中绑定真实平台资源。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	markdownTypes "github.com/ProgramCX/GoQQBot/pkg/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyMarkdownTemplate(ctx *types.Context, templateID string) error {
	return ctx.Reply(message.MarkdownTemplate(templateID,
		markdownTypes.MarkdownParams{Key: "title", Values: []string{"业务通知"}},
	))
}
```

正文、模板 ID 等内容来源必须互斥；模板参数不能和自定义正文混用。模板的变量名需与平台配置一致。

## 图片和布局

公网图片由 QQ 下载转存；请使用平台可访问的地址，并在 Markdown 替代文本中提供适当显示尺寸。无公网地址的图片可以先上传为 Markdown 图片素材，再使用素材的 `Markdown` 或 `MarkdownText` 方法，不要将普通富媒体的 `file_info` 当作 Markdown 图片 URL。

`.VerifyImages()` 要求平台在图片转存失败时中止发送。它无法保证所有终端展示一致。两列布局最多十个按钮；十四个按钮应显式设置 `Columns(3)`。SDK 不自动扩大列数、丢弃按钮或拆成多条消息。

## 另请参阅

[消息类型](README.md) · [读取图片与表情](../rich-content.md) · [官方 Markdown](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/type/markdown.html) · [官方素材管理](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/material-management.html)
