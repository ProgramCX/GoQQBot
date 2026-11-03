# WebSocket 接入

复制 `config.websocket.example.yaml` 为自己的配置文件，填写 `bot.appId` 和
`bot.appSecret`，按已开通权限选择 `connection.websocket.intents`。配置凭证的文件不要提交到版本库。

```sh
go run ./examples/echo -config config.yaml
```

群聊 @机器人发送 `/hello`，或在单聊发送 `/hello`，即可触发示例 handler。
WebSocket 模式不启动 HTTP 监听，不需要公网回调地址或 `server` 配置。
两种模式共用现有的 `OnCommand` / `OnEvent` 和 API 客户端；发消息仍通过 HTTP OpenAPI。

## 配置和迁移

- `connection.mode` 支持 `websocket`、`webhook`，未配置时默认 `webhook`。
- `bot.token` 已弃用且不再必填。SDK 根据 AppID/AppSecret 获取 Access Token。
- `runtime.pool.workers` 默认 10，`queueSize` 默认 100。
  新配置未设置的池字段会回退到旧 `server.pool`，然后应用默认值。
- `runtime.shutdownTimeout` 默认 15s。
- WebSocket 的默认 intents 为 `[GROUP_AND_C2C_EVENT]`。名称必须是官方 intent 名称；
  多项订阅由 SDK 合并位掩码，所需权限需在平台开通。
- `dialTimeout` 默认 10s，也用于 Hello、登录及写入超时。
- 重连采用随机抖动和指数退避，默认从 1s 增长到最多 30s，直到取消。
- `shard.id` 默认 0，`shard.total` 默认 1。每个 Bot 只建立一个分片连接。
  手动部署多个实例时，须配置相同 total 和不同 id，覆盖所需分片；本版本不自动管理多分片。
- 网关地址动态发现，心跳间隔由服务端下发，用户无需填写。
- 可选配置的零值表示使用默认值；负数和不合法的分片、intent、模式会在构造时返回错误。

完整配置见根目录的两份 `config.*.example.yaml`。

## 生命周期

`NewBot(path, handler)` 从文件创建；嵌入其他程序时可以调用
`NewWithConfig(cfg, handler, options...)`。构造器不连接网络、不启动线程池。
配置被复制，构造后修改原配置不会影响实例。

```go
bot, err := goqqrobot.NewWithConfig(cfg, nil)
if err != nil {
    return err
}
// 在启动前注册 handlers。
return bot.Run(ctx)
```

`Run(ctx)` 阻塞运行，调用方取消时返回 `ctx.Err()`。
`Start()` 使用后台 context；`StartWithGracefulShutdown()` 处理进程退出信号。
网络中断会重连；不可恢复的登录错误返回给调用方，启动失败不会继续等待退出信号。

`Shutdown(ctx)` 可并发、可重复调用；它取消连接和 handler context，限时等待任务结束。
`Stop()` 使用配置的关闭时限。关闭前未启动的实例也会被永久关闭；一个 Bot 只能运行一次。
Shutdown 返回超时后，不响应取消的用户 handler 可能仍在执行；Go 无法强制终止它。
handler 可通过 `ctx.RequestContext()` 将取消信号传给下游操作，SDK 消息 API 已使用该 context。

## 事件恢复与过载

WebSocket 维护接收序号、会话 ID 和按投递顺序推进的处理进度。
并发回调乱序完成时，恢复进度不会跳过未结束的前序任务；心跳独立使用最新接收序号。
恢复窗口最多保存 `workers + queueSize` 条尚未推进的事件记录。

队列或恢复窗口满时关闭连接并退避重连，用安全进度请求 Resume；不会把未接受的事件标为处理完成。
Webhook 模式入队失败返回 503，成功入队返回协议 ACK。
任务正常返回、返回错误或 panic 均视为一次处理尝试结束；SDK 不自动重试业务失败。
业务事件并发处理，不保证完成顺序；同一业务需要串行时应由 handler 协调。

会话进度只保存在内存；重连补发可能重复处理，已完成但尚未推进连续进度的事件也可能再次收到。
应用应对扣款、写入等副作用做幂等处理。平台会话过期、进程重启或不可恢复会话后的历史消息完整性
不由本 SDK 保证；本版本没有持久化队列或跨进程恢复功能。

## 内部维护约定

连接资源及 Hello 校验位于 `internal/adapter/websocket.go`，协议循环、认证、心跳及事件派发位于
`websocket_connection.go`；关闭码及错误分类集中在 `websocket_errors.go`，会话和投递水位线在
`websocket_session.go`。这些均是内部实现，不改变调用方入口。

连接状态显式区分 Identify、Resume、等待拒绝关闭原因及 Ready。Identify 收到 op 9 后进入等待状态，
最多等待 `min(2s, dialTimeout)` 获取具体关闭码；期间不再发心跳或接受迟到的 Ready。
心跳 ACK 使用独立状态；全部协议转换和写入由单个连接循环执行。

投递窗口将容量检查与登记放在同一把锁下，包括不含序号的投递。业务完成回调可并发、重复调用，
恢复水位线仅按投递顺序推进。每次重连替换水位线对象，旧连接的迟到回调不会推进新连接的进度。
写入错误通过 `%w` 保留根因；拨号错误与服务端关闭文本仍避免直接回显敏感内容。

## 协议与依赖

- QQ 官方协议：https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html
- WebSocket 库：`github.com/gorilla/websocket v1.5.3`，BSD-2-Clause，无额外运行时依赖。
- 单元测试采用本地 HTTP/WebSocket 替身，不需要真实机器人凭证。
