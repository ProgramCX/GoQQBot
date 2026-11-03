# 配置机器人

文件配置适合独立程序；内存配置适合已有业务系统、环境变量和密钥管理服务。两者最终使用同一套默认值和校验规则。

## 文件配置

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

通过 `goqqrobot.NewBot("config.yaml", nil, options...)` 加载。相对路径依据进程的当前工作目录解析；SDK 不自动搜索配置，也不展开 `${ENV_VAR}`。

## 使用内存配置

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

构造器复制配置及订阅切片；构造后修改原配置不影响机器人。凭证缺失、负数容量、未知模式、非法分片和未知 WebSocket intent 在构造时返回错误。

## 默认值

| 配置 | 默认值 | 作用 |
| --- | --- | --- |
| `connection.mode` | `webhook` | 必须显式选择 WebSocket 才启用网关 |
| `connection.websocket.intents` | `[GROUP_AND_C2C_EVENT]` | 不是自动授予平台权限 |
| `dialTimeout` | `10s` | 拨号、Hello、登录及写入超时 |
| `reconnect.initialDelay` / `maxDelay` | `1s` / `30s` | 重连退避，内部加入随机抖动 |
| `shard.id` / `total` | `0` / `1` | 一个实例仅建立一个分片连接 |
| `runtime.pool.workers` | `10` | 并发处理函数数量 |
| `runtime.pool.queueSize` | `100` | 待处理队列容量 |
| `runtime.shutdownTimeout` | `15s` | 内部关闭等待时间 |
| `server.host` / `port` | `0.0.0.0` / `8080` | 仅 Webhook 使用 |

可选字段为零表示使用默认值，不能通过设置 `queueSize: 0` 创建零容量队列。旧 `server.pool` 仅在对应 `runtime.pool` 字段未设置时作为回退。`bot.token` 被忽略。

## 可配置的 WebSocket 订阅

当前 `IntentMask` 接受：`GUILDS`、`GUILD_MEMBERS`、`GUILD_MESSAGES`、`GUILD_MESSAGE_REACTIONS`、`DIRECT_MESSAGE`、`GROUP_AND_C2C_EVENT`、`INTERACTION`、`MESSAGE_AUDIT`、`FORUMS_EVENT`、`AUDIO_ACTION`、`PUBLIC_GUILD_MESSAGES`。

名称区分大小写。平台新增的 `GROUP_MEMBER_EVENT` 当前尚未加入该映射，不能直接填入 WebSocket 配置；这不等于 Webhook 无法接收平台已订阅的原始事件。订阅名称可接受也不代表每种事件均已有类型化业务模型。

## 另请参阅

- [WebSocket](events/websocket.md)与 [Webhook](events/webhook.md)
- [运行与关闭](lifecycle.md)
- [官方事件订阅与通知](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
