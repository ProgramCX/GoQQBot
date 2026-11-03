// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package types

import "github.com/ProgramCX/GoQQBot/pkg/message"

// Text returns the router-normalized text (surrounding whitespace removed), including the slash for commands.
// Text 返回路由规范化后的文本，移除首尾空白，命令保留斜杠。
// It performs no decoding or I/O. Caller-created contexts must set RawMessage, or use Message for payload content.
// 不执行解码或 I/O；自行构造的上下文需设置 RawMessage，或使用 Message 读取载荷正文。
func (c *Context) Text() string {
	if c == nil {
		return ""
	}
	return c.RawMessage
}

// Message decodes an independent snapshot of the payload body, including original text and attachments.
// Message 将载荷正文解码为独立快照，包括原始文本及附件。
// It needs no API client. It works for caller-declared custom message events; classification remains the caller's responsibility.
// 无需 API 客户端，支持调用方声明的自定义消息事件，分类由调用方负责。
func (c *Context) Message() (*message.Received, error) {
	if c == nil || c.Payload == nil {
		return nil, message.ErrInvalidReceived
	}
	return message.DecodeReceived(c.Payload.T, c.Payload.D)
}
