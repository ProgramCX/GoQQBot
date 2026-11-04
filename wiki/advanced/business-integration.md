# 将机器人接入业务服务

处理函数负责路由、输入和回复。把业务规则、数据库和外部系统访问交给自己的服务。示例用同一个处理函数接收群聊和单聊命令。业务服务无需判断 QQ 的 HTTP 路径。

## 注入业务服务

创建业务服务后，将服务传给注册函数。

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

`GreetingService` 由业务按需定义。SDK 不要求实现这一接口。应用启动时构造服务并连接数据库。机器人退出后，应用关闭自己持有的资源。

## 并发与生命周期

下游 I/O 使用 `ctx.RequestContext()`，也可从它派生带超时的上下文。长期后台任务保存所需标识和数据副本。由应用管理任务的[生命周期](../getting-started/lifecycle.md)。

## 幂等与持久化

订单、扣款和数据库写入需要幂等。用机器人、场景、会话和消息标识确定幂等键。将处理状态写入持久存储。

- 素材与业务关联时，保存素材 ID、URL 和业务键。字段见[素材管理](../server-api/messages/material-management.md)。
- 流式消息由应用保存游标和状态。用法见[单聊流式消息](../server-api/messages/streaming.md)。

## 注意事项

- 处理函数可能并发执行。注入的服务应支持并发访问。共享 map、缓存和素材引用需要同步保护。
- `Context` 包含原子状态。处理函数只传指针，首次使用后不可复制。
- 后台任务可持有标识和数据副本，无需长期持有整个事件 `Context`。
- WebSocket 恢复后可能重复投递。业务用已保存的幂等键和处理状态识别重复请求。
- 展示引用索引 `REFIDX` 和 `TMP_` 并非稳定的全局消息 ID。
- 应用选择数据库和清理策略。SDK 不会在失败后自动重试 POST。

## 另请参阅

[消息接收与富文本](../server-api/messages/rich-content.md) · [生命周期](../getting-started/lifecycle.md) · [原始 API 扩展](raw-api.md)
