package types

type Markdown struct {
	Content          string           `json:"content" validate:"omitempty"`
	CustomTemplateId string           `json:"custom_template_id" validate:"omitempty"`
	Params           []MarkdownParams `json:"params" validate:"omitempty"`
}

type MarkdownParams struct {
	Key    string   `json:"key" validate:"required"`
	Values []string `json:"values" validate:"required"`
}
