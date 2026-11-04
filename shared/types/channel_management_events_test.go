// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package types_test

import (
	"encoding/json"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestChannelManagementEventDecoders(t *testing.T) {
	guild := &types.Context{Payload: &types.Payload{
		T: message.GUILD_CREATE,
		D: json.RawMessage(`{"id":"g","channels":[{"id":"c","guild_id":"g"}],"op_user_id":"operator"}`),
	}}
	change, err := guild.GuildChange()
	if err != nil || change.ID != "g" || change.OpUserID != "operator" || len(change.Channels) != 1 || change.Channels[0].ID != "c" {
		t.Fatalf("guild = %#v (%v)", change, err)
	}
	subchannel := &types.Context{Payload: &types.Payload{
		T: message.CHANNEL_UPDATE,
		D: json.RawMessage(`{"id":"c","guild_id":"g","op_user_id":"operator","speak_permission":2}`),
	}}
	update, err := subchannel.SubchannelChange()
	if err != nil || update.GuildID != "g" || update.OpUserID != "operator" || update.SpeakPermission != 2 {
		t.Fatalf("subchannel = %#v (%v)", update, err)
	}
	member := &types.Context{Payload: &types.Payload{
		T: message.GUILD_MEMBER_UPDATE,
		D: json.RawMessage(`{"guild_id":"g","joined_at":"","roles":["2"],"user":{"id":"u"},"op_user_id":"operator"}`),
	}}
	profile, err := member.GuildMemberChange()
	if err != nil || profile.User == nil || profile.User.ID != "u" || profile.OpUserID != "operator" || len(profile.Roles) != 1 {
		t.Fatalf("member = %#v (%v)", profile, err)
	}
	if _, err := member.SubchannelChange(); err == nil {
		t.Fatal("member event was accepted as a subchannel event")
	}
}
