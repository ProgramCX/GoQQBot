# 运行与关闭

机器人独立运行时，用 `StartWithGracefulShutdown()` 处理系统退出信号。机器人和其他服务一起运行时，把服务的上下文传给 `Run(ctx)`，让它们一起退出。

## 让业务上下文管理运行

创建 Bot 并注册好处理函数后，就可以调用 `Run(ctx)`。

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

`Run` 会一直等到上下文取消、Bot 关闭，或事件连接出错且无法恢复。要判断是不是主动取消，用 `errors.Is(err, context.Canceled)`。如果要判断是不是超时，就检查 `context.DeadlineExceeded`。

## 在服务关闭流程中等待

服务退出时，调用 `Shutdown` 停止接收事件，并等 Bot 完成关闭。

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

`Shutdown` 会先取消事件传输和处理函数的上下文，再等 Bot 退出。处理函数如果要发网络请求或查数据库，把 `ctx.RequestContext()` 传进去，这些操作就能收到同一个取消信号。

## 生命周期约束

创建 Bot、运行 Bot 和关闭 Bot 分别由以下方法负责。

| 操作 | 行为 |
| --- | --- |
| `NewWithConfig` / `NewBot` | 检查配置并创建实例。此时还未联网，也未启动工作协程 |
| `Run(ctx)` | 启动 Bot。运行期间再次调用会返回 `ErrAlreadyStarted` |
| `Start()` | 用后台上下文调用 `Run`。退出时调用 `Stop` 或 `Shutdown` |
| `StartWithGracefulShutdown()` | 收到中断或终止信号后关闭 Bot。关闭时遇到的错误也会返回 |
| `Shutdown(ctx)` | 可以重复调用，也可以并发调用。即使还没调用 `Run`，也会永久关闭这个 Bot |
| `Stop()` | 按 `runtime.shutdownTimeout` 设置等待时限，再调用 `Shutdown` |

## 注意事项

- SDK 会保留处理函数和错误回调各自的首个错误，在退出时一并返回。即使 `errors.Is` 匹配到取消错误，返回值也可能包含其他错误。检查方法见[错误处理](../logging/errors.md)。
- 如果等到超时 Bot 还没关闭，`Shutdown` 会返回上下文错误。忽略取消信号的业务函数可能还在运行。
- 关闭后的 Bot 无法再次运行。调用 `Run` 会收到 `ErrClosed`。需要重启时，重新创建 Bot。
- 如果 API 客户端、HTTP 客户端或日志输出由你传入，就由你负责关闭。SDK 只关闭自己管理的资源。
- 在调用 `Run` 之前注册好处理函数。运行期间，处理函数可能并发执行。
- `*types.Context` 用过之后，请继续使用同一个对象。处理消息时，事件数据和公开字段都按只读数据使用。

## 另请参阅

[配置并发及关闭时限](configuration.md) · [错误处理](../logging/errors.md) · [官方连接维护及事件通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
