# GoQQBot

<p align="center">
  <a href="README.md">English</a> | <a href="README_CN.md">简体中文</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
</p>

一个用 Go 语言编写的轻量级 QQ 机器人框架，基于 QQ Bot API v2 开发，支持 WebSocket 和 Webhook 方式接收事件，提供命令路由、事件处理和多种消息类型支持。

[使用文档](wiki/README.md) · [快速开始](wiki/getting-started/quick-start.md) · [完整目录](wiki/_Sidebar.md)

在仓库根目录运行 `npm ci` 和 `npm run docs:dev` 可预览文档站。构建命令为 `npm run docs:build`，产物位于 `wiki/.vitepress/dist`。GitHub Pages 的发布源选择 GitHub Actions，`文档站` 工作流会在主分支文档更新后部署。

```
   ______          ____    ____     ____           __
  / ____/  ____   / __ \  / __ \   / __ )  ____   / /_
 / / __   / __ \ / / / / / / / /  / __  | / __ \ / __/
/ /_/ /  / /_/ // /_/ / / /_/ /  / /_/ / / /_/ // /_
\____/   \____/ \___\_\ \___\_\ /_____/  \____/ \__/
```


[按钮回调与文本交互指南](docs/interactions.md)：包含覆盖清单、`OnButton` / `OnInteraction`、ACK 和 @/点击命令标签。

## 普通消息也能一处处理

```go
bot.OnMessage(message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
    return ctx.ReplyText("收到：" + ctx.Text())
})
```

事件类型放在第一个参数，与 OnEvent 一致。跨场景统一处理使用 `OnAnyMessage(handler)`。
匹配命令优先；普通文本、未知命令及仅有附件的消息进入 OnMessage。
`ctx.Message()` 可读取发送者、消息 ID 和附件，详见 [接收消息指南](docs/receiving.md)。

## 几行发送 Markdown 和按钮

```go
return ctx.ReplyMarkdown("## 操作成功",
    button.Command("帮助", "/help"),
    button.Command("绑定", "/bind "),
)
```

`button` 来自 `github.com/ProgramCX/GoQQBot/pkg/message/button`。
统一回复入口自动处理群聊/私聊/频道路由和消息序号；支持布局、权限、回调按钮及原始字段扩展。
详见 [消息构造与 GoCortex 迁移](docs/messages.md) 和 [完整菜单示例](examples/menu/main.go)。

## WebSocket 接入

现在可以在配置文件中设置 `connection.mode: websocket`，使用 WebSocket 接收事件。
复制 [WebSocket 配置示例](config.websocket.example.yaml)，填写 AppID/AppSecret 后运行：

```sh
go run ./examples/echo -config config.yaml
```

支持心跳、断线重连、会话恢复和可取消关闭。现有 `OnCommand` / `OnEvent` 注册方式继续可用。
缺省连接方式为 Webhook，旧 `server.pool` 配置仍可读取；`bot.token` 不再必填。
完整说明、默认值与恢复边界见 [接入文档](docs/websocket.md)，另有 [Webhook 配置示例](config.webhook.example.yaml)。

## ✨ 特性

-  **轻量高效** - 基于 Go 语言，性能优异，资源占用低
- **Webhook 支持** - 基于 HTTP Webhook 接收消息，支持 Ed25519 签名验证
- **命令路由** - 灵活的命令注册和路由系统，支持参数解析
- **消息类型** - 支持文本、Markdown、ARK、富媒体等多种消息类型
- **群聊 & 私聊** - 同时支持群聊 (@机器人) 和私聊消息处理
- **事件监听** - 支持好友添加/删除、机器人被拉入/移出群等事件
- **协程池** - 内置任务队列和协程池，高效处理并发消息
- **优雅关闭** - 支持信号监听，平滑关闭服务

## 📦 安装

```bash
go get github.com/ProgramCX/GoQQBot
```

## 🚀 快速开始

### 1. 创建配置文件

创建 `config.yaml` 文件：

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  pool:
    workers: 10      # 工作协程数
    queueSize: 100   # 任务队列大小

bot:
  appId: "your-app-id"
  appSecret: "your-app-secret"
```

> 💡 获取 Bot 配置信息，请前往 [QQ 开放平台](https://q.qq.com/qqbot/)

### 2. 编写你的机器人

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
    // 开发环境使用以下代码
    _, currentFile, _, _ := runtime.Caller(0)
    configPath := filepath.Join(filepath.Dir(currentFile), "config.yaml")

    // 生产环境使用以下代码
	//exePath, _ := os.Executable()
	//exeDir := filepath.Dir(exePath)
	//configPath := filepath.Join(exeDir, "config.yaml")

    // 创建机器人实例
    bot, err := goqqrobot.NewBot(configPath, nil)
    if err != nil {
        log.Printf("创建机器人实例失败: %v", err)
        return
    }

    // 注册群聊命令
    bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, func(ctx *types.Context) error {
        if len(ctx.Args) == 0 { return group.ReplyPlainText(ctx, "Usage: /hello name") }
        err := group.ReplyPlainText(ctx, "你好，"+ctx.Args[0])
        if err != nil {
            log.Printf("回复失败: %v", err)
        }
        return nil
    })

    // 注册私聊命令
    bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
        if len(ctx.Args) == 0 { return c2c.ReplyPlainText(ctx, "Usage: /hello name") }
        err := c2c.ReplyPlainText(ctx, "你好，"+ctx.Args[0])
        if err != nil {
            log.Printf("回复失败: %v", err)
        }
        return nil
    })

    // 监听好友添加事件
    bot.OnEvent(message.FRIEND_ADD, func(ctx *types.Context) error {
        log.Printf("收到好友添加请求")
        return nil
    })

    // 启动机器人（支持优雅关闭）
    if err := bot.StartWithGracefulShutdown(); err != nil {
        if !errors.Is(err, http.ErrServerClosed) {
            log.Fatal(err)
        }
    }
}
```

### 3. 运行

```bash
go run main.go
```

## 📚 API 文档

### 机器人实例

```go
// 从配置文件创建机器人
bot, err := goqqrobot.NewBot("config.yaml", nil)

// 启动服务
bot.Start()

// 带优雅关闭的启动
bot.StartWithGracefulShutdown()

// 停止服务
bot.Stop()
```

### 命令注册

```go
// 注册命令处理器
bot.OnCommand("cmd", message.GROUP_AT_MESSAGE_CREATE, handler)

// 注册事件处理器
bot.OnEvent(message.FRIEND_ADD, handler)
```

### 支持的 Event ID

**私聊事件 (C2C)：**
| Event ID | 说明 |
|---------|------|
| `C2C_MSG_RECEIVE` | 收到私聊消息 |
| `C2C_MESSAGE_CREATE` | 私聊消息创建 |
| `FRIEND_ADD` | 好友添加 |
| `FRIEND_DEL` | 好友删除 |
| `C2C_MSG_REJECT` | 消息被拒绝 |

**群聊事件：**
| Event ID | 说明 |
|---------|------|
| `GROUP_AT_MESSAGE_CREATE` | 群@消息 |
| `GROUP_ADD_ROBOT` | 机器人被添加进群 |
| `GROUP_DEL_ROBOT` | 机器人被移出群 |
| `GROUP_MSG_RECEIVE` | 收到群消息 |
| `GROUP_MSG_REJECT` | 群消息被拒绝 |

### 消息发送 API

**群聊消息：**
```go
import "github.com/ProgramCX/GoQQBot/pkg/api/group"

// 发送纯文本
group.ReplyPlainText(ctx, "消息内容")

// 发送 Markdown
group.ReplyMarkdownContent(ctx, markdownContent)

// 发送富媒体
group.ReplyMedia(ctx, fileInfo)
```

**私聊消息：**
```go
import "github.com/ProgramCX/GoQQBot/pkg/api/c2c"

// 发送纯文本
c2c.ReplyPlainText(ctx, "消息内容")

// 发送 Markdown
c2c.ReplyMarkdownContent(ctx, markdownContent)

// 发送富媒体
c2c.ReplyMedia(ctx, fileInfo)
```

### Context 对象

```go
type Context struct {
    Payload    *Payload   // 原始消息负载
    Command    string     // 命令名（不含/）
    Args       []string   // 命令参数列表
    RawMessage string     // 原始消息内容
}
```

## 📁 项目结构

```
GoQQBot/
├── cmd/                    # 命令行入口
│   └── dev/
│       └── main.go         # 开发环境示例
├── internal/               # 内部代码
│   ├── adapter/            # Webhook 适配器和验证器
│   ├── core/               # 核心功能
│   │   ├── api/            # API 客户端（内部）
│   │   ├── pool/           # 协程池
│   │   └── router/         # 命令路由器
│   └── foundation/         # 基础组件
│       ├── config/         # 配置管理（已移至pkg/）
│       ├── logger/         # 日志系统
│       └── signal.go       # 信号处理
├── pkg/                    # 公开包
│   ├── api/                # 公开 API 客户端
│   │   ├── c2c/            # 私聊消息 API
│   │   ├── channel/        # 频道 API
│   │   └── group/          # 群聊消息 API
│   ├── config/             # 配置管理
│   ├── goqqrobot/          # 机器人核心包
│   └── message/            # 消息类型和工具
├── shared/                 # 共享类型
│   └── types/
└── config.yaml             # 配置文件示例
```

## ⚙️ 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|-------|------|-------|------|
| `server.port` | int | 8080 | Webhook 服务端口 |
| `server.host` | string | "0.0.0.0" | 监听地址 |
| `server.pool.workers` | int | 10 | 工作协程数 |
| `server.pool.queueSize` | int | 100 | 任务队列大小 |
| `bot.appId` | string | - | QQ Bot App ID |
| `bot.token` | string | - | Deprecated; optional and ignored |
| `bot.appSecret` | string | - | QQ Bot App Secret |

## 🔒 Webhook 配置

在 [QQ 开放平台](https://q.qq.com/qqbot/#/developer/webhook-setting) 配置 Webhook：

- **URL**: `https://your-domain/webhook`
- **验证方式**: 选择 Ed25519 签名验证

框架会自动处理验证请求和消息签名验证。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

## 📄 许可证

本项目基于 [MIT](LICENSE) 许可证开源。

## 📖 相关链接

- [QQ Bot API 文档](https://bot.q.qq.com/wiki/develop/api-v2/)
- [QQ 开放平台](https://q.qq.com/qqbot/)
- [群聊与单聊：结构化接收、撤回、卡片和富媒体上传](docs/media.md)（附 14 项按钮验收示例）
- [Markdown 图片素材管理与业务复用](docs/materials.md)（上传、分页查询、删除及按钮示例）
- [日志、错误通知与 panic 处理](docs/diagnostics.md)（默认静默、slog 注入、错误回调与竞态检查）

---

<p align="center">Made with ❤️ by <a href="https://github.com/ProgramCX">ProgramCX</a></p>
