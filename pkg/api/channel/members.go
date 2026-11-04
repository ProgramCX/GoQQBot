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
	"net/url"
	"strconv"
	"time"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

// MemberUser 是管理接口中的频道用户信息，包含平台授权后才返回的关联应用字段。
type MemberUser struct {
	channelmsg.User
	// UnionOpenID 是关联应用的 OpenID。
	UnionOpenID string `json:"union_openid"`
	// UnionUserAccount 是关联应用的用户信息。
	UnionUserAccount string `json:"union_user_account"`
	PublicFlags      uint64 `json:"public_flags"`
	System           bool   `json:"system"`
}

// Member 是频道成员资料。GuildID 和 User 是否填充取决于接口及事件。
type Member struct {
	GuildID  string      `json:"guild_id"`
	User     *MemberUser `json:"user"`
	Nick     string      `json:"nick"`
	Roles    []string    `json:"roles"`
	JoinedAt time.Time   `json:"joined_at"`
	Deaf     bool        `json:"deaf"`
	Mute     bool        `json:"mute"`
	Pending  bool        `json:"pending"`
}

// MemberListOptions 指定频道成员列表的一页。
type MemberListOptions struct {
	// After 是上一页末位成员的用户 ID。首次查询可留空。
	After string
	// Limit 为 1 到 400。零值沿用 QQ 默认值。
	Limit uint32
}

// RoleMemberListOptions 指定身份组成员列表的一页。
type RoleMemberListOptions struct {
	// StartIndex 是上一页的 Next。首次查询可留空。
	StartIndex string
	// Limit 为 1 到 400。零值沿用 QQ 默认值。
	Limit uint32
}

// RoleMemberPage 是身份组成员列表的一页。Next 为 "0" 时已到末页。
type RoleMemberPage struct {
	Members []Member `json:"data"`
	Next    string   `json:"next"`
}

// RemoveMemberOptions 指定移除成员时是否同时拉黑和撤回历史消息。
type RemoveMemberOptions struct {
	AddBlacklist bool `json:"add_blacklist"`
	// DeleteHistoryMessageDays 可设为 3、7、15、30 或 -1。零值不撤回，-1 撤回全部消息。
	DeleteHistoryMessageDays int `json:"delete_history_msg_days"`
}

// OnlineCount 是音视频或直播子频道的在线成员数。
type OnlineCount struct {
	Count uint32 `json:"online_nums"`
}

// ListMembers 查询一页频道成员。使用末位成员的用户 ID 继续查询，空页表示结束。
func (c *Client) ListMembers(ctx context.Context, guildID string, options MemberListOptions) ([]Member, error) {
	path, err := apiwire.Path("/guilds/%s/members", guildID)
	if err != nil {
		return nil, err
	}
	path = memberPagePath(path, "after", options.After, options.Limit)
	result, err := apiwire.Do[[]Member](ctx, c.requester(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// ListRoleMembers 查询一页指定身份组的成员。使用返回的 Next 继续查询。
func (c *Client) ListRoleMembers(ctx context.Context, guildID, roleID string, options RoleMemberListOptions) (*RoleMemberPage, error) {
	path, err := apiwire.Path("/guilds/%s/roles/%s/members", guildID, roleID)
	if err != nil {
		return nil, err
	}
	path = memberPagePath(path, "start_index", options.StartIndex, options.Limit)
	return apiwire.Do[RoleMemberPage](ctx, c.requester(), http.MethodGet, path, nil)
}

func memberPagePath(path, cursorKey, cursor string, limit uint32) string {
	query := url.Values{}
	if cursor != "" {
		query.Set(cursorKey, cursor)
	}
	if limit > 0 {
		query.Set("limit", strconv.FormatUint(uint64(limit), 10))
	}
	if encoded := query.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}

// GetMember 按频道用户 ID 查询成员，含 QQ 向机器人开放的资料。
func (c *Client) GetMember(ctx context.Context, guildID, userID string) (*Member, error) {
	path, err := apiwire.Path("/guilds/%s/members/%s", guildID, userID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Member](ctx, c.requester(), http.MethodGet, path, nil)
}

// RemoveMember 移除频道成员。零值选项只移除成员。
func (c *Client) RemoveMember(ctx context.Context, guildID, userID string, options RemoveMemberOptions) error {
	switch options.DeleteHistoryMessageDays {
	case 0, 3, 7, 15, 30, -1:
	default:
		return errors.New("remove channel member: invalid history message deletion period")
	}
	path, err := apiwire.Path("/guilds/%s/members/%s", guildID, userID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, options)
}

// GetOnlineCount 查询音视频或直播子频道的在线成员数。
func (c *Client) GetOnlineCount(ctx context.Context, channelID string) (*OnlineCount, error) {
	path, err := apiwire.Path("/channels/%s/online_nums", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[OnlineCount](ctx, c.requester(), http.MethodGet, path, nil)
}
