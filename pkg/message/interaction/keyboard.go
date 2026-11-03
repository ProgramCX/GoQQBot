// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

// Package interaction contains raw QQ keyboard, callback and reaction protocol models.
// interaction 包包含原始 QQ 键盘、回调及表态协议模型。
// Higher-level message/button builders can snapshot these models for convenient composition.
// 高级 message/button 构造器可为这些模型创建快照，便于组合。
package interaction

import "encoding/json"

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
	ID string `json:"id,omitempty"`
	// GroupID groups mutually exclusive callback buttons.
	// GroupID 将互斥回调按钮分组。
	GroupID    string           `json:"group_id,omitempty"`
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
	// ButtonStyleDanger uses red text on white.
	// ButtonStyleDanger 使用白底红字。
	ButtonStyleDanger ButtonStyle = 3
	// ButtonStylePrimary uses white text on blue.
	// ButtonStylePrimary 使用蓝底白字。
	ButtonStylePrimary ButtonStyle = 4
)

type ButtonAction struct {
	// Modal optionally requests confirmation before the action.
	// Modal 可选地要求执行操作前确认。
	Modal                *ButtonModal     `json:"modal,omitempty"`
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

// InteractionCreateEvent models all currently documented INTERACTION_CREATE variants.
// InteractionCreateEvent 为当前文档中的所有 INTERACTION_CREATE 变体建模。
// Decode preserves a complete owned raw event for future platform fields.
// Decode 保留完整独立原始事件，以兼容未来平台字段。
type InteractionCreateEvent struct {
	ApplicationID     string           `json:"application_id,omitempty"`
	Raw               json.RawMessage  `json:"-"`
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
	InteractionTypeButton               InteractionType = 11
	InteractionTypeMenu                 InteractionType = 12
	InteractionTypeFeedback             InteractionType = 13
	InteractionTypeClearSession         InteractionType = 14
	InteractionTypeStory                InteractionType = 15
	InteractionTypeSwitchModel          InteractionType = 16
	InteractionTypeUserAuthorize        InteractionType = 18
	InteractionTypeGroupAuthorize       InteractionType = 19
	InteractionTypeGroupAuthorizeStatus InteractionType = 20
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
	FeedbackOpt   string                   `json:"feedback_opt,omitempty"`
	Checked       int                      `json:"checked,omitempty"`
	Action        string                   `json:"action,omitempty"`
	MessageScene  *InteractionMessageScene `json:"message_scene,omitempty"`
	AuthorizeData *AuthorizeData           `json:"authorize_data,omitempty"`
	ButtonData    string                   `json:"button_data,omitempty"`
	ButtonID      string                   `json:"button_id,omitempty"`
	UserID        string                   `json:"user_id,omitempty"`
	FeatureID     string                   `json:"feature_id,omitempty"`
	MessageID     string                   `json:"message_id,omitempty"`
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

// ButtonModal configures a confirmation dialog. Empty captions use platform defaults.
// ButtonModal 配置确认弹窗，空标题使用平台默认值。
type ButtonModal struct {
	// Content activates confirmation when nonempty; empty content leaves the dialog disabled.
	// Content 非空时启用确认，空内容不启用弹窗。
	Content     string `json:"content"`
	ConfirmText string `json:"confirm_text,omitempty"`
	CancelText  string `json:"cancel_text,omitempty"`
}

// InteractionMessageScene contains the platform's extensible key=value context entries.
// InteractionMessageScene 包含平台可扩展的 key=value 上下文条目。
type InteractionMessageScene struct {
	Ext []string `json:"ext"`
}

// AuthorizeData describes a user/group authorization action. OptScene and Scope remain open strings.
// AuthorizeData 描述用户或群授权操作，OptScene 和 Scope 保持为开放字符串。
type AuthorizeData struct {
	OptScene string `json:"opt_scene"`
	Scope    string `json:"scope"`
}
