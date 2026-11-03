package types_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func client(handler func(*http.Request) (*http.Response, error)) *api.Client {
	return api.NewClient("test", "example-secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/app/getAppAccessToken" {
			return response(200, `{"access_token":"test-token","expires_in":7200}`), nil
		}
		return handler(r)
	})}))
}

const sent = `{"id":"reply","timestamp":"2026-11-03T12:00:00+08:00","ext_info":{"ref_idx":"REFIDX_reply"},"future_result":true}`

func TestReplyRoutesAndPayloads(t *testing.T) {
	for _, tc := range []struct {
		name       string
		event      message.EventId
		data, path string
		channel    bool
	}{
		{"group", message.GROUP_AT_MESSAGE_CREATE, `{"id":"message","group_openid":"group"}`, "/v2/groups/group/messages", false},
		{"group-full", message.GROUP_MESSAGE_CREATE, `{"id":"message","group_openid":"group"}`, "/v2/groups/group/messages", false},
		{"c2c", message.C2C_MESSAGE_CREATE, `{"id":"message","author":{"user_openid":"user"}}`, "/v2/users/user/messages", false},
		{"channel", message.AT_MESSAGE_CREATE, `{"id":"message","channel_id":"channel"}`, "/channels/channel/messages", true},
		{"dm", message.DIRECT_MESSAGE_CREATE, `{"id":"message","guild_id":"dm-guild"}`, "/dms/dm-guild/messages", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			event := &types.Context{Payload: &types.Payload{ID: "envelope-event", T: tc.event, D: json.RawMessage(tc.data)}}
			event.Client = client(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != tc.path || r.Method != "POST" {
					t.Fatalf("bad route: %s %s", r.Method, r.URL.Path)
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if string(body["msg_id"]) != `"message"` || body["event_id"] != nil {
					t.Fatal("reply ID selection incorrect")
				}
				if body["content"] != nil || string(body["markdown"]) != `{"content":"## hello"}` || body["keyboard"] == nil {
					t.Fatalf("bad Markdown payload: %s", body)
				}
				if tc.channel {
					if body["msg_type"] != nil || body["msg_seq"] != nil {
						t.Fatal("C2C fields leaked into channel")
					}
				} else {
					if string(body["msg_type"]) != "2" || string(body["msg_seq"]) != "1" {
						t.Fatal("missing type/sequence")
					}
				}
				return response(200, sent), nil
			})
			result, err := event.SendReply(context.Background(), message.Markdown("## hello").Buttons(button.Command("帮助", "/help")))
			if err != nil {
				t.Fatal(err)
			}
			if result.ID != "reply" || result.ExtInfo.RefIndex != "REFIDX_reply" || !strings.Contains(string(result.Raw), "future_result") || calls != 1 {
				t.Fatal("response metadata missing")
			}
		})
	}
}

func TestReplySequencesAreUniqueAcrossConcurrentCalls(t *testing.T) {
	var mu sync.Mutex
	seen := map[int]bool{}
	event := &types.Context{Payload: &types.Payload{T: message.C2C_MESSAGE_CREATE, D: json.RawMessage(`{"id":"message","author":{"user_openid":"user"}}`)}}
	event.Client = client(func(r *http.Request) (*http.Response, error) {
		var body struct {
			Sequence int `json:"msg_seq"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		if seen[body.Sequence] {
			t.Errorf("duplicate sequence %d", body.Sequence)
		}
		seen[body.Sequence] = true
		return response(200, sent), nil
	})
	if err := event.Reply(message.Text("first").Sequence(5)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := event.ReplyText("hello"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	for i := 5; i <= 13; i++ {
		if !seen[i] {
			t.Fatalf("missing sequence %d", i)
		}
	}
}

func TestInteractionUsesDifferentIDsForReplyAndACK(t *testing.T) {
	calls := 0
	event := &types.Context{Payload: &types.Payload{ID: "event-id", T: message.INTERACTION_CREATE, D: json.RawMessage(`{"id":"interaction-id","scene":"group","group_openid":"group"}`)}}
	event.Client = client(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if r.Method == "PUT" {
			if r.URL.Path != "/interactions/interaction-id" || string(body["code"]) != "0" {
				t.Fatal("wrong ACK")
			}
			return response(204, ""), nil
		}
		if string(body["event_id"]) != `"event-id"` || body["msg_id"] != nil || body["msg_seq"] != nil {
			t.Fatal("interaction reply used wrong ID")
		}
		return response(200, sent), nil
	})
	if err := event.AckInteraction(interaction.InteractionCallbackCodeSuccess); err != nil {
		t.Fatal(err)
	}
	if err := event.ReplyMarkdown("done"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal("ACK and message reply must be separate calls")
	}
}

func TestInvalidContextsAndCancellationDoNotSend(t *testing.T) {
	requests := 0
	c := client(func(*http.Request) (*http.Response, error) { requests++; return response(200, sent), nil })
	unsupported := &types.Context{Client: c, Payload: &types.Payload{T: message.C2C_MSG_RECEIVE, D: json.RawMessage(`{"id":"event","openid":"user"}`)}}
	if err := unsupported.ReplyMarkdown("body"); !errors.Is(err, types.ErrUnsupportedReply) {
		t.Fatal(err)
	}
	missing := &types.Context{Client: c, Payload: &types.Payload{T: message.C2C_MESSAGE_CREATE, D: json.RawMessage(`{"id":"message"}`)}}
	if err := missing.ReplyText("body"); !errors.Is(err, types.ErrInvalidReplyContext) {
		t.Fatal(err)
	}
	var nilContext *types.Context
	if err := nilContext.ReplyText("body"); !errors.Is(err, types.ErrInvalidReplyContext) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := unsupported.ReplyContext(ctx, message.Text("body")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if requests != 0 {
		t.Fatal("invalid reply reached HTTP transport")
	}
}

func TestOperationContextAndFailures(t *testing.T) {
	type marker struct{}
	operation := context.WithValue(context.Background(), marker{}, "operation")
	event := &types.Context{Payload: &types.Payload{T: message.C2C_MESSAGE_CREATE, D: json.RawMessage(`{"id":"message","author":{"user_openid":"user"}}`)}}
	sequence := 0
	event.Client = client(func(r *http.Request) (*http.Response, error) {
		if r.Context().Value(marker{}) != "operation" {
			t.Fatal("operation context not propagated")
		}
		sequence++
		var body struct {
			Sequence int `json:"msg_seq"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if sequence != body.Sequence {
			t.Fatal("failed sends must still consume sequence numbers")
		}
		if sequence == 1 {
			return response(429, `{"err_code":40034100}`), nil
		}
		return response(200, `{"id":"reply","timestamp":"bad-time"}`), nil
	})
	err := event.ReplyContext(operation, message.Text("first"))
	var apiError *api.APIError
	if !errors.As(err, &apiError) || apiError.Code != 40034100 {
		t.Fatal(err)
	}
	result, err := event.SendReply(operation, message.Text("second"))
	if err == nil || result == nil || len(result.Raw) == 0 {
		t.Fatal("successful POST decode error must preserve raw result")
	}
	if sequence != 2 {
		t.Fatal("SDK unexpectedly retried a POST")
	}
}

func ExampleContext_ReplyMarkdown() {
	handler := func(event *types.Context) error {
		return event.ReplyMarkdown("## 操作成功",
			button.Command("帮助", "/help"),
			button.Command("绑定", "/bind "),
		)
	}
	_ = handler // Register this handler with bot.OnCommand.
	// 使用 bot.OnCommand 注册此处理函数。
}
func TestExplicitTargetsAndFutureMessageFields(t *testing.T) {
	calls := 0
	c := client(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if string(body["event_id"]) != `"real-event"` || body["msg_id"] != nil || body["msg_seq"] != nil {
			t.Fatal("explicit event target was rewritten")
		}
		if string(body["msg_type"]) != "99" || string(body["future_content"]) != `{"value":"future"}` {
			t.Fatal("future message fields lost")
		}
		return response(200, sent), nil
	})
	msg := message.New(99).WithField("future_content", map[string]string{"value": "future"})
	target := message.ReplyTarget{Scene: message.SceneGroup, ID: "group", EventID: "real-event"}
	if _, err := c.SendReply(context.Background(), target, msg); err != nil {
		t.Fatal(err)
	}
	target.MessageID = "conflicting"
	if _, err := c.SendReply(context.Background(), target, msg); !errors.Is(err, api.ErrInvalidReplyTarget) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("invalid explicit target was sent")
	}
}
func TestInteractionAcknowledgementSingleAttemptAndCodes(t *testing.T) {
	for code := 0; code <= 5; code++ {
		attempts := 0
		event := &types.Context{Payload: &types.Payload{T: message.INTERACTION_CREATE, D: json.RawMessage(`{"id":"interaction-id","type":11}`)}}
		event.Client = client(func(r *http.Request) (*http.Response, error) {
			attempts++
			if r.Method != "PUT" || r.URL.Path != "/interactions/interaction-id" {
				t.Fatal("wrong ACK endpoint")
			}
			var body map[string]int
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if got, exists := body["code"]; !exists || got != code {
				t.Fatal("callback code missing or changed")
			}
			return response(204, ""), nil
		})
		if err := event.AckInteraction(interaction.InteractionCallbackCode(code)); err != nil {
			t.Fatal(err)
		}
		if err := event.AckInteraction(interaction.InteractionCallbackCode(code)); !errors.Is(err, types.ErrInteractionResponseAttempted) {
			t.Fatal("duplicate ACK allowed")
		}
		if attempts != 1 {
			t.Fatal("duplicate ACK sent")
		}
	}
}

func TestInteractionACKIsConcurrentAndFailureAware(t *testing.T) {
	event := &types.Context{Payload: &types.Payload{T: message.INTERACTION_CREATE, D: json.RawMessage(`{"id":"interaction","type":11}`)}}
	var attempts atomic.Int32
	event.Client = client(func(*http.Request) (*http.Response, error) { attempts.Add(1); return response(204, ""), nil })
	var wg sync.WaitGroup
	var succeeded atomic.Int32
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := event.AckInteraction(0)
			if err == nil {
				succeeded.Add(1)
			} else if !errors.Is(err, types.ErrInteractionResponseAttempted) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if attempts.Load() != 1 || succeeded.Load() != 1 {
		t.Fatal("ACK reservation is not atomic")
	}
	failed := &types.Context{Payload: event.Payload}
	failure := errors.New("network failure")
	failed.Client = client(func(*http.Request) (*http.Response, error) { return nil, failure })
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := failed.AckInteractionContext(canceled, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := failed.AckInteraction(0); !errors.Is(err, failure) {
		t.Fatal("canceled context consumed the ACK attempt or network error lost")
	}
	if err := failed.AckInteraction(0); !errors.Is(err, types.ErrInteractionResponseAttempted) {
		t.Fatal("uncertain ACK delivery retried implicitly")
	}
}

func TestInteractionGetterDoesNotNeedAPIClient(t *testing.T) {
	event := &types.Context{Payload: &types.Payload{T: message.INTERACTION_CREATE, D: json.RawMessage(`{"id":"interaction","type":14,"future":true}`)}}
	decoded, err := event.Interaction()
	if err != nil || decoded.Kind() != interaction.InteractionTypeClearSession || decoded.NeedsAcknowledgement() {
		t.Fatal("optional ACK kind was not decoded")
	}
	event.Payload.D[0] = '!'
	if !json.Valid(decoded.Raw) {
		t.Fatal("event getter returned borrowed payload")
	}
}
