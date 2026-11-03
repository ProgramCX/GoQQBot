// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package button

import "github.com/ProgramCX/GoQQBot/pkg/message/interaction"

// PickImage opens the supported C2C client's image picker (anchor=1).
// PickImage 打开支持此功能的单聊客户端图片选择器（anchor=1）。
// Other scenes fall back to prefilling the command unless Outgoing.Strict is used.
// 其他场景降级为预填命令，除非使用 Outgoing.Strict。
func (b Button) PickImage() Button { b.value.Action.Anchor = 1; return b }

// ChooseChannel requests a channel chooser for command buttons in a channel scene.
// ChooseChannel 请求在频道场景中为命令按钮打开频道选择器。
// Other scenes fall back to the current input box unless Outgoing.Strict is used.
// 其他场景降级为当前输入框，除非使用 Outgoing.Strict。
func (b Button) ChooseChannel() Button { b.value.Action.AtBotShowChannelList = true; return b }

// Confirmation copies a dialog configuration, including optional confirm/cancel captions.
// Confirmation 复制确认弹窗配置，包括可选的确认及取消标题。
func (b Button) Confirmation(modal interaction.ButtonModal) Button {
	b.value.Action.Modal = &modal
	return b
}

// ClickLimit sets the legacy click limit (zero is unlimited).
// ClickLimit 设置旧版点击次数上限，零表示不限。
// Deprecated: current platform interfaces mark click_limit deprecated; enforce business idempotency in handlers.
// Deprecated: 当前平台接口已弃用 click_limit，请在处理函数中保证业务幂等。
func (b Button) ClickLimit(limit int) Button { b.value.Action.ClickLimit = limit; return b }
