# 创建第一个机器人

这个机器人通过 WebSocket 接收单聊消息和群聊艾特。你发 `/hello`，它就会在当前会话回复。整个过程不需要公网回调地址。

## 准备配置

在运行程序的目录里创建 `config.yaml`，填写 `appId` 和 `appSecret`。获取方式见[访问凭证](credentials.md)。

```yaml
bot:
  appId: "your-app-id"
  appSecret: "your-app-secret"

connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
```

如果凭证来自环境变量或密钥管理服务，可以用[内存配置](configuration.md#使用内存配置)。

## 编写程序

将完整程序保存为 `main.go`。

```go
package main

import (
	"log/slog"
	"os"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	if err := run(); err != nil {
		slog.Error("机器人退出", "error", err)
		os.Exit(1)
	}
}

func run() error {
	bot, err := goqqrobot.NewBot("config.yaml", nil,
		goqqrobot.WithLogger(slog.Default()),
	)
	if err != nil {
		return err
	}
	hello := func(ctx *types.Context) error {
		return ctx.ReplyText("你好！")
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, hello)
	bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, hello)
	return bot.StartWithGracefulShutdown()
}
```

打开终端，在业务项目目录运行以下命令。

```sh
go mod tidy
go run .
```

启动成功后，打开 QQ，在单聊发 `/hello`。在群聊里，**先艾特机器人，再发 `/hello`**。

`/hello` -> WebSocket -> `OnCommand` -> 回复当前会话

`OnCommand` 依次接收命令、事件类型和处理函数。注册命令时可写 `hello`，不必加 `/`。

## 接收普通消息

想让机器人也回复普通文字，可以在 `StartWithGracefulShutdown` 之前注册下面的处理函数。这个示例只回复带文字的消息。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterMessages(bot *goqqrobot.Bot) {
	bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
		if ctx.Text() == "" {
			return nil
		}
		return ctx.ReplyText("收到：" + ctx.Text())
	})
}
```

收到注册过的命令时，SDK 会先交给命令处理函数。普通文字、未知命令和附件消息会交给 `OnMessage`。附件读取方法见[事件路由](events/routing.md)。

## 注意事项

- 机器人需要在 QQ 开放平台开通对应的单聊或群聊能力，才能收到相应消息。
- YAML 中的环境变量占位符会按原样读取。
- 真实配置留在本机即可。提交到仓库会泄露凭证。
- 用户只发图片时，`ctx.Text()` 可能为空。需要处理图片时，记得读取附件。

## 另请参阅

[事件路由](events/routing.md) · [WebSocket 接入与重连](events/websocket.md) · [日志输出](../logging/write-log.md) · [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
