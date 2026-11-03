# 起步指南

GoQQBot 是用于 QQ 官方机器人的 Go SDK。接收事件可以选择 WebSocket 或 Webhook，发送消息仍调用 QQ 的 HTTPS OpenAPI。同一个机器人实例只选择一种事件接入方式。

第一次使用，按以下顺序阅读：

1. [安装 SDK](installation.md)：Go 版本、模块依赖和本地开发。
2. [快速开始](quick-start.md)：创建机器人，收到 `/hello` 后回复。
3. [获取访问凭证](credentials.md)：AppID、AppSecret 与 Access Token。
4. [配置文件](configuration.md)：连接方式、订阅、并发与关闭超时。
5. [事件订阅与通知](events/README.md)：选择传输方式并注册处理函数。
6. [API 调用指南](api-calls.md)：独立客户端、超时、显式目标及扩展请求。
7. [运行与关闭](lifecycle.md)：嵌入业务服务并管理资源。

新项目优先使用 `goqqrobot.NewWithConfig` / `NewBot`、`OnMessage` 和 `ctx.Reply`。`pkg/api/group`、`pkg/api/c2c` 等旧接口仍可用，但不要在同一消息上混用旧回复接口与新的自动序号接口。

## 最小使用路径

```text
创建配置 → 构造 Bot → 注册处理函数 → Run(ctx) → 取消 / Shutdown(ctx)
```

构造器不会联网或启动工作协程。平台授权、订阅和消息审核必须在 QQ 开放平台完成；SDK 无法通过配置授予权限。

## 另请参阅

- [日志与错误处理](../logging/README.md)
- [消息收发](../server-api/messages/README.md)
- [官方启动接入](https://bot.q.qq.com/wiki/develop/api-v2/)

文档依据：当前工作区 API，核对日期为 2026-11-04。实现覆盖与真实平台验收是两件事，具体边界见各功能页。
