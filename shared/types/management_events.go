package types

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

// GroupJoinRequest 解析当前入群申请事件，可读取审批需要的申请 ID。
func (c *Context) GroupJoinRequest() (*message.GroupJoinRequestEvent, error) {
	return decodeEvent[message.GroupJoinRequestEvent](c, message.GROUP_JOIN_REQUEST)
}

// GroupMemberChange 解析当前群成员加入或退出事件。
func (c *Context) GroupMemberChange() (*message.GroupMemberEvent, error) {
	return decodeEvent[message.GroupMemberEvent](c, message.GROUP_MEMBER_ADD, message.GROUP_MEMBER_REMOVE)
}

// ChannelMessage 解析频道消息或频道私信，保留频道协议的成员、序号和引用字段。
func (c *Context) ChannelMessage() (*channel.Message, error) {
	return decodeEvent[channel.Message](c, message.AT_MESSAGE_CREATE, message.MESSAGE_CREATE, message.DIRECT_MESSAGE_CREATE)
}

// ChannelMessageAudit 解析消息审核事件，结果由事件类型区分。
func (c *Context) ChannelMessageAudit() (*channel.MessageAudited, error) {
	return decodeEvent[channel.MessageAudited](c, message.MESSAGE_AUDIT_PASS, message.MESSAGE_AUDIT_REJECT)
}

// ChannelReaction 解析频道表态添加或移除事件。
func (c *Context) ChannelReaction() (*interaction.MessageReaction, error) {
	return decodeEvent[interaction.MessageReaction](c, message.MESSAGE_REACTION_ADD, message.MESSAGE_REACTION_REMOVE)
}

func decodeEvent[T any](c *Context, events ...message.EventId) (*T, error) {
	if c == nil || c.Payload == nil || !slices.Contains(events, c.Payload.T) {
		return nil, fmt.Errorf("event type does not match decoder")
	}
	var value *T
	if err := json.Unmarshal(c.Payload.D, &value); err != nil {
		return nil, fmt.Errorf("decode event: %w", err)
	}
	if value == nil {
		return nil, fmt.Errorf("event body is null")
	}
	return value, nil
}
