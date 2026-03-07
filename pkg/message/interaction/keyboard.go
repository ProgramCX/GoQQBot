// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package interaction

type Keyboard struct {
	ID      string           `json:"id,omitempty"`
	Content *KeyboardContent `json:"content,omitempty"`
}

type KeyboardContent struct {
	Rows []KeyboardRow `json:"rows"`
}

type KeyboardRow struct {
	Buttons []KeyboardButton `json:"buttons"`
}

type KeyboardButton struct {
	ID         string           `json:"id,omitempty"`
	RenderData ButtonRenderData `json:"render_data"`
	Action     ButtonAction     `json:"action"`
}

type ButtonRenderData struct {
	Label        string      `json:"label"`
	VisitedLabel string      `json:"visited_label"`
	Style        ButtonStyle `json:"style"`
}

type ButtonStyle int

const (
	ButtonStyleGray ButtonStyle = 0
	ButtonStyleBlue ButtonStyle = 1
)

type ButtonAction struct {
	Type                 ActionType       `json:"type"`
	Permission           ButtonPermission `json:"permission"`
	Data                 string           `json:"data"`
	Reply                bool             `json:"reply,omitempty"`
	Enter                bool             `json:"enter,omitempty"`
	Anchor               int              `json:"anchor,omitempty"`
	ClickLimit           int              `json:"click_limit,omitempty"`
	AtBotShowChannelList bool             `json:"at_bot_show_channel_list,omitempty"`
	UnsupportTips        string           `json:"unsupport_tips"`
}

type ActionType int

const (
	ActionTypeJump     ActionType = 0
	ActionTypeCallback ActionType = 1
	ActionTypeCommand  ActionType = 2
)

type ButtonPermission struct {
	Type           PermissionType `json:"type"`
	SpecifyUserIDs []string       `json:"specify_user_ids,omitempty"`
	SpecifyRoleIDs []string       `json:"specify_role_ids,omitempty"`
}

type PermissionType int

const (
	PermissionTypeUser  PermissionType = 0
	PermissionTypeAdmin PermissionType = 1
	PermissionTypeAll   PermissionType = 2
	PermissionTypeRole  PermissionType = 3
)

type InteractionCreateEvent struct {
	ID                string           `json:"id"`
	Type              InteractionType  `json:"type"`
	Scene             InteractionScene `json:"scene"`
	ChatType          ChatType         `json:"chat_type"`
	Timestamp         string           `json:"timestamp"`
	GuildID           string           `json:"guild_id,omitempty"`
	ChannelID         string           `json:"channel_id,omitempty"`
	UserOpenID        string           `json:"user_openid,omitempty"`
	GroupOpenID       string           `json:"group_openid,omitempty"`
	GroupMemberOpenID string           `json:"group_member_openid,omitempty"`
	Data              InteractionData  `json:"data"`
	Version           int              `json:"version"`
}

type InteractionType int

const (
	InteractionTypeButton InteractionType = 11
	InteractionTypeMenu   InteractionType = 12
)

type InteractionScene string

const (
	InteractionSceneC2C   InteractionScene = "c2c"
	InteractionSceneGroup InteractionScene = "group"
	InteractionSceneGuild InteractionScene = "guild"
)

type ChatType int

const (
	ChatTypeGuild ChatType = 0
	ChatTypeGroup ChatType = 1
	ChatTypeC2C   ChatType = 2
)

type InteractionData struct {
	Resolved InteractionResolved `json:"resolved"`
	Type     InteractionType     `json:"type,omitempty"`
}

type InteractionResolved struct {
	ButtonData string `json:"button_data,omitempty"`
	ButtonID   string `json:"button_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	FeatureID  string `json:"feature_id,omitempty"`
	MessageID  string `json:"message_id,omitempty"`
}

type InteractionCallbackCode int

const (
	InteractionCallbackCodeSuccess      InteractionCallbackCode = 0
	InteractionCallbackCodeFailed       InteractionCallbackCode = 1
	InteractionCallbackCodeFrequent     InteractionCallbackCode = 2
	InteractionCallbackCodeDuplicate    InteractionCallbackCode = 3
	InteractionCallbackCodeNoPermission InteractionCallbackCode = 4
	InteractionCallbackCodeAdminOnly    InteractionCallbackCode = 5
)
