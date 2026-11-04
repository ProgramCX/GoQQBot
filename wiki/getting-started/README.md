# 起步指南

GoQQBot 是 QQ 官方机器人的 Go SDK。机器人通过 WebSocket 或 Webhook 接收事件，通过 QQ 的 HTTPS OpenAPI 发送消息。

第一次用 GoQQBot，可以从安装和第一个机器人开始。

1. [安装 SDK](installation.md)。检查 Go 版本，添加模块依赖或接入本地工作区。
2. [快速开始](quick-start.md)。创建机器人，收到 `/hello` 后回复。
3. [获取访问凭证](credentials.md)。准备 AppID、AppSecret 和 Access Token。
4. [配置文件](configuration.md)。设置连接、订阅、并发和关闭超时。
5. [事件订阅与通知](events/README.md)。选择传输方式，注册处理函数。
6. [API 调用指南](api-calls.md)。在事件处理函数之外调用 API，设置发送目标和请求超时。
7. [运行与关闭](lifecycle.md)。把机器人接入已有服务，并随服务一起关闭。

新项目优先使用 `goqqrobot.NewWithConfig` 或 `NewBot`，搭配 `OnMessage` 和 `ctx.Reply`。`pkg/api/group`、`pkg/api/c2c` 等旧接口仍可用。

## 最小使用路径

先创建 Bot，再注册处理函数，最后启动。

```text
创建配置 → 构造 Bot → 注册处理函数 → Run(ctx) → 取消 / Shutdown(ctx)
```

创建 Bot 时，SDK 只检查配置并准备好组件。启动后才会联网和启动工作协程。

## 注意事项

- 同一个机器人实例只选一种事件接入方式。
- 回复同一条消息时，旧接口和新接口选一套用到底。新接口会自动分配序号。
- 授权、事件订阅和消息审核由 QQ 开放平台管理。需要开通权限时，到平台操作。
- 各功能页分别标明 SDK 支持的接口和在 QQ 平台上验收的范围。

## 另请参阅

[日志与错误处理](../logging/README.md) · [消息收发](../server-api/messages/README.md) · [官方启动接入](https://bot.q.qq.com/wiki/develop/api-v2/)

文档按当前工作区的 API 编写，最后核对于 2026-11-04。
