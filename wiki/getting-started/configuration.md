# 配置机器人

机器人独立运行时，可以用 YAML 文件配置。已有业务系统可以从环境变量或密钥管理服务读取配置，再直接传给 SDK。两种方式使用同一套默认值和校验规则。

## 文件配置

用 `goqqrobot.NewBot("config.yaml", nil, options...)` 加载配置。

```yaml
bot:
  appId: "your-app-id"
  appSecret: "your-app-secret"

connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
      - INTERACTION
    dialTimeout: 10s
    reconnect:
      initialDelay: 1s
      maxDelay: 30s
    shard:
      id: 0
      total: 1

runtime:
  pool:
    workers: 10
    queueSize: 100
  shutdownTimeout: 15s
```

传入相对路径时，SDK 会从进程的工作目录查找文件。

## 使用内存配置

在业务代码里填好配置，再用配置创建 Bot。

```go
package guide

import (
	"os"

	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
)

func NewBotFromEnvironment() (*goqqrobot.Bot, error) {
	cfg := config.Config{
		Bot: config.Bot{
			AppId:     os.Getenv("QQ_BOT_APP_ID"),
			AppSecret: os.Getenv("QQ_BOT_APP_SECRET"),
		},
		Connection: config.Connection{
			Mode: "websocket",
			WebSocket: config.WebSocket{
				Intents: []string{"GROUP_AND_C2C_EVENT", "INTERACTION"},
			},
		},
	}
	return goqqrobot.NewWithConfig(cfg, nil)
}
```

SDK 会复制你传入的配置和订阅列表。Bot 创建后，你可以修改手里的配置，Bot 仍会使用原来的值。

## 默认值

下表列出 SDK 在可选字段未设置时采用的值。

| 配置 | 默认值 | 作用 |
| --- | --- | --- |
| `connection.mode` | `webhook` | 设为 `websocket` 时连接网关 |
| `connection.websocket.intents` | `[GROUP_AND_C2C_EVENT]` | 决定网关推送哪些事件 |
| `dialTimeout` | `10s` | 建立连接、等 Hello、登录和写入时使用的超时 |
| `reconnect.initialDelay` / `maxDelay` | `1s` / `30s` | 重连等待时间的初始值和上限。SDK 会加上一点随机延迟 |
| `shard.id` / `total` | `0` / `1` | 每个 Bot 只连接一个分片 |
| `runtime.pool.workers` | `10` | 可同时运行的处理函数数量 |
| `runtime.pool.queueSize` | `100` | 待处理队列容量 |
| `runtime.shutdownTimeout` | `15s` | SDK 等待关闭的时限 |
| `server.host` / `port` | `0.0.0.0` / `8080` | 仅 Webhook 使用 |

## 可配置的 WebSocket 订阅

在配置里填写订阅名称，SDK 会用 `IntentMask` 转成网关需要的值。

- `GUILDS`、`GUILD_MEMBERS`、`GUILD_MESSAGES`、`GUILD_MESSAGE_REACTIONS`。
- `DIRECT_MESSAGE`、`GROUP_MEMBER_EVENT`、`GROUP_AND_C2C_EVENT`、`INTERACTION`。
- `MESSAGE_AUDIT`、`FORUMS_EVENT`、`AUDIO_ACTION`、`PUBLIC_GUILD_MESSAGES`。
- `OPEN_FORUM_EVENT`、`AUDIO_OR_LIVE_CHANNEL_MEMBER`。

## 注意事项

配置前可查[全部 Intent 与下属事件](events/intents.md)，接入选择见[连接方式对比](events/connection.md)。

- 配置文件路径由调用方提供。YAML 中的 `${ENV_VAR}` 会按原样读取。
- 缺少凭证、容量是负数或连接模式写错时，创建 Bot 会失败。分片设置无效或 WebSocket 订阅名称未知时，也会返回错误。
- 可选字段填 0 时会使用默认值。例如 `queueSize: 0` 会用默认队列容量。
- 对应的 `runtime.pool` 字段未设置时，SDK 才会读取旧的 `server.pool`。
- SDK 会忽略 `bot.token`。凭证填写方式见[获取访问凭证](credentials.md)。
- 订阅名称区分大小写。WebSocket 中使用 `GROUP_MEMBER_EVENT` 订阅成员变更和入群申请。Webhook 的订阅在 QQ 平台配置。
- 配置校验通过后，事件权限仍由 QQ 平台检查。SDK 有时能收到事件，但还没有为它定义专门的 Go 类型。

## 另请参阅

[WebSocket](events/websocket.md) · [Webhook](events/webhook.md) · [运行与关闭](lifecycle.md) · [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
