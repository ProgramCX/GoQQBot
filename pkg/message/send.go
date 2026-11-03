// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package message

import (
	"encoding/json"
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
	// INPUT_NOTIFY and CARD are C2C typing status and C2C/group image-text cards.
	// INPUT_NOTIFY 和 CARD 分别为单聊输入状态及单聊或群聊图文卡片。
	INPUT_NOTIFY = TypeMessage(6)
	CARD         = TypeMessage(8)
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
	// GROUP_MESSAGE_CREATE is the opt-in full group-message event.
	// GROUP_MESSAGE_CREATE 是需显式开通的群聊全量消息事件。
	GROUP_MESSAGE_CREATE EventId = "GROUP_MESSAGE_CREATE"
	// INTERACTION_CREATE is a button or menu interaction event.
	// INTERACTION_CREATE 是按钮或菜单交互事件。
	INTERACTION_CREATE EventId = "INTERACTION_CREATE"
	// Channel message events.
	// 频道消息事件。
	AT_MESSAGE_CREATE     EventId = "AT_MESSAGE_CREATE"
	MESSAGE_CREATE        EventId = "MESSAGE_CREATE"
	DIRECT_MESSAGE_CREATE EventId = "DIRECT_MESSAGE_CREATE"
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
	// Reference is the typed display reference. It takes precedence over legacy MessageReference.
	// Reference 是类型化展示引用，优先于旧版 MessageReference。
	Reference *MessageReference `json:"-"`
	QQEventID EventId           `json:"event_id,omitempty" validate:"omitempty"`
	MsgId     string            `json:"msg_id,omitempty" validate:"omitempty"`
	MsgSeq    int               `json:"msg_seq,omitempty" validate:"omitempty"`
}

type GroupSendMessage struct {
	CommonSendMessage
}

type PrivateSendMessage struct {
	CommonSendMessage
	IsWakeup bool `json:"is_wakeup" validate:"omitempty"`
}

// MessageReference holds a REFIDX display reference, distinct from a message ID for recall.
// MessageReference 保存 REFIDX 展示引用，与撤回所用的消息 ID 不同。
type MessageReference struct {
	MessageID string `json:"message_id"`
}

func (m CommonSendMessage) displayReference() *MessageReference {
	if m.Reference != nil {
		return m.Reference
	}
	if m.MessageReference != "" {
		return &MessageReference{MessageID: m.MessageReference}
	}
	return nil
}

// MarshalJSON sends the protocol's reference object while preserving the legacy string field in Go.
// MarshalJSON 发送协议规定的引用对象，同时在 Go 中保留旧版字符串字段。
func (m GroupSendMessage) MarshalJSON() ([]byte, error) {
	type plain GroupSendMessage
	return json.Marshal(struct {
		plain
		Reference *MessageReference `json:"message_reference,omitempty"`
	}{plain(m), m.displayReference()})
}

// MarshalJSON preserves C2C wakeup fields and encodes references as protocol objects.
// MarshalJSON 保留单聊唤醒字段，并将引用编码为协议对象。
func (m PrivateSendMessage) MarshalJSON() ([]byte, error) {
	type plain PrivateSendMessage
	return json.Marshal(struct {
		plain
		Reference *MessageReference `json:"message_reference,omitempty"`
	}{plain(m), m.displayReference()})
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
