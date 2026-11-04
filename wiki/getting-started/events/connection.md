# 选择 WebSocket 或 Webhook

本地开发或没有公网 HTTPS 入口时，通常使用 WebSocket。已有可接收公网 HTTPS 回调的服务时，可以使用 Webhook。两种方式都可复用 SDK 的事件处理函数。

| 比较项 | WebSocket | Webhook |
| --- | --- | --- |
| 连接方向 | SDK 主动连接 QQ 网关 | QQ 向应用发送 HTTPS POST |
| 选择方式 | `connection.mode: websocket` | `connection.mode: webhook` |
| 事件订阅 | 配置 `connection.websocket.intents`，登录时提交 | 在 QQ 开放平台选择事件 |
| 网络要求 | 应用可访问 QQ 的 HTTPS 和 WSS 网关 | QQ 可访问公网回调地址，入口将请求转发到 SDK |
| 连接维护 | SDK 处理登录、心跳和重连，平台允许时恢复会话 | SDK 响应地址验证和签名校验，事件成功入队后确认接收 |

## 启动 WebSocket

在 QQ 机器人运营与管理后台打开 **开发设置 → 事件订阅与回调**，选择 **WebSocket**。在本地配置凭证、连接模式和所需 Intent，再启动程序。遇到登录拒绝时，先看 SDK 输出的网关错误，核对应用权限及订阅项。[运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [WebSocket 配置](websocket.md) · [Intent 对照表](intents.md)

## 启动 Webhook

配置凭证、连接模式和监听地址，再启动程序。用公网 HTTPS 入口将 `POST /webhook` 转发给 SDK，并保留原始请求正文和签名请求头。

在 QQ 机器人运营与管理后台打开 **开发设置 → 事件订阅与回调**，选择 **Webhook**，填写公网 HTTPS 回调地址并完成验证。在 **接收事件配置** 选择 **部分接收** 时，逐项勾选功能页列出的事件。选择 **全部接收** 时，无需逐项勾选。[运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [Webhook 配置](webhook.md)

## 注意事项

- WebSocket 重连可能补发事件。会话进度保存在内存中，进程重启或会话失效后不能保证补齐。
- Webhook 的确认表示事件已入队。后续处理失败需要由应用记录和处理。
- 涉及写入、审批等操作时，应按事件或业务标识去重。
- API 调用权限和事件订阅权限分别由 QQ 检查。能收到事件不代表能调用对应管理接口。

## 另请参阅

[官方连接与事件订阅说明](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html) · [事件入口](README.md) · [运行与关闭](../lifecycle.md)
