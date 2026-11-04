package channel_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestChannelMessageWire(t *testing.T) {
	var sends int
	transport := contentTransport(func(r *http.Request) (*http.Response, error) {
		switch r.URL.EscapedPath() {
		case "/app/getAppAccessToken":
			return contentResponse(http.StatusOK, `{"access_token":"test-token","expires_in":7200}`), nil
		case "/users/@me/dms":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["recipient_id"] != "member" || body["source_guild_id"] != "guild" {
				t.Errorf("direct session body = %v, %v", body, err)
			}
			return contentResponse(http.StatusOK, `{"guild_id":"dm/id","channel_id":"dm-channel","create_time":"1642545606"}`), nil
		case "/dms/dm%2Fid/messages":
			req := &http.Request{Body: r.Body, Header: r.Header}
			if err := req.ParseMultipartForm(4096); err != nil {
				t.Fatal(err)
			}
			defer req.MultipartForm.RemoveAll()
			if got := req.FormValue("content"); got != "hello" {
				t.Errorf("content = %q", got)
			}
			if got := req.FormValue("message_reference"); !strings.Contains(got, `"message_id":"source"`) {
				t.Errorf("reference = %q", got)
			}
			f, h, err := req.FormFile("file_image")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			b, err := io.ReadAll(f)
			if err != nil || h.Filename != "photo.png" || string(b) != "png" {
				t.Errorf("file = %s, %q, %v", h.Filename, b, err)
			}
			return contentResponse(http.StatusOK, `{"id":"posted","guild_id":"dm/id","channel_id":"dm-channel","timestamp":"2021-05-13T14:45:45+08:00","author":{"id":"bot-id","bot":true},"attachments":[{"url":"https://example.test/image.png"}]}`), nil
		case "/dms/dm%2Fid/messages/posted":
			if r.Method != http.MethodDelete || r.URL.Query().Get("hidetip") != "true" {
				t.Errorf("recall = %s %s", r.Method, r.URL)
			}
			return contentResponse(http.StatusNoContent, ""), nil
		case "/channels/a%2Fb/messages":
			sends++
			if sends == 1 {
				return contentResponse(http.StatusOK, `{"id":"legacy","timestamp":"2021-05-13T14:45:45+08:00"}`), nil
			}
			return contentResponse(http.StatusBadRequest, `{"code":304023,"data":{"message_audit":{"audit_id":"audit"}}}`), nil
		case "/channels/a%2Fb/messages/message/reactions/1/203":
			if got := r.URL.Query().Get("cookie"); got != "next+&=" {
				t.Errorf("cookie = %q", got)
			}
			return contentResponse(http.StatusOK, `{"is_end":true}`), nil
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
			return contentResponse(http.StatusNotFound, "{}"), nil
		}
	})
	base := api.NewClient("app", "secret", api.WithHTTPClient(&http.Client{Transport: transport}))
	client := channel.NewClient(base)
	ctx := t.Context()
	dm, err := client.CreateDirectSession(ctx, "guild", "member")
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.SendDirectMessage(ctx, dm.GuildID, channelmsg.ChannelSendMessage{Content: "hello", MessageReference: &channelmsg.MessageReference{MessageID: "source"}}, &channel.LocalImage{Filename: "photo.png", Data: []byte("png")})
	if err != nil || result.ID != "posted" {
		t.Fatalf("result = %#v, %v", result, err)
	}
	sent, err := channel.DecodeSentMessage(result)
	if err != nil || sent.GuildID != "dm/id" || sent.ChannelID != "dm-channel" || sent.Author.ID != "bot-id" || len(sent.Attachments) != 1 {
		t.Fatalf("decoded channel message = %#v, %v", sent, err)
	}
	if err := client.RecallDirectMessage(ctx, dm.GuildID, result.ID, channel.RecallOptions{HideTip: true}); err != nil {
		t.Fatal(err)
	}
	legacy, err := channel.SendPlainText(&types.Context{Context: ctx, Client: base}, "a/b", "hello", "source")
	if err != nil || legacy.Timestamp != time.Date(2021, 5, 13, 6, 45, 45, 0, time.UTC).Unix() {
		t.Fatalf("legacy = %#v, %v", legacy, err)
	}
	_, err = base.SendMessage(ctx, message.Target{Scene: message.SceneChannel, ID: "a/b"}, message.Text("hello"))
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) || apiErr.MessageAuditID != "audit" {
		t.Fatalf("audit error = %#v, %v", apiErr, err)
	}
	if _, err := client.ListReactionUsers(ctx, "a/b", "message", interaction.EmojiTypeSystem, "203", channel.ReactionPage{Cookie: "next+&="}); err != nil {
		t.Fatal(err)
	}
}
