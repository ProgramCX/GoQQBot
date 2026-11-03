// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

// Package types contains raw Markdown and ARK models for advanced protocol integrations.
// types 包包含用于高级协议接入的原始 Markdown 和 ARK 模型。
package types

type Markdown struct {
	// TemplateID is the deprecated platform template identifier. Prefer custom Content.
	// TemplateID 是已弃用的平台模板标识，建议使用自定义 Content。
	TemplateID int `json:"template_id,omitempty"`
	// ForceVerifyImageResource aborts sending if platform image transfer fails.
	// ForceVerifyImageResource 在平台图片转存失败时中止发送。
	ForceVerifyImageResource bool             `json:"force_verify_image_resource,omitempty"`
	Content                  string           `json:"content,omitempty" validate:"omitempty"`
	CustomTemplateId         string           `json:"custom_template_id,omitempty" validate:"omitempty"`
	Params                   []MarkdownParams `json:"params,omitempty" validate:"omitempty"`
}

type MarkdownParams struct {
	Key    string   `json:"key" validate:"required"`
	Values []string `json:"values" validate:"required"`
}
