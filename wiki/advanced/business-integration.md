# 将机器人接入业务服务

群聊和单聊命令可以共用一个处理函数。SDK 根据事件确定回复对象，业务服务负责生成回复内容，并处理业务规则、数据库和外部系统调用。

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

`GreetingService` 是示例为问候功能定义的接口，你可以换成自己的服务。应用启动时创建服务并连接数据库。退出时，再由应用关闭自己打开的资源。

## 并发与生命周期

查询数据库或调用外部服务时，把 `ctx.RequestContext()` 传下去。如果某次操作有自己的时限，可以在这个上下文上加超时。需要长期运行的后台任务由应用负责启动和停止，参见[生命周期](../getting-started/lifecycle.md)。

## 幂等与持久化

WebSocket 恢复后可能重复投递消息。涉及订单、扣款或数据库写入时，可用机器人、场景、会话和消息标识组成幂等键，并保存处理状态。再次收到同一条消息时，先检查对应操作是否已完成。

- 素材与业务关联时，保存素材 ID、URL 和业务键。字段见[素材管理](../server-api/messages/material-management.md)。
- 流式消息由应用保存游标和状态。用法见[单聊流式消息](../server-api/messages/streaming.md)。

## 注意事项

- SDK 可能同时调用多个处理函数。服务里的 map、缓存和素材引用如果会被一起读写，需要加锁或用其他方式同步。
- `Context` 含有原子状态。首次使用后应通过原指针传递，避免复制整个结构。
- 后台任务只保存自己用得上的标识和数据副本。整个事件的 `Context` 留在处理函数里。
- 展示引用索引 `REFIDX` 和 `TMP_` 并非稳定的全局消息 ID。
- 应用选择数据库和清理策略。SDK 不会在失败后自动重试 POST。

## 另请参阅

[消息接收与富文本](../server-api/messages/rich-content.md) · [生命周期](../getting-started/lifecycle.md) · [原始 API 扩展](raw-api.md)
