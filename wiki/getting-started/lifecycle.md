# 运行与关闭

独立机器人可以使用 `StartWithGracefulShutdown()`；嵌入已有服务时使用 `Run(ctx)`，由业务统一管理取消信号。

## 让业务上下文管理运行

下面的完整函数接受已构造并注册好处理函数的机器人：

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

`Run` 阻塞到取消、关闭或不可恢复的传输错误。调用方取消时返回值可通过 `errors.Is(err, context.Canceled)` 或 `context.DeadlineExceeded` 检查。处理函数或观察函数的首个故障还会在退出时合并到返回错误，不能仅因匹配取消错误就断言没有其他故障。

## 在服务关闭流程中等待

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

`Shutdown` 先取消机器人传输及处理函数上下文，再等待关闭；可以并发、重复调用。等待超时时返回上下文错误，但忽略取消信号的业务处理函数仍可能继续执行。需要网络或数据库操作的处理函数应将 `ctx.RequestContext()` 传给下游。

## 生命周期约束

| 操作 | 行为 |
| --- | --- |
| `NewWithConfig` / `NewBot` | 校验和组装，不联网、不启动工作协程 |
| `Run(ctx)` | 启动一次；运行中再次调用返回 `ErrAlreadyStarted` |
| `Start()` | 用后台上下文调用 Run，需另行 Stop/Shutdown |
| `StartWithGracefulShutdown()` | 监听中断及终止信号；保留关闭故障 |
| `Shutdown(ctx)` | 可重复调用；在 Run 前调用也会永久关闭实例 |
| `Stop()` | 使用配置的 `runtime.shutdownTimeout` 调用 Shutdown |

已关闭实例不能重启，再次 Run 返回 `ErrClosed`。需要重新启动时构造新 Bot。注入的 API/HTTP 客户端及日志输出仍归调用方持有，SDK 不关闭这些外部资源。

处理函数可能并发运行；在 Run 前完成注册。`*types.Context` 首次使用后不可复制，载荷及公开字段在处理期间保持只读。

## 另请参阅

- [配置并发及关闭时限](configuration.md)
- [错误处理](../logging/errors.md)
- [官方连接维护及事件通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
