# 群聊富媒体上传

群聊和单聊使用同一套 SDK 上传入口。处理群聊事件时，`ctx.UploadMedia` 自动选中 `group_openid`；无需重复解析群标识或手工拼接口路径。

下面是群聊发送文件的完整处理函数。`fileURL` 由业务提供，`FileTypeFile` 表示作为普通文件上传，不要求平台把它识别为可播放视频。

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

将函数用于 `GROUP_AT_MESSAGE_CREATE`，或平台已开通的 `GROUP_MESSAGE_CREATE`。群内是否必须艾特机器人取决于平台事件订阅和权限，不由上传 API 决定。

## 业务服务中上传

后台任务不需要伪造 `Context`。明确指定 `message.SceneGroup` 和群 OpenID，结果返回给业务保存或发送。

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

上传接口为 `POST /v2/groups/{group_openid}/files`。`Target.ID` 不是普通群号码，必须来自当前机器人的平台 OpenID。不要跨机器人或跨会话复用 `file_info`，也不要把 `file_uuid` 当作它。

## 主要参数及限制

请求字段与 [单聊上传](c2c-upload.md) 相同：`FileType` 为 1..4，URL、UploadID、FileData 三选一，`SrvSendMsg` 默认 false。上传默认不占用被动回复序号，之后发送通过 Context 正常分配 `msg_seq`。

URL 上传是 QQ 服务器下载文件。本机浏览器能下载不代表 QQ 下载节点能完成；遇到平台超时或下载失败，需要结合错误码、下载来源和 trace 排查。SDK 不根据未公开错误码猜测权限或格式原因，也不隐式改用另一种上传方式。

本地文件推荐 [分片上传](multipart-upload.md)。群聊与单聊都走同一上传实现，但各自使用对应会话的预上传、分片完成和合并接口。

## 另请参阅

- [富媒体概述与类型](README.md)
- [群聊消息发送](../group/send.md)
- [QQ 官方：群聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_files.post.html)
