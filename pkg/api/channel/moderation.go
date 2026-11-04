// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// MuteOptions 指定禁言截止时间或持续时长。任一字段填 "0" 可解除禁言。
// 同时设置两个字段时，QQ 优先采用 EndTimestamp。
type MuteOptions struct {
	// EndTimestamp 是 Unix 时间戳，单位为秒。
	EndTimestamp string `json:"mute_end_timestamp,omitempty"`
	// Seconds 是禁言持续时间，单位为秒。
	Seconds string `json:"mute_seconds,omitempty"`
}

func (options MuteOptions) validate() error {
	if options.EndTimestamp == "" && options.Seconds == "" {
		return errors.New("set channel mute: end timestamp or duration is required")
	}
	if options.EndTimestamp != "" {
		if _, err := strconv.ParseUint(options.EndTimestamp, 10, 64); err != nil {
			return errors.New("set channel mute: end timestamp must be an unsigned integer")
		}
	}
	if options.Seconds != "" {
		if _, err := strconv.ParseUint(options.Seconds, 10, 64); err != nil {
			return errors.New("set channel mute: duration must be an unsigned integer")
		}
	}
	return nil
}

// MuteMembersResult 列出操作成功的成员，可能仅包含请求中的部分成员。
type MuteMembersResult struct {
	UserIDs []string `json:"user_ids"`
}

// MessageSettings 是当前机器人在指定频道的消息推送设置。
type MessageSettings struct {
	DisableCreateDM    bool `json:"disable_create_dm"`
	DisablePushMessage bool `json:"disable_push_msg"`
	// ChannelIDs 是可主动推送消息的子频道。
	ChannelIDs []string `json:"channel_ids"`
	// ChannelPushMaxNum 是每个子频道允许主动推送的消息条数上限。
	ChannelPushMaxNum uint32 `json:"channel_push_max_num"`
}

// SetGuildMute 设置频道内非管理员成员的全员禁言。
func (c *Client) SetGuildMute(ctx context.Context, guildID string, options MuteOptions) error {
	if err := options.validate(); err != nil {
		return err
	}
	path, err := apiwire.Path("/guilds/%s/mute", guildID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPatch, path, options)
}

// SetMemberMute 设置指定频道成员的禁言状态。
func (c *Client) SetMemberMute(ctx context.Context, guildID, userID string, options MuteOptions) error {
	if err := options.validate(); err != nil {
		return err
	}
	path, err := apiwire.Path("/guilds/%s/members/%s/mute", guildID, userID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPatch, path, options)
}

// SetMembersMute 批量设置禁言状态。结果只列出 QQ 确认操作成功的用户。
func (c *Client) SetMembersMute(ctx context.Context, guildID string, userIDs []string, options MuteOptions) (*MuteMembersResult, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("set channel members mute: user IDs are required")
	}
	for _, userID := range userIDs {
		if userID == "" {
			return nil, errors.New("set channel members mute: user ID cannot be empty")
		}
	}
	if err := options.validate(); err != nil {
		return nil, err
	}
	path, err := apiwire.Path("/guilds/%s/mute", guildID)
	if err != nil {
		return nil, err
	}
	body := struct {
		MuteOptions
		UserIDs []string `json:"user_ids"`
	}{MuteOptions: options, UserIDs: userIDs}
	return apiwire.Do[MuteMembersResult](ctx, c.requester(), http.MethodPatch, path, body)
}

// GetMessageSettings 查询频道是否允许创建私信、主动推送及其限制。
func (c *Client) GetMessageSettings(ctx context.Context, guildID string) (*MessageSettings, error) {
	path, err := apiwire.Path("/guilds/%s/message/setting", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[MessageSettings](ctx, c.requester(), http.MethodGet, path, nil)
}
