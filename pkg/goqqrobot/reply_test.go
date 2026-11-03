package goqqrobot_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/c2c"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestReplyUsesMessageIDAndEventContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(context.Background(), marker{}, "event-context")
	for _, tc := range []struct {
		event      message.EventId
		data, path string
		reply      func(*types.Context, string) error
	}{
		{message.C2C_MESSAGE_CREATE, `{"id":"message-id","author":{"user_openid":"user"}}`, "/v2/users/user/messages", c2c.ReplyPlainText},
		{message.GROUP_AT_MESSAGE_CREATE, `{"id":"message-id","group_openid":"group"}`, "/v2/groups/group/messages", group.ReplyPlainText},
	} {
		t.Run(string(tc.event), func(t *testing.T) {
			requests := 0
			client := api.NewClient("test", "example-secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				if r.Context().Value(marker{}) != "event-context" {
					t.Error("context was not propagated")
				}
				body := `{"access_token":"token","expires_in":7200}`
				if r.URL.Path != "/app/getAppAccessToken" {
					requests++
					if r.URL.Path != tc.path {
						t.Error("wrong recipient")
					}
					var payload map[string]any
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Fatal(err)
					}
					if _, present := payload["event_id"]; present {
						t.Error("ordinary reply must not send event type as event_id")
					}
					if payload["msg_id"] != "message-id" || payload["content"] != "hello" {
						t.Error("incorrect reply")
					}
					body = `{"id":"reply-id"}`
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})}))
			event := &types.Context{Context: ctx, Payload: &types.Payload{ID: "event-id", T: tc.event, D: json.RawMessage(tc.data)}, Client: client}
			if err := tc.reply(event, "hello"); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatal("reply was not sent")
			}
		})
	}
}
