package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func TestSendModesAndStream(t *testing.T) {
	var bodies []map[string]any
	var paths []string
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/app/getAppAccessToken" {
			return response(`{"access_token":"token","expires_in":7200}`), nil
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, body)
		paths = append(paths, r.URL.Path)
		return response(`{"id":"sent","timestamp":"2026-11-03T20:00:00+08:00","ext_info":{"ref_idx":"REFIDX_sent"},"remain_msg_len":99}`), nil
	})}))
	ctx := context.Background()
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		if _, err := client.SendMessage(ctx, message.Target{Scene: scene, ID: "recipient"}, message.Text("proactive")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := client.SendWakeup(ctx, "recipient", message.Markdown("wakeup")); err != nil {
		t.Fatal(err)
	}
	for i, body := range bodies {
		if body["msg_id"] != nil || body["event_id"] != nil || body["msg_seq"] != nil {
			t.Fatal("proactive sent a reply source")
		}
		if (body["is_wakeup"] == true) != (i == 2) {
			t.Fatal("wrong wakeup mode")
		}
	}
	target := message.ReplyTarget{Scene: message.SceneC2C, ID: "recipient", MessageID: "incoming"}
	first, err := client.SendStreamChunk(ctx, target, message.StreamChunk{InputState: 1, ContentType: "text", Content: "hello"})
	if err != nil || first.ID != "sent" || first.RemainingLength != 99 || first.ExtInfo.RefIndex != "REFIDX_sent" || len(first.Raw) == 0 {
		t.Fatalf("stream response: %v", err)
	}
	if _, err := client.SendStreamChunk(ctx, target, message.StreamChunk{InputState: 10, Index: 1, ContentType: "markdown", Content: "world", StreamMessageID: first.ID, Sequence: 2}); err != nil {
		t.Fatal(err)
	}
	if paths[3] != "/v2/users/recipient/stream_messages" || bodies[3]["index"] != float64(0) || bodies[3]["msg_id"] != "incoming" || bodies[4]["stream_msg_id"] != "sent" || bodies[4]["input_state"] != float64(10) {
		t.Fatal("stream routing or cursor lost")
	}
	target.Scene = message.SceneGroup
	if _, err := client.SendStreamChunk(ctx, target, message.StreamChunk{}); !errors.Is(err, ErrInvalidReplyTarget) {
		t.Fatalf("group stream: %v", err)
	}
	target.Scene = message.SceneC2C
	if _, err := client.SendStreamChunk(ctx, target, message.StreamChunk{InputState: 1, ContentType: "text", Index: 1}); !errors.Is(err, message.ErrInvalidMessage) {
		t.Fatal("missing stream cursor accepted")
	}
}
