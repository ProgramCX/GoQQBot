# 通过 Webhook 接收事件

Webhook 适合有公网 HTTPS 回调地址的业务。SDK 监听 **`POST /webhook`**。SDK 验证回调地址、校验普通事件的 Ed25519 签名，并确认事件入队。

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

- SDK 监听明文 HTTP。反向代理或负载均衡负责终止 HTTPS。转发时保留原始正文和签名请求头。
- 回调路径固定为 `/webhook`，公开配置不提供路径选项。SDK 没有内置健康检查路由，根路径也接收不到回调。
- 确认响应（ACK）只表示事件已入队。共享线程池异步执行业务函数。错误或 panic 由[错误观察入口](../../logging/errors.md)报告。已返回的 ACK 不会改变。
- SDK 为 HTTP 读取和写入各设 5 秒超时。关闭等待时长取自 `runtime.shutdownTimeout`。
- SDK 没有跨进程去重存储。业务需自行处理重复投递。

## 另请参阅

[事件路由](routing.md) · [运行与关闭](../lifecycle.md) · [官方 Webhook 回调及验证](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html) · [官方 Ed25519 签名算法](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html)
