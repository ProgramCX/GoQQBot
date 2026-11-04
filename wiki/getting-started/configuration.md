# 配置机器人

独立程序可用文件配置。已有业务系统可用内存配置，接入环境变量或密钥管理服务。两种方式共用默认值和校验规则。

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

相对路径从进程的当前工作目录算起。

## 使用内存配置

在业务代码中构造配置，再创建机器人。

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

构造器会复制配置和订阅切片。创建机器人后，修改原配置不影响实例。

## 默认值

下表列出 SDK 在可选字段未设置时采用的值。

| 配置 | 默认值 | 作用 |
| --- | --- | --- |
| `connection.mode` | `webhook` | 选择 WebSocket 才启用网关 |
| `connection.websocket.intents` | `[GROUP_AND_C2C_EVENT]` | 设置订阅范围 |
| `dialTimeout` | `10s` | 拨号、Hello、登录及写入超时 |
| `reconnect.initialDelay` / `maxDelay` | `1s` / `30s` | 重连退避，内部加入随机抖动 |
| `shard.id` / `total` | `0` / `1` | 一个实例仅建立一个分片连接 |
| `runtime.pool.workers` | `10` | 并发处理函数数量 |
| `runtime.pool.queueSize` | `100` | 待处理队列容量 |
| `runtime.shutdownTimeout` | `15s` | 内部关闭等待时间 |
| `server.host` / `port` | `0.0.0.0` / `8080` | 仅 Webhook 使用 |

## 可配置的 WebSocket 订阅

`IntentMask` 接受以下订阅名称。

- `GUILDS`、`GUILD_MEMBERS`、`GUILD_MESSAGES`、`GUILD_MESSAGE_REACTIONS`。
- `DIRECT_MESSAGE`、`GROUP_AND_C2C_EVENT`、`INTERACTION`。
- `MESSAGE_AUDIT`、`FORUMS_EVENT`、`AUDIO_ACTION`、`PUBLIC_GUILD_MESSAGES`。

## 注意事项

- SDK 不搜索配置文件，也不展开 `${ENV_VAR}`。
- 凭证缺失、负数容量和未知模式会使构造失败。非法分片和未知 WebSocket intent 也会返回错误。
- 可选字段为零时使用默认值。`queueSize: 0` 不会创建零容量队列。
- 对应的 `runtime.pool` 字段未设置时，SDK 才会读取旧的 `server.pool`。
- SDK 会忽略 `bot.token`。凭证填写方式见[获取访问凭证](credentials.md)。
- 订阅名称区分大小写。当前映射缺少 `GROUP_MEMBER_EVENT`，填入 WebSocket 配置会失败。Webhook 仍可接收平台已订阅的原始事件。
- SDK 接受订阅名称不代表平台已授权。对应事件也可能尚无类型化模型。

## 另请参阅

[WebSocket](events/websocket.md) · [Webhook](events/webhook.md) · [运行与关闭](lifecycle.md) · [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
