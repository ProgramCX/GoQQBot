# GoQQBot

[English](README.md) · [简体中文](README_CN.md)

![Go 1.24.4+](https://img.shields.io/badge/Go-1.24.4%2B-00ADD8?logo=go) [![MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

GoQQBot 是基于 QQ 机器人 API v2 的 Go SDK。可通过 WebSocket 或 Webhook 接收事件，使用同一套回复方法处理 QQ 单聊、群聊及频道消息，也可独立调用群聊、频道和机器人管理接口。

[在线文档](https://programcx.github.io/GoQQBot/) · [快速开始](wiki/getting-started/quick-start.md) · [接口覆盖范围](wiki/reference/coverage.md)

## 安装

```sh
go get github.com/ProgramCX/GoQQBot
```

## 快速开始

在 QQ 机器人运营与管理后台选择 WebSocket 接入，然后创建 `config.yaml`。 [运营与管理后台](https://q.qq.com/qqbot/dashboard/)

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

将下列代码保存为 `main.go`。单聊与群聊共用同一个处理函数，SDK 根据当前会话发送回复。

```go
package main

import (
	"log"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	bot, err := goqqrobot.NewBot("config.yaml", nil)
	if err != nil {
		log.Fatal(err)
	}

	hello := func(ctx *types.Context) error {
		if len(ctx.Args) == 0 {
			return ctx.ReplyText("用法：/hello 名字")
		}
		return ctx.ReplyText("你好，" + ctx.Args[0])
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, hello)
	bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, hello)

	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go run .
```

在 QQ 单聊中发送 `/hello 小明`。在测试群里，先 @机器人再发送命令。通过公网 HTTPS 接收回调时，按 Webhook 指南配置。 [Webhook 配置](wiki/getting-started/events/webhook.md) · [事件订阅对照表](wiki/getting-started/events/intents.md)

## 按功能查找

- [接收消息](wiki/getting-started/events/routing.md)
- [发送消息、Markdown 和富媒体](wiki/server-api/messages/README.md)
- [按钮和互动回调](wiki/server-api/messages/interactions/README.md)
- [群聊管理](wiki/server-api/group-management/README.md)
- [频道管理](wiki/server-api/channel-management/README.md)
- [机器人资料、分享、菜单和指令面板](wiki/server-api/robot/README.md)

## 参与开发

运行 `go test ./...` 和 `go vet ./...` 检查 Go 代码。预览文档时，在仓库根目录运行 `npm ci` 和 `npm run docs:dev`。

项目使用 [MIT](LICENSE) 许可证。
