// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package types

import (
	"encoding/json"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// OpType identifies a QQ gateway or webhook operation.
// OpType 标识 QQ 网关或 Webhook 操作。
type OpType int

// Gateway and webhook operation codes.
// 网关及 Webhook 操作码。
const (
	EventOp          OpType = 0
	HeartbeatOp      OpType = 1
	IdentifyOp       OpType = 2
	ResumeOp         OpType = 6
	ReconnectOp      OpType = 7
	InvalidSessionOp OpType = 9
	HelloOp          OpType = 10
	HeartbeatACKOp   OpType = 11
	VerifyOp         OpType = 13
)

// Payload is the shared gateway/webhook envelope. D remains an owned raw JSON value.
// Payload 是共用的网关或 Webhook 信封，D 保持为独立持有的原始 JSON 值。
type Payload struct {
	ID string `json:"id,omitempty"` // Event ID, distinct from event type T.
	// 事件 ID，与事件类型 T 不同。
	S *int64 `json:"s,omitempty"` // nil when no sequence was supplied.
	// 未提供序号时为 nil。
	Op OpType          `json:"op"`
	D  json.RawMessage `json:"d"`
	T  message.EventId `json:"t,omitempty"`
}
