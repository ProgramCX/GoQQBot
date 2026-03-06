// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package c2c

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ProgramCX/GoQQBot/internal/core/api"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/internal/foundation/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// getPrivateReceiveMessage parses the payload to extract the private receive message.
func getPrivateReceiveMessage(payload types.Payload) (*message.PrivateReceiveMessage, error) {
	pm := message.PrivateReceiveMessage{}
	err := json.Unmarshal(payload.D, &pm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal private receive message failed: %w", err)
	}
	return &pm, nil
}

// buildReply creates a base PrivateSendMessage from the context, automatically setting msg_id and event_id.
func buildReply(c *types.Context, msgType message.TypeMessage) (*message.PrivateSendMessage, error) {
	pm, err := getPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return nil, err
	}
	return &message.PrivateSendMessage{
		CommonSendMessage: message.CommonSendMessage{
			MsgType:   msgType,
			QQEventID: c.Payload.T,
			MsgId:     pm.Id,
		},
	}, nil
}

// send sends a PrivateSendMessage to the user extracted from the context.
func send(c *types.Context, body *message.PrivateSendMessage, configPath string) error {
	pm, err := getPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(api.C2C_SEND_MESSAGE, pm.Author.UserOpenid)
	logger.DEBUG("send private message: %+v", body)
	_, err = api.NewClient(configPath).Post(context.Background(), path, body)
	if err != nil {
		return fmt.Errorf("send private message failed: %w", err)
	}
	return nil
}

// ──────────────────────────────────────
// Plain Text (msg_type = 0)
// ──────────────────────────────────────

// ReplyPlainText replies with a plain text message.
func ReplyPlainText(c *types.Context, text string, configPath string) error {
	body, err := buildReply(c, message.PLAINTEXT)
	if err != nil {
		return err
	}
	body.Content = text
	return send(c, body, configPath)
}

// ──────────────────────────────────────
// Markdown (msg_type = 2)
// ──────────────────────────────────────

// ReplyMarkdownContent replies with a custom markdown content message.
// Use this when sending raw markdown text (custom mode).
func ReplyMarkdownContent(c *types.Context, content string, configPath string) error {
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = &typesMessage.Markdown{Content: content}
	return send(c, body, configPath)
}

// ReplyMarkdownTemplate replies with a markdown template message.
// Use this when referencing a pre-registered markdown template (template mode).
func ReplyMarkdownTemplate(c *types.Context, templateId string, params []typesMessage.MarkdownParams, configPath string) error {
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = &typesMessage.Markdown{
		CustomTemplateId: templateId,
		Params:           params,
	}
	return send(c, body, configPath)
}

// ──────────────────────────────────────
// Markdown + Keyboard (msg_type = 2)
// ──────────────────────────────────────

// ReplyMarkdownWithKeyboard replies with markdown + keyboard.
// Keyboard must be combined with markdown per API requirement.
//
// Both markdown and keyboard support two modes:
//   - Template mode: set Markdown.CustomTemplateId / Keyboard.ID
//   - Custom mode:   set Markdown.Content / Keyboard.Content
func ReplyMarkdownWithKeyboard(c *types.Context, md *typesMessage.Markdown, kb *interaction.Keyboard, configPath string) error {
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
	return send(c, body, configPath)
}

// ──────────────────────────────────────
// Ark (msg_type = 3)
// ──────────────────────────────────────

// ReplyArk replies with an Ark template message.
// Supported template IDs: 23 (link+text), 24 (text+thumbnail), 37 (big image).
func ReplyArk(c *types.Context, ark *typesMessage.Ark, configPath string) error {
	if ark == nil {
		return fmt.Errorf("ark is required")
	}
	body, err := buildReply(c, message.ARK)
	if err != nil {
		return err
	}
	body.Ark = ark
	return send(c, body, configPath)
}

// ──────────────────────────────────────
// Rich Media (msg_type = 7)
// ──────────────────────────────────────

// ReplyMedia replies with a rich media message using file_info obtained from UploadMedia.
// Call UploadMedia first (with srvSendMsg=false) to get the file_info, then pass it here.
func ReplyMedia(c *types.Context, fileInfo string, configPath string) error {
	body, err := buildReply(c, message.MEDIA)
	if err != nil {
		return err
	}
	body.Media = &message.MediaInfo{FileInfo: fileInfo}
	return send(c, body, configPath)
}

// UploadMedia uploads rich media (image/video/voice/file) for use in private chat.
//
// When srvSendMsg is true, the media is sent directly to the user (occupies active message quota).
// When srvSendMsg is false, returns file_info for use with ReplyMedia.
//
// Note: files uploaded via C2C endpoint can only be sent in C2C (private) chats.
func UploadMedia(c *types.Context, fileType message.FileType, url string, srvSendMsg bool, configPath string) (*message.RichMediaResponse, error) {
	pm, err := getPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(api.C2C_FILE_MESSAGE, pm.Author.UserOpenid)
	req := &message.RichMediaRequest{
		FileType:   fileType,
		URL:        url,
		SrvSendMsg: srvSendMsg,
	}
	data, err := api.NewClient(configPath).Post(context.Background(), path, req)
	if err != nil {
		return nil, fmt.Errorf("upload media failed: %w", err)
	}
	var resp message.RichMediaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload media response failed: %w", err)
	}
	return &resp, nil
}

// ──────────────────────────────────────
// Recall (DELETE)
// ──────────────────────────────────────

// RecallMessage recalls a message sent by the bot in a private chat.
// messageId is the message ID returned by the send API.
// Messages older than 2 minutes cannot be recalled.
func RecallMessage(c *types.Context, messageId string, configPath string) error {
	pm, err := getPrivateReceiveMessage(*c.Payload)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(api.C2C_DELETE_MESSAGE, pm.Author.UserOpenid, messageId)
	_, err = api.NewClient(configPath).Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("recall private message failed: %w", err)
	}
	return nil
}

// ──────────────────────────────────────
// Wakeup (is_wakeup = true)
// ──────────────────────────────────────

// SendWakeup sends an interaction recall (wakeup) message to a user.
// This is a proactive message that does NOT require a prior event context.
// is_wakeup is mutually exclusive with msg_id and event_id.
//
// Requires the user to have previously messaged the bot within 30 days.
// Wakeup messages are limited to 1 per period (same day, 1-3 days, 3-7 days, 7-30 days).
func SendWakeup(userOpenid string, text string, msgType message.TypeMessage, configPath string) error {
	body := &message.PrivateSendMessage{
		CommonSendMessage: message.CommonSendMessage{
			Content: text,
			MsgType: msgType,
		},
		IsWakeup: true,
	}
	path := fmt.Sprintf(api.C2C_SEND_MESSAGE, userOpenid)
	logger.DEBUG("send wakeup message to %s: %+v", userOpenid, body)
	_, err := api.NewClient(configPath).Post(context.Background(), path, body)
	if err != nil {
		return fmt.Errorf("send wakeup message failed: %w", err)
	}
	return nil
}
