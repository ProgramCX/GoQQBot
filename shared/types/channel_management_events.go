// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package types

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

// GuildChange 解析机器人加入频道、频道资料变更或机器人退出频道的事件。
func (c *Context) GuildChange() (*channel.Guild, error) {
	return decodeEvent[channel.Guild](c, message.GUILD_CREATE, message.GUILD_UPDATE, message.GUILD_DELETE)
}

// SubchannelChange 解析子频道的创建、更新或删除事件。
func (c *Context) SubchannelChange() (*channel.Channel, error) {
	return decodeEvent[channel.Channel](c, message.CHANNEL_CREATE, message.CHANNEL_UPDATE, message.CHANNEL_DELETE)
}

// GuildMemberChange 解析频道成员加入、更新或退出事件。
func (c *Context) GuildMemberChange() (*channel.GuildMemberEvent, error) {
	return decodeEvent[channel.GuildMemberEvent](c, message.GUILD_MEMBER_ADD, message.GUILD_MEMBER_UPDATE, message.GUILD_MEMBER_REMOVE)
}
