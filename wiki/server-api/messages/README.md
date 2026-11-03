# 消息收发

使用同一个处理函数接收消息，并通过上下文回复当前会话。SDK 自动选择接收方、被动回复来源和发送路径；业务层只负责决定内容。

## 注册一个跨场景命令

下面是可放进业务包的注册函数，`bot` 来自快速开始中的 `NewBot` 或 `NewWithConfig`。请在运行机器人前调用它。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterHello(bot *goqqrobot.Bot) {
	hello := func(ctx *types.Context) error {
		return ctx.ReplyText("你好，我已收到消息。")
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, hello)
	bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, hello)
}
```

在单聊发送 `/hello`，在群里 @机器人后发送 `/hello`。两个事件是否可接收，仍取决于平台权限和订阅配置。

## 选择入口

| 目标 | 入口 | 说明 |
| --- | --- | --- |
| 处理已注册命令 | `OnCommand(command, eventType, handler)` | 命令优先于普通消息处理 |
| 处理某种消息 | `OnMessage(eventType, handler)` | 包括未知命令和仅带附件的消息 |
| 跨消息类型兜底 | `OnAnyMessage(handler)` | 排在命令和指定类型 OnMessage 之后 |
| 读取原始正文与附件 | `ctx.Message()` | 返回独立快照，不执行网络请求 |
| 简单回复 | `ctx.ReplyText`、`ctx.ReplyMarkdown` | 自动选择回复来源及接收方 |
| 组合发送内容 | `ctx.Reply(message.Outgoing)` | 内容构造器可复用，不执行隐式上传 |
| 自行管理回复目标 | `Client.SendReply` | 显式提供 `ReplyTarget`，自行管理多次回复序号 |
| 明确主动发送 | `Client.SendMessage` | 受平台主动消息资格及配额限制 |

## 阅读顺序

- [单聊消息](c2c/README.md)：接收、回复、主动发送及输入状态。
- [群聊消息](group/README.md)：群 @消息、全量消息和跨场景业务复用。
- [频道消息](channel/README.md)：子频道、频道私信与表态。
- [消息类型](types/README.md)：文本、Markdown、图文卡片和 ARK。
- [消息交互](interactions/README.md)：按钮、文本交互、回调和 ACK。
- [富媒体消息](rich-media/README.md)：URL 上传和本地分片上传。
- [Markdown 图片素材](material-management.md)：上传、保存业务引用、复用和删除。
- [读取图片、表情与转发](rich-content.md)：附件、ARKData、MsgElement 和 MessageScene。
- [引用与撤回](reference-and-recall.md)：区分消息 ID、展示索引和被动回复来源。
- [单聊流式消息](streaming.md)：顺序发送和业务恢复。

## 使用边界

`*types.Context` 只按指针传递，第一次使用后不要复制。处理期间其公开字段和载荷保持只读。SDK 的业务处理函数可并发运行；数据库事务、会话串行和业务幂等由调用方负责。

SDK 不自动重试消息 POST。返回错误不一定表示平台未收到请求。需要发送结果时使用 `SendReply`；结果非 nil 且错误非 nil 表示 HTTP 成功但响应解码失败，检查 `Raw` 后再决定恢复方式，不要直接重发。

## 另请参阅

[消息类型](types/README.md) · [引用与撤回](reference-and-recall.md) · [官方消息收发概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/overview.html)
