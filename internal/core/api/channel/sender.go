// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package channel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ProgramCX/GoQQBot/internal/core/api"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message"
	channelMsg "github.com/ProgramCX/GoQQBot/internal/foundation/message/channel"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/internal/foundation/message/types"
)

// sendToChannel sends a ChannelSendMessage to the specified text channel.
func sendToChannel(channelId string, body *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	path := fmt.Sprintf(api.CHANNEL_SEND_MESSAGE, channelId)
	logger.DEBUG("send channel message to %s: %+v", channelId, body)
	data, err := api.NewClient(configPath).Post(context.Background(), path, body)
	if err != nil {
		return nil, fmt.Errorf("send channel message failed: %w", err)
	}
	var resp message.SendMessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal send response failed: %w", err)
	}
	return &resp, nil
}

// sendToDM sends a ChannelSendMessage as a DM (频道私信).
func sendToDM(guildId string, body *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	path := fmt.Sprintf(api.CHANNEL_PRIVATE_MESSAGE, guildId)
	logger.DEBUG("send DM to guild %s: %+v", guildId, body)
	data, err := api.NewClient(configPath).Post(context.Background(), path, body)
	if err != nil {
		return nil, fmt.Errorf("send DM failed: %w", err)
	}
	var resp message.SendMessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal send response failed: %w", err)
	}
	return &resp, nil
}

// ──────────────────────────────────────────────
// Low-level: send custom ChannelSendMessage
// ──────────────────────────────────────────────

// SendMessage sends a custom message to a text channel.
func SendMessage(channelId string, msg *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	return sendToChannel(channelId, msg, configPath)
}

// SendDM sends a custom message as a DM (频道私信).
func SendDM(guildId string, msg *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	return sendToDM(guildId, msg, configPath)
}

// ──────────────────────────────────────────────
// Plain Text
// ──────────────────────────────────────────────

// SendPlainText sends a plain text message to a text channel.
// msgId is the received message ID for passive (reply) messages; pass "" for active messages.
func SendPlainText(channelId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Content: content,
		MsgId:   msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// SendDMPlainText sends a plain text DM to a guild member.
func SendDMPlainText(guildId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Content: content,
		MsgId:   msgId,
	}
	return sendToDM(guildId, body, configPath)
}

// ──────────────────────────────────────────────
// Markdown
// ──────────────────────────────────────────────

// SendMarkdownContent sends a custom markdown content message to a text channel.
func SendMarkdownContent(channelId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Markdown: &typesMessage.Markdown{Content: content},
		MsgId:    msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// SendMarkdownTemplate sends a markdown template message to a text channel.
func SendMarkdownTemplate(channelId string, templateId string, params []typesMessage.MarkdownParams, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Markdown: &typesMessage.Markdown{
			CustomTemplateId: templateId,
			Params:           params,
		},
		MsgId: msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// ──────────────────────────────────────────────
// Markdown + Keyboard
// ──────────────────────────────────────────────

// SendMarkdownWithKeyboard sends markdown + keyboard to a text channel.
// Keyboard must be combined with markdown per API requirement.
//
// Both markdown and keyboard support two modes:
//   - Template mode: set Markdown.CustomTemplateId / Keyboard.ID
//   - Custom mode:   set Markdown.Content / Keyboard.Content
func SendMarkdownWithKeyboard(channelId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{
		Markdown: md,
		Keyboard: kb,
		MsgId:    msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// SendDMMarkdownWithKeyboard sends markdown + keyboard as a DM.
func SendDMMarkdownWithKeyboard(guildId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{
		Markdown: md,
		Keyboard: kb,
		MsgId:    msgId,
	}
	return sendToDM(guildId, body, configPath)
}

// ──────────────────────────────────────────────
// Ark
// ──────────────────────────────────────────────

// SendArk sends an Ark template message to a text channel.
func SendArk(channelId string, ark *typesMessage.Ark, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if ark == nil {
		return nil, fmt.Errorf("ark is required")
	}
	body := &channelMsg.ChannelSendMessage{
		Ark:   ark,
		MsgId: msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// ──────────────────────────────────────────────
// Embed
// ──────────────────────────────────────────────

// SendEmbed sends an Embed message to a text channel.
func SendEmbed(channelId string, embed *channelMsg.MessageEmbed, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if embed == nil {
		return nil, fmt.Errorf("embed is required")
	}
	body := &channelMsg.ChannelSendMessage{
		Embed: embed,
		MsgId: msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

// ──────────────────────────────────────────────
// Recall (DELETE)
// ──────────────────────────────────────────────

// RecallChannelMessage recalls a message in a text channel.
// hideTip controls whether the "message recalled" tip is shown.
func RecallChannelMessage(channelId string, messageId string, hideTip bool, configPath string) error {
	path := fmt.Sprintf(api.CHANNEL_DELETE_MESSAGE, channelId, messageId)
	if hideTip {
		path += "?hidetip=true"
	}
	_, err := api.NewClient(configPath).Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("recall channel message failed: %w", err)
	}
	return nil
}

// RecallDM recalls a DM message (频道私信).
// hideTip controls whether the "message recalled" tip is shown.
func RecallDM(guildId string, messageId string, hideTip bool, configPath string) error {
	path := fmt.Sprintf(api.CHANNEL_PRIVATE_DELETE, guildId, messageId)
	if hideTip {
		path += "?hidetip=true"
	}
	_, err := api.NewClient(configPath).Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("recall DM failed: %w", err)
	}
	return nil
}
