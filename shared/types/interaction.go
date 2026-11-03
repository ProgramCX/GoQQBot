// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package types

import (
	"context"
	"errors"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

// ErrInteractionResponseAttempted prevents a second ACK attempt through the same Context.
// ErrInteractionResponseAttempted 阻止通过同一 Context 再次尝试 ACK。
// A failed first HTTP request may already have reached QQ; explicitly retry at Client level if appropriate.
// 首次 HTTP 请求失败时可能已到达 QQ，必要时请在 Client 层显式重试。
var ErrInteractionResponseAttempted = errors.New("interaction response already attempted on this context")

// Interaction decodes the current event without I/O, returning an independent typed/raw snapshot.
// Interaction 不执行 I/O，将当前事件解码为独立的类型化及原始快照。
// It does not require an API client and preserves undocumented fields in Raw.
// 不需要 API 客户端，未公开字段保留在 Raw 中。
func (c *Context) Interaction() (*interaction.InteractionCreateEvent, error) {
	if c == nil || c.Payload == nil {
		return nil, ErrInvalidReplyContext
	}
	if c.Payload.T != message.INTERACTION_CREATE {
		return nil, ErrUnsupportedReply
	}
	return interaction.Decode(c.Payload.D)
}

// AckInteraction acknowledges d.id with the handler context. It does not send a message or run business logic.
// AckInteraction 使用处理函数上下文确认 d.id，不发送消息或执行业务逻辑。
// Only types 11 and 12 require ACK; other types can be acknowledged explicitly when desired.
// 仅类型 11 和 12 必须 ACK，其他类型可按需显式确认。
func (c *Context) AckInteraction(code interaction.InteractionCallbackCode) error {
	if c == nil {
		return ErrInvalidReplyContext
	}
	return c.AckInteractionContext(c.RequestContext(), code)
}

// AckInteractionContext allows a per-operation deadline. Only one ACK attempt is allowed per Context,
// AckInteractionContext 可设置单次操作截止时间，每个 Context 仅允许一次 ACK 尝试，
// including concurrent calls and failed HTTP attempts. A context already canceled before reservation consumes no attempt.
// 包括并发调用及失败 HTTP 尝试；预留前已取消的上下文不消耗尝试次数。
// This is not cross-event deduplication: repeated delivery creates a different Context and needs application idempotency.
// 这不是跨事件去重，重复投递会创建不同 Context，需要业务层幂等。
func (c *Context) AckInteractionContext(ctx context.Context, code interaction.InteractionCallbackCode) error {
	if c == nil || c.Client == nil || ctx == nil {
		return ErrInvalidReplyContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	event, err := c.Interaction()
	if err != nil {
		return err
	}
	if !c.interactionResponseAttempted.CompareAndSwap(false, true) {
		return ErrInteractionResponseAttempted
	}
	return c.Client.AcknowledgeInteraction(ctx, event.ID, code)
}
