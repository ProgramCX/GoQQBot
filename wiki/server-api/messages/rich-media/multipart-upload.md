# 分片上传本地文件

`UploadFile` 返回前会算好文件哈希，完成预上传，再逐个 PUT 并确认分片，最后合并文件。上传完成后，用返回的 `file_info` 回复用户或发送主动消息（proactive message）。

## 上传并回复本地图片

下面的函数可以放在单聊或群聊的处理函数里。文件由调用方打开和关闭。SDK 只在上传期间读取文件，文件的读写位置不会变。

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

`FileUpload.Reader` 接收 `io.ReaderAt`。可以传入 `bytes.Reader` 或支持随机读取的业务存储。

## 上传流程与底层入口

按以下顺序调用底层接口。

| 阶段 | 方法 | 重要参数 |
| --- | --- | --- |
| 准备 | `Client.PrepareUpload(ctx, target, metadata)` | `UploadPrepareRequest` 的整文件 MD5、SHA1、MD5TenM |
| 上传分片 | `Client.UploadPart(ctx, part, reader, size)` | 原始字节，精确 size |
| 确认分片 | `Client.FinishUploadPart(ctx, target, input)` | UploadID、原始 PartIndex、分片大小与 MD5 |
| 合并 | `Client.UploadMedia(ctx, target, request)` | 选择 UploadID 来源并设置 FileName、FileType |

预上传和分片完成使用会话路径下的 `upload_prepare`、`upload_part_finish`。预上传中的文件大小和分片大小使用十进制字符串。

服务端返回 `UploadPrepareResult`。结果包含 `UploadID`、`BlockSize`、`Parts` 和 `UploadConfig`。`UploadFile` 在传输前校验整份分片计划。

下面的函数组合单个分片的上传和确认。调用方核对偏移量、分片归属和已保存的计划。全部分片完成后，再由调用方合并。

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

断点续传由业务管理。业务保存上传计划、已确认分片和数据版本。恢复前检查临时地址及任务是否过期。

SDK 把服务端的并发与重试建议保留在 `UploadConfig`。业务自行实现底层调度时才会用到这些建议。失败时可沿[错误链](../../../logging/errors.md)识别预上传、存储 PUT、分片完成或合并阶段。

## 注意事项

- `UploadFile` 只上传，不发送。调用期间的数据要保持稳定，SDK 不转码，也不锁文件。
- `Size` 取正数，且与实际可读字节数一致。`Name` 不能为空。`Type` 指定媒体类别。
- 协议规定 `MD5TenM` 使用文件前 **10002432 字节**计算 MD5，与 10 MiB 不同。
- 官方分片索引从零开始。高层接口也接受从一开始的完整连续计划。确认时保留服务端的原始 `Index`。
- 高层使用固定 64 KiB 哈希缓冲区和单个工作协程。遇到首个错误即停止。
- SDK 不自动重试，也不持久化断点。调用返回后，SDK 不会在后台继续上传。
- 取消上下文可中断网络操作，无法强行中断调用方阻塞的 `ReaderAt`。
- 存储 PUT 不带机器人 Authorization 或 Cookie，并拒绝重定向。同步调用返回后，传输不再读取借用的读取器。
- 上传计划和预签名 URL 可能含授权信息。整体写入日志会暴露这些字段。
- 合并等请求可能已在服务端成功，但本地解码失败。重发 POST 前先核对服务端结果。

## 另请参阅

[富媒体概述](README.md) · [QQ 官方：单聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_prepare.post.html) · [QQ 官方：单聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_part_finish.post.html) · [QQ 官方：群聊预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_prepare.post.html) · [QQ 官方：群聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_part_finish.post.html)
