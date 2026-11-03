// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package c2c

import (
	"encoding/json"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/pkg/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func GetPrivateReceiveMessage(payload types.Payload) (*message.PrivateReceiveMessage, error) {
	pm := message.PrivateReceiveMessage{}
	err := json.Unmarshal(payload.D, &pm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal private receive message failed: %w", err)
	}
	return &pm, nil
}

func buildReply(c *types.Context, msgType message.TypeMessage) (*message.PrivateSendMessage, error) {
	pm, err := GetPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return nil, err
	}
	return &message.PrivateSendMessage{
		CommonSendMessage: message.CommonSendMessage{
			MsgType: msgType,
			MsgId:   pm.Id,
		},
	}, nil
}

func send(c *types.Context, body *message.PrivateSendMessage) error {
	pm, err := GetPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(api.C2C_SEND_MESSAGE, pm.Author.UserOpenid)
	_, err = c.Client.Post(c.RequestContext(), path, body)
	if err != nil {
		return fmt.Errorf("send private message failed: %w", err)
	}
	return nil
}

func ReplyPlainText(c *types.Context, text string) error {
	body, err := buildReply(c, message.PLAINTEXT)
	if err != nil {
		return err
	}
	body.Content = text
	return send(c, body)
}

func ReplyMarkdownContent(c *types.Context, content string) error {
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = &typesMessage.Markdown{Content: content}
	return send(c, body)
}

func ReplyMarkdownTemplate(c *types.Context, templateId string, params []typesMessage.MarkdownParams) error {
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = &typesMessage.Markdown{
		CustomTemplateId: templateId,
		Params:           params,
	}
	return send(c, body)
}

func ReplyMarkdownWithKeyboard(c *types.Context, md *typesMessage.Markdown, kb *interaction.Keyboard) error {
	if md == nil {
		return fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return fmt.Errorf("keyboard is required")
	}
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = md
	body.KeyBoards = kb
	return send(c, body)
}

func ReplyArk(c *types.Context, ark *typesMessage.Ark) error {
	if ark == nil {
		return fmt.Errorf("ark is required")
	}
	body, err := buildReply(c, message.ARK)
	if err != nil {
		return err
	}
	body.Ark = ark
	return send(c, body)
}

func ReplyMedia(c *types.Context, fileInfo string) error {
	body, err := buildReply(c, message.MEDIA)
	if err != nil {
		return err
	}
	body.Media = &message.MediaInfo{FileInfo: fileInfo}
	return send(c, body)
}

// UploadMedia uploads via the shared scene-independent engine.
// UploadMedia 通过不依赖场景的共用实现上传。
func UploadMedia(c *types.Context, fileType message.FileType, mediaURL string, srvSendMsg bool) (*message.RichMediaResponse, error) {
	if c == nil {
		return nil, types.ErrInvalidReplyContext
	}
	return c.UploadMedia(c.RequestContext(), message.RichMediaRequest{FileType: fileType, URL: mediaURL, SrvSendMsg: srvSendMsg})
}

// RecallMessage recalls an explicit message ID through the shared API engine.
// RecallMessage 通过共用 API 实现撤回明确的消息 ID。
func RecallMessage(c *types.Context, messageID string) error { return c.Recall(messageID) }

func SendWakeup(c *types.Context, userOpenid string, text string, msgType message.TypeMessage) error {
	body := &message.PrivateSendMessage{
		CommonSendMessage: message.CommonSendMessage{
			Content: text,
			MsgType: msgType,
		},
		IsWakeup: true,
	}
	path := fmt.Sprintf(api.C2C_SEND_MESSAGE, userOpenid)
	_, err := c.Client.Post(c.RequestContext(), path, body)
	if err != nil {
		return fmt.Errorf("send wakeup message failed: %w", err)
	}
	return nil
}
