package types

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// Target returns the current replyable conversation without a reply source.
// Target 返回当前可回复会话，不包含回复来源。
// It also supports button/menu interactions; it does not perform I/O.
// 也支持按钮或菜单交互，不执行 I/O。
func (c *Context) Target() (message.Target, error) {
	target, err := c.replyTarget()
	if err != nil {
		return message.Target{}, err
	}
	return message.Target{Scene: target.Scene, ID: target.ID}, nil
}

// Recall deletes an explicit message ID in the current conversation using the handler context.
// Recall 使用处理函数上下文，删除当前会话中明确指定的消息 ID。
// Use the send result's ID for bot replies, never a REFIDX. It does not implicitly recall the incoming message.
// 撤回机器人回复时使用发送结果 ID，绝不能用 REFIDX；不隐式撤回接收消息。
func (c *Context) Recall(messageID string) error {
	if c == nil {
		return ErrInvalidReplyContext
	}
	return c.RecallContext(c.RequestContext(), messageID)
}

// RecallContext is the cancellable form of Recall. QQ enforces recall permissions and expiry.
// RecallContext 是可取消的 Recall，QQ 检查撤回权限及有效期。
func (c *Context) RecallContext(ctx context.Context, messageID string) error {
	target, err := c.Target()
	if err != nil {
		return err
	}
	return c.Client.RecallMessage(ctx, target, messageID)
}

// UploadMedia uploads into this C2C/group conversation. The zero SrvSendMsg only uploads.
// UploadMedia 上传到当前单聊或群聊会话，SrvSendMsg 为零值时仅上传。
func (c *Context) UploadMedia(ctx context.Context, input message.RichMediaRequest) (*message.RichMediaResponse, error) {
	target, err := c.Target()
	if err != nil {
		return nil, err
	}
	return c.Client.UploadMedia(ctx, target, input)
}

// UploadFile performs a bounded-memory multipart upload into the current C2C/group conversation.
// UploadFile 使用有界内存分片上传到当前单聊或群聊会话。
// It neither closes the caller's reader nor sends the uploaded media.
// 不关闭调用方读取器，也不发送上传的媒体。
func (c *Context) UploadFile(ctx context.Context, file api.FileUpload) (*message.RichMediaResponse, error) {
	target, err := c.Target()
	if err != nil {
		return nil, err
	}
	return c.Client.UploadFile(ctx, target, file)
}
