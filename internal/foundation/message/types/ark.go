package types

// Ark
// reference: https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/type/ark.html
type Ark struct {
	// TemplateID: Template ID
	// 23 Link + Text List Template
	// 24 Text + Thumbnail Template
	// 37 Big Image Template
	TemplateID int `json:"template_id,omitempty"`
	// KV: Key-Value pair list
	KV []ArkKV `json:"kv,omitempty"`
}

type ArkKV struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	// Obj: Array of objects, used when the preset variable is an array
	Obj []ArkObj `json:"obj,omitempty"`
}

type ArkObj struct {
	// ObjKV: List of key-value pairs for the object
	ObjKV []ArkObjKV `json:"obj_kv"`
}

type ArkObjKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
