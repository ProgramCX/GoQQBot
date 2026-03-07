// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package channel

import "time"

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

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Bot      bool   `json:"bot"`
}

type Member struct {
	User     User      `json:"user"`
	Nick     string    `json:"nick"`
	Roles    []string  `json:"roles"`
	JoinedAt time.Time `json:"joined_at"`
}

type MessageEmbed struct {
	Title     string                `json:"title"`
	Prompt    string                `json:"prompt"`
	Thumbnail MessageEmbedThumbnail `json:"thumbnail"`
	Fields    []MessageEmbedField   `json:"fields"`
}

type MessageEmbedThumbnail struct {
	URL string `json:"url"`
}

type MessageEmbedField struct {
	Name string `json:"name"`
}

type MessageAttachment struct {
	URL string `json:"url"`
}

type MessageArk struct {
	TemplateID int            `json:"template_id"`
	KV         []MessageArkKV `json:"kv"`
}

type MessageArkKV struct {
	Key   string          `json:"key"`
	Value string          `json:"value"`
	Obj   []MessageArkObj `json:"obj"`
}

type MessageArkObj struct {
	ObjKV []MessageArkObjKV `json:"obj_kv"`
}

type MessageArkObjKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MessageReference struct {
	MessageID             string `json:"message_id"`
	IgnoreGetMessageError bool   `json:"ignore_get_message_error"`
}

type MessageMarkdown struct {
	TemplateID int                   `json:"template_id"`
	Params     MessageMarkdownParams `json:"params"`
	Content    string                `json:"content"`
}

type MessageMarkdownParams struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type MessageDelete struct {
	Message Message `json:"message"`
	OpUser  User    `json:"op_user"`
}

type MessageAudited struct {
	AuditID      string    `json:"audit_id"`
	MessageID    string    `json:"message_id"`
	GuildID      string    `json:"guild_id"`
	ChannelID    string    `json:"channel_id"`
	AuditTime    time.Time `json:"audit_time"`
	CreateTime   time.Time `json:"create_time"`
	SeqInChannel string    `json:"seq_in_channel"`
}
