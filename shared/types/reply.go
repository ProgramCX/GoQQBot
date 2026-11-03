// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
)

// Reply context errors can be matched with errors.Is; API errors remain available through errors.As.
// 回复上下文错误可通过 errors.Is 匹配，API 错误仍可通过 errors.As 检查。
var (
	ErrInvalidReplyContext = errors.New("missing QQ reply context")
	ErrUnsupportedReply    = errors.New("event does not support automatic reply routing")
)

// ReplyText replies to the current message with plain text.
// ReplyText 用纯文本回复当前消息。
func (c *Context) ReplyText(content string) error { return c.Reply(message.Text(content)) }

// ReplyMarkdown replies with Markdown and optional buttons arranged in two columns.
// ReplyMarkdown 使用 Markdown 及可选按钮回复，按钮按两列排列。
func (c *Context) ReplyMarkdown(content string, buttons ...button.Button) error {
	return c.Reply(message.Markdown(content).Buttons(buttons...))
}

// Reply sends using the handler context and automatic scene/recipient/source selection.
// Reply 使用处理函数上下文发送，自动选择场景、接收方及回复来源。
// Use ReplyContext for a per-operation timeout, or SendReply when response metadata is needed.
// 单次操作超时请使用 ReplyContext，需要响应元数据时使用 SendReply。
func (c *Context) Reply(outgoing message.Outgoing) error {
	if c == nil {
		return ErrInvalidReplyContext
	}
	return c.ReplyContext(c.RequestContext(), outgoing)
}

// ReplyContext sends with the supplied context propagated to authentication and HTTP I/O.
// ReplyContext 将指定上下文传递到认证及 HTTP I/O。
// It shares the same scene validation and sequence counter as Reply.
// 与 Reply 共用场景校验及序号计数器。
func (c *Context) ReplyContext(ctx context.Context, outgoing message.Outgoing) error {
	_, err := c.SendReply(ctx, outgoing)
	return err
}

// SendReply returns send metadata and an owned raw response, routing from the current event.
// SendReply 根据当前事件路由，返回发送元数据及独立持有的原始响应。
// Calls on the same Context reserve unique increasing C2C/group msg_seq values, including failed sends.
// 同一 Context 上的调用预留唯一递增的单聊或群聊 msg_seq，包括失败发送。
// An explicit Outgoing.Sequence is preserved for controlled retries and advances the automatic counter.
// 显式 Outgoing.Sequence 会保留用于受控重试，并推进自动计数器。
// Do not mix automatic replies with legacy send helpers that independently use msg_seq=1.
// 不要将自动回复与独立使用 msg_seq=1 的旧版发送方法混用。
// A non-nil result with an error means HTTP succeeded but response decoding failed; do not blindly resend.
// 同时返回结果和错误表示 HTTP 成功但响应解码失败，不得盲目重发。
func (c *Context) SendReply(ctx context.Context, outgoing message.Outgoing) (*message.ReplyResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil operation context", ErrInvalidReplyContext)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := c.replyTarget()
	if err != nil {
		return nil, err
	}
	if _, err := outgoing.Build(target.Scene); err != nil {
		return nil, err
	}
	if target.MessageID != "" && (target.Scene == message.SceneC2C || target.Scene == message.SceneGroup) {
		sequence := outgoing.SequenceNumber()
		sequence = c.reserveSequence(sequence)
		outgoing = outgoing.Sequence(sequence)
	}
	return c.Client.SendReply(ctx, target, outgoing)
}

func (c *Context) reserveSequence(sequence int) int {
	if sequence == 0 {
		return int(c.replySequence.Add(1))
	}
	for current := c.replySequence.Load(); int64(sequence) > current; current = c.replySequence.Load() {
		if c.replySequence.CompareAndSwap(current, int64(sequence)) {
			break
		}
	}
	return sequence
}

// SendStreamChunk sends a sequential C2C stream update using this event's reply source.
// SendStreamChunk 使用当前事件的回复来源，发送一次顺序单聊流式更新。
// It shares the reply sequence counter. The caller carries result.ID to the next chunk and owns stream ordering.
// 共用回复序号计数器；调用方将 result.ID 传给下一片段，并负责流顺序。
func (c *Context) SendStreamChunk(ctx context.Context, chunk message.StreamChunk) (*message.StreamResult, error) {
	if ctx == nil {
		return nil, ErrInvalidReplyContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := c.replyTarget()
	if err != nil {
		return nil, err
	}
	if target.Scene != message.SceneC2C {
		return nil, ErrUnsupportedReply
	}
	if target.MessageID != "" && chunk.Sequence >= 0 {
		chunk.Sequence = c.reserveSequence(chunk.Sequence)
	}
	return c.Client.SendStreamChunk(ctx, target, chunk)
}

func (c *Context) replyTarget() (message.ReplyTarget, error) {
	if c == nil || c.Payload == nil || c.Client == nil {
		return message.ReplyTarget{}, ErrInvalidReplyContext
	}
	var data struct {
		ID        string `json:"id"`
		GroupID   string `json:"group_openid"`
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
		UserID    string `json:"user_openid"`
		Scene     string `json:"scene"`
		Author    struct {
			UserID string `json:"user_openid"`
		} `json:"author"`
	}
	if err := json.Unmarshal(c.Payload.D, &data); err != nil {
		return message.ReplyTarget{}, fmt.Errorf("decode reply event: %w", err)
	}
	target := message.ReplyTarget{MessageID: data.ID}
	switch c.Payload.T {
	case message.C2C_MESSAGE_CREATE:
		target.Scene = message.SceneC2C
		target.ID = data.Author.UserID
	case message.GROUP_AT_MESSAGE_CREATE, message.GROUP_MESSAGE_CREATE:
		target.Scene = message.SceneGroup
		target.ID = data.GroupID
	case message.AT_MESSAGE_CREATE, message.MESSAGE_CREATE:
		target.Scene = message.SceneChannel
		target.ID = data.ChannelID
	case message.DIRECT_MESSAGE_CREATE:
		target.Scene = message.SceneDirectMessage
		target.ID = data.GuildID
	case message.INTERACTION_CREATE:
		target.MessageID = ""
		target.EventID = c.Payload.ID
		switch data.Scene {
		case "c2c":
			target.Scene = message.SceneC2C
			target.ID = data.UserID
		case "group":
			target.Scene = message.SceneGroup
			target.ID = data.GroupID
		case "guild":
			target.Scene = message.SceneChannel
			target.ID = data.ChannelID
		default:
			return message.ReplyTarget{}, fmt.Errorf("%w: interaction scene %q", ErrUnsupportedReply, data.Scene)
		}
	default:
		return message.ReplyTarget{}, fmt.Errorf("%w: %s; use Client.SendReply with an explicit ReplyTarget for other replyable events", ErrUnsupportedReply, c.Payload.T)
	}
	if target.ID == "" || (target.MessageID == "" && target.EventID == "") {
		return message.ReplyTarget{}, fmt.Errorf("%w: recipient or reply source ID missing", ErrInvalidReplyContext)
	}
	return target, nil
}
