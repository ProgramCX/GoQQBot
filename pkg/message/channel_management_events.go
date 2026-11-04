// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package message

// 频道和子频道生命周期事件使用 GUILDS 订阅。
const (
	GUILD_CREATE   EventId = "GUILD_CREATE"
	GUILD_UPDATE   EventId = "GUILD_UPDATE"
	GUILD_DELETE   EventId = "GUILD_DELETE"
	CHANNEL_CREATE EventId = "CHANNEL_CREATE"
	CHANNEL_UPDATE EventId = "CHANNEL_UPDATE"
	CHANNEL_DELETE EventId = "CHANNEL_DELETE"
)

// 频道成员事件使用 GUILD_MEMBERS 订阅。
const (
	GUILD_MEMBER_ADD    EventId = "GUILD_MEMBER_ADD"
	GUILD_MEMBER_UPDATE EventId = "GUILD_MEMBER_UPDATE"
	GUILD_MEMBER_REMOVE EventId = "GUILD_MEMBER_REMOVE"
)
