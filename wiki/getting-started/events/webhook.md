# 通过 Webhook 接收事件

Webhook 适合有公网 HTTPS 回调地址的业务。QQ 会向 **`POST /webhook`** 发起回调。SDK 负责响应地址验证，并在普通事件通过 Ed25519 验签、进入队列后向 QQ 确认。

## 配置监听

配置 `webhook` 模式后，SDK 会启动 HTTP 服务。

```yaml
bot:
  appId: "your-app-id"
  appSecret: "your-app-secret"

connection:
  mode: webhook

server:
  host: "0.0.0.0"
  port: 8080

runtime:
  pool:
    workers: 10
    queueSize: 100
  shutdownTimeout: 15s
```

沿用[快速开始](../quick-start.md)的 Go 程序，替换配置。在平台管理端填入公网回调地址，如 `https://bot.example.com/webhook`。选择要接收的[事件](README.md)。

    QQ 平台 --HTTPS--> 反向代理 --HTTP--> POST /webhook



## 验证及响应

SDK 自动响应 `op=13` 回调地址验证。普通 `op=0` 事件通过 Ed25519 验签后入队。

| 情况 | 当前 HTTP 行为 |
| --- | --- |
| 非 POST | `405` |
| 超过 4 MiB、读取失败或 JSON 非法 | `400` |
| 回调地址验证失败 | `403` |
| 普通事件签名不合法 | `401` |
| 普通回调不是支持的事件操作 | `400` |
| 队列已满或停止接收 | `503` |
| 成功入队 | `200`，正文 `{"op":12}` |





## 注意事项

回调在 QQ 开放平台订阅。消息和事件名称可查[事件对照表](intents.md)，接入差异见[连接方式对比](connection.md)。

- SDK 的监听端口使用明文 HTTP。由反向代理或负载均衡接收公网 HTTPS 请求，再将原始请求正文和签名请求头转发给 SDK。
- 回调路径固定为 `/webhook`，公开配置不提供路径选项。SDK 没有内置健康检查路由，根路径也接收不到回调。
- 确认响应（ACK）只表示 SDK 已将事件加入队列。业务函数随后由共享的工作协程执行。函数返回错误或发生 panic 时，可通过[错误处理入口](../../logging/errors.md)获知，QQ 已收到的确认不会改变。
- SDK 为 HTTP 读取和写入各设 5 秒超时。关闭等待时长取自 `runtime.shutdownTimeout`。
- SDK 没有跨进程去重存储。业务需自行处理重复投递。

## 另请参阅

[事件路由](routing.md) · [运行与关闭](../lifecycle.md) · [官方 Webhook 回调及验证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html) · [官方 Ed25519 签名算法](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html)
