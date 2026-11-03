// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package interaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidEvent identifies missing IDs or malformed interaction JSON.
// ErrInvalidEvent 表示 ID 缺失或交互 JSON 格式错误。
var ErrInvalidEvent = errors.New("invalid interaction event")

// Decode returns an owned typed event and raw JSON snapshot, allowing unknown event types and fields.
// Decode 返回独立的类型化事件及原始 JSON 快照，允许未知事件类型和字段。
// It accepts d, not the transport envelope. Call Kind for the effective outer/data interaction type.
// 接受 d 而非传输信封，可调用 Kind 获取有效的外层或 data 交互类型。
func Decode(data []byte) (*InteractionCreateEvent, error) {
	var event InteractionCreateEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidEvent, err)
	}
	if strings.TrimSpace(event.ID) == "" {
		return nil, fmt.Errorf("%w: d.id is required", ErrInvalidEvent)
	}
	event.Raw = append(json.RawMessage(nil), data...)
	return &event, nil
}

// Kind returns the outer type, falling back to data.type when absent in a legacy event.
// Kind 返回外层 type，旧版事件缺失时回退到 data.type。
func (e *InteractionCreateEvent) Kind() InteractionType {
	if e == nil {
		return 0
	}
	if e.Type != 0 {
		return e.Type
	}
	return e.Data.Type
}

// NeedsAcknowledgement reports the two documented required-ACK types (11 and 12).
// NeedsAcknowledgement 判断是否为文档要求 ACK 的两种类型（11 和 12）。
// Other types may still be explicitly acknowledged, for example type 14 with code 0.
// 其他类型仍可显式确认，例如类型 14 使用 code 0。
func (e *InteractionCreateEvent) NeedsAcknowledgement() bool {
	return e.Kind() == InteractionTypeButton || e.Kind() == InteractionTypeMenu
}
