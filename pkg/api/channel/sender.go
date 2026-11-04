// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelMsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/pkg/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func sendToChannel(c *types.Context, channelId string, body *channelMsg.ChannelSendMessage) (*message.SendMessageResponse, error) {
	return sendLegacy(c, api.CHANNEL_SEND_MESSAGE, channelId, body)
}

func sendToDM(c *types.Context, guildId string, body *channelMsg.ChannelSendMessage) (*message.SendMessageResponse, error) {
	return sendLegacy(c, api.CHANNEL_PRIVATE_MESSAGE, guildId, body)
}

func sendLegacy(c *types.Context, route, id string, body *channelMsg.ChannelSendMessage) (*message.SendMessageResponse, error) {
	if c == nil || body == nil {
		return nil, fmt.Errorf("event context and message body are required")
	}
	path, err := apiwire.Path(route, id)
	if err != nil {
		return nil, err
	}
	data, err := c.Client.Post(c.RequestContext(), path, body)
	if err != nil {
		return nil, fmt.Errorf("send channel message failed: %w", err)
	}
	var response struct {
		ID        string          `json:"id"`
		Timestamp json.RawMessage `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode successful channel send (message may have been sent): %w", err)
	}
	result := &message.SendMessageResponse{ID: response.ID}
	if len(response.Timestamp) == 0 {
		return result, nil
	}
	if response.Timestamp[0] == '"' {
		var timestamp time.Time
		if err := json.Unmarshal(response.Timestamp, &timestamp); err != nil {
			return nil, fmt.Errorf("decode channel send timestamp (message may have been sent): %w", err)
		}
		result.Timestamp = timestamp.Unix()
	} else if err := json.Unmarshal(response.Timestamp, &result.Timestamp); err != nil {
		return nil, fmt.Errorf("decode channel send timestamp (message may have been sent): %w", err)
	}
	return result, nil
}

func SendMessage(c *types.Context, channelId string, msg *channelMsg.ChannelSendMessage) (*message.SendMessageResponse, error) {
	return sendToChannel(c, channelId, msg)
}

func SendDM(c *types.Context, guildId string, msg *channelMsg.ChannelSendMessage) (*message.SendMessageResponse, error) {
	return sendToDM(c, guildId, msg)
}

func SendPlainText(c *types.Context, channelId string, content string, msgId string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{Content: content, MsgId: msgId}
	return sendToChannel(c, channelId, body)
}

func SendDMPlainText(c *types.Context, guildId string, content string, msgId string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{Content: content, MsgId: msgId}
	return sendToDM(c, guildId, body)
}

func SendMarkdownContent(c *types.Context, channelId string, content string, msgId string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Markdown: &typesMessage.Markdown{Content: content},
		MsgId:    msgId,
	}
	return sendToChannel(c, channelId, body)
}

func SendMarkdownTemplate(c *types.Context, channelId string, templateId string, params []typesMessage.MarkdownParams, msgId string) (*message.SendMessageResponse, error) {
	body := &channelMsg.ChannelSendMessage{
		Markdown: &typesMessage.Markdown{
			CustomTemplateId: templateId,
			Params:           params,
		},
		MsgId: msgId,
	}
	return sendToChannel(c, channelId, body)
}

func SendMarkdownWithKeyboard(c *types.Context, channelId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{Markdown: md, Keyboard: kb, MsgId: msgId}
	return sendToChannel(c, channelId, body)
}

func SendDMMarkdownWithKeyboard(c *types.Context, guildId string, md *typesMessage.Markdown, kb *interaction.Keyboard, msgId string) (*message.SendMessageResponse, error) {
	if md == nil {
		return nil, fmt.Errorf("markdown is required when sending keyboard")
	}
	if kb == nil {
		return nil, fmt.Errorf("keyboard is required")
	}
	body := &channelMsg.ChannelSendMessage{Markdown: md, Keyboard: kb, MsgId: msgId}
	return sendToDM(c, guildId, body)
}

func SendArk(c *types.Context, channelId string, ark *typesMessage.Ark, msgId string) (*message.SendMessageResponse, error) {
	if ark == nil {
		return nil, fmt.Errorf("ark is required")
	}
	body := &channelMsg.ChannelSendMessage{Ark: ark, MsgId: msgId}
	return sendToChannel(c, channelId, body)
}

func SendEmbed(c *types.Context, channelId string, embed *channelMsg.MessageEmbed, msgId string) (*message.SendMessageResponse, error) {
	if embed == nil {
		return nil, fmt.Errorf("embed is required")
	}
	body := &channelMsg.ChannelSendMessage{Embed: embed, MsgId: msgId}
	return sendToChannel(c, channelId, body)
}

func RecallChannelMessage(c *types.Context, channelId string, messageId string, hideTip bool) error {
	if c == nil {
		return fmt.Errorf("event context is required")
	}
	return NewClient(c.Client).RecallChannelMessage(c.RequestContext(), channelId, messageId, RecallOptions{HideTip: hideTip})
}

func RecallDM(c *types.Context, guildId string, messageId string, hideTip bool) error {
	if c == nil {
		return fmt.Errorf("event context is required")
	}
	return NewClient(c.Client).RecallDirectMessage(c.RequestContext(), guildId, messageId, RecallOptions{HideTip: hideTip})
}
