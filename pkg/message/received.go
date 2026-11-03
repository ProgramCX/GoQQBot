// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidReceived marks invalid incoming message JSON.
// ErrInvalidReceived 表示接收消息 JSON 无效。
var ErrInvalidReceived = errors.New("invalid received message")

// Received is an owned common view of C2C, group, channel and channel-DM messages.
// Received 是单聊、群聊、频道及频道私信消息的独立公共视图。
// Content retains original whitespace; Raw preserves the entire event body, including future fields.
// Content 保留原始空白，Raw 保留完整事件正文，包括未来字段。
// Raw and MessageScene may contain platform authorization data; do not log them indiscriminately.
// Raw 和 MessageScene 可能包含平台授权数据，不得随意记录。
type Received struct {
	EventType    EventId              `json:"-"`
	ID           string               `json:"id"`
	Content      string               `json:"content"`
	Timestamp    string               `json:"timestamp"`
	MessageType  int                  `json:"message_type"`
	Author       ReceivedAuthor       `json:"author"`
	GroupOpenID  string               `json:"group_openid,omitempty"`
	ChannelID    string               `json:"channel_id,omitempty"`
	GuildID      string               `json:"guild_id,omitempty"`
	Attachments  []ReceivedAttachment `json:"attachments,omitempty"`
	Mentions     []ReceivedAuthor     `json:"mentions,omitempty"`
	MessageScene *ReceivedScene       `json:"message_scene,omitempty"`
	ArkData      json.RawMessage      `json:"ark_data,omitempty"`
	Elements     []json.RawMessage    `json:"msg_elements,omitempty"`
	// ARK and MsgElements are typed snapshots of ArkData and Elements. Raw fields remain for compatibility.
	// ARK 和 MsgElements 是 ArkData 和 Elements 的类型化快照，原始字段保留以保持兼容。
	ARK         *ARKData        `json:"-"`
	MsgElements []MsgElement    `json:"-"`
	Raw         json.RawMessage `json:"-"`
}

// UnmarshalJSON also populates the typed card and recursive message views.
// UnmarshalJSON 同时填充类型化卡片及递归消息视图。
func (m *Received) UnmarshalJSON(data []byte) error {
	type plain Received
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if len(decoded.ArkData) > 0 {
		if err := json.Unmarshal(decoded.ArkData, &decoded.ARK); err != nil {
			return fmt.Errorf("ark_data: %w", err)
		}
	}
	for _, raw := range decoded.Elements {
		var element MsgElement
		if err := json.Unmarshal(raw, &element); err != nil {
			return fmt.Errorf("msg_elements: %w", err)
		}
		decoded.MsgElements = append(decoded.MsgElements, element)
	}
	decoded.Raw = append(json.RawMessage(nil), data...)
	*m = Received(decoded)
	return nil
}

// ReceivedAuthor keeps scene-specific OpenIDs distinct; these identifiers are not ordinary QQ account numbers.
// ReceivedAuthor 区分各场景的 OpenID；这些标识符不是普通 QQ 号码。
type ReceivedAuthor struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar,omitempty"`
	Bot              bool   `json:"bot"`
	UserOpenID       string `json:"user_openid,omitempty"`
	MemberOpenID     string `json:"member_openid,omitempty"`
	MemberRole       string `json:"member_role,omitempty"`
	UnionOpenID      string `json:"union_openid,omitempty"`
	UnionUserAccount string `json:"union_user_account,omitempty"`
}

// ReceivedAttachment exposes incoming files, images and voice results. URLs are not downloaded automatically.
// ReceivedAttachment 暴露接收的文件、图片及语音结果，不自动下载 URL。
type ReceivedAttachment struct {
	URL          string      `json:"url"`
	Filename     string      `json:"filename"`
	ContentType  ContentType `json:"content_type"`
	Size         int64       `json:"size"`
	Width        int         `json:"width"`
	Height       int         `json:"height"`
	VoiceWAVURL  string      `json:"voice_wav_url,omitempty"`
	ASRReferText string      `json:"asr_refer_text,omitempty"`
}

// ReceivedScene retains open-ended key=value message context entries.
// ReceivedScene 保留可扩展的 key=value 消息上下文条目。
type ReceivedScene struct {
	Source string   `json:"source"`
	Ext    []string `json:"ext"`
}

// IsMessageEvent recognizes the currently supported message event names, excluding interaction/lifecycle events.
// IsMessageEvent 识别当前支持的消息事件名，不包括交互或生命周期事件。
func IsMessageEvent(event EventId) bool {
	switch event {
	case C2C_MESSAGE_CREATE, GROUP_AT_MESSAGE_CREATE, GROUP_MESSAGE_CREATE, AT_MESSAGE_CREATE, MESSAGE_CREATE, DIRECT_MESSAGE_CREATE:
		return true
	default:
		return false
	}
}

// DecodeReceived decodes an owned common view without I/O. It accepts custom event names for future message formats;
// DecodeReceived 不执行 I/O，解码为独立公共视图，并接受未来消息格式的自定义事件名；
// callers classify those events explicitly. Unknown shapes remain available through Raw.
// 调用方需显式分类这些事件，未知结构仍可通过 Raw 获取。
func DecodeReceived(event EventId, data []byte) (*Received, error) {
	if event == INTERACTION_CREATE {
		return nil, fmt.Errorf("%w: use the interaction decoder for INTERACTION_CREATE", ErrInvalidReceived)
	}
	if strings.TrimSpace(string(data)) == "null" {
		return nil, fmt.Errorf("%w: message body must be an object", ErrInvalidReceived)
	}
	var result Received
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidReceived, err)
	}
	result.EventType = event
	result.Raw = append(json.RawMessage(nil), data...)
	return &result, nil
}

// Scene maps known event names to their send scene; custom events return an empty Scene.
// Scene 将已知事件名映射到发送场景，自定义事件返回空 Scene。
func (m *Received) Scene() Scene {
	if m == nil {
		return ""
	}
	switch m.EventType {
	case C2C_MESSAGE_CREATE:
		return SceneC2C
	case GROUP_AT_MESSAGE_CREATE, GROUP_MESSAGE_CREATE:
		return SceneGroup
	case AT_MESSAGE_CREATE, MESSAGE_CREATE:
		return SceneChannel
	case DIRECT_MESSAGE_CREATE:
		return SceneDirectMessage
	default:
		return ""
	}
}

// SenderID selects the sender identifier for this scene without conflating group and C2C OpenIDs.
// SenderID 按场景选择发送者标识，不混用群聊与单聊 OpenID。
// ID is the fallback used by channel/legacy message formats; an unknown/missing sender returns empty.
// ID 是频道及旧版消息格式的回退字段；发送者未知或缺失时返回空串。
func (m *Received) SenderID() string {
	if m == nil {
		return ""
	}
	switch m.Scene() {
	case SceneC2C:
		if m.Author.UserOpenID != "" {
			return m.Author.UserOpenID
		}
	case SceneGroup:
		if m.Author.MemberOpenID != "" {
			return m.Author.MemberOpenID
		}
	}
	return m.Author.ID
}

// ConversationID returns group_openid, user_openid, channel_id or the DM guild_id for known scenes.
// ConversationID 按已知场景返回群 OpenID、用户 OpenID、频道 ID 或私信 guild_id。
func (m *Received) ConversationID() string {
	if m == nil {
		return ""
	}
	switch m.Scene() {
	case SceneGroup:
		return m.GroupOpenID
	case SceneC2C:
		return m.SenderID()
	case SceneChannel:
		return m.ChannelID
	case SceneDirectMessage:
		return m.GuildID
	default:
		return ""
	}
}
