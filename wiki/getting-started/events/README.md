# 事件订阅与通知

QQ 用事件通知消息、交互和机器人生命周期变化。WebSocket 和 Webhook 共用一套事件路由，由数量有限的工作协程处理。业务函数可以同时用于两种接入方式。

## 选择接入方式

通过 `connection.mode` 选择一种接入方式。

| 场景 | 选择 | 需要准备 |
| --- | --- | --- |
| 本地开发、没有公网回调地址 | [WebSocket](websocket.md) | 出站网络、平台准入及对应 intents |
| 已部署公网 HTTPS 服务 | [Webhook](webhook.md) | 回调地址、平台订阅及 HTTPS 转发 |

## 选对处理入口

根据事件类型和任务选择处理入口。优先级见[路由顺序](routing.md)。

| 需求 | 入口 |
| --- | --- |
| `/hello` 这类命令 | `OnCommand("hello", eventType, handler)` |
| 指定类型的普通消息或附件 | `OnMessage(eventType, handler)` |
| 处理所有已知类型中未匹配的消息 | `OnAnyMessage(handler)` |
| 生命周期等原始事件 | `OnEvent(eventType, handler)` |
| 按钮、菜单、其他交互 | `OnButton` / `OnMenu` / `OnInteraction` |

## 事件 ID 与类型

按用途读取事件字段。

- `ctx.Payload.T` 是事件类型。
- `ctx.Payload.ID` 是外层事件对象的 ID。
- `ctx.Payload.D` 是事件数据。

用 `ctx.Message()` 读消息，用 `ctx.Interaction()` 读交互。

## 注意事项

- WebSocket 的 intents 不会替代 Webhook 管理端的订阅。SDK 无法补回平台未投递的事件。
- 命令优先处理。`OnMessage` 只接收未匹配命令的消息。
- `OnAnyMessage` 不处理交互或生命周期事件。
- 普通消息的 `d.id`、引用索引和交互确认用的 `d.id` 用途不同。分别按对应接口取值。
- 原始载荷可能包含授权信息。记录整个 Payload 可能泄露这些信息。

## 另请参阅

[事件路由与非命令消息](routing.md) · [消息收发](../../server-api/messages/README.md) · [消息交互](../../server-api/messages/interactions/README.md) · [官方事件结构及订阅](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html)
