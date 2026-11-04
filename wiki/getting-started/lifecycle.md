# 运行与关闭

独立机器人可用 `StartWithGracefulShutdown()`。嵌入已有服务时，用 `Run(ctx)` 让业务统一管理取消信号。

## 让业务上下文管理运行

下面的函数接受已创建并注册好处理函数的机器人。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
)

func Serve(ctx context.Context, bot *goqqrobot.Bot) error {
	return bot.Run(ctx)
}
```

`Run` 会阻塞，直到取消、关闭或发生不可恢复的传输错误。调用方取消时，可用 `errors.Is(err, context.Canceled)` 检查返回值。超时则检查 `context.DeadlineExceeded`。

## 在服务关闭流程中等待

调用 `Shutdown` 停止接收事件并等待关闭。

```go
package guide

import (
	"context"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
)

func StopBot(bot *goqqrobot.Bot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return bot.Shutdown(ctx)
}
```

`Shutdown` 先取消传输和处理函数的上下文，再等待关闭。多次或并发调用都受支持。网络和数据库操作可以接收 `ctx.RequestContext()`，随处理函数取消。

## 生命周期约束

下表列出各方法管理的阶段。

| 操作 | 行为 |
| --- | --- |
| `NewWithConfig` / `NewBot` | 校验和组装，不联网、不启动工作协程 |
| `Run(ctx)` | 启动一次。运行中再次调用返回 `ErrAlreadyStarted` |
| `Start()` | 用后台上下文调用 Run，需另行 Stop/Shutdown |
| `StartWithGracefulShutdown()` | 监听中断及终止信号，保留关闭故障 |
| `Shutdown(ctx)` | 可重复调用。在 Run 前调用也会永久关闭实例 |
| `Stop()` | 使用配置的 `runtime.shutdownTimeout` 调用 Shutdown |

## 注意事项

- 处理函数或观察函数的首个故障会并入退出错误。返回值匹配取消错误时，仍可能含有其他故障。检查方法见[错误处理](../logging/errors.md)。
- 等待超时时，`Shutdown` 返回上下文错误。忽略取消信号的业务函数仍可能继续执行。
- 已关闭实例再次运行会返回 `ErrClosed`。重新启动时，创建新的 Bot。
- 注入的 API 客户端、HTTP 客户端和日志输出由调用方管理。SDK 不关闭这些外部资源。
- 处理函数可能并发运行。注册在 `Run` 前完成。
- `*types.Context` 首次使用后不可复制。处理期间，载荷及公开字段保持只读。

## 另请参阅

[配置并发及关闭时限](configuration.md) · [错误处理](../logging/errors.md) · [官方连接维护及事件通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
