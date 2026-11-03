# 事件订阅与通知

QQ 通过事件通知消息、交互和机器人生命周期变化。SDK 将 WebSocket 与 Webhook 接收的事件送入同一套路由和有界线程池，业务处理函数不需要判断底层传输方式。

## 选择接入方式

| 场景 | 选择 | 需要准备 |
| --- | --- | --- |
| 本地开发、没有公网回调地址 | [WebSocket](websocket.md) | 出站网络、平台准入及对应 intents |
| 已部署公网 HTTPS 服务 | [Webhook](webhook.md) | 回调地址、平台订阅及 HTTPS 转发 |

配置中的 `connection.mode` 二选一。WebSocket 配置里的 intents 不会替代 Webhook 管理端的事件订阅；SDK 接收后也不能补回平台未投递的事件。

## 选对处理入口

| 需求 | 入口 |
| --- | --- |
| `/hello` 这类命令 | `OnCommand("hello", eventType, handler)` |
| 指定类型的普通消息或附件 | `OnMessage(eventType, handler)` |
| 所有已知消息类型兜底 | `OnAnyMessage(handler)` |
| 生命周期等原始事件 | `OnEvent(eventType, handler)` |
| 按钮、菜单、其他交互 | `OnButton` / `OnMenu` / `OnInteraction` |

`OnMessage` 是命令之后的消息兜底，不是每条消息都会执行的观察器；`OnAnyMessage` 也不包括交互或生命周期事件。

## 事件 ID 与类型

`ctx.Payload.T` 是事件类型，`ctx.Payload.ID` 是事件信封 ID，`ctx.Payload.D` 是事件数据。普通消息的 `d.id`、引用索引和交互确认用的 `d.id` 都有各自含义，不要混用。

原始载荷可能包含授权信息。读取消息内容优先使用 `ctx.Message()`，读取交互使用 `ctx.Interaction()`，不要把整个 Payload 直接打进日志。

## 另请参阅

- [事件路由与非命令消息](routing.md)
- [消息收发](../../server-api/messages/README.md)
- [消息交互](../../server-api/messages/interactions/README.md)
- [官方事件结构及订阅](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
