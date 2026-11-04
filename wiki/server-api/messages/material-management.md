# 管理 Markdown 图片素材

图片上传一次后，可在同一机器人的 Markdown 消息中复用。保存平台返回的素材引用即可。

## 上传本地图片并展示

下面的处理函数可用于单聊或群聊。上传不自动发送。展示由 `ctx.Reply` 完成。

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

`image.Markdown(...)` 返回不可变的 `message.Outgoing`。消息可继续[组合按钮](interactions/buttons.md)、文本片段或图片转存校验。`MarkdownText(...)` 返回行内语法，可嵌入自己的 Markdown 模板。

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

复用只发送素材引用，不发起新的上传请求。

## 与业务服务结合

已有图表字节时，可用 `bytes.Reader` 省去临时文件。下面的函数不依赖接收事件或路由器。业务可以保存返回的素材。

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

`Reader` 是 SDK 借用的 `io.ReaderAt`。使用文件路径的版本会自行打开并关闭文件。高层接口同步计算哈希、上传及确认，使用有界哈希缓冲区。

应用可创建 `*api.Client`，再通过 `goqqrobot.WithAPIClient(client)` 注入 Bot。业务服务与处理函数可以共用认证缓存。业务服务无需引用 `internal` 包，也无需实现通用存储接口。

应用可按业务键保存以下记录。SDK 不会自动缓存这些记录。

| 业务字段 | 来源及用途 |
| --- | --- |
| robot / AppID | 标记素材归属 |
| reportID 或其他业务键 | 将素材与业务对象关联 |
| imageID | `ImageMaterial.ID`，查询和删除用的标识 |
| URL | `ImageMaterial.URL`，完整 Markdown 素材引用 |
| width / height | 可选显示尺寸，由业务持久化 |

素材过期后，由业务决定是否重建图片、上传及更新记录。

## 完整 API

高层方法适合一次完成上传。需要分步控制时，可调用对应阶段的方法。

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

需要独立管理上传阶段时，可组合三个底层方法。下面用稳定的 PNG 字节演示。三个步骤使用相同的文件名、大小、MIME 和数据。

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

Prepare 的 `FileSize` 使用 JSON 数字。普通富媒体预上传则使用十进制字符串。`MD5` 是 32 位十六进制（128 位 ÷ 4）。`SHA1` 是 40 位十六进制（160 位 ÷ 4）。

`PutMarkdownImage` 返回去掉外围引号的 ETag MD5。Complete 使用该值。

## 分页查询

`ListMarkdownImages(op, "")` 取首页并返回 `ImageMaterialPage`。SDK 每次请求一页并编码游标。`NextCursor` 为空表示末页。

下面的函数最多读取 `maxPages` 页。业务可使用返回的游标继续加载。分页函数不修改素材。

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

`Total` 是 `uint64`。`CreatedAt` 和 `ExpiresAt` 是以 Unix 秒表示的 `int64`。SDK 接受 JSON 数字和十进制字符串，无需经过浮点转换。

## 删除素材

业务确认素材不再需要后，可以按明确的素材 ID 删除。

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

应用结合自身存储事务，决定何时删除业务记录或标记失效。

## 显示尺寸及手机端排查

高层上传可读取 JPEG、PNG 和 GIF 图片头中的宽高。SDK 用 Width 和 Height 构造带尺寸的 Markdown 图片语法。

WebP、列表结果或数据库恢复的记录可能没有尺寸。可以显式调用 `image.WithDimensions(width, height).Markdown("说明")`。`WithDimensions` 返回独立副本。

手机仅显示替代文本（alt text）时，先检查新消息，再删除素材。

- 核对实际图片尺寸。
- 核对手机 QQ 版本。
- 核对平台转存结果，并与电脑端对比。

## 大小、配额与有效期

按官方配额及服务端返回的有效期管理素材。

| 项目 | 当前文档及 SDK 行为 |
| --- | --- |
| 单图大小 | 官方不超过 10 MB。SDK 按 10 MiB（10 × 1024 × 1024 = 10485760 字节）校验，平台最终判定 |
| 格式 | PNG、JPEG、GIF、WebP。扩展名与 MIME 要一致，SDK 不转码 |
| 数量 | 官方当前每机器人默认一百张，显式删除可释放配额 |
| 上传任务 | Prepare 的 UploadID 四小时有效 |
| 预签名地址 | 使用返回的 ExpiresIn 有效期 |
| 图片素材 | 自最后一次消息成功引用起，连续九十天未使用则过期 |

`TTL` 是以秒计的滑动窗口。查询结果的 `ExpiresAt` 是服务器给出的当前到期时间。

用 `errors.Is(err, api.ErrInvalidImageMaterial)` 判断本地错误。用 `errors.As` 提取平台错误 `*api.APIError`，见[错误处理](../../logging/errors.md)。

## 注意事项

- 素材接口返回 `qqbot-media://image/v1/...`。普通富媒体的 `file_info` 无法代替素材引用。
- 使用完整的平台素材值。临时 PreviewURL 不适合替换素材 URL。手工拼接素材 token 缺少平台返回的引用依据。
- `Reader` 中的数据在调用期间要保持稳定。SDK 不关闭借用的读取器。
- 高层接口不缓存、不在后台上传，也不自动重试或删除素材。SDK 不根据同名文件或 MD5 去重。
- PUT 使用与 Prepare 一致的 Content-Type。请求不带机器人认证或 Cookie，也不跟随重定向。
- Prepare 或 Complete 返回了结果但 err 也不是 nil，表示 HTTP 已成功，但响应无法正常使用。任务或素材可能已经创建。
- 上述情况会停止高层流程，不自动重发。使用底层方法时，业务保存所需状态并核对平台结果。
- Raw、PresignedURL 和 PreviewURL 可能含临时鉴权信息。整体写入日志会暴露这些字段。
- SDK 不进行无界分页。整数非法、`Total` 为负或数值溢出时，SDK 返回错误。
- 删除接口使用素材 ID，URL 和消息 ID 不适用。素材删除后，新消息无法引用。已发送消息的转存副本不受此删除接口影响。
- 不存在的素材 ID 会保留平台错误，SDK 不会当作成功处理。
- 显式宽高均需为正，错误在 `MarkdownText` 或 `Build` 时返回。两个尺寸均为零时输出无尺寸语法。
- 尺寸是 SDK 本地显示元数据。素材 API 不保证返回尺寸，SDK 也不自动持久化尺寸。
- 自动补尺寸不代表所有客户端已经实机通过。临时 PreviewURL 不适合替换永久素材引用。
- 首次上传返回的 TTL 不能作为永久本地倒计时。素材只供所属机器人使用。
- 重新上传、更新映射和清理失效记录由应用决定。SDK 不隐式重传。

## 另请参阅

[普通富媒体上传](rich-media/README.md) · [Markdown 消息](types/markdown.md) · [按钮组合](interactions/buttons.md) · [QQ 官方：素材管理](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/material-management.html)
