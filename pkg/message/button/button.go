// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

// Package button constructs immutable QQ keyboard buttons. Constructors prefill commands;
// button 包构造不可变 QQ 键盘按钮，构造器默认预填命令；
// automatic sending must be explicitly requested; message builders fall back to prefill outside C2C.
// 自动发送需显式请求，消息构造器在单聊之外降级为预填。
package button

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

// ErrInvalidButton marks a locally invalid button or layout; no network call is made.
// ErrInvalidButton 表示本地按钮或布局无效，不执行网络调用。
var ErrInvalidButton = errors.New("invalid button or keyboard")

// Button is an immutable value. Chained methods return independent values safe to reuse concurrently.
// Button 是不可变值，链式方法返回可并发复用的独立值。
// The zero value is invalid. Build returns an owned protocol snapshot.
// 零值无效，Build 返回独立持有的协议快照。
type Button struct{ value interaction.KeyboardButton }

func newButton(label, data string, kind interaction.ActionType) Button {
	return Button{interaction.KeyboardButton{
		RenderData: interaction.ButtonRenderData{Label: label},
		Action:     interaction.ButtonAction{Type: kind, Data: data, Permission: interaction.ButtonPermission{Type: interaction.PermissionTypeAll}, UnsupportTips: "Please use a client that supports buttons."},
	}}
}

// Command creates a command button that prefills data, preserving spaces and arguments exactly.
// Command 创建预填 data 的命令按钮，精确保留空格及参数。
func Command(label, data string) Button { return newButton(label, data, interaction.ActionTypeCommand) }

// Link creates a jump button. The destination may be HTTP(S) or a platform-supported app scheme.
// Link 创建跳转按钮，目标可以是 HTTP(S) 或平台支持的应用协议。
func Link(label, destination string) Button {
	return newButton(label, destination, interaction.ActionTypeJump)
}

// Callback creates a button that emits INTERACTION_CREATE. A handler must acknowledge that event.
// Callback 创建产生 INTERACTION_CREATE 的按钮，处理函数必须确认该事件。
func Callback(label, data string) Button {
	return newButton(label, data, interaction.ActionTypeCallback)
}

// Raw copies an existing protocol button, preserving advanced settings.
// Raw 复制现有协议按钮并保留高级配置。
func Raw(value interaction.KeyboardButton) Button { return Button{clone(value)} }

// ID assigns a stable ID. Omitted IDs are generated uniquely within each keyboard.
// ID 指定稳定标识，省略时在每个键盘内生成唯一 ID。
func (b Button) ID(id string) Button { b.value.ID = id; return b }

// Style selects a platform render style. Unknown styles are left to the platform for forward compatibility.
// Style 选择平台渲染样式，未知样式交由平台处理以便向前兼容。
func (b Button) Style(style interaction.ButtonStyle) Button {
	b.value.RenderData.Style = style
	return b
}

// Primary selects the blue outline style supported by existing QQ clients.
// Primary 选择现有 QQ 客户端支持的蓝色边框样式。
func (b Button) Primary() Button { return b.Style(interaction.ButtonStyleBlue) }

// Visited sets the label shown after a click; an empty value keeps the platform default.
// Visited 设置点击后显示的标签，空值保留平台默认行为。
func (b Button) Visited(label string) Button { b.value.RenderData.VisitedLabel = label; return b }

// AutoSend requests immediate command sending where supported. Message builders keep it in C2C,
// AutoSend 在支持时请求立即发送命令，消息构造器在单聊中保留它，
// fall back to prefill elsewhere, or reject it when Outgoing.Strict is selected.
// 在其他场景降级为预填；选择 Outgoing.Strict 时则拒绝。
func (b Button) AutoSend() Button { b.value.Action.Enter = true; return b }

// Prefill disables automatic sending; commands are inserted into the input box.
// Prefill 关闭自动发送，将命令插入输入框。
func (b Button) Prefill() Button { b.value.Action.Enter = false; return b }

// Reply attaches a reference to the original message when sending a command.
// Reply 在发送命令时附加对原消息的引用。
func (b Button) Reply() Button { b.value.Action.Reply = true; return b }

// Group assigns an exclusive callback-button group. It is invalid for command and jump buttons.
// Group 设置互斥回调按钮组，不适用于命令和跳转按钮。
func (b Button) Group(id string) Button { b.value.GroupID = id; return b }

// Confirm asks for confirmation before the action, with platform-default button captions.
// Confirm 在执行操作前请求确认，按钮标题使用平台默认值。
func (b Button) Confirm(text string) Button {
	b.value.Action.Modal = &interaction.ButtonModal{Content: text}
	return b
}

// Users restricts the button to specific platform user IDs and copies the supplied slice.
// Users 将按钮限制为指定平台用户 ID，并复制传入切片。
func (b Button) Users(ids ...string) Button {
	b.value.Action.Permission = interaction.ButtonPermission{Type: interaction.PermissionTypeUser, SpecifyUserIDs: append([]string(nil), ids...)}
	return b
}

// Roles restricts the button to channel roles; message.Build rejects it outside channel scenes.
// Roles 将按钮限制为指定频道身份组；message.Build 在频道外拒绝此配置。
func (b Button) Roles(ids ...string) Button {
	b.value.Action.Permission = interaction.ButtonPermission{Type: interaction.PermissionTypeRole, SpecifyRoleIDs: append([]string(nil), ids...)}
	return b
}

// AdminOnly restricts the button to administrators.
// AdminOnly 将按钮限制为仅管理员可用。
func (b Button) AdminOnly() Button {
	b.value.Action.Permission = interaction.ButtonPermission{Type: interaction.PermissionTypeAdmin}
	return b
}

// All permits everyone to use the button.
// All 允许所有人使用按钮。
func (b Button) All() Button {
	b.value.Action.Permission = interaction.ButtonPermission{Type: interaction.PermissionTypeAll}
	return b
}

// UnsupportedTips sets fallback text for clients that cannot display the button.
// UnsupportedTips 设置不支持显示按钮的客户端使用的提示文本。
func (b Button) UnsupportedTips(text string) Button { b.value.Action.UnsupportTips = text; return b }

func clone(v interaction.KeyboardButton) interaction.KeyboardButton {
	v.Action.Permission.SpecifyUserIDs = append([]string(nil), v.Action.Permission.SpecifyUserIDs...)
	v.Action.Permission.SpecifyRoleIDs = append([]string(nil), v.Action.Permission.SpecifyRoleIDs...)
	if v.Action.Modal != nil {
		m := *v.Action.Modal
		v.Action.Modal = &m
	}
	return v
}

// Build validates scene-independent constraints and returns an owned protocol button.
// Build 校验与场景无关的约束，并返回独立持有的协议按钮。
func (b Button) Build() (interaction.KeyboardButton, error) {
	v := clone(b.value)
	invalid := func(field string) (interaction.KeyboardButton, error) {
		return interaction.KeyboardButton{}, fmt.Errorf("%w: %s", ErrInvalidButton, field)
	}
	if strings.TrimSpace(v.RenderData.Label) == "" || utf8.RuneCountInString(v.RenderData.Label) > 10 {
		return invalid("label must contain 1..10 characters")
	}
	if v.Action.Type >= interaction.ActionTypeJump && v.Action.Type <= interaction.ActionTypeCommand && strings.TrimSpace(v.Action.Data) == "" {
		return invalid("action data is required")
	}
	if v.Action.Type != interaction.ActionTypeCommand && (v.Action.Enter || v.Action.Reply || v.Action.Anchor != 0 || v.Action.AtBotShowChannelList) {
		return invalid("enter/reply/anchor/channel chooser require a command action")
	}
	if v.GroupID != "" && v.Action.Type != interaction.ActionTypeCallback {
		return invalid("group_id requires a callback action")
	}
	if v.Action.Permission.Type == interaction.PermissionTypeUser && len(v.Action.Permission.SpecifyUserIDs) == 0 {
		return invalid("user permission requires IDs")
	}
	if v.Action.Permission.Type == interaction.PermissionTypeRole && len(v.Action.Permission.SpecifyRoleIDs) == 0 {
		return invalid("role permission requires IDs")
	}
	if v.Action.ClickLimit < 0 {
		return invalid("click_limit must not be negative")
	}
	if m := v.Action.Modal; m != nil {
		if utf8.RuneCountInString(m.Content) > 40 || utf8.RuneCountInString(m.ConfirmText) > 4 || utf8.RuneCountInString(m.CancelText) > 4 {
			return invalid("confirmation text/captions exceed limits")
		}
		lower := strings.ToLower(m.Content)
		if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
			return invalid("confirmation must not contain a URL")
		}
	}
	return v, nil
}

// Keyboard arranges buttons left to right in 1..5 columns, at most five rows.
// Keyboard 按从左到右的顺序排列按钮，列数为 1..5，最多五行。
// IDs are deterministic layout IDs; use ID for stable callback identities across different layouts.
// ID 根据布局确定生成；跨布局需要稳定回调标识时请显式调用 ID。
func Keyboard(columns int, buttons ...Button) (*interaction.Keyboard, error) {
	if columns < 1 || columns > 5 {
		return nil, fmt.Errorf("%w: columns must be in 1..5", ErrInvalidButton)
	}
	if len(buttons) > columns*5 {
		rowCount := 1 + (len(buttons)-1)/columns
		if len(buttons) <= 25 {
			return nil, fmt.Errorf("%w: %d buttons in %d columns require %d rows (maximum 5); use Columns(%d) or paginate", ErrInvalidButton, len(buttons), columns, rowCount, 1+(len(buttons)-1)/5)
		}
		return nil, fmt.Errorf("%w: %d buttons exceed the maximum of 25 per keyboard; paginate the menu", ErrInvalidButton, len(buttons))
	}
	var rows [][]Button
	for start := 0; start < len(buttons); start += columns {
		rows = append(rows, buttons[start:min(start+columns, len(buttons))])
	}
	return Rows(rows...)
}

// Rows builds explicitly arranged rows, each containing 1..5 buttons, with at most five rows.
// Rows 创建显式布局，每行 1..5 个按钮，最多五行。
// It rejects duplicate explicit IDs and generates IDs without colliding with them.
// 拒绝重复的显式 ID，生成的 ID 不会与它们冲突。
func Rows(rows ...[]Button) (*interaction.Keyboard, error) {
	if len(rows) == 0 || len(rows) > 5 {
		return nil, fmt.Errorf("%w: keyboard requires 1..5 rows", ErrInvalidButton)
	}
	result := &interaction.Keyboard{Content: &interaction.KeyboardContent{}}
	ids := make(map[string]bool)
	for _, row := range rows {
		if len(row) == 0 || len(row) > 5 {
			return nil, fmt.Errorf("%w: row requires 1..5 buttons", ErrInvalidButton)
		}
		out := interaction.KeyboardRow{}
		for _, b := range row {
			v, err := b.Build()
			if err != nil {
				return nil, err
			}
			if v.ID != "" {
				if ids[v.ID] {
					return nil, fmt.Errorf("%w: duplicate button ID", ErrInvalidButton)
				}
				ids[v.ID] = true
			}
			out.Buttons = append(out.Buttons, v)
		}
		result.Content.Rows = append(result.Content.Rows, out)
	}
	next := 1
	for i := range result.Content.Rows {
		for j := range result.Content.Rows[i].Buttons {
			v := &result.Content.Rows[i].Buttons[j]
			if v.ID != "" {
				continue
			}
			for ids["button_"+strconv.Itoa(next)] {
				next++
			}
			v.ID = "button_" + strconv.Itoa(next)
			ids[v.ID] = true
			next++
		}
	}
	return result, nil
}
