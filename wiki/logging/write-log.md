# 如何写入日志文件

用标准库 `log/slog` 创建日志器，再通过 `goqqrobot.WithLogger` 传给 Bot。日志文件由应用创建和关闭。

## 同时写入文件和控制台

先按[快速开始](../getting-started/quick-start.md)准备 `config.yaml`，再运行完整程序。

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

`os.O_APPEND` 保留已有日志，启动时不会截断文件。只写文件时，将 `io.MultiWriter(file, os.Stdout)` 改为 `file`。业务代码也可使用同一个 `logger`。为 Bot 设置日志器不会修改全局的 `slog.Default()`。

## 格式与级别

用下面的函数创建 JSON 日志器。返回的 `LevelVar` 可在运行中调整级别。

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

用 `level.Set(slog.LevelDebug)` 开启 Debug，再用 `level.Set(slog.LevelInfo)` 恢复 Info。命令注册等诊断信息使用 Debug 级别。调整级别不改变消息处理行为。

## 日志轮转

生产部署可把标准输出交给运行环境收集。应用也可以提供实现 `io.Writer` 的轮转输出，再交给 slog Handler。

## 注意事项

- 处理函数在退出前仍可能写日志。等机器人退出后再关闭文件，避免写入失败。多个服务共用日志器时，由外层服务[统一管理关闭](../getting-started/lifecycle.md)。
- 传入 `WithLogger(nil)` 或省略该选项，SDK 都会保持静默。
- SDK 不轮转文件，也不管理日志保留周期。

## 另请参阅

[错误处理及 panic 观察](errors.md) · [运行与关闭](../getting-started/lifecycle.md) · [官方 TraceID 说明](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html)
