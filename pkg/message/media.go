// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package message

// FileType identifies the uploaded media category. Oversize media may be downgraded by QQ to files.
// FileType 标识上传媒体类别；QQ 可能将过大媒体降级为文件。
type FileType int

const (
	FileTypeImage FileType = 1 // png/jpg
	// PNG/JPG 图片。
	FileTypeVideo FileType = 2 // mp4
	// MP4 视频。
	FileTypeVoice FileType = 3 // silk
	// SILK 语音。
	FileTypeFile FileType = 4 // arbitrary files
	// 任意文件。
)

// RichMediaRequest selects exactly one source: URL, UploadID, or protocol FileData.
// RichMediaRequest 必须从 URL、UploadID 或协议 FileData 中选择唯一来源。
// SrvSendMsg explicitly opts into proactive sending; its zero value only uploads.
// SrvSendMsg 显式启用主动发送，零值仅上传。
// FileData is forwarded as-is: QQ documents a string and a 5M limit without specifying its encoding.
// FileData 原样转发：QQ 文档规定它为字符串且上限 5M，但未指定编码。
// Prefer UploadFile's unambiguous binary multipart flow for local files.
// 本地文件优先使用 UploadFile 的明确二进制分片上传流程。
type RichMediaRequest struct {
	FileType   FileType `json:"file_type"`
	URL        string   `json:"url,omitempty"`
	SrvSendMsg bool     `json:"srv_send_msg"`
	UploadID   string   `json:"upload_id,omitempty"`
	FileName   string   `json:"file_name,omitempty"`
	FileData   string   `json:"file_data,omitempty"`
}

// RichMediaResponse contains opaque file_info, its lifetime in seconds (zero means long-lived), and optional send ID.
// RichMediaResponse 包含不透明 file_info、以秒计的有效期（零表示长期）及可选发送 ID。
type RichMediaResponse struct {
	FileUUID string `json:"file_uuid"`
	FileInfo string `json:"file_info"`
	TTL      int    `json:"ttl"`
	ID       string `json:"id,omitempty"`
	RawURL   string `json:"raw_url,omitempty"`
}

// UploadPrepareRequest contains protocol checksums encoded as lowercase hexadecimal.
// UploadPrepareRequest 包含以小写十六进制编码的协议校验和。
// MD5TenM covers the first 10002432 bytes, not 10 MiB. Sizes are decimal strings on the wire.
// MD5TenM 覆盖前 10002432 字节而非 10 MiB；大小在协议中编码为十进制字符串。
type UploadPrepareRequest struct {
	FileType FileType `json:"file_type"`
	FileSize string   `json:"file_size"`
	FileName string   `json:"file_name"`
	MD5      string   `json:"md5"`
	SHA1     string   `json:"sha1"`
	MD5TenM  string   `json:"md5_10m"`
}

// UploadPart describes a server-assigned part index and its temporary signed storage URL. Do not log it.
// UploadPart 描述服务端分配的分片索引及临时签名存储 URL，不得记录该 URL。
// Documentation specifies zero-based indices; UploadFile also accepts complete one-based plans.
// 文档规定索引从零开始，UploadFile 也接受完整的从一开始的计划。
// Preserve Index unchanged when acknowledging a transferred part.
// 确认已传输分片时保留 Index 不变。
type UploadPart struct {
	Index        int    `json:"index"`
	PresignedURL string `json:"presigned_url"`
	BlockSize    string `json:"block_size"`
}

// UploadConfig describes server recommendations. The synchronous UploadFile helper uses one worker and no retries.
// UploadConfig 描述服务端建议，同步 UploadFile 仅使用一个工作协程且不重试。
type UploadConfig struct {
	Concurrency  int `json:"concurrency"`
	RetryTimeout int `json:"retry_timeout"`
	RetryDelay   int `json:"retry_delay"`
}

// UploadPrepareResult is an upload plan. Keep it to implement explicit resumable uploads with the low-level APIs.
// UploadPrepareResult 是上传计划，可结合底层 API 实现显式断点续传。
type UploadPrepareResult struct {
	UploadID     string       `json:"upload_id"`
	BlockSize    string       `json:"block_size"`
	Parts        []UploadPart `json:"parts"`
	UploadConfig UploadConfig `json:"upload_config"`
}

// UploadPartFinishRequest acknowledges one successfully transferred part.
// UploadPartFinishRequest 确认一个已成功传输的分片。
type UploadPartFinishRequest struct {
	UploadID  string `json:"upload_id"`
	PartIndex int    `json:"part_index"`
	BlockSize string `json:"block_size"`
	MD5       string `json:"md5"`
}

// Target addresses a conversation independently of an incoming event or reply source.
// Target 独立于接收事件或回复来源，指定目标会话。
type Target struct {
	Scene Scene
	ID    string
}
