// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package interaction

type MessageReaction struct {
	UserID    string         `json:"user_id"`
	GuildID   string         `json:"guild_id"`
	ChannelID string         `json:"channel_id"`
	Target    ReactionTarget `json:"target"`
	Emoji     Emoji          `json:"emoji"`
}

type ReactionTarget struct {
	ID   string             `json:"id"`
	Type ReactionTargetType `json:"type"`
}

type Emoji struct {
	ID   string    `json:"id"`
	Type EmojiType `json:"type"`
}

type ReactionTargetType int

const (
	ReactionTargetTypeMessage ReactionTargetType = 0
	ReactionTargetTypeFeed    ReactionTargetType = 1
	ReactionTargetTypeComment ReactionTargetType = 2
	ReactionTargetTypeReply   ReactionTargetType = 3
)

type EmojiType int

const (
	EmojiTypeSystem EmojiType = 1
	EmojiTypeCustom EmojiType = 2
)

type GetReactionUsersResponse struct {
	Users  []ReactionUser `json:"users"`
	Cookie string         `json:"cookie"`
	IsEnd  bool           `json:"is_end"`
}

type ReactionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}
