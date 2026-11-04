package api

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// ErrInvalidMedia marks invalid upload input or an inconsistent upload plan.
// ErrInvalidMedia 表示上传输入无效或上传计划不一致。
var ErrInvalidMedia = errors.New("invalid rich media operation")

func conversationPath(ctx context.Context, c *Client, target message.Target) (string, error) {
	if c == nil || ctx == nil || strings.TrimSpace(target.ID) == "" {
		return "", fmt.Errorf("%w: client, context and recipient are required", ErrInvalidReplyTarget)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var kind string
	switch target.Scene {
	case message.SceneC2C:
		kind = "users"
	case message.SceneGroup:
		kind = "groups"
	default:
		return "", fmt.Errorf("%w: operation requires C2C or group", ErrInvalidReplyTarget)
	}
	return "/v2/" + kind + "/" + url.PathEscape(target.ID), nil
}

// RecallMessage deletes an explicit message in the selected conversation. It never guesses the message ID.
// RecallMessage 删除指定会话中的明确消息，绝不猜测消息 ID。
// C2C permits bot messages only; group member recalls require bot administrator permission.
// 单聊仅允许撤回机器人消息；群成员消息撤回要求机器人具有管理员权限。
// QQ enforces the two-minute window and permissions; errors remain inspectable as *APIError.
// QQ 执行两分钟时限及权限检查；错误仍可通过 *APIError 检查。
func (c *Client) RecallMessage(ctx context.Context, target message.Target, messageID string) error {
	if target.Scene == message.SceneChannel || target.Scene == message.SceneDirectMessage {
		return c.RecallChannelMessage(ctx, target, messageID, false)
	}
	path, err := conversationPath(ctx, c, target)
	if err != nil {
		return err
	}
	if strings.TrimSpace(messageID) == "" {
		return fmt.Errorf("%w: message ID is required", ErrInvalidReplyTarget)
	}
	_, err = c.Delete(ctx, path+"/messages/"+url.PathEscape(messageID))
	return err
}

// RecallChannelMessage 撤回子频道或频道私信中的指定消息，可选择隐藏提示。
func (c *Client) RecallChannelMessage(ctx context.Context, target message.Target, messageID string, hideTip bool) error {
	var route string
	switch target.Scene {
	case message.SceneChannel:
		route = CHANNEL_DELETE_MESSAGE
	case message.SceneDirectMessage:
		route = CHANNEL_PRIVATE_DELETE
	default:
		return fmt.Errorf("%w: recall requires a channel or direct session", ErrInvalidReplyTarget)
	}
	path, err := apiwire.Path(route, target.ID, messageID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidReplyTarget, err)
	}
	if hideTip {
		path += "?hidetip=true"
	}
	_, err = c.Delete(ctx, path)
	return err
}

// UploadMedia uploads from a URL, a prepared upload ID, or protocol FileData. It never retries POSTs.
// UploadMedia 从 URL、准备好的上传 ID 或协议 FileData 上传；绝不重试 POST。
// Only SrvSendMsg=true sends proactively. A non-nil result with error means success response decoding failed;
// 仅 SrvSendMsg=true 时主动发消息；同时返回结果和错误表示成功响应解码失败，
// the upload or proactive message may already exist, so do not blindly retry.
// 上传或主动消息可能已存在，因此不得盲目重试。
func (c *Client) UploadMedia(ctx context.Context, target message.Target, input message.RichMediaRequest) (*message.RichMediaResponse, error) {
	path, err := conversationPath(ctx, c, target)
	if err != nil {
		return nil, err
	}
	if input.FileType < 1 || input.FileType > 4 {
		return nil, fmt.Errorf("%w: file_type must be in 1..4", ErrInvalidMedia)
	}
	sources := 0
	if input.URL != "" {
		u, err := url.Parse(input.URL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return nil, fmt.Errorf("%w: HTTP media URL required", ErrInvalidMedia)
		}
		sources++
	}
	if input.UploadID != "" {
		sources++
	}
	if input.FileData != "" {
		sources++
	}
	if sources != 1 || ((input.UploadID != "" || input.FileData != "") && strings.TrimSpace(input.FileName) == "") {
		return nil, fmt.Errorf("%w: exactly one source and a file name for local uploads are required", ErrInvalidMedia)
	}
	data, err := c.Post(ctx, path+"/files", input)
	if err != nil {
		operation := "upload media"
		if input.UploadID != "" {
			operation = "merge media upload"
		}
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	result := new(message.RichMediaResponse)
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful upload response (operation may have completed): %w", err)
	}
	return result, nil
}

// PrepareUpload obtains a part plan without transferring bytes. It can be composed with UploadPart,
// PrepareUpload 获取分片计划但不传输字节，可与 UploadPart、
// FinishUploadPart and UploadMedia(UploadID) for caller-managed resumability.
// FinishUploadPart 及 UploadMedia(UploadID) 组合，由调用方管理断点续传。
func (c *Client) PrepareUpload(ctx context.Context, target message.Target, input message.UploadPrepareRequest) (*message.UploadPrepareResult, error) {
	path, err := conversationPath(ctx, c, target)
	if err != nil {
		return nil, err
	}
	size, sizeErr := strconv.ParseInt(input.FileSize, 10, 64)
	validHash := func(value string, size int) bool {
		decoded, err := hex.DecodeString(value)
		return err == nil && len(decoded) == size
	}
	if sizeErr != nil || size <= 0 || strings.TrimSpace(input.FileName) == "" || input.FileType < 1 || input.FileType > 4 || !validHash(input.MD5, 16) || !validHash(input.SHA1, 20) || !validHash(input.MD5TenM, 16) {
		return nil, fmt.Errorf("%w: invalid prepare metadata or checksums", ErrInvalidMedia)
	}
	data, err := c.Post(ctx, path+"/upload_prepare", input)
	if err != nil {
		return nil, fmt.Errorf("prepare media upload: %w", err)
	}
	result := new(message.UploadPrepareResult)
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("decode upload plan: %w", err)
	}
	return result, nil
}

// FinishUploadPart acknowledges a successful PUT. PartIndex zero is sent explicitly.
// FinishUploadPart 确认一次成功的 PUT；PartIndex 为零时也显式发送。
func (c *Client) FinishUploadPart(ctx context.Context, target message.Target, input message.UploadPartFinishRequest) error {
	path, err := conversationPath(ctx, c, target)
	if err != nil {
		return err
	}
	size, sizeErr := strconv.ParseInt(input.BlockSize, 10, 64)
	hash, hashErr := hex.DecodeString(input.MD5)
	if input.UploadID == "" || input.PartIndex < 0 || sizeErr != nil || size <= 0 || hashErr != nil || len(hash) != md5.Size {
		return fmt.Errorf("%w: invalid part completion", ErrInvalidMedia)
	}
	_, err = c.Post(ctx, path+"/upload_part_finish", input)
	if err != nil {
		return fmt.Errorf("finish media upload part %d: %w", input.PartIndex, err)
	}
	return nil
}

// UploadPart PUTs exactly size bytes to a signed storage URL without QQ authentication or cookies.
// UploadPart 向签名存储 URL 精确 PUT 指定 size 字节，不携带 QQ 认证或 Cookie。
// It uses the injected HTTP transport/timeout, rejects redirects, and performs no retries.
// 使用注入的 HTTP 传输及超时，拒绝重定向，不执行重试。
// The reader is borrowed for this synchronous call; the caller retains resource ownership.
// 同步调用期间借用读取器，资源所有权仍归调用方。
func (c *Client) UploadPart(ctx context.Context, part message.UploadPart, reader io.Reader, size int64) error {
	if c == nil || ctx == nil || reader == nil || size <= 0 || part.Index < 0 {
		return fmt.Errorf("%w: invalid part transfer", ErrInvalidMedia)
	}
	_, err := c.putStorage(ctx, part.PresignedURL, reader, size, "application/octet-stream")
	if err != nil {
		return fmt.Errorf("upload part %d: %w", part.Index, err)
	}
	return nil
}

// putStorage is shared by conversation media and Markdown image materials.
// putStorage 由会话富媒体与 Markdown 图片素材共用。
// It returns an owned header value, fences borrowed readers, and never sends bot authentication.
// 返回独立持有的响应头值，对借用读取器设置访问边界，绝不发送机器人认证。
func (c *Client) putStorage(ctx context.Context, signedURL string, reader io.Reader, size int64, contentType string) (string, error) {
	u, err := url.Parse(signedURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", fmt.Errorf("%w: signed upload URL must use HTTPS", ErrInvalidMedia)
	}
	body := &uploadReader{reader: io.LimitReader(reader, size)}
	defer body.stop()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, body)
	if err != nil {
		return "", fmt.Errorf("%w: invalid upload request", ErrInvalidMedia)
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)
	storage := *c.httpClient
	storage.Jar = nil
	storage.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := storage.Do(req)
	if err != nil {
		// url.Error includes the signed URL; unwrap it before returning a diagnostic.
		// url.Error 包含签名 URL；返回诊断前先解包。
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return "", fmt.Errorf("storage upload: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &APIError{StatusCode: resp.StatusCode}
	}
	return resp.Header.Get("ETag"), nil
}

// A transport may finish asynchronously. Closing a response does not necessarily
// 传输可能异步结束；关闭响应不一定能
// stop all reads of its request, so fence the borrowed reader before returning it.
// 停止读取请求，因此归还借用读取器前需封闭访问。
type uploadReader struct {
	mu     sync.Mutex
	reader io.Reader
	closed bool
	active sync.WaitGroup
}

func (r *uploadReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return 0, io.EOF
	}
	r.active.Add(1)
	r.mu.Unlock()
	defer r.active.Done()
	return r.reader.Read(p)
}

func (r *uploadReader) stop() {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	r.active.Wait()
}

// FileUpload describes borrowed, stable random-access file data. UploadFile never closes Reader.
// FileUpload 描述借用且稳定的随机访问文件数据；UploadFile 从不关闭 Reader。
// The caller must not modify the file until the call returns. Size must match the available bytes.
// 调用返回前调用方不得修改文件；Size 必须与可用字节数匹配。
type FileUpload struct {
	Reader io.ReaderAt
	Size   int64
	Name   string
	Type   message.FileType
}

// UploadFile synchronously hashes, prepares, transfers and merges a local file using bounded memory.
// UploadFile 使用有界内存同步完成本地文件哈希、准备、传输及合并。
// It uses one worker, stops on the first error, and never retries or sends a message automatically.
// 使用单个工作协程，首次出错即停止，不自动重试或发送消息。
// Supply an operation deadline. For resumption or concurrency, compose the low-level upload methods instead.
// 请设置操作截止时间；需要续传或并发时，组合底层上传方法。
func (c *Client) UploadFile(ctx context.Context, target message.Target, file FileUpload) (*message.RichMediaResponse, error) {
	if _, err := conversationPath(ctx, c, target); err != nil {
		return nil, err
	}
	if file.Reader == nil || file.Size <= 0 || strings.TrimSpace(file.Name) == "" || file.Type < 1 || file.Type > 4 {
		return nil, fmt.Errorf("%w: reader, positive size, name and file type are required", ErrInvalidMedia)
	}
	fullMD5, fullSHA1, prefixMD5 := md5.New(), sha1.New(), md5.New()
	reader := io.NewSectionReader(file.Reader, 0, file.Size)
	buffer := make([]byte, 64<<10)
	var total int64
	for total < file.Size {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := reader.Read(buffer)
		if n > 0 {
			_, _ = fullMD5.Write(buffer[:n])
			_, _ = fullSHA1.Write(buffer[:n])
			if total < 10002432 {
				_, _ = prefixMD5.Write(buffer[:min(int64(n), 10002432-total)])
			}
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF && total == file.Size {
				break
			}
			return nil, fmt.Errorf("hash media: %w", err)
		}
		if n == 0 {
			return nil, fmt.Errorf("hash media: %w", io.ErrNoProgress)
		}
	}
	plan, err := c.PrepareUpload(ctx, target, message.UploadPrepareRequest{FileType: file.Type, FileName: file.Name, FileSize: strconv.FormatInt(file.Size, 10), MD5: hex.EncodeToString(fullMD5.Sum(nil)), SHA1: hex.EncodeToString(fullSHA1.Sum(nil)), MD5TenM: hex.EncodeToString(prefixMD5.Sum(nil))})
	if err != nil {
		return nil, err
	}
	block, err := strconv.ParseInt(plan.BlockSize, 10, 64)
	if err != nil || block <= 0 || plan.UploadID == "" || int64(len(plan.Parts)) != 1+(file.Size-1)/block {
		return nil, fmt.Errorf("%w: inconsistent upload plan", ErrInvalidMedia)
	}
	// Validate the entire manifest before transferring any bytes; response order is not significant.
	// 传输字节前校验完整清单；响应顺序不影响结果。
	// The documented indices start at zero. Also accept a complete one-based plan,
	// 文档索引从零开始；也接受完整的从一开始的计划，
	// keeping the server's index for completion and using only its slot for file offsets.
	// 完成确认保留服务端索引，仅使用其位置计算文件偏移。
	indexBase := 1
	for _, part := range plan.Parts {
		if part.Index == 0 {
			indexBase = 0
			break
		}
	}
	parts := make([]message.UploadPart, len(plan.Parts))
	seen := make([]bool, len(parts))
	for _, part := range plan.Parts {
		slot := part.Index - indexBase
		if slot < 0 || slot >= len(parts) {
			return nil, fmt.Errorf("%w: part index %d outside %d..%d for %d parts", ErrInvalidMedia, part.Index, indexBase, indexBase+len(parts)-1, len(parts))
		}
		if seen[slot] {
			return nil, fmt.Errorf("%w: duplicate part index %d in %d parts", ErrInvalidMedia, part.Index, len(parts))
		}
		size := min(block, file.Size-int64(slot)*block)
		partSize, err := strconv.ParseInt(part.BlockSize, 10, 64)
		u, urlErr := url.Parse(part.PresignedURL)
		if err != nil || size != partSize || urlErr != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return nil, fmt.Errorf("%w: inconsistent part metadata", ErrInvalidMedia)
		}
		seen[slot] = true
		parts[slot] = part
	}
	for slot, part := range parts {
		offset := int64(slot) * block
		size := min(block, file.Size-offset)
		hash := md5.New()
		// Transports may return a response before they finish reading the request body.
		// 传输可能在请求正文读取结束前就返回响应。
		// Hash locally first so no hash state is shared with an HTTP transport goroutine.
		// 先在本地计算哈希，避免与 HTTP 传输协程共享哈希状态。
		section := io.NewSectionReader(file.Reader, offset, size)
		var hashed int64
		for hashed < size {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			n, readErr := section.Read(buffer)
			if n > 0 {
				_, _ = hash.Write(buffer[:n])
				hashed += int64(n)
			}
			if readErr != nil {
				if readErr == io.EOF && hashed == size {
					break
				}
				return nil, fmt.Errorf("hash media part %d: %w", part.Index, readErr)
			}
			if n == 0 {
				return nil, fmt.Errorf("hash media part %d: %w", part.Index, io.ErrNoProgress)
			}
		}
		if err := c.UploadPart(ctx, part, io.NewSectionReader(file.Reader, offset, size), size); err != nil {
			return nil, err
		}
		if err := c.FinishUploadPart(ctx, target, message.UploadPartFinishRequest{UploadID: plan.UploadID, PartIndex: part.Index, BlockSize: strconv.FormatInt(size, 10), MD5: hex.EncodeToString(hash.Sum(nil))}); err != nil {
			return nil, fmt.Errorf("finish media part %d: %w", part.Index, err)
		}
	}
	return c.UploadMedia(ctx, target, message.RichMediaRequest{FileType: file.Type, UploadID: plan.UploadID, FileName: file.Name})
}
