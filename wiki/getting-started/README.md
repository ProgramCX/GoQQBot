# 起步指南

GoQQBot 是 QQ 官方机器人的 Go SDK。用 WebSocket 或 Webhook 接收事件，用 QQ 的 HTTPS OpenAPI 发送消息。

第一次使用可按以下顺序阅读。

1. [安装 SDK](installation.md)。检查 Go 版本，添加模块依赖或接入本地工作区。
2. [快速开始](quick-start.md)。创建机器人，收到 `/hello` 后回复。
3. [获取访问凭证](credentials.md)。准备 AppID、AppSecret 和 Access Token。
4. [配置文件](configuration.md)。设置连接、订阅、并发和关闭超时。
5. [事件订阅与通知](events/README.md)。选择传输方式，注册处理函数。
6. [API 调用指南](api-calls.md)。设置独立客户端、超时和显式目标，扩展请求。
7. [运行与关闭](lifecycle.md)。接入业务服务，管理资源。

新项目优先使用 `goqqrobot.NewWithConfig` 或 `NewBot`，搭配 `OnMessage` 和 `ctx.Reply`。`pkg/api/group`、`pkg/api/c2c` 等旧接口仍可用。

## 最小使用路径

依次创建机器人、注册处理函数并启动。

```text
创建配置 → 构造 Bot → 注册处理函数 → Run(ctx) → 取消 / Shutdown(ctx)
```

构造器不联网，也不启动工作协程。

## 注意事项

- 同一个机器人实例只选一种事件接入方式。
- 同一消息的回复应统一使用旧接口或新接口。新接口会自动分配序号。
- 平台授权、订阅和消息审核由 QQ 开放平台处理。修改 SDK 配置无法授予权限。
- 实现覆盖与真实平台验收分别核对，具体边界见各功能页。

## 另请参阅

[日志与错误处理](../logging/README.md) · [消息收发](../server-api/messages/README.md) · [官方启动接入](https://bot.q.qq.com/wiki/develop/api-v2/)

文档依据为当前工作区的 API，核对日期为 2026-11-04。
