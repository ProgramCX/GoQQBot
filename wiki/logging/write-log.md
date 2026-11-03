# 如何写入日志文件

使用标准库 `log/slog` 创建日志器，通过 `goqqrobot.WithLogger` 注入。文件由应用创建和关闭，SDK 不接管所有权。

## 同时写入文件和控制台

下面是完整程序。先按[快速开始](../getting-started/quick-start.md)准备 `config.yaml`：

```go
package main

import (
	"io"
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
	file, err := os.OpenFile("bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	logger := slog.New(slog.NewTextHandler(io.MultiWriter(file, os.Stdout), nil))
	bot, err := goqqrobot.NewBot("config.yaml", nil, goqqrobot.WithLogger(logger))
	if err != nil {
		return err
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("你好")
	})
	return bot.StartWithGracefulShutdown()
}
```

`os.O_APPEND` 保留既有日志，而不是启动时截断文件。只写文件时将 `io.MultiWriter(file, os.Stdout)` 改为 `file`。应用自己写业务日志也应使用同一个 `logger`，注入不会改写全局 `slog.Default()`。

文件在机器人退出后关闭；若应用有多个服务共用日志器，应由更外层的服务生命周期统一关闭。

## 格式与级别

下面的完整函数创建 JSON 日志器，返回 `LevelVar` 供应用在运行中调整：

```go
package guide

import (
	"io"
	"log/slog"
)

func NewLogger(output io.Writer) (*slog.Logger, *slog.LevelVar) {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	return slog.New(handler), level
}
```

调用方可用 `level.Set(slog.LevelDebug)` 开启 Debug，使用 `level.Set(slog.LevelInfo)` 恢复。命令注册等诊断信息位于 Debug；开启级别不改变业务消息处理行为。传 `WithLogger(nil)` 或省略该选项会保持 SDK 静默。

## 日志轮转

SDK 不做文件轮转或日志保留管理。生产部署可写标准输出交给运行环境收集，或由应用提供实现 `io.Writer` 的轮转输出，再交给 slog Handler。不要让日志器在关闭后仍被运行中的处理函数使用。

## 另请参阅

- [错误处理及 panic 观察](errors.md)
- [运行与关闭](../getting-started/lifecycle.md)
- [官方 TraceID 说明](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
