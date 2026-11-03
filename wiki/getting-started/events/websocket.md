# 通过 WebSocket 接收事件

WebSocket 模式由 SDK 动态发现网关，完成 Hello、Identify/Resume、心跳及重连。它不启动 HTTP 监听，不需要 `server` 配置。

## 配置连接

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

配合[快速开始](../quick-start.md)中的程序运行；不需要自己发送心跳或缓存 Access Token。`INTERACTION` 用于需要回调的交互，普通命令按钮的预填并不等同于交互回调。

## 重连与会话恢复

临时连接故障使用带随机抖动的指数退避。成功登录后恢复初始退避；取消上下文会结束连接、读取协程及等待。不可恢复的认证、分片、订阅权限或 IP 限制错误直接返回，不无限重连。

SDK 区分最新接收序号与业务完成水位线：心跳使用前者，Resume 使用按投递顺序连续完成的后者。并发处理函数乱序结束，不会让恢复位置跳过尚未结束的前序任务。重连时旧完成回调与新连接的水位线隔离。

Identify 被 op 9 拒绝时，SDK 最多等待 `min(2s, dialTimeout)` 获取服务端 close frame 的具体原因，避免把 IP 白名单限制简单当成凭证错误。没有原因时也会按时退出。

## 并发与过载

每个实例只建立一个分片连接。接收消息帧上限为 **4 MiB**。恢复窗口最多保存 `workers + queueSize` 条尚未推进的投递记录。

队列或恢复窗口满时关闭连接、退避并用安全进度恢复，不把未接受的事件当作处理完成。处理函数正常返回、返回错误或 panic 都表示一次处理尝试结束；业务失败不会由 SDK 自动重试。

会话及进度只在内存中，重放可能重复投递。已完成但尚未推进连续水位线的事件也可能再次出现；需要业务幂等。进程重启或平台会话失效后的历史完整性不由 SDK 保证。

`shard.total` 多实例部署时必须一致，各实例使用不同 `id`；SDK 不自动启动或调度全部分片。

## 另请参阅

- [配置及已支持的 intents](../configuration.md)
- [运行与关闭](../lifecycle.md)
- [错误处理](../../logging/errors.md)
- [官方连接维护、Identify、Resume 及心跳](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
