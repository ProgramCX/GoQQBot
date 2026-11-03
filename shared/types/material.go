package types

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// UploadMarkdownImage uploads a reusable image owned by this bot, without sending or requiring a message event.
// UploadMarkdownImage 上传机器人持有的可复用图片，不发送消息，也不要求消息事件。
// For background business services, use Client.UploadMarkdownImage directly.
// 后台业务服务可直接使用 Client.UploadMarkdownImage。
func (c *Context) UploadMarkdownImage(ctx context.Context, file api.ImageUpload) (*message.ImageMaterial, error) {
	if c == nil || c.Client == nil {
		return nil, ErrInvalidReplyContext
	}
	return c.Client.UploadMarkdownImage(ctx, file)
}

// UploadMarkdownImageFile opens, uploads and closes a local image without sending it.
// UploadMarkdownImageFile 打开、上传并关闭本地图片，不发送图片。
func (c *Context) UploadMarkdownImageFile(ctx context.Context, path string) (*message.ImageMaterial, error) {
	if c == nil || c.Client == nil {
		return nil, ErrInvalidReplyContext
	}
	return c.Client.UploadMarkdownImageFile(ctx, path)
}
