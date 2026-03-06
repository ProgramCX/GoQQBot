// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package message

// FileType represents the type of rich media file.
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/send-receive/rich-media.html
type FileType int

const (
	FileTypeImage FileType = 1 // png/jpg
	FileTypeVideo FileType = 2 // mp4
	FileTypeVoice FileType = 3 // silk
	FileTypeFile  FileType = 4 // not yet supported
)

// RichMediaRequest is the request body for uploading rich media files.
// POST /v2/users/{openid}/files (C2C)
// POST /v2/groups/{group_openid}/files (Group)
type RichMediaRequest struct {
	FileType   FileType `json:"file_type"`
	URL        string   `json:"url"`
	SrvSendMsg bool     `json:"srv_send_msg"`
}

// RichMediaResponse is the response from uploading rich media files.
type RichMediaResponse struct {
	FileUUID string `json:"file_uuid"`
	FileInfo string `json:"file_info"`
	TTL      int    `json:"ttl"`
	ID       string `json:"id,omitempty"` // Only returned when srv_send_msg is true
}
