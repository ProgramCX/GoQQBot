# GoQQBot 使用文档

GoQQBot 用于接收 QQ 机器人事件和回复消息。你可以将按钮、富媒体与业务服务组合起来。文档对应当前仓库源码。最低 Go 版本为 **1.24.4**，以仓库的 `go.mod` 为准。

目录按 [QQ 官方 API v2](https://bot.q.qq.com/wiki/develop/api-v2/) 划分模块。每页介绍一项任务，组织方式参考 [Gin 的使用文档](https://gin-gonic.com/zh-cn/docs/logging/write-log/)。内容以 SDK 调用为主。完整字段表见官方文档。

## 开始使用

1. [安装 SDK](getting-started/installation.md)，或使用本地 `replace` 接入尚未发布的工作区。
2. [运行第一个机器人](getting-started/quick-start.md)，选择 WebSocket 或 Webhook。
3. [注册命令与普通消息](getting-started/events/routing.md)，按事件类型复用 handler。
4. [发送 Markdown 和按钮](server-api/messages/interactions/buttons.md)，再接入自己的业务服务。

## 起步指南

按接入阶段查找配置和运行方法。

| 主题 | 页面 |
| --- | --- |
| 启动接入 | [安装](getting-started/installation.md)、[快速开始](getting-started/quick-start.md) |
| 获取访问凭证 | [Access Token 与凭证](getting-started/credentials.md) |
| API 调用指南 | [客户端、超时与错误](getting-started/api-calls.md) |
| 事件订阅与通知 | [概述](getting-started/events/README.md)、[WebSocket](getting-started/events/websocket.md)、[Webhook](getting-started/events/webhook.md) |
| SDK 配置与生命周期 | [配置参考](getting-started/configuration.md)、[运行与关闭](getting-started/lifecycle.md) |

## 服务端接口

按消息场景或管理任务查找接口。

| 官方模块 | 文档 |
| --- | --- |
| 消息收发 | [概述](server-api/messages/README.md)、[单聊](server-api/messages/c2c/README.md)、[群聊](server-api/messages/group/README.md)、[频道](server-api/messages/channel/README.md) |
| 消息类型 | [纯文本](server-api/messages/types/text.md)、[Markdown](server-api/messages/types/markdown.md)、[图文卡片](server-api/messages/types/cards.md)、[ARK](server-api/messages/types/ark.md) |
| 引用和撤回 | [消息索引、回复来源和撤回 ID](server-api/messages/reference-and-recall.md) |
| 富文本接收 | [图片、表情、ARKData、MsgElement 和 MessageScene](server-api/messages/rich-content.md) |
| 流式消息 | [单聊顺序更新](server-api/messages/streaming.md) |
| 消息交互 | [概述](server-api/messages/interactions/README.md)、[按钮](server-api/messages/interactions/buttons.md)、[回调](server-api/messages/interactions/button-callbacks.md)、[文本交互](server-api/messages/interactions/text.md) |
| 富媒体消息 | [概述](server-api/messages/rich-media/README.md)、[单聊](server-api/messages/rich-media/c2c-upload.md)、[群聊](server-api/messages/rich-media/group-upload.md)、[分片步骤](server-api/messages/rich-media/multipart-upload.md) |
| 素材管理 | [Markdown 图片上传、复用、查询和删除](server-api/messages/material-management.md) |
| 机器人 | [资料与频道列表](server-api/robot/README.md)、[分享和事件](server-api/robot/sharing-and-events.md)、[菜单及指令面板](server-api/messages/interactions/menus-and-panels.md) |
| 群聊管理 | [概述](server-api/group-management/README.md)、[成员](server-api/group-management/members.md)、[入群审批](server-api/group-management/join-requests.md) |
| 频道管理 | [概述](server-api/channel-management/README.md)、[资料及子频道](server-api/channel-management/guilds-and-channels.md)、[成员与权限](server-api/channel-management/members-and-permissions.md)、[内容](server-api/channel-management/content.md) |

## 日志、业务接入和验收

完成接入后，可按业务需要查看以下页面。

- [日志概述](logging/README.md)、[写入日志文件](logging/write-log.md)、[错误与 panic](logging/errors.md)。
- [业务服务接入](advanced/business-integration.md)、[原始 API 与字段扩展](advanced/raw-api.md)、[测试消息发送](advanced/testing.md)。
- [运行示例与人工验收](examples/README.md)、[常见问题](faq/README.md)、[SDK 覆盖范围](reference/coverage.md)。

管理类文档标明尚未封装的接口。对应条目提供官方入口或应用层调用示例。

## 文档维护

完整侧边导航见 [_Sidebar.md](_Sidebar.md)。Go 代码块均可独立编译。

- `package guide` 示例可放入业务包。
- 标注完整程序的 `package main` 示例可单独运行。
- 示例通过函数参数或本地配置接收参数。

在仓库根目录运行以下命令。脚本检查链接和代码围栏。Go 示例放入带本地 `replace` 的临时模块中编译。

```sh
python wiki/_tools/check_docs.py
```

[SDK rules](../.codex/rules/go-sdk.md)记录项目维护规范。历史核对记录见 [docs](../docs/api-coverage.md)。

## 注意事项

- 管理类示例中的原始调用不代表 SDK 已提供完整封装。
- 平台权限、消息审核、客户端显示及接口配额由 QQ 平台决定。
- 检查脚本不连接机器人，也不发消息。编译通过只表示示例与当前 SDK 兼容。QQ 实机联调需另外执行。

## 另请参阅

[文档风格指南](STYLE.md)
