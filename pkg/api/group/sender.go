// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package group

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ProgramCX/GoQQBot/internal/core/api"
	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	typesMessage "github.com/ProgramCX/GoQQBot/pkg/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func getGroupReceiveMessage(payload types.Payload) (*message.GroupReceiveMessage, error) {
	gm := message.GroupReceiveMessage{}
	err := json.Unmarshal(payload.D, &gm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal group receive message failed: %w", err)
	}
	return &gm, nil
}

func buildReply(c *types.Context, msgType message.TypeMessage) (*message.GroupSendMessage, error) {
	gm, err := getGroupReceiveMessage(*c.Payload)
	if err != nil {
		return nil, err
	}
	return &message.GroupSendMessage{
		CommonSendMessage: message.CommonSendMessage{
			MsgType:   msgType,
			QQEventID: c.Payload.T,
			MsgId:     gm.Id,
		},
	}, nil
}

func send(c *types.Context, body *message.GroupSendMessage, configPath string) error {
	gm, err := getGroupReceiveMessage(*c.Payload)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(api.GROUP_SEND_MESSAGE, gm.GroupOpenid)
	logger.DEBUG("send group message: %+v", body)
	_, err = api.NewClient(configPath).Post(context.Background(), path, body)
	if err != nil {
		return fmt.Errorf("send group message failed: %w", err)
	}
	return nil
}

func ReplyPlainText(c *types.Context, text string, configPath string) error {
	body, err := buildReply(c, message.PLAINTEXT)
	if err != nil {
		return err
	}
	body.Content = text
	return send(c, body, configPath)
}

func ReplyMarkdownContent(c *types.Context, content string, configPath string) error {
	body, err := buildReply(c, message.MARKDOWN)
	if err != nil {
		return err
	}
	body.Markdown = &typesMessage.Markdown{Content: content}
	return send(c, body, configPath)
}

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

func ReplyMedia(c *types.Context, fileInfo string, configPath string) error {
	body, err := buildReply(c, message.MEDIA)
	if err != nil {
		return err
	}
	body.Media = &message.MediaInfo{FileInfo: fileInfo}
	return send(c, body, configPath)
}

func UploadMedia(c *types.Context, fileType message.FileType, url string, srvSendMsg bool, configPath string) (*message.RichMediaResponse, error) {
	gm, err := getGroupReceiveMessage(*c.Payload)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(api.GROUP_FILE_MESSAGE, gm.GroupOpenid)
	req := &message.RichMediaRequest{
		FileType:   fileType,
		URL:        url,
		SrvSendMsg: srvSendMsg,
	}
	data, err := api.NewClient(configPath).Post(context.Background(), path, req)
	if err != nil {
		return nil, fmt.Errorf("upload group media failed: %w", err)
	}
	var resp message.RichMediaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload media response failed: %w", err)
	}
	return &resp, nil
}

func RecallMessage(c *types.Context, messageId string, configPath string) error {
	gm, err := getGroupReceiveMessage(*c.Payload)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(api.GROUP_DELETE_MESSAGE, gm.GroupOpenid, messageId)
	_, err = api.NewClient(configPath).Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("recall group message failed: %w", err)
	}
	return nil
}
