// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

// Guild 表示频道资料。生命周期事件可能只提供部分字段。
type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	OwnerID     string `json:"owner_id"`
	Owner       bool   `json:"owner"`
	MemberCount int    `json:"member_count"`
	MaxMembers  int    `json:"max_members"`
	Description string `json:"description"`
	// JoinedAt 保留 QQ 返回的加入时间。
	JoinedAt     string    `json:"joined_at"`
	Channels     []Channel `json:"channels"`
	UnionWorldID string    `json:"union_world_id"`
	UnionOrgID   string    `json:"union_org_id"`
	// OpUserID 是发起生命周期变更的用户。
	OpUserID string `json:"op_user_id"`
}

// Channel 表示子频道或分组。生命周期事件可能只提供部分字段。
type Channel struct {
	ID              string `json:"id"`
	GuildID         string `json:"guild_id"`
	Name            string `json:"name"`
	Type            int    `json:"type"`
	SubType         int    `json:"sub_type"`
	Position        int    `json:"position"`
	ParentID        string `json:"parent_id"`
	OwnerID         string `json:"owner_id"`
	PrivateType     int    `json:"private_type"`
	SpeakPermission int    `json:"speak_permission"`
	ApplicationID   string `json:"application_id"`
	// Permissions 是用十进制字符串表示的权限位图。
	Permissions string `json:"permissions"`
	// OpUserID 是发起生命周期变更的用户。
	OpUserID string `json:"op_user_id"`
}

// GuildMemberEvent 表示频道成员加入、更新或退出。
type GuildMemberEvent struct {
	GuildID string   `json:"guild_id"`
	User    *User    `json:"user"`
	Nick    string   `json:"nick"`
	Roles   []string `json:"roles"`
	// JoinedAt 保留 QQ 返回的时间文本。
	JoinedAt string `json:"joined_at"`
	OpUserID string `json:"op_user_id"`
}
