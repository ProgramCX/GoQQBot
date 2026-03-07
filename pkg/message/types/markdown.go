// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package types

type Markdown struct {
	Content          string           `json:"content,omitempty" validate:"omitempty"`
	CustomTemplateId string           `json:"custom_template_id,omitempty" validate:"omitempty"`
	Params           []MarkdownParams `json:"params,omitempty" validate:"omitempty"`
}

type MarkdownParams struct {
	Key    string   `json:"key" validate:"required"`
	Values []string `json:"values" validate:"required"`
}
