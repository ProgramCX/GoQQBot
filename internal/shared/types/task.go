// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package types

import (
	"context"
)

type Task struct {
	Ctx     context.Context
	Payload *Payload
	Done    func()
}
