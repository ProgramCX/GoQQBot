package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

// DirectSession 是机器人与同频道成员建立的私信会话。
type DirectSession struct {
	// GuildID 用于发送或撤回该会话中的私信。
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	// CreateTime 是 QQ 返回的 Unix 秒字符串。
	CreateTime string `json:"create_time"`
}

// CreateDirectSession 与同一频道内的成员建立私信会话。
func (c *Client) CreateDirectSession(ctx context.Context, guildID, recipientID string) (*DirectSession, error) {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(recipientID) == "" {
		return nil, fmt.Errorf("source guild and recipient IDs are required")
	}
	body := struct {
		RecipientID   string `json:"recipient_id"`
		SourceGuildID string `json:"source_guild_id"`
	}{recipientID, guildID}
	return apiwire.Do[DirectSession](ctx, c.requester(), http.MethodPost, "/users/@me/dms", body)
}

// LocalImage 是随频道消息直接上传的图片。发送期间调用方应保持 Data 不变。
type LocalImage struct {
	Filename string
	Data     []byte
}

// DecodeSentMessage parses the full QQ response from a successful channel or direct message send.
// DecodeSentMessage 将成功发送的子频道或频道私信响应解析为完整消息。
func DecodeSentMessage(result *message.ReplyResult) (*channelmsg.Message, error) {
	if result == nil || len(result.Raw) == 0 {
		return nil, fmt.Errorf("channel send response is required")
	}
	var sent *channelmsg.Message
	if err := json.Unmarshal(result.Raw, &sent); err != nil {
		return nil, fmt.Errorf("decode channel message response: %w", err)
	}
	if sent == nil || sent.ID == "" {
		return nil, fmt.Errorf("channel send response has no message ID")
	}
	return sent, nil
}

// SendChannelMessage 发送完整的频道消息。传入 image 时使用 multipart 表单。
// 成功发送后若响应解析失败，会同时返回含 Raw 的结果和错误。
func (c *Client) SendChannelMessage(ctx context.Context, channelID string, body channelmsg.ChannelSendMessage, image *LocalImage) (*message.ReplyResult, error) {
	return c.send(ctx, api.CHANNEL_SEND_MESSAGE, channelID, body, image)
}

// SendDirectMessage 向已建立的频道私信会话发送消息，支持与频道消息相同的表单。
// 成功发送后若响应解析失败，会同时返回含 Raw 的结果和错误。
func (c *Client) SendDirectMessage(ctx context.Context, sessionID string, body channelmsg.ChannelSendMessage, image *LocalImage) (*message.ReplyResult, error) {
	return c.send(ctx, api.CHANNEL_PRIVATE_MESSAGE, sessionID, body, image)
}

func (c *Client) send(ctx context.Context, route, id string, body channelmsg.ChannelSendMessage, image *LocalImage) (*message.ReplyResult, error) {
	path, err := apiwire.Path(route, id)
	if err != nil {
		return nil, err
	}
	if body.MsgId != "" && body.EventId != "" {
		return nil, fmt.Errorf("supply at most one msg_id or event_id")
	}
	var data []byte
	if image == nil {
		data, err = c.requester().Post(ctx, path, body)
	} else {
		data, err = c.sendImage(ctx, path, body, image)
	}
	if err != nil {
		return nil, err
	}
	result := &message.ReplyResult{Raw: data}
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful channel send (message may have been sent): %w", err)
	}
	return result, nil
}

func (c *Client) sendImage(ctx context.Context, path string, body channelmsg.ChannelSendMessage, image *LocalImage) ([]byte, error) {
	if strings.TrimSpace(image.Filename) == "" || strings.ContainsAny(image.Filename, "\r\n\x00") || len(image.Data) == 0 || body.Image != "" {
		return nil, fmt.Errorf("local image requires filename and data, and cannot be combined with an image URL")
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	var payload bytes.Buffer
	form := multipart.NewWriter(&payload)
	for key, raw := range fields {
		value := string(raw)
		if len(raw) > 0 && raw[0] == '"' {
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
		}
		if err := form.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	part, err := form.CreateFormFile("file_image", image.Filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(image.Data); err != nil {
		return nil, err
	}
	if err := form.Close(); err != nil {
		return nil, err
	}
	return c.requester().PostMultipart(ctx, path, form, &payload)
}

// RecallOptions 控制频道撤回时是否隐藏提示。
type RecallOptions struct {
	HideTip bool
}

// RecallChannelMessage 撤回指定子频道中的消息，权限由 QQ 检查。
func (c *Client) RecallChannelMessage(ctx context.Context, channelID, messageID string, options RecallOptions) error {
	return c.requester().RecallChannelMessage(ctx, message.Target{Scene: message.SceneChannel, ID: channelID}, messageID, options.HideTip)
}

// RecallDirectMessage 撤回机器人在指定私信会话中发出的消息。
func (c *Client) RecallDirectMessage(ctx context.Context, sessionID, messageID string, options RecallOptions) error {
	return c.requester().RecallChannelMessage(ctx, message.Target{Scene: message.SceneDirectMessage, ID: sessionID}, messageID, options.HideTip)
}
