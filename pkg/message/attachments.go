package message

import "strings"

// IsImage classifies only the declared MIME type (including the legacy "image" value).
// IsImage 仅根据声明的 MIME 类型分类，包括旧版 image 值。
// It does not infer an image or sticker from text, file extensions or remote URL contents.
// 不根据文本、扩展名或远程 URL 内容推断图片或表情包。
func (a ReceivedAttachment) IsImage() bool {
	kind := strings.ToLower(string(a.ContentType))
	return kind == "image" || strings.HasPrefix(kind, "image/")
}

// AllAttachments returns an owned flattened snapshot of top-level and nested quote/forward attachments.
// AllAttachments 返回顶层及嵌套引用或转发附件的独立扁平快照。
// Use MsgElements directly when original authors and message indices are needed. No URLs are downloaded.
// 需要原作者及消息索引时直接使用 MsgElements，不下载 URL。
func (m *Received) AllAttachments() []MessageAttachment {
	if m == nil {
		return nil
	}
	result := append([]MessageAttachment(nil), m.Attachments...)
	var stack []*MsgElement
	for i := len(m.MsgElements) - 1; i >= 0; i-- {
		stack = append(stack, &m.MsgElements[i])
	}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		result = append(result, node.Attachments...)
		for i := len(node.MsgElements) - 1; i >= 0; i-- {
			stack = append(stack, &node.MsgElements[i])
		}
	}
	return result
}

// Images returns image attachments from the message and recursive quotes/forwards.
// Images 返回当前消息及递归引用或转发中的图片附件。
// A sticker is included only if QQ actually delivered it as an image attachment; no image URL is invented.
// 仅在 QQ 实际下发图片附件时包含表情包，不虚构图片 URL。
func (m *Received) Images() []MessageAttachment {
	var result []MessageAttachment
	for _, attachment := range m.AllAttachments() {
		if attachment.IsImage() {
			result = append(result, attachment)
		}
	}
	return result
}
