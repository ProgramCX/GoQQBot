# 通过 WebSocket 接收事件

WebSocket 适合通过网关接收事件。SDK 动态发现网关，处理 Hello、Identify/Resume、心跳和重连。该模式无需配置 `server`，也不会启动 HTTP 监听。

## 配置连接

在配置中选择 `websocket` 并填写事件 intents。

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
```

配合[快速开始](../quick-start.md)中的程序运行。SDK 发送心跳并缓存 Access Token。订阅 `INTERACTION` 可接收[交互回调](../../server-api/messages/interactions/button-callbacks.md)。普通命令按钮只预填时不会触发交互回调。

## 重连与会话恢复

临时连接故障使用带随机抖动的指数退避。

- 登录成功后，退避恢复到初始值。
- 取消上下文会结束连接、读取协程及等待。

心跳使用最新接收序号。Resume 使用业务完成水位线（watermark）。水位线按投递顺序推进，只越过连续完成的任务。回调乱序完成时，恢复位置仍会保留未结束的前序任务。旧连接的完成回调不会推进新连接的水位线。



## 并发与过载

恢复窗口最多保存 `workers + queueSize` 条尚未推进的投递记录。

队列或恢复窗口满时，SDK 关闭连接并退避。重连使用安全进度。未接受的事件不会标记为处理完成。处理函数正常返回、返回错误或 panic，都算一次处理尝试结束。





## 注意事项

- 不可恢复的认证、分片、订阅权限或 IP 限制错误会直接返回。SDK 不会继续重连。
- Identify 收到 `op 9` 后，SDK 等待服务端关闭帧。等待上限取 `2s` 和 `dialTimeout` 的较小值。具体原因可区分 IP 白名单限制与凭证错误。收不到原因也会按时退出。
- SDK 将接收消息帧限制为 **4 MiB**。SDK 不自动重试业务失败。
- 会话和进度仅保存在内存中。重放可能重复投递。已完成但水位线未推进的事件也可能再出现。业务需保证幂等。进程重启或平台会话失效后，SDK 无法保证历史消息完整。
- 每个实例只建立一个分片连接。多实例使用相同的 `shard.total`，并分配不同的 `id`。SDK 不自动启动或调度全部分片。

## 另请参阅

[配置及已支持的 intents](../configuration.md) · [运行与关闭](../lifecycle.md) · [错误处理](../../logging/errors.md) · [官方连接维护、Identify、Resume 及心跳](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
