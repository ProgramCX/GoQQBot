// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package message

import (
	"fmt"

	"github.com/ProgramCX/GoQQBot/internal/foundation/message/interaction"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message/types"
)

type TypeMessage int
type EventId string

type GroupEventId EventId
type PrivateEventId EventId

// Message Types
const (
	PLAINTEXT TypeMessage = TypeMessage(0)
	MARKDOWN  TypeMessage = TypeMessage(2)
	ARK                   = TypeMessage(3)
	EMBED                 = TypeMessage(4)
	MEDIA                 = TypeMessage(7)
)

// EventIds
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/send-receive/send.html
const (
	INTERACTION_CREATE EventId = "INTERACTION_CREATE"
	C2C_MSG_RECEIVE    EventId = "C2C_MSG_RECEIVE"
	FRIEND_ADD         EventId = "FRIEND_ADD"
	GROUP_ADD_ROBOT    EventId = "GROUP_ADD_ROBOT"
	GROUP_MSG_RECEIVE  EventId = "GROUP_MSG_RECEIVE"
)

type CommonSendMessage struct {
	Content          string               `json:"content" validate:"omitempty"`
	MsgType          string               `json:"msg_type" validate:"required"`
	Markdown         types.Markdown       `json:"markdown" validate:"omitempty"`
	KeyBoards        interaction.Keyboard `json:"keyboard" validate:"omitempty"`
	Ark              types.Ark            `json:"ark" validate:"omitempty"`
	MessageReference string               `json:"message_reference" validate:"omitempty"` // Not Supported Yet
	QQEventID        EventId              `json:"event_id" validate:"omitempty"`
	MsgId            string               `json:"msg_id" validate:"omitempty"`
	MsgSeq           int                  `json:"msg_seq" validate:"omitempty"`
}

type GroupSendMessage struct {
	CommonSendMessage
}

type PrivateSendMessage struct {
	CommonSendMessage
	IsWakeup bool `json:"is_wakeup" validate:"omitempty"`
}

func (e GroupEventId) Valid() bool {
	switch e {
	case GroupEventId(GROUP_ADD_ROBOT), GroupEventId(GROUP_MSG_RECEIVE), GroupEventId(INTERACTION_CREATE):
		return true
	default:
		return false
	}
}

func (e PrivateEventId) Valid() bool {
	switch e {
	case PrivateEventId(C2C_MSG_RECEIVE), PrivateEventId(FRIEND_ADD):
		return true
	default:
		return false
	}
}

func (e GroupEventId) SetGroupEventId(id EventId) error {
	if !e.Valid() {
		return fmt.Errorf("GroupEventId %s is not valid", id)
	}
	return nil
}
