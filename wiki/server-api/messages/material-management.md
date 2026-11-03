# Markdown 图片素材管理

上传一次图片，保存平台返回的素材引用，之后在同一机器人的 Markdown 消息中复用。该接口返回 `qqbot-media://image/v1/...`，与普通富媒体的 `file_info` 不同，两者不能互换。

## 上传本地图片并展示

下面是完整处理函数，可用于单聊或群聊。上传不自动发送，展示由 `ctx.Reply` 完成。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func replyChart(ctx *types.Context, imagePath string) error {
	image, err := ctx.UploadMarkdownImageFile(ctx.RequestContext(), imagePath)
	if err != nil {
		return err
	}
	return ctx.Reply(image.Markdown("业务图表"))
}
```

`image.Markdown(...)` 返回普通的不可变 `message.Outgoing`，可以继续组合按钮、文本片段，或设置图片转存校验。调用 `MarkdownText(...)` 可获取行内语法，嵌入自己的 Markdown 模板。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func replySavedChart(ctx *types.Context, image message.ImageMaterial) error {
	inline, err := image.MarkdownText("分析图表")
	if err != nil {
		return err
	}
	return ctx.ReplyMarkdown("## 分析结果\n" + inline + "\n统计口径见业务说明。")
}
```

复用只发送引用，不产生新的上传请求。请传完整的平台素材值，不用临时 PreviewURL 替换其 URL，也不要手工拼接素材 token。

## 与业务服务结合

业务已有图表字节时，直接用 `bytes.Reader`，不必落地临时文件。下面函数不依赖接收事件或路由器，返回素材供业务持久化。

```go
package guide

import (
	"bytes"
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func uploadChart(op context.Context, client *api.Client, pngBytes []byte) (*message.ImageMaterial, error) {
	return client.UploadMarkdownImage(op, api.ImageUpload{
		Reader:      bytes.NewReader(pngBytes),
		Size:        int64(len(pngBytes)),
		Name:        "chart.png",
		ContentType: "image/png",
	})
}
```

`Reader` 为借用的 `io.ReaderAt`，数据在调用期间必须稳定，SDK 不关闭它。文件路径版本则显式打开并关闭文件。高层同步计算哈希、上传及确认，使用有界哈希缓冲区；不缓存、不后台上传、不自动重试或删除素材。

由应用创建 `*api.Client`，通过 `goqqrobot.WithAPIClient(client)` 注入 Bot，可以让业务服务与处理器共用认证缓存。SDK 不要求业务服务引用 `internal` 包或实现通用存储接口。

建议按业务键保存以下记录；它是应用数据模型建议，不是 SDK 的自动缓存：

| 业务字段 | 来源及用途 |
| --- | --- |
| robot / AppID | 标记素材归属，不能跨机器人复用 |
| reportID 或其他业务键 | 将素材与业务对象关联 |
| imageID | `ImageMaterial.ID`，查询和删除用的标识 |
| URL | `ImageMaterial.URL`，完整 Markdown 素材引用 |
| width / height | 可选显示尺寸，需自行持久化 |

过期后是否重新生成图片、上传和更新记录，由业务明确决定。SDK 不基于同名文件或 MD5 自动去重。

## 完整 API

| Client 方法 | 对应操作 |
| --- | --- |
| `PrepareMarkdownImage(ctx, metadata)` | `POST /v2/media-images/prepare` |
| `PutMarkdownImage(ctx, plan, file)` | 向预签名 URL PUT 图片二进制 |
| `CompleteMarkdownImage(ctx, input)` | `POST /v2/media-images/complete` |
| `ListMarkdownImages(ctx, cursor)` | `GET /v2/media-images` 的一页 |
| `DeleteMarkdownImage(ctx, imageID)` | `DELETE /v2/media-images/{image_id}/delete` |
| `UploadMarkdownImage(ctx, file)` | 组合哈希、准备、PUT 和确认 |
| `UploadMarkdownImageFile(ctx, path)` | 打开文件并调用组合流程，结束关闭文件 |

## 手工组合准备、PUT 与确认

仅在需要独立管理上传阶段时使用底层方法。下面以稳定的 PNG 字节为例，保持文件名、大小、MIME 和数据在三个步骤中一致。

```go
package guide

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func uploadChartInSteps(op context.Context, client *api.Client, pngBytes []byte) (*message.ImageMaterial, error) {
	md5Sum := md5.Sum(pngBytes)
	sha1Sum := sha1.Sum(pngBytes)
	file := api.ImageUpload{
		Reader:      bytes.NewReader(pngBytes),
		Size:        int64(len(pngBytes)),
		Name:        "chart.png",
		ContentType: "image/png",
	}
	digest := hex.EncodeToString(md5Sum[:])
	plan, err := client.PrepareMarkdownImage(op, message.ImagePrepareRequest{
		FileName:    file.Name,
		FileSize:    uint64(file.Size),
		ContentType: file.ContentType,
		MD5:         digest,
		SHA1:        hex.EncodeToString(sha1Sum[:]),
	})
	if err != nil {
		return nil, err
	}
	etag, err := client.PutMarkdownImage(op, *plan, file)
	if err != nil {
		return nil, err
	}
	if etag != digest {
		return nil, fmt.Errorf("图片上传 ETag 与原始文件 MD5 不一致")
	}
	return client.CompleteMarkdownImage(op, message.ImageCompleteRequest{
		UploadID: plan.UploadID,
		MD5:      etag,
	})
}
```

Prepare 的 FileSize 是 JSON 数字，与普通富媒体预上传的十进制字符串不同；MD5 为 32 位、SHA1 为 40 位十六进制。PUT 使用与 Prepare 一致的 Content-Type，不带机器人认证和 Cookie，并拒绝重定向。`PutMarkdownImage` 返回去掉外围引号后的 ETag MD5，Complete 必须使用这个值。

如果 Prepare 或 Complete 同时返回非 nil 结果和错误，HTTP 已成功但响应不能正常使用，任务或素材可能已创建。高层调用会停止，不会盲目重发；底层调用方应保留必要业务状态并核对平台结果。Raw、PresignedURL 和 PreviewURL 可能含临时鉴权信息，不整体写日志。

## 分页查询

`ListMarkdownImages(op, "")` 取首页，返回 `ImageMaterialPage`。SDK 每次仅请求一页、正确编码游标，不进行无界遍历。

下面函数获取最多 maxPages 页，返回累计素材和下一游标，便于业务继续加载；不修改素材。

```go
package guide

import (
	"context"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func listImages(op context.Context, client *api.Client, cursor string, maxPages int) ([]message.ImageMaterial, string, error) {
	if maxPages < 1 {
		return nil, cursor, fmt.Errorf("maxPages 必须为正数")
	}
	var images []message.ImageMaterial
	for i := 0; i < maxPages; i++ {
		page, err := client.ListMarkdownImages(op, cursor)
		if err != nil {
			return images, cursor, err
		}
		images = append(images, page.Images...)
		cursor = page.NextCursor
		if cursor == "" {
			break
		}
	}
	return images, cursor, nil
}
```

`Total` 是 uint64，`CreatedAt` 和 `ExpiresAt` 是 Unix 秒 int64。平台这些字段可能以 JSON 数字或十进制字符串返回，SDK 均接受，不经浮点转换；非法整数、负的 Total 或溢出返回错误。`NextCursor` 为空表示末页。

## 删除素材

删除明确的素材 ID，不要传 URL 或消息 ID。下面函数适合业务已确认不再需要该素材后调用；不是自动清理策略。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func deleteSelectedImage(op context.Context, client *api.Client, image message.ImageMaterial) error {
	return client.DeleteMarkdownImage(op, image.ID)
}
```

删除后不能在新消息中引用它，已发送消息的转存副本不受此删除接口影响。不存在的 ID 保留平台错误，不被静默当作成功；业务记录何时删除或标记失效，应与自己的存储事务配合。

## 显示尺寸及手机端排查

高层上传可从 JPEG、PNG 和 GIF 图片头读取 Width、Height，构造带尺寸的 Markdown 图片语法。WebP、列表结果或数据库恢复的记录可能没有尺寸，可显式使用 `image.WithDimensions(width, height).Markdown("说明")`。

`WithDimensions` 返回独立副本，宽高都必须为正；校验在 `MarkdownText` 或 `Build` 时返回。尺寸是 SDK 本地显示元数据，不是素材 API 保证返回的字段，也不会自动持久化。两个尺寸都为零时输出无尺寸语法。

若手机只显示替代文字，在删除前检查新发送的素材消息，核对实际图片尺寸、手机 QQ 版本及平台转存结果，并与电脑端比较。自动补尺寸不代表所有客户端均已实机通过；不要自动用临时 PreviewURL 替换永久素材引用。

## 大小、配额与有效期

| 项目 | 当前文档及 SDK 行为 |
| --- | --- |
| 单图大小 | 官方不超过 10 MB，SDK 按 10 MiB（10485760 字节）校验，平台最终判定 |
| 格式 | PNG、JPEG、GIF、WebP，文件扩展名须与 MIME 一致，SDK 不转码 |
| 数量 | 官方当前每机器人默认一百张，可通过显式删除释放配额 |
| 上传任务 | Prepare 的 UploadID 四小时有效 |
| 预签名地址 | 使用返回的 ExpiresIn 有效期 |
| 图片素材 | 距最后成功消息引用连续九十天未使用则过期 |

`TTL` 是滑动窗口秒数，不是固定过期时刻；查询的 `ExpiresAt` 才是当前服务器给出的到期时间。不要把首次上传的 TTL 变成永久本地倒计时，也不要跨机器人复用素材。

本地错误通过 `errors.Is(err, api.ErrInvalidImageMaterial)` 判断，平台错误通过 `errors.As` 提取 `*api.APIError`。是否重新上传、更新映射或清理失效记录由应用选择，SDK 不隐式重传。

## 另请参阅

- [普通富媒体上传](rich-media/README.md)
- [Markdown 消息](types/markdown.md)
- [按钮组合](interactions/buttons.md)
- [QQ 官方：素材管理](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/material-management.html)
