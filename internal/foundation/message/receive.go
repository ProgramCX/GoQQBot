// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package message

type ContentType string

const (
	IMAGE_JPEG ContentType = "image/jpeg"
	IMAGE_PNG  ContentType = "image/png"
	IMAGE_GIF  ContentType = "image/gif"
	FILE       ContentType = "file"
	VIDEO_MP4  ContentType = "video/mp4"
	VOICE      ContentType = "voice"
)

type CommonReceiveMessage struct {
	Id string `json:"id"`
	//Author      Author       `json:"author"`
	Content     string       `json:"content"`
	Timestamp   string       `json:"timestamp"`
	Attachments []Attachment `json:"attachments"`
}

type GroupReceiveMessage struct {
	CommonReceiveMessage
	Author MemberAuthor `json:"author"`
}

type PrivateReceiveMessage struct {
	CommonReceiveMessage
	Author UserAuthor `json:"author"`
}

type UserAuthor struct {
	UserOpenid string `json:"user_openid"`
}

type MemberAuthor struct {
	MemberOpenid string `json:"member_openid"`
}
type Attachment struct {
	ContentType ContentType `json:"content_type"`
	Filename    string      `json:"filename"`
	Height      int         `json:"height"`
	Width       int         `json:"width"`
	Size        int         `json:"size"`
	Url         string      `json:"url"`
}
