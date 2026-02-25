// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package interaction

// MessageReaction
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/emoji.html#message-reaction-add
type MessageReaction struct {
	UserID string `json:"user_id"`

	GuildID string `json:"guild_id"`

	ChannelID string `json:"channel_id"`

	Target ReactionTarget `json:"target"`

	Emoji Emoji `json:"emoji"`
}

type ReactionTarget struct {
	ID string `json:"id"`
	// Type:
	// 0 Message
	// 1 Feed
	// 2 Comment
	// 3 Reply
	Type ReactionTargetType `json:"types"`
}

type Emoji struct {
	ID string `json:"id"`
	// Type:
	// 1 System Emoji
	// 2 Custom Emoji
	Type EmojiType `json:"types"`
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

// GetReactionUsersResponse
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/trans/emoji.html#get-reaction-users
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
