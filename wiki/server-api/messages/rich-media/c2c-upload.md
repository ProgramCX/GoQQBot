# 单聊富媒体上传

在单聊消息处理函数中调用 `ctx.UploadMedia`，由当前事件确定用户及会话，随后通过普通回复入口发送。以下函数接收一个由业务提供、可被 QQ 平台下载的公网图片 URL。

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

将此函数用于 `C2C_MESSAGE_CREATE` 的处理器；URL 可来自业务配置或自己的存储服务，不要直接把未验证的任意外部输入交给平台下载。相同处理函数也可用于群聊，上传和发送均保持在当前会话。

## 参数与协议

| 字段 | 说明 |
| --- | --- |
| `FileType` | 图片、视频、语音或文件，使用对应 `FileType` 常量 |
| `URL` | 平台下载的 HTTP(S) 地址，不能同时设置 UploadID 或 FileData |
| `SrvSendMsg` | 默认 false；true 请求直接主动发送 |
| `FileName` | 使用 UploadID 或 FileData 时必填，URL 模式无需自行推断远端文件名 |

底层上传请求为 `POST /v2/users/{user_openid}/files`。SDK 校验 URL 的基本格式，但不能证明 QQ 下载节点可访问该主机或接受其内容。平台可能因下载失败、上传超时、格式或审核拒绝请求。

如果图片来自本地，请使用 [分片上传](multipart-upload.md)，不要把本机文件路径当成 URL。SDK 不会在 URL 上传失败后自动下载和改走本地上传。

## 脱离消息处理器调用

业务服务无需持有路由器或 SDK Context。使用自己持有的 API 客户端及明确会话目标即可上传，之后再由业务决定发送时机。

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

此函数仅上传，不构造被动回复来源，也不主动发送。后续发送可使用 `Client.SendReply` 配合真实事件来源，或按平台资格显式使用 `Client.SendMessage`。

## 错误与重试

输入不合法可用 `errors.Is(err, api.ErrInvalidMedia)` 判断。平台错误保留为 `*api.APIError`，通过 `errors.As` 获取 HTTP 状态、错误码和 trace；平台提供的说明可显式读取 `DiagnosticMessage()`。

如果上传返回非 nil 结果及非 nil 错误，可能是 HTTP 成功后的响应解码失败；任务甚至主动消息可能已存在，不要盲目重发。SDK 不自动重试 POST。

## 另请参阅

- [富媒体概述](README.md)
- [本地分片上传](multipart-upload.md)
- [QQ 官方：单聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_files.post.html)
