# 将机器人接入业务服务

让 handler 处理路由、输入和回复，把业务规则、数据库和外部系统访问放在自己的服务中。下面同一个函数同时处理群聊和单聊命令，不需要在业务服务里判断 QQ HTTP 路径。

## 注入业务服务

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type GreetingService interface {
	Greet(context.Context, string) (string, error)
}

func RegisterGreeting(bot *goqqrobot.Bot, service GreetingService) {
	handler := func(ctx *types.Context) error {
		if len(ctx.Args) != 1 {
			return ctx.ReplyText("用法：/hello 名字")
		}
		reply, err := service.Greet(ctx.RequestContext(), ctx.Args[0])
		if err != nil {
			return err
		}
		return ctx.ReplyText(reply)
	}
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, handler)
	bot.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, handler)
}
```

`GreetingService` 是业务自行定义的最小接口，并非 SDK 规定的框架接口。应用启动时负责构造服务、连接数据库，在机器人退出后关闭自己持有的资源。

## 并发与生命周期

处理函数可能并发执行，注入的服务必须允许相应并发访问；共享 map、缓存和素材引用需要同步保护。`Context` 含原子状态，只能传指针，首次使用后不可复制。

下游 I/O 使用 `ctx.RequestContext()` 或由它派生的超时上下文。需要长期后台任务时，保存业务所需的标识及数据副本，由应用管理任务生命周期，不长期持有整个事件 `Context`。

## 幂等与持久化

WebSocket 恢复可能重复投递。会产生订单、扣款或数据库写入的操作，应按机器人、场景、会话和消息标识确定业务幂等键，并持久化处理状态。不要把展示引用索引 `REFIDX` 或 `TMP_` 视为稳定的全局消息 ID。

素材与业务关联时保存素材 ID、URL 和业务键；流式消息保存自己的流游标和状态。SDK 不会替应用选择数据库、设置自动清理策略或在失败后偷偷重试 POST。

## 另请参阅

- [消息接收与富文本](../server-api/messages/rich-content.md)
- [生命周期](../getting-started/lifecycle.md)
- [原始 API 扩展](raw-api.md)
