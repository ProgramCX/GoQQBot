// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package types

import (
	"encoding/json"

	"github.com/ProgramCX/GoQQBot/internal/foundation/message"
)

var (
	EventOp  = OpType(0)  // Event
	VerifyOp = OpType(13) // WebhookValidation
)

type OpType int

// Payload QQ webhook Request structure
type Payload struct {
	Op OpType          `json:"op"`
	D  json.RawMessage `json:"d"` // Raw Message
	T  message.EventId `json:"t"` // MessageType
}
