# 创建第一个机器人

本例通过 WebSocket 接收群聊艾特和 QQ 单聊消息，用户发送 `/hello` 后回复。无需公网回调地址；需要机器人已经开通相应能力。

## 准备配置

在程序启动目录创建 `config.yaml`，替换两个占位值：

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

SDK 不会自动展开 YAML 中的环境变量。使用环境变量或密钥管理服务时，参见[内存配置](configuration.md#使用内存配置)。不要把真实配置提交到版本库。

## 编写程序

保存为 `main.go`：

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

运行：

```sh
go mod tidy
go run .
```

在单聊发送 `/hello`，或在群里 **艾特机器人后发送 `/hello`**。`OnCommand` 的参数顺序是“命令、事件类型、处理函数”；命令可写 `hello`，不必加 `/`。

## 接收普通消息

在 `StartWithGracefulShutdown` 之前调用下面的注册函数：

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

此处仅演示回复有正文的消息。已注册命令优先，其他文字、未知命令和附件消息由 `OnMessage` 接收。图片可能没有文本，实际业务不要仅凭 `ctx.Text()` 非空判断是否有消息，附件读取见[事件路由](events/routing.md)。

## 另请参阅

- [事件路由](events/routing.md)
- [WebSocket 接入与重连](events/websocket.md)
- [日志输出](../logging/write-log.md)
- [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
