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

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// Role 描述频道身份组及其人数。
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Color 是 ARGB 颜色的整数值。
	Color uint32 `json:"color"`
	// Hoist 为 1 时在成员列表中单独展示。
	Hoist       uint32 `json:"hoist"`
	MemberCount uint32 `json:"number"`
	MemberLimit uint32 `json:"member_limit"`
}

// RoleList 是一个频道的身份组列表及身份组数量上限。
type RoleList struct {
	GuildID   string `json:"guild_id"`
	Roles     []Role `json:"roles"`
	RoleLimit string `json:"role_num_limit"`
}

// RoleOptions 指定创建或更新时要提交的字段。nil 字段会省略。
type RoleOptions struct {
	Name  *string `json:"name,omitempty"`
	Color *uint32 `json:"color,omitempty"`
	// Hoist 取 0 表示不单独展示，取 1 表示单独展示。
	Hoist *uint32 `json:"hoist,omitempty"`
}

// RoleResult 包含创建或更新后 QQ 返回的身份组。
type RoleResult struct {
	GuildID string `json:"guild_id"`
	RoleID  string `json:"role_id"`
	Role    Role   `json:"role"`
}

// RoleMembershipOptions 指定身份组操作关联的子频道。子频道管理员身份组（ID "5"）需要填写 ChannelID。
type RoleMembershipOptions struct {
	ChannelID string
}

// ListRoles 获取频道身份组和 QQ 返回的数量上限。
func (c *Client) ListRoles(ctx context.Context, guildID string) (*RoleList, error) {
	path, err := apiwire.Path("/guilds/%s/roles", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[RoleList](ctx, c.requester(), http.MethodGet, path, nil)
}

// CreateRole 在频道中创建身份组。至少指定一个字段。
func (c *Client) CreateRole(ctx context.Context, guildID string, options RoleOptions) (*RoleResult, error) {
	path, err := apiwire.Path("/guilds/%s/roles", guildID)
	if err != nil {
		return nil, err
	}
	return c.writeRole(ctx, http.MethodPost, path, options)
}

// UpdateRole 修改指定身份组。只提交 options 中非 nil 的字段。
func (c *Client) UpdateRole(ctx context.Context, guildID, roleID string, options RoleOptions) (*RoleResult, error) {
	path, err := apiwire.Path("/guilds/%s/roles/%s", guildID, roleID)
	if err != nil {
		return nil, err
	}
	return c.writeRole(ctx, http.MethodPatch, path, options)
}

func (c *Client) writeRole(ctx context.Context, method, path string, options RoleOptions) (*RoleResult, error) {
	if options.Name == nil && options.Color == nil && options.Hoist == nil {
		return nil, errors.New("write channel role: at least one field is required")
	}
	return apiwire.Do[RoleResult](ctx, c.requester(), method, path, options)
}

// DeleteRole 删除指定频道身份组。
func (c *Client) DeleteRole(ctx context.Context, guildID, roleID string) error {
	path, err := apiwire.Path("/guilds/%s/roles/%s", guildID, roleID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}

// AddMemberRole 将频道成员加入指定身份组。
func (c *Client) AddMemberRole(ctx context.Context, guildID, userID, roleID string, options RoleMembershipOptions) error {
	return c.writeMemberRole(ctx, http.MethodPut, guildID, userID, roleID, options)
}

// RemoveMemberRole 将频道成员从指定身份组移除。
func (c *Client) RemoveMemberRole(ctx context.Context, guildID, userID, roleID string, options RoleMembershipOptions) error {
	return c.writeMemberRole(ctx, http.MethodDelete, guildID, userID, roleID, options)
}

func (c *Client) writeMemberRole(ctx context.Context, method, guildID, userID, roleID string, options RoleMembershipOptions) error {
	path, err := apiwire.Path("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID)
	if err != nil {
		return err
	}
	if roleID == "5" && options.ChannelID == "" {
		return errors.New("write channel member role: channel ID required for subchannel administrator")
	}
	var body any
	if options.ChannelID != "" {
		type channelRef struct {
			ID string `json:"id"`
		}
		body = struct {
			Channel channelRef `json:"channel"`
		}{Channel: channelRef{ID: options.ChannelID}}
	}
	return apiwire.Exec(ctx, c.requester(), method, path, body)
}
