// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

// Package message describes QQ events and outgoing messages. Outgoing builders are immutable
// message 包描述 QQ 事件及发送消息，发送构造器是不可变
// values; they validate local protocol constraints without I/O. Platform permissions remain server-enforced.
// 值，仅校验本地协议约束，不执行 I/O；平台权限仍由服务端检查。
package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/pkg/message/text"
	markdownTypes "github.com/ProgramCX/GoQQBot/pkg/message/types"
)

// Scene selects the protocol used to send a reply.
// Scene 选择发送回复所用的协议场景。
type Scene string

// Supported message scenes. DirectMessage means channel DM, not QQ C2C.
// 支持的消息场景；DirectMessage 表示频道私信，而非 QQ 单聊。
const (
	SceneC2C           Scene = "c2c"
	SceneGroup         Scene = "group"
	SceneChannel       Scene = "channel"
	SceneDirectMessage Scene = "channel_dm"
)

// ErrInvalidMessage marks local construction errors. API failures are returned separately.
// ErrInvalidMessage 表示本地构造错误，API 故障单独返回。
var ErrInvalidMessage = errors.New("invalid outgoing message")

// ReplyTarget identifies a recipient and exactly one reply source. An empty source is rejected
// ReplyTarget 指定接收方及唯一回复来源；来源为空时拒绝发送，
// to prevent accidentally sending a proactive message when handling an unsupported event.
// 防止处理不支持的事件时意外发送主动消息。
type ReplyTarget struct {
	Scene Scene
	// ID is user_openid, group_openid, channel_id, or the DM session's guild_id for the chosen scene.
	// ID 根据场景分别为 user_openid、group_openid、channel_id 或私信会话 guild_id。
	ID string
	// MessageID is d.id of the source message, mutually exclusive with EventID.
	// MessageID 是来源消息的 d.id，与 EventID 互斥。
	MessageID string
	// EventID is the event envelope's id, never its type name or interaction d.id.
	// EventID 是事件信封的 id，不是事件类型名或交互 d.id。
	EventID string
}

// ReplyResult describes a successful send. Raw preserves newly added platform response fields.
// ReplyResult 描述成功发送结果，Raw 保留平台新增响应字段。
// It is separate from the legacy SendMessageResponse to retain that type's source compatibility.
// 与旧版 SendMessageResponse 分开，以保留旧类型的源码兼容性。
type ReplyResult struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ExtInfo   struct {
		RefIndex string `json:"ref_idx"`
	} `json:"ext_info"`
	Raw json.RawMessage `json:"-"`
}

// Outgoing is an immutable description, safe for concurrent reuse and branching.
// Outgoing 是不可变消息描述，可并发复用及派生。
// New types can use New and WithField without extending the high-level reply signatures.
// 新类型可使用 New 和 WithField，无需扩展高级回复函数签名。
// Build and send methods report chained construction errors; constructors never perform I/O.
// Build 和发送方法报告链式构造错误；构造器从不执行 I/O。
type Outgoing struct {
	fields      map[string]json.RawMessage
	kind        TypeMessage
	err         error
	columns     int
	buttons     []button.Button
	rows        [][]button.Button
	sequence    int
	strictScene bool
	fragments   []text.Fragment
}

// New starts a message with an explicit protocol type, including future platform message types.
// New 使用明确的协议类型创建消息，也支持未来的平台消息类型。
func New(kind TypeMessage) Outgoing {
	raw, _ := json.Marshal(kind) // An integer always encodes as JSON.
	// 整数始终按 JSON 编码。
	return Outgoing{kind: kind, columns: 2, fields: map[string]json.RawMessage{"msg_type": raw}}
}

// Text creates a plain-text message.
// Text 创建纯文本消息。
func Text(content string) Outgoing { return New(PLAINTEXT).WithField("content", content) }

// Markdown creates a custom Markdown message, without a template application requirement in C2C/group scenes.
// Markdown 创建自定义 Markdown 消息，单聊和群聊场景无需申请模板。
func Markdown(content string) Outgoing {
	return New(MARKDOWN).WithField("markdown", markdownTypes.Markdown{Content: content})
}

// MarkdownTemplate creates a legacy template message. Prefer Markdown for new C2C/group integrations.
// MarkdownTemplate 创建旧版模板消息，新单聊和群聊接入建议使用 Markdown。
func MarkdownTemplate(id string, params ...markdownTypes.MarkdownParams) Outgoing {
	return New(MARKDOWN).WithField("markdown", markdownTypes.Markdown{CustomTemplateId: id, Params: params})
}

// Media creates a C2C/group rich-media message from a previously uploaded file_info value.
// Media 使用已上传的 file_info 创建单聊或群聊富媒体消息。
// The upload must belong to the target scene and remain within its TTL.
// 上传结果必须属于目标场景且仍在 TTL 有效期内。
func Media(fileInfo string) Outgoing {
	return New(MEDIA).WithField("media", MediaInfo{FileInfo: fileInfo})
}

func (m Outgoing) fail(err error) Outgoing {
	if m.err == nil {
		m.err = err
	}
	return m
}
func (m Outgoing) copyFields() map[string]json.RawMessage {
	fields := make(map[string]json.RawMessage, len(m.fields))
	for k, v := range m.fields {
		fields[k] = append(json.RawMessage(nil), v...)
	}
	return fields
}

// WithField snapshots an advanced protocol field immediately, including fields introduced after this SDK.
// WithField 立即为高级协议字段创建快照，包括此 SDK 发布后引入的字段。
// Routing/reply fields (msg_id, event_id, msg_seq, is_wakeup) and msg_type are managed separately.
// 路由和回复字段（msg_id、event_id、msg_seq、is_wakeup）及 msg_type 单独管理。
// Marshal errors are returned by Build/Reply. Known invalid field combinations are still rejected.
// 序列化错误由 Build/Reply 返回，已知无效字段组合仍会被拒绝。
func (m Outgoing) WithField(name string, value any) Outgoing {
	switch name {
	case "", "msg_type", "msg_id", "event_id", "msg_seq", "is_wakeup":
		return m.fail(fmt.Errorf("%w: field %q is managed by the reply API", ErrInvalidMessage, name))
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return m.fail(fmt.Errorf("%w: encode %s: %w", ErrInvalidMessage, name, err))
	}
	m.fields = m.copyFields()
	m.fields[name] = raw
	return m
}

// Buttons appends buttons for automatic left-to-right layout. The default is two columns.
// Buttons 追加按钮并自动从左到右排列，默认两列。
// It cannot be combined with Row or a raw/template Keyboard.
// 不能与 Row 或原始或模板 Keyboard 混用。
// Build rejects layouts above five rows: the default fits ten buttons; use Columns(3) for up to fifteen.
// Build 拒绝超过五行的布局：默认最多十个按钮，Columns(3) 可容纳十五个。
func (m Outgoing) Buttons(buttons ...button.Button) Outgoing {
	m.buttons = append(append([]button.Button(nil), m.buttons...), buttons...)
	return m
}

// Columns sets the automatic layout width, in 1..5. It may appear before or after Buttons.
// Columns 设置自动排列列数，范围 1..5，可在 Buttons 前后调用。
// Build reports overflow; columns are never widened implicitly and buttons are never dropped.
// Build 报告布局超限，不隐式增加列数，也不丢弃按钮。
func (m Outgoing) Columns(columns int) Outgoing { m.columns = columns; return m }

// Strict rejects unsupported scene-specific button/text command hints instead of falling back to prefill.
// Strict 拒绝场景不支持的按钮或文本命令提示，不降级为输入框预填。
// Defaults adapt enter/anchor and channel choosers to their supported scenes. Permissions are never relaxed.
// 默认按场景适配 enter/anchor 及频道选择器，绝不放宽权限。
func (m Outgoing) Strict() Outgoing { m.strictScene = true; return m }

// Row appends an explicitly arranged row; use it instead of Buttons for uneven layouts.
// Row 追加显式布局的一行；不等长布局请用它替代 Buttons。
func (m Outgoing) Row(buttons ...button.Button) Outgoing {
	m.rows = append(append([][]button.Button(nil), m.rows...), append([]button.Button(nil), buttons...))
	return m
}

// Keyboard snapshots a raw keyboard. Template ID and Content are mutually exclusive.
// Keyboard 为原始键盘创建快照，模板 ID 与 Content 互斥。
func (m Outgoing) Keyboard(keyboard *interaction.Keyboard) Outgoing {
	return m.WithField("keyboard", keyboard)
}

// KeyboardTemplate uses a previously configured keyboard template ID.
// KeyboardTemplate 使用已配置的键盘模板 ID。
func (m Outgoing) KeyboardTemplate(id string) Outgoing {
	return m.Keyboard(&interaction.Keyboard{ID: id})
}

// Reference adds a display quote using a message index (REFIDX), not the msg_id used for passive replies.
// Reference 使用消息索引 REFIDX 添加展示引用，而非被动回复使用的 msg_id。
func (m Outgoing) Reference(index string) Outgoing {
	if strings.TrimSpace(index) == "" {
		return m.fail(fmt.Errorf("%w: reference index is required", ErrInvalidMessage))
	}
	return m.WithField("message_reference", MessageReference{MessageID: index})
}

// Ark snapshots a legacy ARK template. Availability depends on the target scene and platform permissions.
// Ark 为旧版 ARK 模板创建快照，是否可用取决于目标场景和平台权限。
// Incoming ARKData is a different protocol and cannot be passed as an outgoing template.
// 接收的 ARKData 属于另一套协议，不能作为发送模板传入。
func Ark(content markdownTypes.Ark) Outgoing { return New(ARK).WithField("ark", content) }

// VerifyImages asks the platform to fail a Markdown send if image transfer fails.
// VerifyImages 要求平台在图片转存失败时使 Markdown 发送失败。
func (m Outgoing) VerifyImages() Outgoing {
	var markdown map[string]json.RawMessage
	if err := json.Unmarshal(m.fields["markdown"], &markdown); err != nil || markdown == nil {
		return m.fail(fmt.Errorf("%w: VerifyImages requires Markdown", ErrInvalidMessage))
	}
	markdown["force_verify_image_resource"] = json.RawMessage("true")
	return m.WithField("markdown", markdown)
}

// Sequence supplies an explicit positive C2C/group reply sequence for controlled retries.
// Sequence 提供明确且为正数的单聊或群聊回复序号，用于受控重试。
// Omit it for context-owned automatic numbering. Failed requests still consume automatic numbers.
// 省略时由上下文自动编号；失败请求仍消耗自动序号。
func (m Outgoing) Sequence(sequence int) Outgoing {
	if sequence <= 0 {
		return m.fail(fmt.Errorf("%w: sequence must be positive", ErrInvalidMessage))
	}
	m.sequence = sequence
	return m
}

// SequenceNumber returns the explicit sequence, or zero when automatic numbering was requested.
// SequenceNumber 返回显式序号，自动编号时返回零。
func (m Outgoing) SequenceNumber() int { return m.sequence }

// Build returns an owned payload with scene-specific constraints validated. It does not add reply IDs.
// Build 校验场景约束并返回独立持有的载荷，不添加回复 ID。
// Unknown extension fields and render styles are preserved for forward compatibility.
// 保留未知扩展字段及渲染样式，以便向前兼容。
func (m Outgoing) Build(scene Scene) (map[string]json.RawMessage, error) {
	if m.err != nil {
		return nil, m.err
	}
	invalid := func(reason string) (map[string]json.RawMessage, error) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidMessage, reason)
	}
	if m.fields == nil {
		return invalid("zero message; use Text, Markdown or New")
	}
	switch scene {
	case SceneC2C, SceneGroup, SceneChannel, SceneDirectMessage:
	default:
		return invalid("unsupported scene")
	}
	if m.columns < 1 || m.columns > 5 {
		return invalid("columns must be in 1..5")
	}
	fields := m.copyFields()
	if len(m.fragments) > 0 {
		var content string
		var markdown map[string]json.RawMessage
		switch m.kind {
		case PLAINTEXT:
			if err := json.Unmarshal(fields["content"], &content); err != nil {
				return invalid("text fragments require text content")
			}
		case MARKDOWN:
			if err := json.Unmarshal(fields["markdown"], &markdown); err != nil || markdown == nil {
				return invalid("Markdown fragments require a Markdown object")
			}
			if raw, ok := markdown["content"]; ok {
				if err := json.Unmarshal(raw, &content); err != nil {
					return invalid("Markdown content must be a string")
				}
			}
		default:
			return invalid("fragments require a text or Markdown message")
		}
		for _, fragment := range m.fragments {
			rendered, err := fragment.Render(text.Options{Scene: string(scene), Markdown: m.kind == MARKDOWN, Strict: m.strictScene})
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidMessage, err)
			}
			content += rendered
		}
		encoded, err := json.Marshal(content)
		if err != nil {
			return nil, err
		}
		if m.kind == PLAINTEXT {
			fields["content"] = encoded
		} else {
			markdown["content"] = encoded
			fields["markdown"], err = json.Marshal(markdown)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(fields) == 1 {
		return invalid("a message content field is required")
	}
	contentExists := func(key string) bool { v, ok := fields[key]; return ok && string(v) != "null" }
	// Known body types select one content object; extension message types remain open.
	// 已知正文类型选择唯一内容对象，扩展消息类型保持开放。
	selected := map[TypeMessage]string{PLAINTEXT: "content", MARKDOWN: "markdown", MEDIA: "media", CARD: "card", INPUT_NOTIFY: "input_notify"}
	if key, known := selected[m.kind]; known {
		for _, other := range []string{"content", "markdown", "media", "card", "input_notify", "ark"} {
			if other != key && contentExists(other) {
				return invalid("message type conflicts with " + other)
			}
		}
	}
	switch m.kind {
	case INPUT_NOTIFY:
		if scene != SceneC2C {
			return invalid("typing status requires C2C")
		}
		var notify InputNotify
		if json.Unmarshal(fields["input_notify"], &notify) != nil || notify.Type != 1 || notify.Seconds < 1 || notify.Seconds > 60 {
			return invalid("typing status requires input_type=1 and seconds in 1..60")
		}
	case CARD:
		if scene != SceneC2C && scene != SceneGroup {
			return invalid("card requires C2C or group")
		}
		var card struct {
			Type    string          `json:"type"`
			Content json.RawMessage `json:"content"`
		}
		if json.Unmarshal(fields["card"], &card) != nil || card.Type == "" || len(card.Content) == 0 || string(card.Content) == "null" {
			return invalid("card type and content are required")
		}
		if card.Type == "tuwen" {
			var content ImageTextCard
			if json.Unmarshal(card.Content, &content) != nil {
				return invalid("invalid image-text card")
			}
			if err := validateImageTextCard(content); err != nil {
				return nil, err
			}
		}
	case PLAINTEXT:
		var text string
		if json.Unmarshal(fields["content"], &text) != nil || strings.TrimSpace(text) == "" {
			return invalid("text content is required")
		}
		if contentExists("markdown") {
			return invalid("text and Markdown cannot be combined")
		}
	case MARKDOWN:
		var markdown markdownTypes.Markdown
		if err := json.Unmarshal(fields["markdown"], &markdown); err != nil {
			return invalid("markdown object is required")
		}
		sources := 0
		if strings.TrimSpace(markdown.Content) != "" {
			sources++
		}
		if markdown.CustomTemplateId != "" {
			sources++
		}
		if markdown.TemplateID != 0 {
			sources++
		}
		if sources != 1 {
			return invalid("Markdown requires exactly one content/template source")
		}
		if markdown.Content != "" && len(markdown.Params) > 0 {
			return invalid("template params cannot accompany raw Markdown")
		}
		if contentExists("content") || contentExists("ark") || contentExists("media") || contentExists("card") {
			return invalid("Markdown must not contain content/ark/media/card")
		}
	case MEDIA:
		if scene == SceneChannel || scene == SceneDirectMessage {
			return invalid("file_info media is only supported in C2C/group scenes")
		}
		var media MediaInfo
		if json.Unmarshal(fields["media"], &media) != nil || media.FileInfo == "" {
			return invalid("media.file_info is required")
		}
	}
	_, rawKeyboard := fields["keyboard"]
	if (len(m.buttons) > 0 && len(m.rows) > 0) || (rawKeyboard && (len(m.buttons) > 0 || len(m.rows) > 0)) {
		return invalid("choose one keyboard layout: Buttons, Row, or Keyboard")
	}
	var keyboard *interaction.Keyboard
	var err error
	if len(m.buttons) > 0 {
		keyboard, err = button.Keyboard(m.columns, m.buttons...)
	}
	if len(m.rows) > 0 {
		keyboard, err = button.Rows(m.rows...)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: keyboard: %w", ErrInvalidMessage, err)
	}
	if keyboard != nil {
		fields["keyboard"], err = json.Marshal(keyboard)
		if err != nil {
			return nil, err
		}
		rawKeyboard = true
	}
	if rawKeyboard {
		if err := json.Unmarshal(fields["keyboard"], &keyboard); err != nil || keyboard == nil {
			return invalid("keyboard object is required")
		}
		if (keyboard.ID == "") == (keyboard.Content == nil) {
			return invalid("keyboard requires template ID or content, exclusively")
		}
		if keyboard.Content != nil {
			// Validate a raw keyboard through the same path without rewriting its extension fields or IDs.
			// 原始键盘通过相同路径校验，不改写其扩展字段或 ID。
			rows := make([][]button.Button, len(keyboard.Content.Rows))
			for i, row := range keyboard.Content.Rows {
				for _, v := range row.Buttons {
					if m.strictScene && scene != SceneC2C && (v.Action.Enter || v.Action.Anchor != 0) {
						return invalid("button enter/anchor is supported only in C2C")
					}
					if m.strictScene && scene != SceneChannel && v.Action.AtBotShowChannelList {
						return invalid("channel chooser requires a channel scene")
					}
					if scene != SceneChannel && scene != SceneDirectMessage && (v.Action.Permission.Type == interaction.PermissionTypeRole || len(v.Action.Permission.SpecifyRoleIDs) > 0) {
						return invalid("button roles are supported only in channel scenes")
					}
					rows[i] = append(rows[i], button.Raw(v))
				}
			}
			if _, err := button.Rows(rows...); err != nil {
				return nil, fmt.Errorf("%w: keyboard: %w", ErrInvalidMessage, err)
			}
			if !m.strictScene {
				adapted, err := adaptKeyboard(fields["keyboard"], scene)
				if err != nil {
					return nil, fmt.Errorf("%w: adapt keyboard: %w", ErrInvalidMessage, err)
				}
				fields["keyboard"] = adapted
			}
		}
	}
	if scene == SceneChannel || scene == SceneDirectMessage {
		if m.sequence != 0 {
			return invalid("msg_seq is not supported in channel scenes")
		}
		delete(fields, "msg_type")
	}
	return fields, nil
}

// adaptKeyboard changes only command hints; raw extension fields and numeric values remain intact.
// adaptKeyboard 仅修改命令提示，保留原始扩展字段及数值。
// All JSON shapes have already been checked by the keyboard decoder and button validator above.
// 所有 JSON 结构已由上方的键盘解码器和按钮校验器检查。
func adaptKeyboard(raw json.RawMessage, scene Scene) (json.RawMessage, error) {
	var keyboard map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keyboard); err != nil {
		return nil, err
	}
	var content map[string]json.RawMessage
	if err := json.Unmarshal(keyboard["content"], &content); err != nil {
		return nil, err
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(content["rows"], &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		var buttons []map[string]json.RawMessage
		if err := json.Unmarshal(row["buttons"], &buttons); err != nil {
			return nil, err
		}
		for _, b := range buttons {
			var action map[string]json.RawMessage
			if err := json.Unmarshal(b["action"], &action); err != nil {
				return nil, err
			}
			var kind interaction.ActionType
			if rawKind, ok := action["type"]; ok {
				if err := json.Unmarshal(rawKind, &kind); err != nil {
					return nil, err
				}
			}
			if kind != interaction.ActionTypeCommand {
				continue
			}
			if scene != SceneC2C {
				delete(action, "enter")
				delete(action, "anchor")
			}
			if scene != SceneChannel {
				delete(action, "at_bot_show_channel_list")
			}
			encoded, err := json.Marshal(action)
			if err != nil {
				return nil, err
			}
			b["action"] = encoded
		}
		encoded, err := json.Marshal(buttons)
		if err != nil {
			return nil, err
		}
		row["buttons"] = encoded
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	content["rows"] = encoded
	encoded, err = json.Marshal(content)
	if err != nil {
		return nil, err
	}
	keyboard["content"] = encoded
	return json.Marshal(keyboard)
}

// Append appends scene-aware text interaction fragments without modifying the original message.
// Append 追加感知场景的文本交互片段，不修改原消息。
// It supports text/Markdown content, not templates. Raw initial content remains unchanged.
// 支持文本及 Markdown 正文，不支持模板；原始初始正文保持不变。
func (m Outgoing) Append(fragments ...text.Fragment) Outgoing {
	m.fragments = append(append([]text.Fragment(nil), m.fragments...), fragments...)
	return m
}

// MarkdownParts constructs Markdown entirely from scene-aware interaction and plain-text fragments.
// MarkdownParts 完全使用感知场景的交互片段及纯文本片段构造 Markdown。
func MarkdownParts(fragments ...text.Fragment) Outgoing { return Markdown("").Append(fragments...) }

// TextParts constructs plain text entirely from scene-aware interaction and plain-text fragments.
// TextParts 完全使用感知场景的交互片段及纯文本片段构造纯文本。
func TextParts(fragments ...text.Fragment) Outgoing { return Text("").Append(fragments...) }
