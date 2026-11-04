# GoQQBot

[English](README.md) · [简体中文](README_CN.md)

![Go 1.24.4+](https://img.shields.io/badge/Go-1.24.4%2B-00ADD8?logo=go) [![MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

GoQQBot is a Go SDK for QQ Bot API v2. Receive events over WebSocket or Webhook, use a shared reply API for direct, group, and channel messages, and call the group, channel, and robot management APIs from standalone services.

[Documentation (Chinese)](https://programcx.github.io/GoQQBot/) · [Quick start](wiki/getting-started/quick-start.md) · [API coverage](wiki/reference/coverage.md)

## Install

```sh
go get github.com/ProgramCX/GoQQBot
```

## Quick start

Select WebSocket in the QQ bot dashboard, then create `config.yaml`. [QQ bot dashboard](https://q.qq.com/qqbot/dashboard/)

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

Save the following as `main.go`. Direct and group chats share one handler. The SDK routes each reply to the current conversation.

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
			return ctx.ReplyText("Usage: /hello name")
		}
		return ctx.ReplyText("Hello, " + ctx.Args[0])
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

Send `/hello Alice` in a QQ direct chat, or mention the bot before sending the command in a test group. For public HTTPS callbacks, follow the Webhook setup guide. [Webhook setup](wiki/getting-started/events/webhook.md) · [Event subscriptions](wiki/getting-started/events/intents.md)

## Guides

- [Receive messages](wiki/getting-started/events/routing.md)
- [Send messages, Markdown, and media](wiki/server-api/messages/README.md)
- [Buttons and callbacks](wiki/server-api/messages/interactions/README.md)
- [Group management](wiki/server-api/group-management/README.md)
- [Channel management](wiki/server-api/channel-management/README.md)
- [Robot profile, sharing, menus, and panels](wiki/server-api/robot/README.md)

## Contributing

Run `go test ./...` and `go vet ./...` to check Go changes. To preview the documentation, run `npm ci` and `npm run docs:dev` from the repository root.

Licensed under [MIT](LICENSE).
