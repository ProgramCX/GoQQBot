// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

// ErrInvalidReplyTarget marks a missing recipient, conflicting reply IDs or unsupported scene.
// ErrInvalidReplyTarget 表示接收方缺失、回复 ID 冲突或场景不支持。
var ErrInvalidReplyTarget = errors.New("invalid reply target")

// SendReply sends one message to an explicit scene/recipient. It is the shared engine for Context replies.
// SendReply 向明确的场景及接收方发送一条消息，是 Context 回复的共用实现。
// It never retries message POSTs automatically. For controlled retries, use Outgoing.Sequence.
// 不自动重试消息 POST；受控重试可使用 Outgoing.Sequence。
// A successful send followed by a decode failure returns a non-nil result with Raw and a non-nil error;
// 发送成功但解码失败时，返回包含 Raw 的非 nil 结果及非 nil 错误；
// do not blindly resend because the message may already have been delivered.
// 消息可能已送达，不得盲目重发。
func (c *Client) SendReply(ctx context.Context, target message.ReplyTarget, outgoing message.Outgoing) (*message.ReplyResult, error) {
	return c.sendMessage(ctx, target, outgoing, false, false)
}

// SendMessage explicitly sends proactively, subject to QQ permissions and quotas.
// SendMessage 显式主动发送消息，受 QQ 权限及配额约束。
// Prefer SendReply when replying to an event; this method never invents a passive reply source.
// 回复事件时优先使用 SendReply；此方法从不虚构被动回复来源。
func (c *Client) SendMessage(ctx context.Context, target message.Target, outgoing message.Outgoing) (*message.ReplyResult, error) {
	return c.sendMessage(ctx, message.ReplyTarget{Scene: target.Scene, ID: target.ID}, outgoing, true, false)
}

// SendWakeup sends an explicit C2C wakeup message without msg_id/event_id.
// SendWakeup 显式发送不带 msg_id/event_id 的单聊唤醒消息。
// QQ enforces eligibility and quotas. It shares validation and response handling with SendReply.
// QQ 检查资格及配额；校验和响应处理与 SendReply 共用。
func (c *Client) SendWakeup(ctx context.Context, userOpenID string, outgoing message.Outgoing) (*message.ReplyResult, error) {
	return c.sendMessage(ctx, message.ReplyTarget{Scene: message.SceneC2C, ID: userOpenID}, outgoing, true, true)
}

func (c *Client) sendMessage(ctx context.Context, target message.ReplyTarget, outgoing message.Outgoing, proactive, wakeup bool) (*message.ReplyResult, error) {
	if c == nil || ctx == nil {
		return nil, fmt.Errorf("%w: API client and operation context are required", ErrInvalidReplyTarget)
	}
	if strings.TrimSpace(target.ID) == "" {
		return nil, fmt.Errorf("%w: recipient ID is required", ErrInvalidReplyTarget)
	}
	if !proactive && (target.MessageID == "") == (target.EventID == "") {
		return nil, fmt.Errorf("%w: supply exactly one msg_id or event_id", ErrInvalidReplyTarget)
	}
	var route string
	switch target.Scene {
	case message.SceneC2C:
		route = C2C_SEND_MESSAGE
	case message.SceneGroup:
		route = GROUP_SEND_MESSAGE
	case message.SceneChannel:
		route = CHANNEL_SEND_MESSAGE
	case message.SceneDirectMessage:
		route = CHANNEL_PRIVATE_MESSAGE
	default:
		return nil, fmt.Errorf("%w: unsupported scene %q", ErrInvalidReplyTarget, target.Scene)
	}
	body, err := outgoing.Build(target.Scene)
	if err != nil {
		return nil, err
	}
	put := func(key string, value any) { raw, _ := json.Marshal(value); body[key] = raw } // Only strings, integers and booleans are passed.
	// 仅传入字符串、整数及布尔值。
	if target.MessageID != "" {
		put("msg_id", target.MessageID)
	} else if target.EventID != "" {
		put("event_id", target.EventID)
	}
	if wakeup {
		put("is_wakeup", true)
	}
	if target.Scene == message.SceneC2C || target.Scene == message.SceneGroup {
		if target.MessageID != "" {
			sequence := outgoing.SequenceNumber()
			if sequence == 0 {
				sequence = 1
			}
			put("msg_seq", sequence)
		} else if outgoing.SequenceNumber() != 0 {
			return nil, fmt.Errorf("%w: msg_seq requires msg_id", message.ErrInvalidMessage)
		}
	}
	data, err := c.Post(ctx, fmt.Sprintf(route, url.PathEscape(target.ID)), body)
	if err != nil {
		return nil, err
	}
	result := &message.ReplyResult{Raw: append(json.RawMessage(nil), data...)}
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful reply response (message may have been sent): %w", err)
	}
	return result, nil
}

// SendStreamChunk sends one C2C streaming update without hidden buffering, retries or goroutines.
// SendStreamChunk 发送一次单聊流式更新，不隐式缓冲、重试或启动协程。
// It requires a reply source unless chunk.Wakeup explicitly requests a wakeup message.
// 必须提供回复来源，除非 chunk.Wakeup 显式请求唤醒消息。
// A non-nil result with an error means HTTP succeeded; consult Raw before deciding how to recover.
// 同时返回结果和错误表示 HTTP 成功；决定恢复方式前请检查 Raw。
func (c *Client) SendStreamChunk(ctx context.Context, target message.ReplyTarget, chunk message.StreamChunk) (*message.StreamResult, error) {
	path, err := conversationPath(ctx, c, message.Target{Scene: target.Scene, ID: target.ID})
	if err != nil {
		return nil, err
	}
	if target.Scene != message.SceneC2C || (target.MessageID != "" && target.EventID != "") || (!chunk.Wakeup && target.MessageID == "" && target.EventID == "") {
		return nil, fmt.Errorf("%w: streaming requires C2C and a reply source or explicit wakeup", ErrInvalidReplyTarget)
	}
	if (chunk.InputMode != "" && chunk.InputMode != "append" && chunk.InputMode != "replace") || (chunk.InputState != 1 && chunk.InputState != 10) || chunk.Index < 0 || chunk.Sequence < 0 || (chunk.ContentType != "text" && chunk.ContentType != "markdown") || (chunk.Index > 0 && chunk.StreamMessageID == "") || (chunk.Index == 0 && chunk.StreamMessageID != "") {
		return nil, fmt.Errorf("%w: invalid stream chunk", message.ErrInvalidMessage)
	}
	body := struct {
		message.StreamChunk
		MessageID string `json:"msg_id,omitempty"`
		EventID   string `json:"event_id,omitempty"`
	}{chunk, target.MessageID, target.EventID}
	data, err := c.Post(ctx, path+"/stream_messages", body)
	if err != nil {
		return nil, err
	}
	result := &message.StreamResult{ReplyResult: message.ReplyResult{Raw: append(json.RawMessage(nil), data...)}}
	if err := json.Unmarshal(data, result); err != nil {
		return result, fmt.Errorf("decode successful stream response (message may have been sent): %w", err)
	}
	return result, nil
}

// AcknowledgeInteraction acknowledges an interaction's d.id, not the envelope event ID.
// AcknowledgeInteraction 确认交互的 d.id，而不是信封事件 ID。
// Callback handlers should call it promptly; command callback scenarios have a three-second deadline.
// 回调处理函数应及时调用；命令回调场景有三秒截止时间。
func (c *Client) AcknowledgeInteraction(ctx context.Context, id string, code interaction.InteractionCallbackCode) error {
	if c == nil || ctx == nil || strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: interaction ID is required", ErrInvalidReplyTarget)
	}
	_, err := c.Put(ctx, fmt.Sprintf(INTERACTION_CALLBACK, url.PathEscape(id)), struct {
		Code interaction.InteractionCallbackCode `json:"code"`
	}{code})
	return err
}
