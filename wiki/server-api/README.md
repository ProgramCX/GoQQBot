# 服务端接口

本目录沿用 QQ 官方服务端接口的模块划分。示例优先使用 `Context` 的便捷方法；后台任务或独立服务使用 `api.Client`，不需要启动机器人事件连接。

| 官方模块 | 从这里开始 | 当前 SDK 范围 |
| --- | --- | --- |
| 消息收发 | [消息收发概述](messages/README.md) | 单聊、群聊、频道收发及消息构造 |
| 消息类型 | [消息类型](messages/types/README.md) | 文本、Markdown、卡片、媒体及旧版 ARK |
| 消息交互 | [交互概述](messages/interactions/README.md) | 按钮、文本交互、回调路由和 ACK |
| 富媒体消息 | [上传概述](messages/rich-media/README.md) | URL 上传、本地分片上传及底层分片步骤 |
| 图片素材管理 | [Markdown 图片素材](messages/material-management.md) | 准备、传输、确认、查询、复用和删除 |
| 机器人 | [机器人接口](robot/README.md) | 资料、分享、菜单管理尚无类型化封装 |
| 群聊管理 | [群聊管理](group-management/README.md) | 查询、成员管理和审批尚无类型化封装 |
| 频道管理 | [频道管理](channel-management/README.md) | 管理与内容类接口主要尚未封装 |

“已封装”表示存在可调用的 SDK 方法，不代表当前机器人已经获得平台权限，也不代表所有客户端效果均已实机验证。`Client.Request` 可以调用某条路径，不等于该模块已有专用请求模型、响应模型或业务校验。

## 按调用方式选择入口

| 需求 | 入口 |
| --- | --- |
| 回复当前消息 | `ctx.ReplyText`、`ctx.Reply`、`ctx.SendReply` |
| 后台向明确来源回复 | `client.SendReply` 与 `message.ReplyTarget` |
| 显式主动发送 | `client.SendMessage`，平台检查主动消息资格 |
| 上传后交给业务保存 | `client.UploadFile`、`client.UploadMarkdownImage` |
| 调用尚未封装的接口 | [原始 API 扩展](../advanced/raw-api.md) |

## 另请参阅

- [API 调用、超时与自定义 HTTP 客户端](../getting-started/api-calls.md)
- [SDK 覆盖范围](../reference/coverage.md)
- [QQ 官方服务端接口导航](https://bot.q.qq.com/wiki/develop/api-v2/)
