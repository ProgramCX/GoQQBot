package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

// Guild 表示机器人可访问的频道。
type Guild = channelmsg.Guild

// GuildListOptions 控制单次分页。Before 与 After 同时填写时 QQ 优先使用 Before。
type GuildListOptions struct {
	Before string
	After  string
	// Limit 为零时采用 QQ 默认值，平台上限为 100。
	Limit int
}

type guildList struct {
	Guilds []Guild `json:"guilds"`
}

type channelList struct {
	Channels []Channel `json:"channels"`
}

func (c *channelList) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &c.Channels)
	}
	type object channelList
	return json.Unmarshal(data, (*object)(c))
}

func (g *guildList) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &g.Guilds)
	}
	type object guildList
	return json.Unmarshal(data, (*object)(g))
}

// ListGuilds 查询当前机器人的一页频道。
func (c *Client) ListGuilds(ctx context.Context, options GuildListOptions) ([]Guild, error) {
	q := url.Values{}
	if options.Before != "" {
		q.Set("before", options.Before)
	}
	if options.After != "" {
		q.Set("after", options.After)
	}
	if options.Limit != 0 {
		q.Set("limit", strconv.Itoa(options.Limit))
	}
	path := "/users/@me/guilds"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	result, err := apiwire.Do[guildList](ctx, c.requester(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return result.Guilds, nil
}

// Channel 表示频道下的子频道或分组。
type Channel = channelmsg.Channel

// ChannelPatch 只修改非 nil 字段，允许显式设置零值。
type ChannelPatch struct {
	Name            *string `json:"name,omitempty"`
	Position        *int    `json:"position,omitempty"`
	ParentID        *string `json:"parent_id,omitempty"`
	PrivateType     *int    `json:"private_type,omitempty"`
	SpeakPermission *int    `json:"speak_permission,omitempty"`
}

// ChannelCreate 设置新子频道的属性。nil 数值字段使用平台默认值。
type ChannelCreate struct {
	Name            string   `json:"name,omitempty"`
	Position        *int     `json:"position,omitempty"`
	ParentID        string   `json:"parent_id,omitempty"`
	PrivateType     *int     `json:"private_type,omitempty"`
	SpeakPermission *int     `json:"speak_permission,omitempty"`
	Type            *int     `json:"type,omitempty"`
	SubType         *int     `json:"sub_type,omitempty"`
	PrivateUserIDs  []string `json:"private_user_ids,omitempty"`
	ApplicationID   string   `json:"application_id,omitempty"`
}

// GetGuild 查询指定频道的基本信息。
func (c *Client) GetGuild(ctx context.Context, guildID string) (*Guild, error) {
	path, err := apiwire.Path("/guilds/%s", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Guild](ctx, c.requester(), http.MethodGet, path, nil)
}

// GetChannel 查询指定子频道的资料。
func (c *Client) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	path, err := apiwire.Path("/channels/%s", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Channel](ctx, c.requester(), http.MethodGet, path, nil)
}

// CreateChannel 在指定频道中创建子频道，平台检查管理权限。
func (c *Client) CreateChannel(ctx context.Context, guildID string, body ChannelCreate) (*Channel, error) {
	path, err := apiwire.Path("/guilds/%s/channels", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Channel](ctx, c.requester(), http.MethodPost, path, body)
}

// UpdateChannel 更新指定子频道的部分属性。
func (c *Client) UpdateChannel(ctx context.Context, channelID string, body ChannelPatch) (*Channel, error) {
	path, err := apiwire.Path("/channels/%s", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Channel](ctx, c.requester(), http.MethodPatch, path, body)
}

// DeleteChannel 删除指定子频道。
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	path, err := apiwire.Path("/channels/%s", channelID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}

// ListChannels 查询指定频道下的子频道。
func (c *Client) ListChannels(ctx context.Context, guildID string) ([]Channel, error) {
	path, err := apiwire.Path("/guilds/%s/channels", guildID)
	if err != nil {
		return nil, err
	}
	result, err := apiwire.Do[channelList](ctx, c.requester(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return result.Channels, nil
}
