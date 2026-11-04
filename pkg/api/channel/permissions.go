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

// PermissionMask 是子频道权限位图，以十进制字符串传输。
type PermissionMask uint64

const (
	// PermissionView 允许查看子频道。
	PermissionView PermissionMask = 1 << iota
	// PermissionManage 表示可管理子频道。权限修改接口不支持变更此位。
	PermissionManage
	// PermissionSpeak 允许在子频道发言。
	PermissionSpeak
)

// Permissions 是用户或身份组在指定子频道上的权限。
type Permissions struct {
	ChannelID string         `json:"channel_id"`
	UserID    string         `json:"user_id,omitempty"`
	RoleID    string         `json:"role_id,omitempty"`
	Mask      PermissionMask `json:"permissions,string"`
}

// PermissionChanges 指定授予或移除哪些权限。两者包含同一位时，QQ 移除该权限。
type PermissionChanges struct {
	Add    PermissionMask `json:"add,string"`
	Remove PermissionMask `json:"remove,string"`
}

// APIIdentifier 通过 HTTP 方法和官方路径模板标识需要授权的接口。
type APIIdentifier struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// APIPermission 描述机器人在指定频道可使用或申请授权的接口。
type APIPermission struct {
	APIIdentifier
	Description string `json:"desc"`
	// AuthStatus 为 1 时表示已授权。
	AuthStatus int `json:"auth_status"`
}

// APIPermissionList 是频道可用接口权限的集合。
type APIPermissionList struct {
	APIs []APIPermission `json:"apis"`
}

// APIPermissionRequest 指定在何处发送授权请求及其用途。
type APIPermissionRequest struct {
	ChannelID   string        `json:"channel_id"`
	API         APIIdentifier `json:"api_identify"`
	Description string        `json:"desc"`
}

// APIPermissionDemand 是 QQ 返回的频道接口授权请求。创建请求不代表已经获得授权。
type APIPermissionDemand struct {
	GuildID string `json:"guild_id"`
	APIPermissionRequest
	Title string `json:"title"`
}

// GetMemberPermissions 获取用户的子频道权限。
func (c *Client) GetMemberPermissions(ctx context.Context, channelID, userID string) (*Permissions, error) {
	path, err := apiwire.Path("/channels/%s/members/%s/permissions", channelID, userID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Permissions](ctx, c.requester(), http.MethodGet, path, nil)
}

// UpdateMemberPermissions 授予或移除用户的子频道权限。
func (c *Client) UpdateMemberPermissions(ctx context.Context, channelID, userID string, changes PermissionChanges) error {
	path, err := apiwire.Path("/channels/%s/members/%s/permissions", channelID, userID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPut, path, changes)
}

// GetRolePermissions 获取身份组的子频道权限。
func (c *Client) GetRolePermissions(ctx context.Context, channelID, roleID string) (*Permissions, error) {
	path, err := apiwire.Path("/channels/%s/roles/%s/permissions", channelID, roleID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Permissions](ctx, c.requester(), http.MethodGet, path, nil)
}

// UpdateRolePermissions 授予或移除身份组的子频道权限。
func (c *Client) UpdateRolePermissions(ctx context.Context, channelID, roleID string, changes PermissionChanges) error {
	path, err := apiwire.Path("/channels/%s/roles/%s/permissions", channelID, roleID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPut, path, changes)
}

// ListAPIPermissions 查询机器人在频道的可用接口及授权状态。
func (c *Client) ListAPIPermissions(ctx context.Context, guildID string) (*APIPermissionList, error) {
	path, err := apiwire.Path("/guilds/%s/api_permission", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[APIPermissionList](ctx, c.requester(), http.MethodGet, path, nil)
}

// RequestAPIPermission 在指定子频道发送接口授权链接。
func (c *Client) RequestAPIPermission(ctx context.Context, guildID string, request APIPermissionRequest) (*APIPermissionDemand, error) {
	if request.ChannelID == "" || request.API.Path == "" || request.API.Method == "" {
		return nil, errors.New("request channel API permission: channel ID, API path and method are required")
	}
	path, err := apiwire.Path("/guilds/%s/api_permission/demand", guildID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[APIPermissionDemand](ctx, c.requester(), http.MethodPost, path, request)
}
