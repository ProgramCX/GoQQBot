package api

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif" // Register standard image headers for optional local display dimensions.
	// 注册标准图片格式头解码器，用于可选的本地显示尺寸读取。
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// ErrInvalidImageMaterial marks invalid metadata, storage ETags or material responses.
// ErrInvalidImageMaterial 表示素材元数据、存储 ETag 或素材响应不合法。
// Platform errors remain inspectable as *APIError through errors.As.
// 可通过 errors.As 将平台错误提取为 *APIError。
var ErrInvalidImageMaterial = errors.New("invalid Markdown image material")

// MaxImageMaterialSize is the documented per-image limit, interpreted as 10 MiB.
// MaxImageMaterialSize 是文档规定的单张图片大小上限，按 10 MiB 解释。
const MaxImageMaterialSize int64 = 10 << 20

// ImageUpload borrows stable random-access image data for a synchronous upload. The caller retains ownership.
// ImageUpload 在同步上传期间借用稳定的随机访问图片数据，所有权仍归调用方。
// ContentType may be empty to infer it from Name (.png, .jpg/.jpeg, .gif, .webp); if supplied it must match.
// ContentType 为空时根据 Name 推断（.png、.jpg/.jpeg、.gif、.webp）；显式指定时必须匹配。
type ImageUpload struct {
	Reader      io.ReaderAt
	Size        int64
	Name        string
	ContentType string
}

func imageOperation(ctx context.Context, c *Client) error {
	if ctx == nil || c == nil {
		return fmt.Errorf("%w: client and operation context are required", ErrInvalidImageMaterial)
	}
	return ctx.Err()
}

func imageContentType(name, declared string) (string, error) {
	types := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp"}
	expected := types[strings.ToLower(filepath.Ext(name))]
	if expected == "" || (declared != "" && declared != expected) {
		return "", fmt.Errorf("%w: image filename and MIME type must match png/jpeg/gif/webp", ErrInvalidImageMaterial)
	}
	return expected, nil
}

func validImageHash(value string, size int) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == size
}

// PrepareMarkdownImage registers an image upload without transferring bytes or creating a message.
// PrepareMarkdownImage 登记图片上传，不传输字节或创建消息。
// Upload tasks expire after four hours; the returned storage URL has its own ExpiresIn lifetime.
// 上传任务四小时后过期；返回的存储 URL 有独立的 ExpiresIn 有效期。
// It never retries. A non-nil result with error means registration succeeded but the response was invalid.
// 不执行重试；同时返回结果和错误表示登记成功，但响应不合法。
func (c *Client) PrepareMarkdownImage(ctx context.Context, input message.ImagePrepareRequest) (*message.ImageUploadPlan, error) {
	if err := imageOperation(ctx, c); err != nil {
		return nil, err
	}
	if _, err := imageContentType(input.FileName, input.ContentType); err != nil {
		return nil, err
	}
	if input.ContentType == "" || input.FileSize == 0 || input.FileSize > uint64(MaxImageMaterialSize) || !validImageHash(input.MD5, md5.Size) || !validImageHash(input.SHA1, sha1.Size) {
		return nil, fmt.Errorf("%w: prepare requires MIME type, size in 1..10 MiB, MD5 and SHA1", ErrInvalidImageMaterial)
	}
	data, err := c.Post(ctx, "/v2/media-images/prepare", input)
	if err != nil {
		return nil, fmt.Errorf("prepare Markdown image: %w", err)
	}
	result := &message.ImageUploadPlan{Raw: append(json.RawMessage(nil), data...)}
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful image registration (task may exist): %w", err)
	}
	if result.UploadID == "" || result.PresignedURL == "" {
		return result, fmt.Errorf("%w: successful prepare response lacks upload ID or storage URL", ErrInvalidImageMaterial)
	}
	return result, nil
}

// PutMarkdownImage transfers image bytes with the declared MIME type and returns the storage ETag MD5.
// PutMarkdownImage 使用声明的 MIME 类型传输图片字节，并返回存储 ETag MD5。
// The same Name, Size and ContentType must have been used at prepare time. The reader is never closed.
// Name、Size 和 ContentType 必须与准备阶段一致；不关闭读取器。
// It rejects redirects, sends no bot credentials/cookies to storage, and performs no retries.
// 拒绝重定向，不向存储发送机器人凭证或 Cookie，也不执行重试。
func (c *Client) PutMarkdownImage(ctx context.Context, plan message.ImageUploadPlan, file ImageUpload) (string, error) {
	if err := imageOperation(ctx, c); err != nil {
		return "", err
	}
	contentType, err := imageContentType(file.Name, file.ContentType)
	if err != nil {
		return "", err
	}
	if file.Reader == nil || file.Size <= 0 || file.Size > MaxImageMaterialSize || plan.UploadID == "" {
		return "", fmt.Errorf("%w: reader, upload ID and image size in 1..10 MiB are required", ErrInvalidImageMaterial)
	}
	etag, err := c.putStorage(ctx, plan.PresignedURL, io.NewSectionReader(file.Reader, 0, file.Size), file.Size, contentType)
	if err != nil {
		return "", fmt.Errorf("PUT Markdown image: %w", err)
	}
	etag = strings.TrimSpace(etag)
	if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		etag = etag[1 : len(etag)-1]
	}
	if !validImageHash(etag, md5.Size) {
		return "", fmt.Errorf("%w: storage response lacks a valid MD5 ETag; bytes may already be uploaded", ErrInvalidImageMaterial)
	}
	return strings.ToLower(etag), nil
}

// CompleteMarkdownImage finalizes uploaded bytes. Input.MD5 is the ETag returned by PutMarkdownImage.
// CompleteMarkdownImage 确认已上传字节；Input.MD5 是 PutMarkdownImage 返回的 ETag。
// A non-nil result with error means HTTP succeeded; do not blindly retry or upload again.
// 同时返回结果和错误表示 HTTP 请求成功；不得盲目重试或再次上传。
func (c *Client) CompleteMarkdownImage(ctx context.Context, input message.ImageCompleteRequest) (*message.ImageMaterial, error) {
	if err := imageOperation(ctx, c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.UploadID) == "" || !validImageHash(input.MD5, md5.Size) {
		return nil, fmt.Errorf("%w: upload ID and MD5 ETag required", ErrInvalidImageMaterial)
	}
	data, err := c.Post(ctx, "/v2/media-images/complete", input)
	if err != nil {
		return nil, fmt.Errorf("complete Markdown image: %w", err)
	}
	result := &message.ImageMaterial{Raw: append(json.RawMessage(nil), data...)}
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful image completion (material may exist): %w", err)
	}
	if result.ID == "" {
		return result, fmt.Errorf("%w: successful completion response lacks image ID", ErrInvalidImageMaterial)
	}
	if _, err := result.MarkdownText(""); err != nil {
		return result, fmt.Errorf("%w: successful completion response lacks valid image reference", ErrInvalidImageMaterial)
	}
	return result, nil
}

// ListMarkdownImages fetches one page of this bot's valid images, newest first.
// ListMarkdownImages 按最新优先顺序获取当前机器人有效图片的一页。
// Empty cursor selects the first page. Pagination and durable business-key mappings remain caller-owned.
// 空游标选择首页；翻页及业务键的持久化映射由调用方管理。
func (c *Client) ListMarkdownImages(ctx context.Context, cursor string) (*message.ImageMaterialPage, error) {
	if err := imageOperation(ctx, c); err != nil {
		return nil, err
	}
	path := "/v2/media-images"
	if cursor != "" {
		path += "?" + url.Values{"cursor": []string{cursor}}.Encode()
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list Markdown images: %w", err)
	}
	result := &message.ImageMaterialPage{Raw: append(json.RawMessage(nil), data...)}
	if strings.TrimSpace(string(data)) == "null" {
		return nil, fmt.Errorf("%w: image list response must be an object", ErrInvalidImageMaterial)
	}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("decode image list: %w", err)
	}
	if result.Images == nil {
		result.Images = []message.ImageMaterial{}
	}
	return result, nil
}

// DeleteMarkdownImage explicitly deletes one bot-owned image and releases its quota.
// DeleteMarkdownImage 显式删除一张机器人持有的图片并释放配额。
// Its reference immediately stops working for new messages; already-sent copies remain unaffected.
// 其引用立即不能用于新消息；已发送的副本不受影响。
// A platform not-found error is preserved, not silently treated as success. No retries are performed.
// 保留平台的未找到错误，不将其静默视为成功；不执行重试。
func (c *Client) DeleteMarkdownImage(ctx context.Context, imageID string) error {
	if err := imageOperation(ctx, c); err != nil {
		return err
	}
	if strings.TrimSpace(imageID) == "" {
		return fmt.Errorf("%w: image ID required", ErrInvalidImageMaterial)
	}
	_, err := c.Delete(ctx, "/v2/media-images/"+url.PathEscape(imageID)+"/delete")
	if err != nil {
		return fmt.Errorf("delete Markdown image: %w", err)
	}
	return nil
}

type imageContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r imageContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// UploadMarkdownImage synchronously hashes, registers, transfers and completes one reusable image.
// UploadMarkdownImage 同步完成哈希、登记、传输和确认，生成一张可复用图片。
// It does not send a message, cache, delete other images, retry POSTs or launch background work.
// 不发送消息、不缓存、不删除其他图片、不重试 POST，也不启动后台任务。
// Supply a deadline and stable input; cancellation cannot forcibly interrupt a caller's blocked ReaderAt.
// 请提供截止时间及稳定输入；取消无法强制中断调用方阻塞的 ReaderAt。
func (c *Client) UploadMarkdownImage(ctx context.Context, file ImageUpload) (*message.ImageMaterial, error) {
	if err := imageOperation(ctx, c); err != nil {
		return nil, err
	}
	contentType, err := imageContentType(file.Name, file.ContentType)
	if err != nil {
		return nil, err
	}
	if file.Reader == nil || file.Size <= 0 || file.Size > MaxImageMaterialSize {
		return nil, fmt.Errorf("%w: reader and size in 1..10 MiB required", ErrInvalidImageMaterial)
	}
	md5sum, sha1sum := md5.New(), sha1.New()
	n, err := io.CopyBuffer(io.MultiWriter(md5sum, sha1sum), imageContextReader{ctx, io.NewSectionReader(file.Reader, 0, file.Size)}, make([]byte, 64<<10))
	if err != nil {
		return nil, fmt.Errorf("hash Markdown image: %w", err)
	}
	if n != file.Size {
		return nil, fmt.Errorf("hash Markdown image: %w", io.ErrUnexpectedEOF)
	}
	checksum := hex.EncodeToString(md5sum.Sum(nil))
	plan, err := c.PrepareMarkdownImage(ctx, message.ImagePrepareRequest{FileName: file.Name, FileSize: uint64(file.Size), ContentType: contentType, MD5: checksum, SHA1: hex.EncodeToString(sha1sum.Sum(nil))})
	if err != nil {
		return nil, err
	}
	etag, err := c.PutMarkdownImage(ctx, *plan, file)
	if err != nil {
		return nil, err
	}
	if etag != checksum {
		return nil, fmt.Errorf("%w: storage ETag differs from original image MD5; check file stability", ErrInvalidImageMaterial)
	}
	result, err := c.CompleteMarkdownImage(ctx, message.ImageCompleteRequest{UploadID: plan.UploadID, MD5: etag})
	if err != nil {
		return result, err
	}
	// Dimensions are local display hints, not a second validation gate after successful upload.
	// 尺寸是本地显示提示，不作为上传成功后的第二道校验门槛。
	// Unsupported headers (e.g. WebP without a registered decoder) remain usable via WithDimensions.
	// 不支持的格式头（例如未注册解码器的 WebP）仍可通过 WithDimensions 使用。
	config, _, decodeErr := image.DecodeConfig(imageContextReader{ctx, io.NewSectionReader(file.Reader, 0, file.Size)})
	if decodeErr == nil && config.Width > 0 && config.Height > 0 {
		result.Width = config.Width
		result.Height = config.Height
	}
	return result, nil
}

// UploadMarkdownImageFile explicitly opens a local file, uploads it and closes the owned handle.
// UploadMarkdownImageFile 显式打开本地文件、上传并关闭所持有的文件句柄。
// Use UploadMarkdownImage for bytes.Reader, generated charts, object storage or caller-owned file handles.
// bytes.Reader、生成图表、对象存储或调用方文件句柄可使用 UploadMarkdownImage。
func (c *Client) UploadMarkdownImageFile(ctx context.Context, path string) (*message.ImageMaterial, error) {
	if err := imageOperation(ctx, c); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open Markdown image: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat Markdown image: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: regular image file required", ErrInvalidImageMaterial)
	}
	return c.UploadMarkdownImage(ctx, ImageUpload{Reader: file, Size: info.Size(), Name: info.Name()})
}
