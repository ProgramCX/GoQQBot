# 富媒体消息

发送图片、视频、语音或普通文件分为两步：先上传到目标会话，取得不透明的 `file_info`，再使用 `message.Media(fileInfo)` 发送。上传成功不代表消息已经发送。

| 数据来源 | 推荐入口 | 行为 |
| --- | --- | --- |
| 公网 HTTP(S) URL | `Context.UploadMedia` 或 `Client.UploadMedia` | QQ 平台下载指定 URL |
| 本地文件、内存对象、随机访问存储 | `Context.UploadFile` 或 `Client.UploadFile` | SDK 计算哈希并按服务端计划分片上传 |
| 已准备的上传任务 | `PrepareUpload`、`UploadPart`、`FinishUploadPart`、`UploadMedia` | 业务自行管理上传阶段和恢复状态 |
| Markdown 中可复用图片 | [图片素材管理](../material-management.md) | 返回 `qqbot-media://...`，不是 `file_info` |

## 选择媒体类别

| 类型 | 常量 | 格式提示 |
| --- | --- | --- |
| 1 图片 | `message.FileTypeImage` | PNG/JPEG 等平台支持图片 |
| 2 视频 | `message.FileTypeVideo` | 平台支持的视频格式，例如 MP4 |
| 3 语音 | `message.FileTypeVoice` | 以平台当前允许的语音格式为准 |
| 4 文件 | `message.FileTypeFile` | 普通文件，权限和限制由平台决定 |

SDK 不转码，也不根据扩展名把文件变成符合平台要求的媒体。文件大小、音视频编码、内容审核及最终展示由 QQ 决定。不要把“上传文件成功”当成“视频可播放”的验收结果。

## 上传结果的有效范围

`RichMediaResponse.FileInfo` 应用于相应的场景和会话，不能假定单聊、群聊、不同会话间都可以复用。`TTL` 为秒，零表示长期有效；其他值到期后由业务显式重新上传。

`FileUUID` 是文件标识，`RawURL` 是可选下载地址，不要把临时下载地址当成永久业务引用。SDK 不自动缓存、刷新或跨会话迁移上传结果。

`SrvSendMsg` 默认为 false，只上传；显式为 true 才请求平台直接主动发送，受资格与配额限制。推荐使用默认值，再用 `ctx.Reply` 明确发送被动回复。

## 选择指南

- [单聊 URL 上传与发送](c2c-upload.md)
- [群聊 URL 上传与业务服务集成](group-upload.md)
- [本地文件、预上传及分片完成](multipart-upload.md)

`URL`、`UploadID`、`FileData` 必须三选一。官方将 `file_data` 描述为字符串和 5M 限制，但没有明确编码；SDK 仅原样透传，不宣称已验证任意二进制编码。本地文件优先使用分片入口，不要猜测 Base64 或其他编码。

## 另请参阅

- [消息收发](../README.md)
- [Markdown 图片素材](../material-management.md)
- [QQ 官方：富媒体消息概述](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/rich-media.html)
