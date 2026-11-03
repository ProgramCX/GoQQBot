// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/pkg/message/types"
)

type ChannelSendMessage struct {
	Content          string                `json:"content,omitempty"`
	Embed            *MessageEmbed         `json:"embed,omitempty"`
	Ark              *types.Ark            `json:"ark,omitempty"`
	MessageReference *MessageReference     `json:"message_reference,omitempty"`
	Image            string                `json:"image,omitempty"`
	MsgId            string                `json:"msg_id,omitempty"`
	EventId          string                `json:"event_id,omitempty"`
	Markdown         *types.Markdown       `json:"markdown,omitempty"`
	Keyboard         *interaction.Keyboard `json:"keyboard,omitempty"`
}
