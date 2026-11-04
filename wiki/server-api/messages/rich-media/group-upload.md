# 上传群聊富媒体

群聊和单聊都可以用 `ctx.UploadMedia` 上传文件。群聊时，SDK 会从事件里读取 `group_openid`，并拼好请求路径。

下面的函数发送群聊文件。`fileURL` 由业务提供。`FileTypeFile` 按普通文件上传，不要求平台识别为可播放视频。

```go
package guide

import (
	"context"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func replyGroupFile(ctx *types.Context, fileURL string) error {
	op, cancel := context.WithTimeout(ctx.RequestContext(), time.Minute)
	defer cancel()
	uploaded, err := ctx.UploadMedia(op, message.RichMediaRequest{
		FileType: message.FileTypeFile,
		URL:      fileURL,
	})
	if err != nil {
		return err
	}
	return ctx.ReplyContext(op, message.Media(uploaded.FileInfo))
}
```

将函数用于 `GROUP_AT_MESSAGE_CREATE`，或平台已开通的 `GROUP_MESSAGE_CREATE`。注册方式见 [接收群聊消息](../group/receive.md)。

## 业务服务中上传

指定 `message.SceneGroup` 和群 OpenID 即可上传。后台任务也能直接上传，省去构造 `Context` 的步骤。上传结果可以保存下来，也可以接着用于发送。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func uploadGroupVoice(op context.Context, client *api.Client, groupOpenID, voiceURL string) (*message.RichMediaResponse, error) {
	return client.UploadMedia(op, message.Target{
		Scene: message.SceneGroup,
		ID:    groupOpenID,
	}, message.RichMediaRequest{
		FileType: message.FileTypeVoice,
		URL:      voiceURL,
	})
}
```

上传请求为 `POST /v2/groups/{group_openid}/files`。

## 主要参数及限制

字段与 [单聊上传](c2c-upload.md) 相同。

- `FileType` 使用平台类型 1..4。
- `URL`、`UploadID`、`FileData` 三选一。
- `SrvSendMsg` 默认 false。

上传默认不占用被动消息（passive message）的回复序号。随后发送时，`Context` 正常分配 `msg_seq`。

本地文件推荐 [分片上传](multipart-upload.md)。群聊与单聊共用上传实现。预上传、分片完成和合并各用目标会话的接口。

## 注意事项

- 群内是否需要艾特机器人，由平台订阅及权限决定。上传接口不改变接收条件。
- `Target.ID` 使用当前机器人的平台群 OpenID。普通群号码不适用。
- `file_info` 用于当前机器人和会话。`file_uuid` 不能代替 `file_info`。
- 文件由 QQ 服务器从 URL 下载。即使本机浏览器能打开地址，QQ 也可能下载失败。
- 平台超时或下载失败时，结合错误码、下载来源和 trace 排查。SDK 不猜测未公开错误码对应的权限或格式问题。SDK 也不自动切换上传方式。

## 另请参阅

[富媒体概述与类型](README.md) · [群聊消息发送](../group/send.md) · [QQ 官方：群聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_files.post.html)
