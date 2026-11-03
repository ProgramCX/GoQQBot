package message

import (
	"encoding/json"
	"strings"
)

// MessageAttachment describes incoming media without downloading it. It is compatible with ReceivedAttachment.
// MessageAttachment 描述接收的媒体而不下载，与 ReceivedAttachment 兼容。
type MessageAttachment = ReceivedAttachment

// MessageScene carries platform context, including sensitive authorization entries.
// MessageScene 携带平台上下文，其中可能包含敏感授权条目。
type MessageScene = ReceivedScene

// Lookup returns the first matching key=value entry, preserving equals signs in the value.
// Lookup 返回首个匹配的 key=value 条目，并保留值中的等号。
func (s *ReceivedScene) Lookup(key string) (string, bool) {
	if s != nil {
		for _, entry := range s.Ext {
			k, value, ok := strings.Cut(entry, "=")
			if ok && k == key {
				return value, true
			}
		}
	}
	return "", false
}

// MessageIndex returns msg_idx for display references, never for message recall.
// MessageIndex 返回用于展示引用的 msg_idx，绝不能用于撤回消息。
func (s *ReceivedScene) MessageIndex() string { value, _ := s.Lookup("msg_idx"); return value }

// ReferencedIndex returns ref_msg_idx when the incoming message quotes another message.
// ReferencedIndex 在接收消息引用其他消息时返回 ref_msg_idx。
func (s *ReceivedScene) ReferencedIndex() string { value, _ := s.Lookup("ref_msg_idx"); return value }

// ARKData describes an incoming structured card, not the legacy outgoing ARK template.
// ARKData 描述接收的结构化卡片，不是旧版发送 ARK 模板。
// Fields retains arbitrary JSON values (including arrays and objects) and future card types.
// Fields 保留任意 JSON 值，包括数组、对象和未来卡片类型。
type ARKData struct {
	Prompt string                     `json:"prompt"`
	Type   string                     `json:"ark_type"`
	Name   string                     `json:"ark_name"`
	Fields map[string]json.RawMessage `json:"fields"`
}

// Field decodes a named card field into a caller-owned destination. Missing fields return false, nil.
// Field 将指定卡片字段解码到调用方持有的目标；字段缺失返回 false、nil。
func (a *ARKData) Field(name string, dst any) (bool, error) {
	if a == nil {
		return false, nil
	}
	data, ok := a.Fields[name]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(data, dst)
}

// MsgElement describes a recursive quote, parallel message, or forwarded chat entry.
// MsgElement 描述递归引用、并列消息或合并转发中的条目。
// Raw preserves future element fields. Mutations affect only this decoded message snapshot.
// Raw 保留未来元素字段，修改仅影响本次解码的消息快照。
type MsgElement struct {
	MessageIndex string              `json:"msg_idx"`
	Author       ReceivedAuthor      `json:"author"`
	MessageType  int                 `json:"message_type"`
	Content      string              `json:"content"`
	Attachments  []MessageAttachment `json:"attachments,omitempty"`
	ARK          *ARKData            `json:"ark_data,omitempty"`
	MsgElements  []MsgElement        `json:"msg_elements,omitempty"`
	Raw          json.RawMessage     `json:"-"`
}

// UnmarshalJSON retains an independent copy of the original element.
// UnmarshalJSON 保留原始元素的独立副本。
func (m *MsgElement) UnmarshalJSON(data []byte) error {
	type plain MsgElement
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	decoded.Raw = append(json.RawMessage(nil), data...)
	*m = MsgElement(decoded)
	return nil
}
