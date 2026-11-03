# 文档导航

- [文档首页](README.md)

## 起步指南

- [起步指南](getting-started/README.md)
- [调用 OpenAPI](getting-started/api-calls.md)
- [配置机器人](getting-started/configuration.md)
- [获取访问凭证](getting-started/credentials.md)
- [安装 SDK](getting-started/installation.md)
- [运行与关闭](getting-started/lifecycle.md)
- [创建第一个机器人](getting-started/quick-start.md)
- [事件订阅与通知](getting-started/events/README.md)
  - [注册命令与普通消息处理函数](getting-started/events/routing.md)
  - [通过 Webhook 接收事件](getting-started/events/webhook.md)
  - [通过 WebSocket 接收事件](getting-started/events/websocket.md)

## 服务端接口

- [服务端接口](server-api/README.md)
- [消息收发](server-api/messages/README.md)
  - [Markdown 图片素材管理](server-api/messages/material-management.md)
  - [引用与撤回消息](server-api/messages/reference-and-recall.md)
  - [读取图片、表情、卡片与转发](server-api/messages/rich-content.md)
  - [流式发送单聊消息](server-api/messages/streaming.md)
  - [单聊消息](server-api/messages/c2c/README.md)
    - [接收单聊消息](server-api/messages/c2c/receive.md)
    - [发送单聊消息](server-api/messages/c2c/send.md)
  - [群聊消息](server-api/messages/group/README.md)
    - [接收群聊消息](server-api/messages/group/receive.md)
    - [发送群聊消息](server-api/messages/group/send.md)
  - [频道消息](server-api/messages/channel/README.md)
    - [发送及撤回频道消息](server-api/messages/channel/messages.md)
    - [添加和读取频道表情表态](server-api/messages/channel/reactions.md)
  - [消息类型](server-api/messages/types/README.md)
    - [使用 ARK 模板](server-api/messages/types/ark.md)
    - [发送图文卡片](server-api/messages/types/cards.md)
    - [发送 Markdown](server-api/messages/types/markdown.md)
    - [回复纯文本](server-api/messages/types/text.md)
  - [消息交互](server-api/messages/interactions/README.md)
    - [处理按钮回调](server-api/messages/interactions/button-callbacks.md)
    - [发送按钮](server-api/messages/interactions/buttons.md)
    - [快捷菜单与指令面板](server-api/messages/interactions/menus-and-panels.md)
    - [文本交互](server-api/messages/interactions/text.md)
  - [富媒体消息](server-api/messages/rich-media/README.md)
    - [单聊富媒体上传](server-api/messages/rich-media/c2c-upload.md)
    - [群聊富媒体上传](server-api/messages/rich-media/group-upload.md)
    - [本地文件与分片上传](server-api/messages/rich-media/multipart-upload.md)
- [机器人](server-api/robot/README.md)
  - [分享链接与机器人事件](server-api/robot/sharing-and-events.md)
- [群聊管理](server-api/group-management/README.md)
  - [入群申请与自动审批](server-api/group-management/join-requests.md)
  - [查询和管理群成员](server-api/group-management/members.md)
- [频道管理](server-api/channel-management/README.md)
  - [频道内容管理](server-api/channel-management/content.md)
  - [频道资料与子频道管理](server-api/channel-management/guilds-and-channels.md)
  - [频道成员、身份组与权限](server-api/channel-management/members-and-permissions.md)

## 日志

- [日志与错误处理](logging/README.md)
- [处理错误与 panic](logging/errors.md)
- [如何写入日志文件](logging/write-log.md)

## 业务接入与扩展

- [业务接入与扩展](advanced/README.md)
- [将机器人接入业务服务](advanced/business-integration.md)
- [调用未封装的官方接口](advanced/raw-api.md)
- [测试消息发送](advanced/testing.md)

## 示例与验收

- [示例程序与人工验收](examples/README.md)

## 参考

- [SDK 覆盖范围](reference/coverage.md)

## 常见问题

- [常见问题](faq/README.md)
