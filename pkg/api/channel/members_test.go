// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
)

type memberTransport func(*http.Request) (*http.Response, error)

func (f memberTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMemberManagementWireContracts(t *testing.T) {
	transport := memberTransport(func(r *http.Request) (*http.Response, error) {
		status := http.StatusOK
		var body string
		switch {
		case r.URL.Path == "/app/getAppAccessToken":
			body = `{"access_token":"test-token","expires_in":7200}`
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/guilds/guild%2Fid/members":
			if r.URL.Query().Get("after") != "cursor+&=" || r.URL.Query().Get("limit") != "100" {
				t.Errorf("wrong member cursor: %q", r.URL.RawQuery)
			}
			body = `[{"user":{"id":"next","union_openid":"linked"},"joined_at":"2025-01-02T12:00:00+08:00"}]`
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/g/roles/r/members":
			if r.URL.Query().Get("start_index") != "index+&=" || r.URL.Query().Has("after") {
				t.Errorf("wrong role member cursor: %q", r.URL.RawQuery)
			}
			body = `{"data":[{"user":{"id":"member"}}],"next":"0"}`
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/g/roles/r":
			data, _ := io.ReadAll(r.Body)
			if string(data) != `{"hoist":0}` {
				t.Errorf("role patch must preserve explicit zero and omit unset fields: %s", data)
			}
			body = `{"role_id":"r","role":{"id":"r","hoist":0}}`
		case r.Method == http.MethodPut && r.URL.Path == "/channels/c/roles/r/permissions":
			data, _ := io.ReadAll(r.Body)
			if string(data) != `{"add":"5","remove":"0"}` {
				t.Errorf("permissions must be decimal strings: %s", data)
			}
			status = http.StatusNoContent
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/g/members/u/roles/5":
			data, _ := io.ReadAll(r.Body)
			if string(data) != `{"channel":{"id":"c"}}` {
				t.Errorf("subchannel admin removal needs a body: %s", data)
			}
			status = http.StatusNoContent
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/g/mute":
			var input struct {
				Seconds string   `json:"mute_seconds"`
				UserIDs []string `json:"user_ids"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Seconds != "0" || len(input.UserIDs) != 2 {
				t.Errorf("wrong batch unmute body: %#v (%v)", input, err)
			}
			body = `{"user_ids":["u1"]}`
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			status = http.StatusNotFound
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	client := channel.NewClient(api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: transport})))
	ctx := t.Context()
	members, err := client.ListMembers(ctx, "guild/id", channel.MemberListOptions{After: "cursor+&=", Limit: 100})
	if err != nil || len(members) != 1 || members[0].User == nil || members[0].User.ID != "next" || members[0].User.UnionOpenID != "linked" || members[0].JoinedAt.IsZero() {
		t.Fatalf("members = %#v (%v)", members, err)
	}
	page, err := client.ListRoleMembers(ctx, "g", "r", channel.RoleMemberListOptions{StartIndex: "index+&="})
	if err != nil || len(page.Members) != 1 || page.Next != "0" {
		t.Fatalf("role members = %#v (%v)", page, err)
	}
	var hidden uint32
	if _, err := client.UpdateRole(ctx, "g", "r", channel.RoleOptions{Hoist: &hidden}); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateRolePermissions(ctx, "c", "r", channel.PermissionChanges{Add: channel.PermissionView | channel.PermissionSpeak}); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveMemberRole(ctx, "g", "u", "5", channel.RoleMembershipOptions{ChannelID: "c"}); err != nil {
		t.Fatal(err)
	}
	result, err := client.SetMembersMute(ctx, "g", []string{"u1", "u2"}, channel.MuteOptions{Seconds: "0"})
	if err != nil || len(result.UserIDs) != 1 || result.UserIDs[0] != "u1" {
		t.Fatalf("batch unmute = %#v (%v)", result, err)
	}
	if _, err := client.SetMembersMute(ctx, "g", nil, channel.MuteOptions{Seconds: "120"}); err == nil {
		t.Fatal("empty batch would submit a guild-wide mute request")
	}
}
