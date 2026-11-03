package message

import (
	"fmt"
	"net/url"
	"unicode/utf8"
)

// MessageCard is an outgoing card envelope. WithField can carry future card types and content schemas.
// MessageCard 是发送卡片的信封，可通过 WithField 扩展未来卡片类型及内容结构。
type MessageCard struct {
	Type    string `json:"type"`
	Content any    `json:"content"`
}

// ImageTextCard is the documented tuwen content. Both URLs must be HTTPS.
// ImageTextCard 对应文档中的 tuwen 内容，两个 URL 都必须使用 HTTPS。
type ImageTextCard struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PictureURL  string `json:"pic_url"`
}

// Card snapshots an outgoing image-text card; Build/Reply report validation errors.
// Card 创建发送图文卡片的快照，Build/Reply 报告校验错误。
func Card(content ImageTextCard) Outgoing {
	return New(CARD).WithField("card", MessageCard{Type: "tuwen", Content: content})
}

// InputNotify is the C2C typing indicator. Seconds must be in 1..60.
// InputNotify 是单聊输入状态提示，Seconds 必须为 1..60。
type InputNotify struct {
	Type    int `json:"input_type"`
	Seconds int `json:"input_second"`
}

// Typing describes a C2C typing indicator; it performs no I/O or background work.
// Typing 描述单聊输入状态提示，不执行 I/O 或后台任务。
func Typing(seconds int) Outgoing {
	return New(INPUT_NOTIFY).WithField("input_notify", InputNotify{Type: 1, Seconds: seconds})
}

func validateImageTextCard(card ImageTextCard) error {
	if n := utf8.RuneCountInString(card.Title); n < 1 || n > 50 {
		return fmt.Errorf("%w: card title requires 1..50 characters", ErrInvalidMessage)
	}
	if n := utf8.RuneCountInString(card.Description); n < 1 || n > 200 {
		return fmt.Errorf("%w: card description requires 1..200 characters", ErrInvalidMessage)
	}
	for _, value := range []string{card.URL, card.PictureURL} {
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "https" || u.Host == "" || utf8.RuneCountInString(value) > 1024 {
			return fmt.Errorf("%w: card links require HTTPS and at most 1024 characters", ErrInvalidMessage)
		}
	}
	return nil
}
