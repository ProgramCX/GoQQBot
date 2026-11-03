package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/internal/core/router"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type mockTransport func(*http.Request) (*http.Response, error)

func (f mockTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type request struct {
	method, path string
	body         map[string]json.RawMessage
}
type harness struct {
	suite          *suite
	router         *router.CommandRouter
	client         *api.Client
	requests       []request
	t              *testing.T
	eventIndex     int
	rejectMentions bool
}

func newHarness(t *testing.T, options suiteOptions) *harness {
	h := &harness{t: t, suite: newSuite(options), router: router.New(nil)}
	h.client = api.NewClient("test-app", "test-secret", api.WithHTTPClient(&http.Client{Transport: mockTransport(func(r *http.Request) (*http.Response, error) {
		body := `{"access_token":"test-token","expires_in":7200}`
		if r.URL.Path != "/app/getAppAccessToken" {
			record := request{method: r.Method, path: r.URL.Path}
			if err := json.NewDecoder(r.Body).Decode(&record.body); err != nil {
				return nil, err
			}
			h.requests = append(h.requests, record)
			if h.rejectMentions && strings.Contains(string(record.body["content"])+string(record.body["markdown"]), "qqbot-at-everyone") {
				return &http.Response{StatusCode: 403, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"err_code":40034105}`))}, nil
			}
			body = `{"id":"reply","timestamp":"2026-11-03T20:00:00+08:00"}`
			if r.Method == http.MethodPut {
				body = `{}`
			}
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}))
	h.suite.register(h.router)
	return h
}
func (h *harness) dispatch(scene, command, id string, kind int) []request {
	h.t.Helper()
	start := len(h.requests)
	h.eventIndex++
	d := map[string]any{"id": fmt.Sprintf("source-%d", h.eventIndex), "scene": scene}
	switch scene {
	case "c2c":
		d["user_openid"] = "user"
	case "group":
		d["group_openid"] = "group"
		d["group_member_openid"] = "member"
	case "guild":
		d["channel_id"] = "channel"
		d["guild_id"] = "guild"
	}
	event := message.INTERACTION_CREATE
	if command != "" {
		d["content"] = command
		switch scene {
		case "c2c":
			event = message.C2C_MESSAGE_CREATE
			d["author"] = map[string]string{"user_openid": "user"}
		case "group":
			event = message.GROUP_AT_MESSAGE_CREATE
			d["author"] = map[string]string{"member_openid": "member"}
		case "guild":
			event = message.AT_MESSAGE_CREATE
			d["author"] = map[string]string{"id": "user"}
		}
		if strings.Contains(command, "picker") || strings.Contains(command, "附件测试") {
			d["attachments"] = []any{map[string]string{"content_type": "image/png", "url": "https://example.invalid/test.png"}}
		}
	} else {
		d["type"] = kind
		d["data"] = map[string]any{"type": kind, "resolved": map[string]string{"button_id": id, "button_data": id, "feature_id": "demo-menu", "user_id": "user"}}
	}
	raw, err := json.Marshal(d)
	if err != nil {
		h.t.Fatal(err)
	}
	h.router.Handle(&types.Task{Ctx: context.Background(), Payload: &types.Payload{ID: "INTERACTION_CREATE:envelope", T: event, D: raw}, Client: h.client})
	return h.requests[start:]
}
func markdownOf(r request) string {
	var md struct {
		Content string `json:"content"`
	}
	_ = json.Unmarshal(r.body["markdown"], &md)
	return md.Content
}
func keyboardOf(t *testing.T, r request) *interaction.Keyboard {
	t.Helper()
	if r.body["keyboard"] == nil {
		return nil
	}
	var k interaction.Keyboard
	if err := json.Unmarshal(r.body["keyboard"], &k); err != nil {
		t.Fatal(err)
	}
	return &k
}

func TestEveryRenderedCaseHasAWorkingHandler(t *testing.T) {
	for _, scene := range []string{"c2c", "group", "guild"} {
		t.Run(scene, func(t *testing.T) {
			h := newHarness(t, suiteOptions{TemplateID: "test-template", MiniAppURL: "mqqapi://test-app", RoleID: "test-role", ChannelID: "test-channel", MenuFeature: "demo-menu"})
			pending := h.dispatch(scene, "/interact", "", 0)
			visited := map[string]bool{}
			commands := map[string]bool{}
			styles := map[int]bool{}
			permissions := map[int]bool{}
			jumpCount, callbackCount, templateCount := 0, 0, 0
			tags := regexp.MustCompile(`<qqbot-cmd-(?:input|enter) text="([^"]+)"`)
			for len(pending) > 0 {
				record := pending[0]
				pending = pending[1:]
				if record.method != "POST" {
					continue
				}
				if record.body["msg_id"] == nil && record.body["event_id"] == nil {
					t.Fatal("reply did not carry a valid source")
				}
				for _, match := range tags.FindAllStringSubmatch(markdownOf(record), -1) {
					command, err := url.QueryUnescape(match[1])
					if err != nil {
						t.Fatal(err)
					}
					if !commands[command] {
						commands[command] = true
						requests := h.dispatch(scene, command, "", 0)
						if len(requests) != 1 || requests[0].method != "POST" {
							t.Fatalf("text command %q has no reply handler", command)
						}
						pending = append(pending, requests...)
					}
				}
				kb := keyboardOf(t, record)
				if kb == nil {
					continue
				}
				if kb.ID != "" {
					templateCount++
					continue
				}
				if kb.Content == nil || len(kb.Content.Rows) > 5 {
					t.Fatal("invalid keyboard layout")
				}
				hasJump, hasUIConfirmation := false, false
				for _, row := range kb.Content.Rows {
					for _, b := range row.Buttons {
						styles[int(b.RenderData.Style)] = true
						permissions[int(b.Action.Permission.Type)] = true
						if b.ID == "ui-confirm" {
							hasUIConfirmation = true
						}
						switch b.Action.Type {
						case interaction.ActionTypeJump:
							jumpCount++
							hasJump = true
							// Link buttons perform client navigation and intentionally produce no backend callback.
							// 链接按钮由客户端执行跳转，按设计不产生后端回调。
						case interaction.ActionTypeCommand:
							if b.ID == "cmd-auto" && b.Action.Enter != (scene == "c2c") {
								t.Fatal("AutoSend scene projection incorrect")
							}
							if b.ID == "cmd-picker" && (b.Action.Anchor == 1) != (scene == "c2c") {
								t.Fatal("image picker projection incorrect")
							}
							if b.ID == "cmd-chooser" && b.Action.AtBotShowChannelList != (scene == "guild") {
								t.Fatal("channel chooser projection incorrect")
							}
							if !commands[b.Action.Data] {
								commands[b.Action.Data] = true
								requests := h.dispatch(scene, b.Action.Data, "", 0)
								if len(requests) != 1 {
									t.Fatalf("command %q has no result", b.Action.Data)
								}
								pending = append(pending, requests...)
							}
						case interaction.ActionTypeCallback:
							if visited[b.ID] {
								continue
							}
							visited[b.ID] = true
							callbackCount++
							before := h.suite.counts[interaction.InteractionTypeButton]
							requests := h.dispatch(scene, "", b.ID, 11)
							if h.suite.counts[interaction.InteractionTypeButton] != before {
								t.Fatalf("button %s fell through to generic handler instead of its registered handler", b.ID)
							}
							if len(requests) != 2 || requests[0].method != "PUT" || requests[1].method != "POST" {
								t.Fatalf("button %s: want ACK then reply, got %+v", b.ID, requests)
							}
							if !strings.HasPrefix(requests[0].path, "/interactions/source-") {
								t.Fatal("ACK used the envelope ID instead of d.id")
							}
							code := 0
							if strings.HasPrefix(b.ID, "ack-") {
								code, _ = strconv.Atoi(strings.TrimPrefix(b.ID, "ack-"))
							}
							if string(requests[0].body["code"]) != strconv.Itoa(code) {
								t.Fatalf("button %s sent wrong ACK code", b.ID)
							}
							pending = append(pending, requests...)
						}
					}
				}
				if hasJump && !hasUIConfirmation {
					t.Fatal("client-only navigation demo has no paired manual-confirmation handler")
				}
			}
			for _, test := range h.suite.cases {
				if !visited["case-"+test.id] {
					t.Fatalf("case %s is unreachable from the menu", test.id)
				}
			}
			for code := 0; code <= 5; code++ {
				if !visited["ack-"+strconv.Itoa(code)] {
					t.Fatal("ACK case is unreachable")
				}
			}
			for _, style := range []int{0, 1, 3, 4} {
				if !styles[style] {
					t.Fatalf("style %d not demonstrated", style)
				}
			}
			for _, permission := range []int{0, 1, 2} {
				if !permissions[permission] {
					t.Fatalf("permission %d not demonstrated", permission)
				}
			}
			if scene == "guild" && !permissions[3] {
				t.Fatal("role restriction not demonstrated")
			}
			if templateCount != 1 || jumpCount < 2 {
				t.Fatal("configured template or navigation cases were not exercised")
			}
			for _, expected := range []string{"/itest prefill 参数", "/itest auto", "/itest quote", "/itest picker", "/itest chooser", "/itest input 中文 + &", "/itest enter"} {
				if !commands[expected] {
					t.Fatalf("command %q not exercised", expected)
				}
			}
			t.Logf("%s: %d catalog cases, %d callback handlers, %d command inputs, %d client-only links; all ACK/reply routes verified", scene, len(h.suite.cases), callbackCount, len(commands), jumpCount)
		})
	}
}

func TestNativeEventHandlersAcknowledgeOnlyRequiredKinds(t *testing.T) {
	h := newHarness(t, suiteOptions{MenuFeature: "demo-menu"})
	for _, kind := range []interaction.InteractionType{11, 12, 13, 14, 15, 16, 18, 19, 20} {
		requests := h.dispatch("c2c", "", "external-platform-button", int(kind))
		if kind == 11 || kind == 12 {
			if len(requests) != 2 || requests[0].method != "PUT" || requests[1].method != "POST" {
				t.Fatalf("native type %d missing ACK/result", kind)
			}
		} else if len(requests) != 0 {
			t.Fatalf("native type %d unexpectedly sent an ACK/message", kind)
		}
		if h.suite.counts[kind] != 1 {
			t.Fatalf("native type %d not recorded exactly once", kind)
		}
	}
	requests := h.dispatch("c2c", "", "page-platform", 11)
	body := markdownOf(requests[len(requests)-1])
	for _, kind := range []int{11, 12, 13, 14, 15, 16, 18, 19, 20} {
		if !strings.Contains(body, fmt.Sprintf("type=%d：1", kind)) {
			t.Fatalf("type %d is not visible in native event report", kind)
		}
	}
}

func TestMissingPlatformFixturesGiveAnExplicitResult(t *testing.T) {
	h := newHarness(t, suiteOptions{})
	for _, id := range []string{"case-template", "case-miniapp", "case-role", "case-channel"} {
		requests := h.dispatch("guild", "", id, 11)
		if len(requests) != 2 || markdownOf(requests[1]) == "" {
			t.Fatalf("case %s silently skipped missing fixture", id)
		}
		if kb := keyboardOf(t, requests[1]); kb == nil || kb.Content == nil {
			t.Fatal("unsupported case has no return path")
		}
	}
}
func TestPlatformDenialProducesVisibleFailureFeedback(t *testing.T) {
	h := newHarness(t, suiteOptions{})
	h.rejectMentions = true
	requests := h.dispatch("group", "", "case-everyone-md", 11)
	if len(requests) != 3 || requests[0].method != "PUT" || requests[1].method != "POST" || requests[2].method != "POST" {
		t.Fatal("want ACK, rejected feature request, visible error reply")
	}
	feedback := markdownOf(requests[2])
	if !strings.Contains(feedback, "40034105") || !strings.Contains(feedback, "未通过") {
		t.Fatal("platform rejection was hidden or reported as passed")
	}
}

func TestOrdinaryMessagesAndAttachmentOnlyHaveReplies(t *testing.T) {
	h := newHarness(t, suiteOptions{})
	for _, scene := range []string{"c2c", "group", "guild"} {
		for _, content := range []string{"普通文字", "/unknown argument"} {
			requests := h.dispatch(scene, content, "", 0)
			if len(requests) != 1 || requests[0].body["content"] == nil {
				t.Fatal("OnMessage did not produce a plain-text result")
			}
		}
	}
	raw := json.RawMessage(`{"id":"image","author":{"user_openid":"user"},"attachments":[{"url":"https://example.invalid/image","content_type":"image/png"}]}`)
	start := len(h.requests)
	h.router.Handle(&types.Task{Ctx: context.Background(), Payload: &types.Payload{T: message.C2C_MESSAGE_CREATE, D: raw}, Client: h.client})
	if len(h.requests) != start+1 || !strings.Contains(string(h.requests[start].body["content"]), "附件 1 个") {
		t.Fatal("empty-text attachment was dropped")
	}
}
func TestMessageFallbackDoesNotEchoBotMessages(t *testing.T) {
	h := newHarness(t, suiteOptions{})
	h.router.Handle(&types.Task{Ctx: context.Background(), Payload: &types.Payload{T: message.MESSAGE_CREATE, D: json.RawMessage(`{"id":"message","channel_id":"channel","content":"bot output","author":{"id":"bot","bot":true}}`)}, Client: h.client})
	if len(h.requests) != 0 {
		t.Fatal("example must not echo bot output into a loop")
	}
}
