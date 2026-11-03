package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMediaMenuBuildsAllButtonsInBothScenes(t *testing.T) {
	opts := options{}
	for i, name := range []string{"image", "video", "voice", "file"} {
		opts.fixtures = append(opts.fixtures, fixture{name: name, label: name, kind: message.FileType(i + 1)})
	}
	cases := mediaCases(opts)
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		t.Run(string(scene), func(t *testing.T) {
			body, err := mediaMenu(cases).Build(scene)
			if err != nil {
				t.Fatal(err)
			}
			var keyboard interaction.Keyboard
			if err := json.Unmarshal(body["keyboard"], &keyboard); err != nil {
				t.Fatal(err)
			}
			if keyboard.Content == nil || len(keyboard.Content.Rows) != 5 {
				t.Fatal("14-button menu must fit in five rows")
			}
			seen := map[string]bool{}
			for _, row := range keyboard.Content.Rows {
				if len(row.Buttons) < 1 || len(row.Buttons) > 3 {
					t.Fatal("invalid menu row width")
				}
				for _, b := range row.Buttons {
					if seen[b.ID] || b.Action.Data != "/"+b.ID {
						t.Fatal("duplicate or misrouted menu button")
					}
					seen[b.ID] = true
				}
			}
			if len(seen) != 14 {
				t.Fatalf("menu lost buttons: %d", len(seen))
			}
			for _, item := range cases {
				if !seen[item.command] || item.handle == nil {
					t.Fatalf("missing button or handler: %s", item.command)
				}
			}
		})
	}
}

func TestEveryMediaMenuCaseHasWorkingHandlerInBothScenes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.bin")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	opts := options{cardURL: "https://example.test/page", cardPicture: "https://example.test/image"}
	for i, name := range []string{"image", "video", "voice", "file"} {
		opts.fixtures = append(opts.fixtures, fixture{name: name, label: name, url: "https://example.test/" + name, path: path, kind: message.FileType(i + 1)})
	}
	cases := mediaCases(opts)
	if len(cases) != 14 {
		t.Fatalf("missing menu cases: %d", len(cases))
	}
	seen := map[string]bool{}
	for _, item := range cases {
		if seen[item.command] || item.handle == nil || item.label == "" {
			t.Fatalf("invalid case: %s", item.command)
		}
		seen[item.command] = true
		for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
			t.Run(item.command+"/"+string(scene), func(t *testing.T) {
				var paths []string
				client := api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
					body := `{}`
					if r.URL.Path == "/app/getAppAccessToken" {
						body = `{"access_token":"token","expires_in":7200}`
					} else {
						paths = append(paths, r.Method+" "+r.URL.Path)
						switch {
						case strings.HasSuffix(r.URL.Path, "/upload_prepare"):
							body = `{"upload_id":"task","block_size":"3","parts":[{"index":0,"block_size":"3","presigned_url":"https://storage.test/part"}]}`
						case r.Method == "PUT":
							if _, err := io.Copy(io.Discard, r.Body); err != nil {
								t.Fatal(err)
							}
						case strings.HasSuffix(r.URL.Path, "/files"):
							body = `{"file_info":"opaque","ttl":60}`
						case strings.HasSuffix(r.URL.Path, "/messages"), strings.HasSuffix(r.URL.Path, "/stream_messages"):
							body = `{"id":"sent","timestamp":"2026-11-03T20:00:00+08:00"}`
						case strings.Contains(r.URL.Path, "/messages/"):
							if !strings.HasSuffix(r.URL.Path, "/sent") {
								t.Fatal("recalled user message instead of bot reply")
							}
						}
					}
					return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
				})}))
				event := message.C2C_MESSAGE_CREATE
				data := `{"id":"incoming","author":{"user_openid":"user"},"message_scene":{"ext":["msg_idx=REFIDX_current"]}}`
				if scene == message.SceneGroup {
					event = message.GROUP_AT_MESSAGE_CREATE
					data = `{"id":"incoming","group_openid":"group","message_scene":{"ext":["msg_idx=REFIDX_current"]}}`
				}
				ctx := &types.Context{Context: context.Background(), Client: client, Payload: &types.Payload{T: event, D: json.RawMessage(data)}}
				if err := item.handle(ctx); err != nil {
					t.Fatal(err)
				}
				if len(paths) == 0 {
					t.Fatal("handler did not call an API")
				}
				if strings.HasSuffix(item.command, "-local") && len(paths) != 5 {
					t.Fatalf("local path incomplete: %v", paths)
				}
				if item.command == "media-recall" && len(paths) != 3 {
					t.Fatalf("recall path incomplete: %v", paths)
				}
			})
		}
	}
}
