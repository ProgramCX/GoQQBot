# 本地文件与分片上传

使用 `UploadFile` 同步完成哈希、预上传、分片 PUT、分片确认和合并。它只上传，不发送；得到 `file_info` 后再选择回复或主动发送。

## 上传并回复本地图片

以下函数可用于单聊或群聊处理器。调用方打开并关闭文件；SDK 在同步调用期间借用它，不改变文件偏移。

```go
package guide

import (
	"context"
	"os"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func replyLocalImage(ctx *types.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	op, cancel := context.WithTimeout(ctx.RequestContext(), 2*time.Minute)
	defer cancel()
	uploaded, err := ctx.UploadFile(op, api.FileUpload{
		Reader: file,
		Size:   stat.Size(),
		Name:   stat.Name(),
		Type:   message.FileTypeImage,
	})
	if err != nil {
		return err
	}
	return ctx.ReplyContext(op, message.Media(uploaded.FileInfo))
}
```

`FileUpload.Reader` 是 `io.ReaderAt`，也可使用 `bytes.Reader` 或实现随机读取的业务存储。`Size` 必须为正且与实际可读数据一致，`Name` 必须非空，`Type` 为媒体类别。调用期间数据必须稳定，SDK 不替你转码或锁文件。

## 上传流程与底层入口

| 阶段 | 方法 | 重要参数 |
| --- | --- | --- |
| 准备 | `Client.PrepareUpload(ctx, target, metadata)` | `UploadPrepareRequest` 的整文件 MD5、SHA1、MD5TenM |
| 上传分片 | `Client.UploadPart(ctx, part, reader, size)` | 原始字节，精确 size |
| 确认分片 | `Client.FinishUploadPart(ctx, target, input)` | UploadID、原始 PartIndex、分片大小与 MD5 |
| 合并 | `Client.UploadMedia(ctx, target, request)` | 选择 UploadID 来源并设置 FileName、FileType |

预上传和分片完成接口分别为会话路径下的 `upload_prepare`、`upload_part_finish`。预上传的文件和分片大小使用十进制字符串；`MD5TenM` 是文件前 **10002432 字节**的 MD5，不是 10 MiB。

服务端返回 `UploadPrepareResult`，包含 UploadID、BlockSize、Parts 和 UploadConfig。`UploadFile` 先校验整份分片计划，再传输数据。官方索引从零开始；高层也接受完整连续的从一开始的计划，确认时保留服务器返回的原始 Index，不能随意减一。

底层上传单个分片可组合为下面的完整函数。调用方负责确认 offset、分片归属和已保存计划一致，并在全部分片完成后合并；本函数不隐式完成整个上传。

```go
package guide

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func transferPart(op context.Context, client *api.Client, target message.Target, uploadID string, part message.UploadPart, source io.ReaderAt, offset int64) error {
	size, err := strconv.ParseInt(part.BlockSize, 10, 64)
	if err != nil {
		return err
	}
	if source == nil || size <= 0 || offset < 0 {
		return fmt.Errorf("分片源、大小或偏移不合法")
	}
	hash := md5.New()
	copied, err := io.Copy(hash, io.NewSectionReader(source, offset, size))
	if err != nil {
		return err
	}
	if copied != size {
		return io.ErrUnexpectedEOF
	}
	if err := client.UploadPart(op, part, io.NewSectionReader(source, offset, size), size); err != nil {
		return err
	}
	return client.FinishUploadPart(op, target, message.UploadPartFinishRequest{
		UploadID:  uploadID,
		PartIndex: part.Index,
		BlockSize: part.BlockSize,
		MD5:       hex.EncodeToString(hash.Sum(nil)),
	})
}
```

## 并发、恢复和失败语义

高层上传使用固定 64 KiB 哈希缓冲区及单个工作协程，首错即停；不自动重试、不持久化断点、不在后台继续上传。服务端并发与重试建议保留在 UploadConfig 中，只有业务明确实现底层调度时才使用。

需要断点续传时，业务保存上传计划、已确认分片和数据版本，恢复前检查临时地址及任务是否过期。取消 context 可中断网络步骤，但不能强制打断调用方实现的阻塞 `ReaderAt`。

存储 PUT 不携带机器人 Authorization 或 Cookie，拒绝重定向；借用读取器在同步调用返回后不再交给传输继续读取。上传计划和预签名 URL 可能含授权信息，不要整体写入日志。

失败阶段会通过错误链标识预上传、存储 PUT、分片完成或合并。合并等 HTTP 成功后解码失败存在结果不确定性，不得盲目重发 POST。

## 另请参阅

- [富媒体概述](README.md)
- [QQ 官方：单聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_prepare.post.html)
- [QQ 官方：单聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_part_finish.post.html)
- [QQ 官方：群聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_prepare.post.html)
- [QQ 官方：群聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_part_finish.post.html)
