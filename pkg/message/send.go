// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package message

import (
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/pkg/message/types"
)

type TypeMessage int
type EventId string

type GroupEventId EventId
type PrivateEventId EventId

const (
	PLAINTEXT TypeMessage = TypeMessage(0)
	MARKDOWN  TypeMessage = TypeMessage(2)
	ARK                   = TypeMessage(3)
	EMBED                 = TypeMessage(4)
	MEDIA                 = TypeMessage(7)
)

const (
	C2C_MSG_RECEIVE    EventId = "C2C_MSG_RECEIVE"
	C2C_MESSAGE_CREATE EventId = "C2C_MESSAGE_CREATE"
	FRIEND_ADD         EventId = "FRIEND_ADD"
	FRIEND_DEL         EventId = "FRIEND_DEL"
	C2C_MSG_REJECT     EventId = "C2C_MSG_REJECT"
)

const (
	GROUP_ADD_ROBOT          EventId = "GROUP_ADD_ROBOT"
	GROUP_DEL_ROBOT          EventId = "GROUP_DEL_ROBOT"
	GROUP_MSG_RECEIVE        EventId = "GROUP_MSG_RECEIVE"
	GROUP_MSG_REJECT         EventId = "GROUP_MSG_REJECT"
	SUBSCRIBE_MESSAGE_STATUS EventId = "SUBSCRIBE_MESSAGE_STATUS"
	GROUP_AT_MESSAGE_CREATE  EventId = "GROUP_AT_MESSAGE_CREATE"
)

type MediaInfo struct {
	FileInfo string `json:"file_info"`
}

type SendMessageResponse struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
}

type CommonSendMessage struct {
	Content          string                `json:"content,omitempty" validate:"omitempty"`
	MsgType          TypeMessage           `json:"msg_type" validate:"required"`
	Markdown         *types.Markdown       `json:"markdown,omitempty" validate:"omitempty"`
	KeyBoards        *interaction.Keyboard `json:"keyboard,omitempty" validate:"omitempty"`
	Ark              *types.Ark            `json:"ark,omitempty" validate:"omitempty"`
	Media            *MediaInfo            `json:"media,omitempty" validate:"omitempty"`
	MessageReference string                `json:"message_reference,omitempty" validate:"omitempty"`
	QQEventID        EventId               `json:"event_id,omitempty" validate:"omitempty"`
	MsgId            string                `json:"msg_id,omitempty" validate:"omitempty"`
	MsgSeq           int                   `json:"msg_seq,omitempty" validate:"omitempty"`
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
	case GroupEventId(GROUP_ADD_ROBOT),
		GroupEventId(GROUP_AT_MESSAGE_CREATE),
		GroupEventId(GROUP_DEL_ROBOT),
		GroupEventId(GROUP_MSG_RECEIVE),
		GroupEventId(GROUP_MSG_REJECT),
		GroupEventId(SUBSCRIBE_MESSAGE_STATUS):
		return true
	default:
		return false
	}
}

func (e PrivateEventId) Valid() bool {
	switch e {
	case PrivateEventId(C2C_MSG_RECEIVE),
		PrivateEventId(FRIEND_ADD),
		PrivateEventId(C2C_MESSAGE_CREATE),
		PrivateEventId(FRIEND_DEL),
		PrivateEventId(C2C_MSG_REJECT):
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
