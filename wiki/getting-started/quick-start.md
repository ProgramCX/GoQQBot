# 创建第一个机器人

用 WebSocket 接收单聊消息和群聊艾特。机器人收到 `/hello` 后回复。运行程序无需公网回调地址。

## 准备配置

在程序启动目录创建 `config.yaml`，填写 `appId` 和 `appSecret`。获取方式见[访问凭证](credentials.md)。

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

使用环境变量或密钥管理服务时，改用[内存配置](configuration.md#使用内存配置)。

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

在业务项目目录运行以下命令。

```sh
go mod tidy
go run .
```

在单聊发送 `/hello`。在群聊中，**艾特机器人后发送 `/hello`**。

`/hello` -> WebSocket -> `OnCommand` -> 回复当前会话

`OnCommand` 依次接收命令、事件类型和处理函数。注册命令时可写 `hello`，不必加 `/`。

## 接收普通消息

在 `StartWithGracefulShutdown` 之前调用注册函数。示例只回复带正文的消息。

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

已注册命令优先处理。其他文字、未知命令和附件消息由 `OnMessage` 接收。附件读取方法见[事件路由](events/routing.md)。

## 注意事项

- 机器人需已开通相应的单聊和群聊能力。
- SDK 不展开 YAML 中的环境变量。
- 提交真实配置会泄露凭证。将真实配置留在本地。
- 图片消息可能没有文字。只检查 `ctx.Text()` 会漏掉仅含附件的消息。

## 另请参阅

[事件路由](events/routing.md) · [WebSocket 接入与重连](events/websocket.md) · [日志输出](../logging/write-log.md) · [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
