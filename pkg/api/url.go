// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package api

var (
	C2C_SEND_MESSAGE   = "/v2/users/%s/messages"
	C2C_FILE_MESSAGE   = "/v2/users/%s/files"
	C2C_DELETE_MESSAGE = "/v2/users/%s/messages/%s"

	GROUP_FILE_MESSAGE   = "/v2/groups/%s/files"
	GROUP_SEND_MESSAGE   = "/v2/groups/%s/messages"
	GROUP_DELETE_MESSAGE = "/v2/groups/%s/messages/%s"

	CHANNEL_SEND_MESSAGE    = "/channels/%s/messages"
	CHANNEL_PRIVATE_MESSAGE = "/dms/%s/messages"
	CHANNEL_DELETE_MESSAGE  = "/channels/%s/messages/%s"
	CHANNEL_PRIVATE_DELETE  = "/dms/%s/messages/%s"

	CHANNEL_EMOJI = "/channels/%s/messages/%s/reactions/%s/%s"

	INTERACTION_CALLBACK = "/interactions/%s"
)
