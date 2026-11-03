# Markdown 图片素材管理

依据 [QQ 官方素材管理文档](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/material-management.html)，这套接口管理机器人自己的 Markdown 图片，返回 `qqbot-media://image/v1/...`。它与群聊/单聊上传返回的 `file_info` 是两个独立协议，不能互换。

## 最常用的三行

在处理器中上传本地图片并回复：

```go
image, err := ctx.UploadMarkdownImageFile(ctx.RequestContext(), "chart.png")
if err != nil { return err }
return ctx.Reply(image.Markdown("业务图表"))
```

Upload 只上传，不发送消息。`image.Markdown` 返回已有的不可变 `message.Outgoing`，可继续 `.Buttons(...)` 或 `.Append(...)`。整条回复仍由 `ctx.Reply` 自动选择群聊/单聊接收者、事件来源和序号。

组合到调用方的 Markdown 模板中：

```go
inline, err := image.MarkdownText("图表说明")
if err != nil { return err }
return ctx.ReplyMarkdown("## 分析结果\n" + inline + "\n后续业务正文")
```

图片引用保持平台返回值不变，说明文字做 Markdown 转义。不把临时 PreviewURL 当成永久图片链接。

高层 UploadMarkdownImage/UploadMarkdownImageFile 对可识别的本地 JPEG、PNG、GIF 读取尺寸，保存到返回值的 Width、Height；随后 Markdown/MarkdownText 会输出 `![说明 #宽px #高px](qqbot-media://...)`。这样可以对照排查不同 QQ 客户端对无尺寸图片语法的处理差异，不代表已保证每个客户端版本都支持素材引用。

低层完成接口、列表查询或从数据库恢复的记录可能没有尺寸。可显式设置，WebP 也可用同一个入口：

```go
return ctx.Reply(image.WithDimensions(800, 400).Markdown("业务图表"))
```

WithDimensions 返回独立副本，两个值必须为正，错误在 MarkdownText/Build 时返回。Width、Height 是 SDK 本地展示信息，不假装成服务端响应字段；需要跨进程保留时由业务单独保存。两个值均为零时保留原无尺寸语法。不要把 `#800px #400px` 塞进说明文字绕过转义。

## 和业务服务组合

有图表字节、内存文件或自己管理的存储时，不需要写入临时磁盘文件，也不需要 SDK Context：

```go
image, err := client.UploadMarkdownImage(op, api.ImageUpload{
    Reader: bytes.NewReader(pngBytes),
    Size: int64(len(pngBytes)),
    Name: "chart.png",
})
```

`Reader` 为 `io.ReaderAt`，调用期间数据须保持稳定，结束后所有权仍属于调用方。哈希与上传各读取一次原文件，另按需读取图片头获取可选尺寸；使用固定 64 KiB 哈希缓冲区，不复制整个图片；提供的 reader 也需要让读操作能够返回。ContentType 默认按扩展名推断，可显式指定并校验一致性。路径版本显式打开文件并负责关闭，Reader 版本不关闭调用方资源。

推荐让业务服务持有自己创建的 `*api.Client`，并用 `goqqrobot.WithAPIClient(client)` 注入 Bot，共用认证缓存。业务服务可定义自己消费的小接口，不需要导入 `internal` 包或持有消息路由器。例如：

```go
type ImageStore interface {
    UploadMarkdownImage(context.Context, api.ImageUpload) (*message.ImageMaterial, error)
    ListMarkdownImages(context.Context, string) (*message.ImageMaterialPage, error)
    DeleteMarkdownImage(context.Context, string) error
}
```

SDK 不预设数据库、缓存键、清理策略或自动去重。业务可用 `reportID -> {image.ID, image.URL}` 保存引用，同一个机器人在有效期内复用它；再次回复 `image.Markdown(...)` 不会产生上传请求。素材属于机器人，不属于单条消息或群会话，不要跨机器人复用。

ImageMaterial 的 `TTL` 是滑动有效期窗口，单位秒，不是固定过期时刻。查询结果中的 `ExpiresAt` 是服务器给出的 Unix 秒时间；不要把首次上传的 TTL 当作永久的本地倒计时。收到 `40035009` 等失效错误时，由业务显式决定是否更新自己的记录、重新上传，SDK 不自动重传。

## 完整 API

| Client 方法 | 协议操作 |
| --- | --- |
| `PrepareMarkdownImage(ctx, message.ImagePrepareRequest)` | `POST /v2/media-images/prepare` |
| `PutMarkdownImage(ctx, plan, api.ImageUpload)` | HTTPS PUT 原始图片字节到预签名 URL |
| `CompleteMarkdownImage(ctx, message.ImageCompleteRequest)` | `POST /v2/media-images/complete` |
| `ListMarkdownImages(ctx, cursor)` | `GET /v2/media-images?cursor=...` |
| `DeleteMarkdownImage(ctx, imageID)` | `DELETE /v2/media-images/{image_id}/delete` |
| `UploadMarkdownImage(ctx, api.ImageUpload)` | 计算校验和并按上述三步完整上传 |
| `UploadMarkdownImageFile(ctx, path)` | 显式读取本地文件并完整上传 |

上传 metadata 的 `file_size` 为 JSON 数字，与普通富媒体预上传中的十进制字符串不同。Prepare 中的 MD5 和 SHA1 为十六进制。Put 使用与 Prepare **完全一致**的 Content-Type，且不携带机器人 Authorization 或 Cookie，不接受重定向。Complete 使用 PUT 响应 ETag 去掉双引号后的 MD5；高层 Upload 还会将其与原始文件 MD5 对照，避免上传过程中数据变化。

完整低层组合：

```go
plan, err := client.PrepareMarkdownImage(op, metadata)
if err != nil { return err }
etag, err := client.PutMarkdownImage(op, *plan, file)
if err != nil { return err }
image, err := client.CompleteMarkdownImage(op, message.ImageCompleteRequest{
    UploadID: plan.UploadID, MD5: etag,
})
```

低层调用需保持同一份 Name、Size、ContentType 和文件数据，任务、重试及恢复状态由业务管理。高层调用没有隐藏重试、goroutine、缓存或自动发送。每个网络步骤均传播 context。建议调用方用 `context.WithTimeout` 设置总体期限。

分页只请求一页，游标按 URL 查询参数正确编码：

平台的 total、created_at、expires_at 可能使用 JSON 数字或十进制字符串。SDK 同时接受这两种形式，公开字段仍为 uint64/int64，不经过浮点转换；负的 total、溢出及非整数不会静默变成零。

```go
page, err := client.ListMarkdownImages(op, "")
if err != nil { return err }
// 处理 page.Images；根据业务需要继续读取，避免无界遍历。
if page.NextCursor != "" {
    next, err := client.ListMarkdownImages(op, page.NextCursor)
    if err != nil { return err }
    return service.HandleImagePage(op, next)
}
```

删除需要明确 image_id，不能传 qqbot-media URL；不存在的 ID 返回平台错误，不会被转成成功。删除后新消息不能引用该资源，已发送消息的转存副本不受影响。

## 限制和错误

- 单图 10 MB，SDK 按 `10 * 1024 * 1024` 字节校验；支持 PNG、JPEG、GIF、WebP，扩展名与 MIME 必须一致。SDK 不转码，文件内容仍须为真实合法图片。
- 官方说明每个机器人默认最多 100 张；上传任务有效期四小时，预签名 URL 以 ExpiresIn 为准，素材连续 90 天没有被消息引用则过期。配额及有效期由平台最终判定。
- PreviewURL、PresignedURL、Raw 可能含临时鉴权数据，不应整包记录或持久化成业务引用。
- 本地输入错误可通过 `errors.Is(err, api.ErrInvalidImageMaterial)` 判断；API 错误保留为 `*api.APIError`，平台说明通过 `DiagnosticMessage()` 读取。
- Prepare/Complete 返回非 nil 结果并伴随错误时，HTTP 已成功但响应不能正常使用。Raw 保留服务器响应；任务或素材可能已创建，不要盲目重新上传。可按 image_id 或查询接口进行业务核对。
- 其他失败立即停止，失败的 PUT 不会被 Complete，ETag 缺失或不匹配也不会继续；调用方明确决定重试或清理。

## 按钮验收

在仓库根目录启动：

```powershell
go run ./examples/materials -config config.yaml -image-file "E:\Document\GoQQBot\test\image.jpg"
```

停止先前的测试机器人进程后再启动，避免多个同凭证连接干扰。单聊发送 `/materials`，群聊发送 `@机器人 /materials`，包含五个按钮及对应处理器：上传并展示、复用当前素材、查询素材首页、删除当前素材、接收结构说明。直接发送图片、表情、引用消息会进入有类型的 OnMessage 并显示接收结构摘要。

示例只在内存中保留当前素材，不自动删除历史素材；重启后可查询平台列表，但业务映射应由正式应用持久化。再次点上传会显式执行新的上传流程，并替换内存中的当前引用。删除只作用于这个进程记住的当前素材。

手机端只显示 `[说明文字]` 时，先在删除素材之前检查新发送的“上传并展示”和“复用当前素材”，并与电脑端比较。本地 JPEG/PNG/GIF 会带上实际尺寸；如果仍显示替代文字，需要结合手机 QQ 版本和平台实际转存结果继续验证，不能直接断言为客户端不支持，也不要把临时 PreviewURL 自动替换成永久引用。删除后的客户端缓存可能影响观察，应与删除前的显示分开记录。

模拟测试从实际生成的菜单按钮找到处理器，覆盖群聊/单聊的上传、复用、列表和删除流程，也覆盖 MIME、ETag、游标编码和失败停止；它不代表 QQ 平台的真实验收已经完成。
