// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package types

type Ark struct {
	TemplateID int     `json:"template_id,omitempty"`
	KV         []ArkKV `json:"kv,omitempty"`
}

type ArkKV struct {
	Key   string   `json:"key"`
	Value string   `json:"value,omitempty"`
	Obj   []ArkObj `json:"obj,omitempty"`
}

type ArkObj struct {
	ObjKV []ArkObjKV `json:"obj_kv"`
}

type ArkObjKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
