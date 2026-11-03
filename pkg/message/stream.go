package message

// StreamChunk describes one C2C streaming update. Send updates sequentially and carry the preceding
// StreamChunk 描述一次单聊流式更新，需顺序发送，并将上次
// response ID in StreamMessageID. The caller owns ordering, persistence and recovery after uncertain failures.
// 响应 ID 放入 StreamMessageID；调用方负责顺序、持久化及不确定失败后的恢复。
type StreamChunk struct {
	InputMode string `json:"input_mode,omitempty"` // append (default) or replace; replace must preserve the already-sent prefix.
	// append 为默认追加模式，replace 为替换模式；替换必须保留已发送前缀。
	InputState int `json:"input_state"` // 1 generating, 10 finished.
	// 1 表示生成中，10 表示已完成。
	Index int `json:"index"` // Starts at zero; increase for subsequent updates.
	// 从零开始，后续更新递增。
	ContentType string `json:"content_type"` // text or markdown.
	// 文本或 Markdown。
	Content         string `json:"content_raw"`
	StreamMessageID string `json:"stream_msg_id,omitempty"`
	Sequence        int    `json:"msg_seq,omitempty"`
	// Wakeup explicitly requests a streaming wakeup message, subject to platform eligibility and quota.
	// Wakeup 显式请求流式唤醒消息，受平台资格及配额约束。
	Wakeup bool `json:"is_wakeup,omitempty"`
}

// StreamResult contains the current streaming ID and remaining character budget.
// StreamResult 包含当前流式 ID 及剩余字符预算。
// Raw and send-response decode error semantics are the same as ReplyResult.
// Raw 及发送响应解码错误的语义与 ReplyResult 相同。
type StreamResult struct {
	ReplyResult
	RemainingLength int `json:"remain_msg_len"`
}
