package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
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

func TestMaterialMenuButtonsAndLifecycleInBothScenes(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+a0ZkAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(data)
	etag := hex.EncodeToString(sum[:])
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		t.Run(string(scene), func(t *testing.T) {
			demo := &materialDemo{file: path}
			cases := demo.cases()
			body, err := materialMenu(cases).Build(scene)
			if err != nil {
				t.Fatal(err)
			}
			var keyboard interaction.Keyboard
			if err := json.Unmarshal(body["keyboard"], &keyboard); err != nil {
				t.Fatal(err)
			}
			if keyboard.Content == nil || len(keyboard.Content.Rows) != 3 {
				t.Fatal("menu does not fit layout")
			}
			var actions []string
			for _, row := range keyboard.Content.Rows {
				for _, b := range row.Buttons {
					actions = append(actions, strings.TrimPrefix(b.Action.Data, "/"))
				}
			}
			if len(actions) != 5 {
				t.Fatal("missing material buttons")
			}
			prepares, completes, deletes, lists, sends := 0, 0, 0, 0, 0
			client := api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				text := `{}`
				switch {
				case r.URL.Path == "/app/getAppAccessToken":
					text = `{"access_token":"token","expires_in":7200}`
				case r.URL.Path == "/v2/media-images/prepare":
					prepares++
					text = `{"upload_id":"task","presigned_url":"https://storage.test/image","expires_in":3600}`
				case r.URL.Host == "storage.test":
					if r.Header.Get("Content-Type") != "image/png" || r.Header.Get("Authorization") != "" {
						t.Fatal("incorrect material storage PUT")
					}
					_, _ = io.Copy(io.Discard, r.Body)
					return &http.Response{StatusCode: 200, Header: http.Header{"Etag": []string{`"` + etag + `"`}}, Body: io.NopCloser(strings.NewReader(""))}, nil
				case r.URL.Path == "/v2/media-images/complete":
					completes++
					text = `{"image_id":"created","url":"qqbot-media://image/v1/reference","ttl":7776000}`
				case r.URL.Path == "/v2/media-images":
					lists++
					text = `{"images":[{"image_id":"created","url":"qqbot-media://image/v1/reference","file_name":"image.png","expires_at":"9999999999"}],"total":"1","next_cursor":""}`
				case r.Method == "DELETE":
					deletes++
					if r.URL.Path != "/v2/media-images/created/delete" {
						t.Fatal("deleted an unrelated image")
					}
				case strings.HasSuffix(r.URL.Path, "/messages"):
					sends++
					var body struct {
						Markdown struct {
							Content string `json:"content"`
						} `json:"markdown"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if strings.Contains(body.Markdown.Content, "qqbot-media://") && !strings.Contains(body.Markdown.Content, "#1px #1px") {
						t.Fatal("upload/reuse reply lost detected PNG dimensions")
					}
					text = `{"id":"sent","timestamp":"2026-11-03T23:00:00+08:00"}`
				default:
					t.Fatalf("unexpected endpoint: %s", r.URL.Path)
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(text))}, nil
			})}))
			event := message.C2C_MESSAGE_CREATE
			payload := `{"id":"incoming","author":{"user_openid":"user"}}`
			if scene == message.SceneGroup {
				event = message.GROUP_AT_MESSAGE_CREATE
				payload = `{"id":"incoming","group_openid":"group"}`
			}
			ctx := &types.Context{Context: context.Background(), Client: client, Payload: &types.Payload{T: event, D: json.RawMessage(payload)}}
			if err := ctx.Reply(materialMenu(cases)); err != nil {
				t.Fatal(err)
			}
			for _, command := range actions {
				found := false
				for _, item := range cases {
					if item.command == command {
						found = true
						if err := item.handler(ctx); err != nil {
							t.Fatal(err)
						}
					}
				}
				if !found {
					t.Fatalf("button has no handler: %s", command)
				}
			}
			if prepares != 1 || completes != 1 || lists != 1 || deletes != 1 || sends != 6 {
				t.Fatalf("lifecycle mismatch: prepare=%d complete=%d list=%d delete=%d send=%d", prepares, completes, lists, deletes, sends)
			}
			if demo.current() != nil {
				t.Fatal("deleted material remains active")
			}
			if err := demo.reuse(ctx); err != nil || prepares != 1 {
				t.Fatal("reuse after delete silently uploaded an image")
			}
		})
	}
}
