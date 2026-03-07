// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package types

import "github.com/ProgramCX/GoQQBot/pkg/api"

type Context struct {
	Payload    *Payload
	Command    string
	Args       []string
	RawMessage string
	Client     *api.Client
}
