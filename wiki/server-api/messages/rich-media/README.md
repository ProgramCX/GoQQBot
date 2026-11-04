# 发送富媒体消息

图片、视频、语音和普通文件都先上传，再发送。上传到目标会话后，平台返回不透明的 `file_info`。用 `message.Media(fileInfo)` 构造消息。

| 数据来源 | 推荐入口 | 行为 |
| --- | --- | --- |
| 公网 HTTP(S) URL | `Context.UploadMedia` 或 `Client.UploadMedia` | QQ 平台下载指定 URL |
| 本地文件、内存对象、随机访问存储 | `Context.UploadFile` 或 `Client.UploadFile` | SDK 计算哈希并按服务端计划分片上传 |
| 已准备的上传任务 | `PrepareUpload`、`UploadPart`、`FinishUploadPart`、`UploadMedia` | 业务管理上传阶段和恢复状态 |
| Markdown 中可复用图片 | [图片素材管理](../material-management.md) | 返回 `qqbot-media://...`，与 `file_info` 用途不同 |

## 选择媒体类别

按官方协议中的媒体类别选择对应常量。

| 类型 | 常量 | 格式提示 |
| --- | --- | --- |
| 1 图片 | `message.FileTypeImage` | PNG/JPEG 等平台支持图片 |
| 2 视频 | `message.FileTypeVideo` | 平台支持的视频格式，例如 MP4 |
| 3 语音 | `message.FileTypeVoice` | 以平台当前允许的语音格式为准 |
| 4 文件 | `message.FileTypeFile` | 普通文件，权限和限制由平台决定 |

## 上传结果的有效范围

发送时使用目标会话对应的 `RichMediaResponse.FileInfo`。结果还包含下列字段。

- `TTL` 以秒计。零表示长期有效。其他值到期后由业务重新上传。
- `FileUUID` 是文件标识。
- `RawURL` 是可选下载地址。

`SrvSendMsg` 默认为 false，只上传。推荐保留默认值，再用 `ctx.Reply` 发送被动消息（passive message）。

## 选择指南

先看文件在哪里，再按单聊或群聊选择上传方法。

- [单聊 URL 上传与发送](c2c-upload.md)
- [群聊 URL 上传与业务服务集成](group-upload.md)
- [本地文件、预上传及分片完成](multipart-upload.md)

## 注意事项

- 上传成功只确认上传结果。发送和视频播放需要分别验证。
- SDK 不转码。更改扩展名无法转换媒体格式。QQ 决定文件大小、音视频编码、审核及展示要求。
- `file_info` 应用于对应场景及会话。单聊、群聊和不同会话之间的复用没有保证。
- 临时下载地址不适合保存为永久业务引用。SDK 不自动缓存、刷新或跨会话迁移上传结果。
- `SrvSendMsg` 为 true 时请求主动消息（proactive message）。平台检查资格和配额，见 [调用 OpenAPI](../../../getting-started/api-calls.md)。
- `URL`、`UploadID`、`FileData` 三选一。官方将 `file_data` 描述为字符串，限制为 5M，但未明确编码。
- SDK 会原样传递 `file_data`。官方没有明确说明是否支持 Base64 等编码，SDK 也没有验证任意二进制编码。上传本地文件时可以直接用分片上传。

## 另请参阅

[消息收发](../README.md) · [Markdown 图片素材](../material-management.md) · [QQ 官方：富媒体消息概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/rich-media.html)
