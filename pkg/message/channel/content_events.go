package channel

import "encoding/json"

// ForumEvent 表示主题、评论、回复或审核事件。开放论坛事件只包含前三个标识字段。
type ForumEvent struct {
	GuildID   string            `json:"guild_id"`
	ChannelID string            `json:"channel_id"`
	AuthorID  string            `json:"author_id"`
	Thread    *ForumThreadEvent `json:"thread_info,omitempty"`
	Post      *ForumPostEvent   `json:"post_info,omitempty"`
	Reply     *ForumReplyEvent  `json:"reply_info,omitempty"`
	// 以下字段只在审核事件中提供。Result 为 0 表示审核通过。
	ThreadID  string  `json:"thread_id,omitempty"`
	PostID    string  `json:"post_id,omitempty"`
	ReplyID   string  `json:"reply_id,omitempty"`
	AuditType uint32  `json:"type,omitempty"`
	Result    *uint32 `json:"result,omitempty"`
	ErrorText string  `json:"err_msg,omitempty"`
}

// UnmarshalJSON 将平台发送的字符串或数字标识统一解析为字符串。
func (e *ForumEvent) UnmarshalJSON(data []byte) error {
	type plain ForumEvent
	var result ForumEvent
	wire := struct {
		*plain
		GuildID   eventIdentifier `json:"guild_id"`
		ChannelID eventIdentifier `json:"channel_id"`
		AuthorID  eventIdentifier `json:"author_id"`
	}{plain: (*plain)(&result)}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	result.GuildID = string(wire.GuildID)
	result.ChannelID = string(wire.ChannelID)
	result.AuthorID = string(wire.AuthorID)
	*e = result
	return nil
}

type eventIdentifier string

func (id *eventIdentifier) UnmarshalJSON(data []byte) error {
	if len(data) != 0 && data[0] == '"' {
		return json.Unmarshal(data, (*string)(id))
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*id = eventIdentifier(number)
	return nil
}

// ForumThreadEvent 保存主题标识及内容。标题和正文保留原始 JSON，可能是字符串或富文本数组。
type ForumThreadEvent struct {
	ThreadID string          `json:"thread_id"`
	Title    json.RawMessage `json:"title"`
	Content  json.RawMessage `json:"content"`
	DateTime string          `json:"date_time"`
}

// ForumPostEvent 保存主题下的评论及所属主题。Content 保留原始 JSON。
type ForumPostEvent struct {
	ThreadID string          `json:"thread_id"`
	PostID   string          `json:"post_id"`
	Content  json.RawMessage `json:"content"`
	DateTime string          `json:"date_time"`
}

// ForumReplyEvent 保存对评论的回复及上级标识。Content 保留原始 JSON。
type ForumReplyEvent struct {
	ThreadID string          `json:"thread_id"`
	PostID   string          `json:"post_id"`
	ReplyID  string          `json:"reply_id"`
	Content  json.RawMessage `json:"content"`
	DateTime string          `json:"date_time"`
}

// AudioAction 表示音频播放或机器人上下麦事件。事件类型在外层事件对象中。
type AudioAction struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	AudioURL  string `json:"audio_url"`
	Text      string `json:"text"`
}

// AudioLiveMember 表示成员进入或退出音视频或直播子频道。
type AudioLiveMember struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	// ChannelType 为 2 表示音视频子频道，为 5 表示直播子频道。
	ChannelType int `json:"channel_type"`
}
