// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

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
