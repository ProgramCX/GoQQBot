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
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelMsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/pkg/message/types"
)

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

func SendMessage(channelId string, msg *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	return sendToChannel(channelId, msg, configPath)
}

func SendDM(guildId string, msg *channelMsg.ChannelSendMessage, configPath string) (*message.SendMessageResponse, error) {
	return sendToDM(guildId, msg, configPath)
}

func SendPlainText(channelId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{Content: content, MsgId: msgId}
	return sendToChannel(channelId, body, configPath)
}

func SendDMPlainText(guildId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{Content: content, MsgId: msgId}
	return sendToDM(guildId, body, configPath)
}

func SendMarkdownContent(channelId string, content string, msgId string, configPath string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Markdown: &typesMessage.Markdown{Content: content},
		MsgId:    msgId,
	}
	return sendToChannel(channelId, body, configPath)
}

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

func SendMarkdownWithKeyboard(channelId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{Markdown: md, Keyboard: kb, MsgId: msgId}
	return sendToChannel(channelId, body, configPath)
}

func SendDMMarkdownWithKeyboard(guildId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{Markdown: md, Keyboard: kb, MsgId: msgId}
	return sendToDM(guildId, body, configPath)
}

func SendArk(channelId string, ark *typesMessage.Ark, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if ark == nil {
		return nil, fmt.Errorf("ark is required")
	}
	body := &channelMsg.ChannelSendMessage{Ark: ark, MsgId: msgId}
	return sendToChannel(channelId, body, configPath)
}

func SendEmbed(channelId string, embed *channelMsg.MessageEmbed, msgId string, configPath string) (*message.SendMessageResponse, error) {
	if embed == nil {
		return nil, fmt.Errorf("embed is required")
	}
	body := &channelMsg.ChannelSendMessage{Embed: embed, MsgId: msgId}
	return sendToChannel(channelId, body, configPath)
}

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
