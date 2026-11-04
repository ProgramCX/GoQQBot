# 使用消息交互

按钮和可点击指令可以触发业务操作。按点击后的行为选择处理入口。

| 需要实现的行为 | 构造入口 | 接收入口 |
| --- | --- | --- |
| 点击后预填命令 | `button.Command` | 用户发送后进入 `OnCommand` |
| 单聊点击直接发送命令 | `Command(...).AutoSend()` | `OnCommand` |
| 点击后通知服务端 | `button.Callback` | `OnButton` 或 `OnInteraction` |
| 打开网页或平台支持的应用链接 | `button.Link` | 客户端跳转，不产生按钮回调 |
| 正文内可点击命令 | `text.Input`、`text.Enter` | 用户发送后进入 `OnCommand` |
| 快捷菜单操作 | 平台侧配置菜单 | `OnMenu` |

## 准备订阅

[WebSocket](../../../getting-started/events/websocket.md) 同时订阅消息与互动事件。[Webhook](../../../getting-started/events/webhook.md) 在开放平台配置对应订阅。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
      - INTERACTION
```

## 按任务阅读

- [发送按钮、布局和权限](buttons.md)
- [处理按钮回调与确认事件](button-callbacks.md)
- [艾特用户、可点击文字和系统表情](text.md)
- [接收快捷菜单与指令面板边界](menus-and-panels.md)

## 交互与消息的区别

回复会向会话发消息。ACK 向平台确认互动处理结果。先判断业务权限，再及时 ACK，最后处理耗时操作。

每个互动事件只选择一个处理函数。

- 先找匹配具体按钮或菜单 ID 的处理函数。
- 再找匹配事件类型的处理函数。
- 最后使用 `OnEvent(message.INTERACTION_CREATE, ...)`。

在仓库根目录运行完整人工验收示例。

```sh
go run ./examples/interactions -config config.yaml
```

在单聊发送 `/interact`。群聊按平台要求艾特机器人后发送同一命令。

## 注意事项

- SDK 配置不会替应用申请权限、创建菜单或更改平台订阅。
- 消息回复和 ACK 分属不同操作，无法互相替代。SDK 不自动选择成功响应。
- 按钮展示、平台权限和客户端兼容性需要在真实 QQ 中验收。本地测试无法验证这些结果。

## 另请参阅

[消息收发](../README.md) · [QQ 官方：消息交互概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/overview.html) · [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html)
