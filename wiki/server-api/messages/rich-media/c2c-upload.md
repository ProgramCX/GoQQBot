# 单聊富媒体上传

单聊富媒体可先用 `ctx.UploadMedia` 上传，再用普通回复入口发送。SDK 从当前事件确定用户和会话。下例接收业务提供的图片 URL，地址需能被 QQ 下载。

```go
package guide

import (
	"context"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func replyImageURL(ctx *types.Context, imageURL string) error {
	op, cancel := context.WithTimeout(ctx.RequestContext(), 30*time.Second)
	defer cancel()
	uploaded, err := ctx.UploadMedia(op, message.RichMediaRequest{
		FileType: message.FileTypeImage,
		URL:      imageURL,
	})
	if err != nil {
		return err
	}
	return ctx.ReplyContext(op, message.Media(uploaded.FileInfo))
}
```

在 `C2C_MESSAGE_CREATE` 的处理函数里调用示例。URL 可来自业务配置或自有存储。同一个函数也适用于[群聊](group-upload.md)。上传与发送都使用当前会话。

## 参数与协议

单聊上传使用 `POST /v2/users/{user_openid}/files`。

| 字段 | 说明 |
| --- | --- |
| `FileType` | 图片、视频、语音或文件，使用对应 `FileType` 常量 |
| `URL` | 平台下载的 HTTP(S) 地址。从 URL、`UploadID`、`FileData` 中三选一。 |
| `SrvSendMsg` | 默认为 `false`。设为 `true` 请求直接发送主动消息（proactive message）。 |
| `FileName` | 用 `UploadID` 或 `FileData` 时填写。URL 模式无需自行推断远端文件名。 |



本地图片使用[分片上传](multipart-upload.md)。

## 脱离消息处理器调用

业务服务可以用 API 客户端和明确的会话目标上传。服务无需持有路由器或 SDK Context。由业务选择发送时机。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func uploadUserVideo(op context.Context, client *api.Client, userOpenID, videoURL string) (*message.RichMediaResponse, error) {
	return client.UploadMedia(op, message.Target{
		Scene: message.SceneC2C,
		ID:    userOpenID,
	}, message.RichMediaRequest{
		FileType: message.FileTypeVideo,
		URL:      videoURL,
	})
}
```

示例函数只上传。函数不构造被动回复来源，也不主动发送。[发送单聊消息](../c2c/send.md)时，可用 `Client.SendReply` 配合真实事件来源。具备平台资格时也可用 `Client.SendMessage`。

## 错误与重试

用 `errors.Is(err, api.ErrInvalidMedia)` 判断输入是否合法。平台错误保留为 `*api.APIError`。用 `errors.As` 读取 HTTP 状态、错误码和 trace。`DiagnosticMessage()` 可读取平台提供的说明。



## 注意事项

- 将未经验证的地址交给平台，会触发对用户指定地址的下载。
- SDK 只校验 URL 的基本格式。QQ 下载节点可能访问不到地址，或因超时、格式及审核拒绝请求。
- 本机路径无法作为公网 URL 上传。URL 上传失败后，SDK 不会自动下载文件或改走本地上传。
- 上传返回了结果但 err 也不是 nil。此时 HTTP 请求可能已成功，响应解码却失败了。上传任务或主动消息可能已存在。重发可能产生重复。SDK 不自动重试 POST。

## 另请参阅

[富媒体概述](README.md) · [本地分片上传](multipart-upload.md) · [QQ 官方：单聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_files.post.html)
