// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package message

type FileType int

const (
	FileTypeImage FileType = 1 // png/jpg
	FileTypeVideo FileType = 2 // mp4
	FileTypeVoice FileType = 3 // silk
	FileTypeFile  FileType = 4 // not yet supported
)

type RichMediaRequest struct {
	FileType   FileType `json:"file_type"`
	URL        string   `json:"url"`
	SrvSendMsg bool     `json:"srv_send_msg"`
}

type RichMediaResponse struct {
	FileUUID string `json:"file_uuid"`
	FileInfo string `json:"file_info"`
	TTL      int    `json:"ttl"`
	ID       string `json:"id,omitempty"`
}
