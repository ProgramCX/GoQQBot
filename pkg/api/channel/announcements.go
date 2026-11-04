// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"context"
	"net/http"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// AnnouncementType 表示全局公告的类别。
type AnnouncementType uint32

// 频道成员公告与欢迎公告的协议值。
const (
	AnnouncementMember  AnnouncementType = 0
	AnnouncementWelcome AnnouncementType = 1
)

// RecommendedChannel 指定推荐的子频道和推荐语。
type RecommendedChannel struct {
	ChannelID string `json:"channel_id"`
	Introduce string `json:"introduce"`
}

// AnnouncementInput 以现有消息或推荐子频道列表设置全局公告。
// 指定 MessageID 时需同时指定 ChannelID，平台优先使用消息公告。
type AnnouncementInput struct {
	MessageID           string               `json:"message_id,omitempty"`
	ChannelID           string               `json:"channel_id,omitempty"`
	Type                AnnouncementType     `json:"announces_type,omitempty"`
	RecommendedChannels []RecommendedChannel `json:"recommend_channels,omitempty"`
}

// Announcement 保存平台返回的全局公告设置。
type Announcement struct {
	GuildID             string               `json:"guild_id"`
	MessageID           string               `json:"message_id"`
	ChannelID           string               `json:"channel_id"`
	Type                AnnouncementType     `json:"announces_type"`
	RecommendedChannels []RecommendedChannel `json:"recommend_channels"`
}

// PinnedMessages 保存子频道当前的精华消息 ID。
type PinnedMessages struct {
	GuildID    string   `json:"guild_id"`
	ChannelID  string   `json:"channel_id"`
	MessageIDs []string `json:"message_ids"`
}

// CreateAnnouncement 设置频道全局公告。推荐子频道列表由本次请求整体替换。
func (c *Client) CreateAnnouncement(ctx context.Context, guildID string, input AnnouncementInput) (*Announcement, error) {
	path, err := apiwire.Path("/guilds/%s/announces", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Announcement](ctx, c.requester(), http.MethodPost, path, input)
}

// DeleteAnnouncement 删除指定消息公告。传入 all 可删除全局公告，包括推荐子频道公告。
func (c *Client) DeleteAnnouncement(ctx context.Context, guildID, messageID string) error {
	path, err := apiwire.Path("/guilds/%s/announces/%s", guildID, messageID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}

// AddPin 将可见消息设为精华，并返回操作后的精华消息列表。
func (c *Client) AddPin(ctx context.Context, channelID, messageID string) (*PinnedMessages, error) {
	path, err := apiwire.Path("/channels/%s/pins/%s", channelID, messageID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[PinnedMessages](ctx, c.requester(), http.MethodPut, path, nil)
}

// ListPins 查询子频道当前的精华消息。
func (c *Client) ListPins(ctx context.Context, channelID string) (*PinnedMessages, error) {
	path, err := apiwire.Path("/channels/%s/pins", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[PinnedMessages](ctx, c.requester(), http.MethodGet, path, nil)
}

// DeletePin 移除指定精华消息。传入 all 可移除子频道内的全部精华。
func (c *Client) DeletePin(ctx context.Context, channelID, messageID string) error {
	path, err := apiwire.Path("/channels/%s/pins/%s", channelID, messageID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}
