# SDK 覆盖范围

此页列出根据 SDK 源码核对过的功能，并标明哪些接口尚无专用方法。核对日期为 2026-11-04。

## 按官方模块查找

按官方分类查看实现情况和使用方法。

| 官方分类 | 实现情况 | 使用文档 |
| --- | --- | --- |
| 访问凭证与调用 | 客户端缓存令牌，并发调用共用刷新。支持自定义 HTTP 客户端。 | [凭证](../getting-started/credentials.md)、[API 调用](../getting-started/api-calls.md) |
| 事件订阅与通知 | WebSocket、Webhook，统一路由和生命周期 | [事件接入](../getting-started/events/README.md) |
| 单聊、群聊消息 | 接收、回复、显式主动发送、撤回、引用、媒体等 | [消息收发](../server-api/messages/README.md) |
| 单聊流式消息 | 显式逐次发送，业务负责顺序及恢复 | [流式消息](../server-api/messages/streaming.md) |
| 消息类型和交互 | 消息构造、按钮、文本标签、回调和 ACK | [类型](../server-api/messages/types/README.md)、[交互](../server-api/messages/interactions/README.md) |
| 富媒体与图片素材 | 支持 URL 和本地分片上传。Markdown 素材有独立流程。 | [富媒体](../server-api/messages/rich-media/README.md)、[素材](../server-api/messages/material-management.md) |
| 机器人资料、分享、菜单/面板管理 | 尚无专用方法，可显式调用原始 API | [机器人](../server-api/robot/README.md) |
| 群聊管理 | 尚无完整领域封装 | [群聊管理](../server-api/group-management/README.md) |
| 频道消息和表态 | 新回复入口及旧版专用方法并存 | [频道消息](../server-api/messages/channel/README.md) |
| 频道管理和内容管理 | 主要尚未封装 | [频道管理](../server-api/channel-management/README.md) |

## 使用前应了解的边界

平台文档变化后，按具体字段表和当前源码重新核对。

## 注意事项

- 此页依据当前工作区源码。已发布版本和真实 QQ 验收结果需分别核对。
- 当前 intent 映射缺少 `GROUP_MEMBER_EVENT`。配置此名称会失败。注册 `OnEvent` 不会订阅事件。
- 部分生命周期事件不支持 `Context.Reply` 自动路由。此时按官方允许的来源填写 `ReplyTarget`。
- 频道旧版返回模型可能无法解码字符串时间戳。表态列表旧方法未正确转义特殊分页 cookie。使用前查看[表态说明](../server-api/messages/channel/reactions.md)。
- 引用索引可能不稳定。多层引用可能变成 `TMP_` 或文本摘要。SDK 保留收到的数据。平台未下发的消息链无法保证恢复。
- 平台未必下发内置表情的图片 URL 或小程序的访问 URL。无论读取原始 JSON 还是 SDK 解析后的结构，都无法获得 QQ 未提供的数据。
- Markdown 图片素材提供尺寸辅助。手机端显示仍需实机验证。
- WebSocket 进度保存在内存中，恢复时可能重复投递。SDK 不提供持久化队列或跨进程去重，也不自动重试业务 POST。

- `WithField`、自定义事件名和 `Client.Request` 提供扩展入口。各模块是否提供专用方法与数据结构，见上表。

## 另请参阅

[详细接口核对表](../../docs/api-coverage.md) · [官方 API v2](https://bot.q.qq.com/wiki/develop/api-v2/) · [常见问题](../faq/README.md)
