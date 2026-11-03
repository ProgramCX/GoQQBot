# 通过 Webhook 接收事件

Webhook 模式适合有公网 HTTPS 回调地址的业务。SDK 内置 HTTP 服务监听 **`POST /webhook`**，负责回调地址验证、普通事件 Ed25519 验签及入队确认。

## 配置监听

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

使用[快速开始](../quick-start.md)同一份 Go 程序，替换配置即可。平台管理端填写对外地址，例如 `https://bot.example.com/webhook`，并选择要接收的事件。

SDK 监听明文 HTTP；通过反向代理或负载均衡终止 HTTPS，并保留原始请求正文及签名请求头。当前没有公开配置来更改回调路径，也没有内置健康检查路由；不要把根路径当作回调地址。

## 验证及响应

平台的 `op=13` 回调地址验证由 SDK 自动响应。普通事件需通过 Ed25519 签名验证，再接受 `op=0` 派发。

| 情况 | 当前 HTTP 行为 |
| --- | --- |
| 非 POST | `405` |
| 超过 4 MiB、读取失败或 JSON 非法 | `400` |
| 回调地址验证失败 | `403` |
| 普通事件签名不合法 | `401` |
| 普通回调不是支持的事件操作 | `400` |
| 队列已满或停止接收 | `503` |
| 成功入队 | `200`，正文 `{"op":12}` |

成功入队的 ACK 不代表业务已经执行成功。业务函数由共享线程池异步执行；其错误或 panic 通过[错误观察入口](../../logging/errors.md)报告，而不是改写已经返回的平台 ACK。

HTTP 读写超时各为 5 秒；关闭时使用 `runtime.shutdownTimeout` 等待。平台重复投递需要业务自行幂等；SDK 没有跨进程事件去重存储。

## 另请参阅

- [事件路由](routing.md)
- [运行与关闭](../lifecycle.md)
- [官方 Webhook 回调及验证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
- [官方 Ed25519 签名算法](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html)
