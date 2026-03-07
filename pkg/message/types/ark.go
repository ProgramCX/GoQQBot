// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

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
