# GoQQBot

<p align="center">
  <a href="README.md">English</a> | <a href="README_CN.md">简体中文</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
</p>

A lightweight QQ Bot framework written in Go, based on QQ Bot API v2. It supports WebSocket and Webhook event receiving, command routing, event handling, and multiple message types.

[Usage wiki (Chinese)](wiki/README.md) · [Quick start](wiki/getting-started/quick-start.md) · [Full navigation](wiki/_Sidebar.md)

```
   ______          ____    ____     ____           __
  / ____/  ____   / __ \  / __ \   / __ )  ____   / /_
 / / __   / __ \ / / / / / / / /  / __  | / __ \ / __/
/ /_/ /  / /_/ // /_/ / / /_/ /  / /_/ / / /_/ // /_
\____/   \____/ \___\_\ \___\_\ /_____/  \____/ \__/
```


[Interaction guide (Chinese)](docs/interactions.md): coverage matrix, callback routing, one-attempt ACK, mentions and clickable text commands.

## Ordinary messages

```go
bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
    return ctx.ReplyText("Received: " + ctx.Text())
})
```

OnMessage takes the event type first, consistently with OnEvent. Use `OnAnyMessage(handler)` for a cross-scene fallback.
Registered commands take precedence; other text, unknown commands and attachment-only messages reach the message handler.
Use `ctx.Message()` for sender/message IDs and attachments; see the [receiving guide (Chinese)](docs/receiving.md).

## Markdown and buttons in a few lines

```go
return ctx.ReplyMarkdown("## Done",
    button.Command("Help", "/help"),
    button.Link("Website", "https://example.com"),
)
```

Import `button` from `github.com/ProgramCX/GoQQBot/pkg/message/button`.
Unified replies infer group/C2C/channel recipients and manage reply sequences.
Immutable builders support layouts, permissions, callbacks and raw protocol extensions.
See the [message guide and migration examples (Chinese)](docs/messages.md)
and the [compilable menu example](examples/menu/main.go).

## WebSocket support

Set `connection.mode: websocket` to receive events over WebSocket.
Copy the [WebSocket configuration](config.websocket.example.yaml), fill in AppID/AppSecret, then run:

```sh
go run ./examples/echo -config config.yaml
```

Supports heartbeats, reconnection, session resume and cancellable shutdown.
Existing `OnCommand` / `OnEvent` registrations remain available. The default mode is Webhook;
legacy `server.pool` settings are supported and `bot.token` is no longer required.
See the [connection guide (Chinese)](docs/websocket.md) for defaults, lifecycle and replay limitations,
and the [Webhook configuration](config.webhook.example.yaml) for HTTP callbacks.

## ✨ Features

-  **Lightweight & Efficient** - Built with Go for excellent performance and low resource usage
- **Webhook Support** - Receive messages via HTTP Webhook with Ed25519 signature verification
- **Command Routing** - Flexible command registration and routing system with parameter parsing
- **Message Types** - Support text, Markdown, ARK, rich media, and more
- **Group & Private Chat** - Handle both group (@bot) and private chat messages
- **Event Listening** - Listen to friend add/remove, bot added/removed from groups, etc.
- **Goroutine Pool** - Built-in task queue and worker pool for efficient concurrent processing
- **Graceful Shutdown** - Signal handling for smooth server shutdown

## 📦 Installation

```bash
go get github.com/ProgramCX/GoQQBot
```

## 🚀 Quick Start

### 1. Create Config File

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  pool:
    workers: 10      # Number of worker goroutines
    queueSize: 100   # Task queue size

bot:
  appId: "your-app-id"
  appSecret: "your-app-secret"
```

> 💡 Get your Bot credentials from [QQ Open Platform](https://q.qq.com/qqbot/)

### 2. Write Your Bot

```go
package main

import (
    "errors"
    "log"
    "net/http"
    "path/filepath"
    "runtime"

    "github.com/ProgramCX/GoQQBot/pkg/api/c2c"
    "github.com/ProgramCX/GoQQBot/pkg/api/group"
    "github.com/ProgramCX/GoQQBot/pkg/message"
    "github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
    "github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	// In development environment
    _, currentFile, _, _ := runtime.Caller(0)
    configPath := filepath.Join(filepath.Dir(currentFile), "config.yaml")

	// In production environment, use the following line instead
    //exePath, _ := os.Executable()
	//exeDir := filepath.Dir(exePath)
	//configPath := filepath.Join(exeDir, "bot.yaml")

    // Create bot instance
    bot, err := goqqrobot.NewBot(configPath, nil)
    if err != nil {
        log.Printf("Error in creating new bot instance: %v", err)
        return
    }

    // Register group command
    bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, func(ctx *types.Context) error {
        if len(ctx.Args) == 0 { return group.ReplyPlainText(ctx, "Usage: /hello name") }
        err := group.ReplyPlainText(ctx, "Hello, "+ctx.Args[0])
        if err != nil {
            log.Printf("Reply failed: %v", err)
        }
        return nil
    })

    // Register private chat command
    bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
        if len(ctx.Args) == 0 { return c2c.ReplyPlainText(ctx, "Usage: /hello name") }
        err := c2c.ReplyPlainText(ctx, "Hello, "+ctx.Args[0])
        if err != nil {
            log.Printf("Reply failed: %v", err)
        }
        return nil
    })

    // Listen to friend add event
    bot.OnEvent(message.FRIEND_ADD, func(ctx *types.Context) error {
        log.Printf("New friend request received")
        return nil
    })

    // Start bot with graceful shutdown
    if err := bot.StartWithGracefulShutdown(); err != nil {
        if !errors.Is(err, http.ErrServerClosed) {
            log.Fatal(err)
        }
    }
}
```

### 3. Run

```bash
go run main.go
```

## 📚 API Documentation

### Bot Instance

```go
// Create bot from config file
bot, err := goqqrobot.NewBot("config.yaml", nil)

// Start server
bot.Start()

// Start with graceful shutdown
bot.StartWithGracefulShutdown()

// Stop server
bot.Stop()
```

### Command Registration

```go
// Register command handler
bot.OnCommand("cmd", message.GROUP_AT_MESSAGE_CREATE, handler)

// Register event handler
bot.OnEvent(message.FRIEND_ADD, handler)
```

### Supported Event IDs

**Private Chat (C2C) Events:**
| Event ID | Description |
|---------|-------------|
| `C2C_MSG_RECEIVE` | Private message received |
| `C2C_MESSAGE_CREATE` | Private message created |
| `FRIEND_ADD` | Friend added |
| `FRIEND_DEL` | Friend deleted |
| `C2C_MSG_REJECT` | Message rejected |

**Group Events:**
| Event ID | Description |
|---------|-------------|
| `GROUP_AT_MESSAGE_CREATE` | Group @ message |
| `GROUP_ADD_ROBOT` | Bot added to group |
| `GROUP_DEL_ROBOT` | Bot removed from group |
| `GROUP_MSG_RECEIVE` | Group message received |
| `GROUP_MSG_REJECT` | Group message rejected |

### Message Sending APIs

**Group Messages:**
```go
import "github.com/ProgramCX/GoQQBot/pkg/api/group"

// Send plain text
group.ReplyPlainText(ctx, "message content")

// Send Markdown
group.ReplyMarkdownContent(ctx, markdownContent)

// Send rich media
group.ReplyMedia(ctx, fileInfo)
```

**Private Messages:**
```go
import "github.com/ProgramCX/GoQQBot/pkg/api/c2c"

// Send plain text
c2c.ReplyPlainText(ctx, "message content")

// Send Markdown
c2c.ReplyMarkdownContent(ctx, markdownContent)

// Send rich media
c2c.ReplyMedia(ctx, fileInfo)
```

### Context Object

```go
type Context struct {
    Payload    *Payload   // Raw message payload
    Command    string     // Command name (without /)
    Args       []string   // Command arguments
    RawMessage string     // Raw message content
}
```

## 📁 Project Structure

```
GoQQBot/
├── cmd/                    # Command entry points
│   └── dev/
│       └── main.go         # Development example
├── internal/               # Internal code
│   ├── adapter/            # Webhook adapter and validator
│   ├── core/               # Core functionality
│   │   ├── api/            # API clients (internal)
│   │   ├── pool/           # Goroutine pool
│   │   └── router/         # Command router
│   └── foundation/         # Foundation components
│       ├── config/         # Configuration management (moved to pkg/)
│       ├── logger/         # Logging system
│       └── signal.go       # Signal handling
├── pkg/                    # Public packages
│   ├── api/                # Public API clients
│   │   ├── c2c/            # Private message APIs
│   │   ├── channel/        # Channel APIs
│   │   └── group/          # Group message APIs
│   ├── config/             # Configuration management
│   ├── goqqrobot/          # Bot core package
│   └── message/            # Message types and utilities
├── shared/                 # Shared types
│   └── types/
└── config.yaml             # Example configuration
```

## ⚙️ Configuration

| Option | Type | Default | Description |
|-------|------|---------|-------------|
| `server.port` | int | 8080 | Webhook server port |
| `server.host` | string | "0.0.0.0" | Listen address |
| `server.pool.workers` | int | 10 | Number of workers |
| `server.pool.queueSize` | int | 100 | Task queue size |
| `bot.appId` | string | - | QQ Bot App ID |
| `bot.token` | string | - | Deprecated; optional and ignored |
| `bot.appSecret` | string | - | QQ Bot App Secret |

## 🔒 Webhook Setup

Configure Webhook in [QQ Open Platform](https://q.qq.com/qqbot/#/developer/webhook-setting):

- **URL**: `https://your-domain/webhook`
- **Verification**: Select Ed25519 signature verification

The framework automatically handles validation requests and message signature verification.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the [MIT](LICENSE) License.

## 📖 Links

- [QQ Bot API Documentation](https://bot.q.qq.com/wiki/develop/api-v2/)
- [QQ Open Platform](https://q.qq.com/qqbot/)
- [C2C/group typed messages, recall, cards and multipart media uploads](docs/media.md) (includes 14 interactive acceptance cases)
- [Reusable Markdown image materials](docs/materials.md) (upload, pagination, deletion and interactive examples)
- [Logging, error reporting and panic handling](docs/diagnostics.md) (silent defaults, slog injection and race checks)

---

<p align="center">Made with ❤️ by <a href="https://github.com/ProgramCX">ProgramCX</a></p>
