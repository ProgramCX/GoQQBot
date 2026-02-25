// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package channel

import "time"

// Message
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/template/model.html#message
type Message struct {
	ID               string              `json:"id"`
	ChannelID        string              `json:"channel_id"`
	GuildID          string              `json:"guild_id"`
	Content          string              `json:"content"`
	Timestamp        time.Time           `json:"timestamp"`
	EditedTimestamp  time.Time           `json:"edited_timestamp"`
	MentionEveryone  bool                `json:"mention_everyone"`
	Author           User                `json:"author"`
	Attachments      []MessageAttachment `json:"attachments"`
	Embeds           []MessageEmbed      `json:"embeds"`
	Mentions         []User              `json:"mentions"`
	Member           Member              `json:"member"`
	Ark              MessageArk          `json:"ark"`
	Seq              int                 `json:"seq"`
	SeqInChannel     string              `json:"seq_in_channel"`
	MessageReference MessageReference    `json:"message_reference"`
}

// User (Placeholder)
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Bot      bool   `json:"bot"`
}

// Member (Placeholder)
type Member struct {
	User     User      `json:"user"`
	Nick     string    `json:"nick"`
	Roles    []string  `json:"roles"`
	JoinedAt time.Time `json:"joined_at"`
}

// MessageEmbed
type MessageEmbed struct {
	Title     string                `json:"title"`
	Prompt    string                `json:"prompt"`
	Thumbnail MessageEmbedThumbnail `json:"thumbnail"`
	Fields    []MessageEmbedField   `json:"fields"`
}

// MessageEmbedThumbnail
type MessageEmbedThumbnail struct {
	URL string `json:"url"`
}

// MessageEmbedField
type MessageEmbedField struct {
	Name string `json:"name"`
}

// MessageAttachment
type MessageAttachment struct {
	URL string `json:"url"`
}

// MessageArk
type MessageArk struct {
	TemplateID int            `json:"template_id"`
	KV         []MessageArkKV `json:"kv"`
}

// MessageArkKV
type MessageArkKV struct {
	Key   string          `json:"key"`
	Value string          `json:"value"`
	Obj   []MessageArkObj `json:"obj"`
}

// MessageArkObj
type MessageArkObj struct {
	ObjKV []MessageArkObjKV `json:"obj_kv"`
}

// MessageArkObjKV
type MessageArkObjKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MessageReference
type MessageReference struct {
	MessageID             string `json:"message_id"`
	IgnoreGetMessageError bool   `json:"ignore_get_message_error"`
}

// MessageMarkdown
type MessageMarkdown struct {
	TemplateID int                   `json:"template_id"`
	Params     MessageMarkdownParams `json:"params"`
	Content    string                `json:"content"`
}

// MessageMarkdownParams
type MessageMarkdownParams struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// MessageDelete
type MessageDelete struct {
	Message Message `json:"message"`
	OpUser  User    `json:"op_user"`
}

// MessageAudited
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/template/model.html#messageaudited
type MessageAudited struct {
	AuditID      string    `json:"audit_id"`
	MessageID    string    `json:"message_id"`
	GuildID      string    `json:"guild_id"`
	ChannelID    string    `json:"channel_id"`
	AuditTime    time.Time `json:"audit_time"`
	CreateTime   time.Time `json:"create_time"`
	SeqInChannel string    `json:"seq_in_channel"`
}
