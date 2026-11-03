# 消息交互

用按钮和消息正文中的可点击指令，让用户触发业务操作。先选择交互方式，再选择处理入口：回调按钮、普通命令和网页跳转产生的行为不同。

| 需要实现的行为 | 构造入口 | 接收入口 |
| --- | --- | --- |
| 点击后预填命令 | `button.Command` | 用户发送后进入 `OnCommand` |
| 单聊点击直接发送命令 | `Command(...).AutoSend()` | `OnCommand` |
| 点击后通知服务端 | `button.Callback` | `OnButton` 或 `OnInteraction` |
| 打开网页或平台支持的应用链接 | `button.Link` | 客户端跳转，不产生按钮回调 |
| 正文内可点击命令 | `text.Input`、`text.Enter` | 用户发送后进入 `OnCommand` |
| 快捷菜单操作 | 平台侧配置菜单 | `OnMenu` |

## 准备订阅

WebSocket 配置同时订阅消息与互动事件；Webhook 在开放平台配置相应订阅。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
      - INTERACTION
```

配置不会替你申请权限、创建菜单或更改平台订阅。

## 按任务阅读

- [发送按钮、布局和权限](buttons.md)
- [处理按钮回调与确认事件](button-callbacks.md)
- [艾特用户、可点击文字和系统表情](text.md)
- [接收快捷菜单与指令面板边界](menus-and-panels.md)

## 交互与消息的区别

消息回复是向会话发送内容，ACK 是向平台确认互动处理结果；二者不能互相替代。回调按钮应及时 ACK，再执行耗时工作。SDK 不自动选择成功响应，业务需要先完成权限判断。

具体按钮或菜单 ID 的处理函数优先于类型处理函数，最后才进入 `OnEvent(message.INTERACTION_CREATE, ...)`；每个互动事件只选择一个处理函数。

完整人工验收示例在仓库根目录运行：

```sh
go run ./examples/interactions -config config.yaml
```

在单聊发送 `/interact`，或在群聊按平台要求艾特机器人后发送 `/interact`。按钮展示、平台权限与客户端兼容性需要真实 QQ 验证，本地测试不能替代实机验收。

## 另请参阅

- [消息收发](../README.md)
- [QQ 官方：消息交互概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/overview.html)
- [QQ 官方：互动事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/interaction_create.html)
