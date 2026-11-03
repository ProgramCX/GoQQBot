// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

// Package text constructs QQ text-interaction fragments, including mentions and clickable commands.
// text 包构造 QQ 文本交互片段，包括艾特及可点击命令。
// Fragments are immutable. Render or a message builder validates the target scene and URL-encodes commands once.
// 片段不可变，Render 或消息构造器校验目标场景，并仅对命令进行一次 URL 编码。
package text

import (
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"unicode/utf8"
)

// ErrInvalidFragment marks empty or oversized fragment arguments.
// ErrInvalidFragment 表示片段参数为空或超长。
var ErrInvalidFragment = errors.New("invalid text interaction")

// ErrUnsupportedFragment marks a meaningful interaction unavailable in the target scene/format.
// ErrUnsupportedFragment 表示有意义的交互在目标场景或格式中不可用。
var ErrUnsupportedFragment = errors.New("unsupported text interaction")

// Options describes a render target. Scene is c2c, group, channel, or channel_dm.
// Options 描述渲染目标，Scene 为 c2c、group、channel 或 channel_dm。
// Strict rejects Enter outside C2C; default rendering falls back to an Input command.
// Strict 在单聊外拒绝 Enter，默认渲染降级为 Input 命令。
type Options struct {
	Scene    string
	Markdown bool
	Strict   bool
}

// Fragment is an immutable text or platform interaction tag. Zero renders as empty plain text.
// Fragment 是不可变文本或平台交互标签，零值渲染为空纯文本。
// Non-command interactions are not silently reinterpreted in unsupported scenes.
// 不会在不支持的场景中静默改写非命令交互的语义。
type Fragment struct {
	kind      int
	value     string
	label     string
	reference bool
}

const (
	plain = iota
	atUser
	atEveryone
	enter
	input
	channel
	emoji
)

// Plain preserves text verbatim, including Markdown syntax. Escape untrusted Markdown at application level.
// Plain 原样保留文本，包括 Markdown 语法；不可信 Markdown 需在业务层转义。
func Plain(value string) Fragment { return Fragment{kind: plain, value: value} }

// AtUser mentions a member_openid in group Markdown, or a user_openid in channel text/Markdown.
// AtUser 在群 Markdown 中艾特 member_openid，在频道文本或 Markdown 中艾特 user_openid。
func AtUser(id string) Fragment { return Fragment{kind: atUser, value: id} }

// AtEveryone creates a group/channel mention tag. Platform administrator/permission/daily limits still apply.
// AtEveryone 创建群或频道全体艾特标签，平台管理员、权限及每日次数限制仍适用。
func AtEveryone() Fragment { return Fragment{kind: atEveryone} }

// Enter creates a clickable Markdown command, automatically sent in supported C2C clients.
// Enter 创建可点击 Markdown 命令，在支持的单聊客户端中自动发送。
// Group and channel scenes fall back to Input by default. Supply raw, not pre-encoded, command text.
// 群聊及频道默认降级为 Input，请提供原始命令文本而非预编码文本。
func Enter(command string) Fragment { return Fragment{kind: enter, value: command} }

// Input creates a clickable Markdown command that prefills the input box. Raw text is URL-encoded once.
// Input 创建预填输入框的可点击 Markdown 命令，原始文本仅进行一次 URL 编码。
func Input(command string) Fragment { return Fragment{kind: input, value: command} }

// Label sets Input's optional show text. It is not supported on other fragment types.
// Label 设置 Input 的可选显示文本，其他片段类型不支持。
func (f Fragment) Label(label string) Fragment { f.label = label; return f }

// Reference makes Input prefill a quote of the original message; it is invalid on other fragments.
// Reference 让 Input 预填原消息引用，其他片段不适用。
func (f Fragment) Reference() Fragment { f.reference = true; return f }

// Channel links to a channel within the current guild; only channel scenes can render it.
// Channel 链接当前频道服务器内的子频道，仅频道场景可渲染。
func Channel(id string) Fragment { return Fragment{kind: channel, value: id} }

// Emoji renders a type-1 system emoji tag in channel scenes. Use ordinary text for type-2 emoji.
// Emoji 在频道场景渲染类型 1 系统表情标签，类型 2 表情请使用普通文本。
func Emoji(id string) Fragment { return Fragment{kind: emoji, value: id} }

// Render builds one tag for the target, escaping attribute values and encoding command text/show.
// Render 为目标构造一个标签，转义属性值并编码命令文本及显示文本。
// It does not parse or rewrite tags supplied via Plain.
// 不解析或改写通过 Plain 传入的标签。
func (f Fragment) Render(options Options) (string, error) {
	switch options.Scene {
	case "c2c", "group", "channel", "channel_dm":
	default:
		return "", fmt.Errorf("%w: unknown scene", ErrUnsupportedFragment)
	}
	isChannel := options.Scene == "channel" || options.Scene == "channel_dm"
	unsupported := func(reason string) (string, error) { return "", fmt.Errorf("%w: %s", ErrUnsupportedFragment, reason) }
	invalid := func(reason string) (string, error) { return "", fmt.Errorf("%w: %s", ErrInvalidFragment, reason) }
	if f.kind != input && (f.label != "" || f.reference) {
		return invalid("show/reference apply only to Input")
	}
	if f.kind != plain && f.kind != atEveryone && strings.TrimSpace(f.value) == "" {
		return invalid("argument is required")
	}
	escape := html.EscapeString
	encode := func(value string) string { return strings.ReplaceAll(url.QueryEscape(value), "+", "%20") }
	switch f.kind {
	case plain:
		return f.value, nil
	case atUser:
		if !isChannel && !(options.Scene == "group" && options.Markdown) {
			return unsupported("AtUser requires group Markdown or a channel scene")
		}
		return `<qqbot-at-user id="` + escape(f.value) + `" />`, nil
	case atEveryone:
		if options.Scene != "group" && !isChannel {
			return unsupported("AtEveryone requires group/channel")
		}
		return `<qqbot-at-everyone />`, nil
	case enter, input:
		if !options.Markdown {
			return unsupported("command tags require Markdown")
		}
		if utf8.RuneCountInString(f.value) > 100 || utf8.RuneCountInString(f.label) > 100 {
			return invalid("text/show may contain at most 100 characters before encoding")
		}
		if f.kind == enter && options.Scene == "c2c" {
			return `<qqbot-cmd-enter text="` + encode(f.value) + `" />`, nil
		}
		if f.kind == enter && options.Strict {
			return unsupported("Enter is supported only in C2C")
		}
		result := `<qqbot-cmd-input text="` + encode(f.value) + `"`
		if f.label != "" {
			result += ` show="` + encode(f.label) + `"`
		}
		if f.reference {
			result += ` reference="true"`
		}
		return result + ` />`, nil
	case channel:
		if !isChannel {
			return unsupported("Channel links require a channel scene")
		}
		return `<#` + escape(f.value) + `>`, nil
	case emoji:
		if !isChannel {
			return unsupported("system emoji tags require a channel scene")
		}
		return `<emoji:` + escape(f.value) + `>`, nil
	default:
		return invalid("unknown fragment")
	}
}
