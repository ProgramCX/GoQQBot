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

type contentTransport func(*http.Request) (*http.Response, error)

func (f contentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func contentResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestContentWireContracts(t *testing.T) {
	transport := contentTransport(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.URL.Path == "/app/getAppAccessToken":
			return contentResponse(http.StatusOK, `{"access_token":"test-token","expires_in":7200}`), nil
		case r.Method == http.MethodPost && r.URL.EscapedPath() == "/channels/schedule%2Fid/schedules":
			var input struct {
				Schedule struct {
					Start string `json:"start_timestamp"`
					End   string `json:"end_timestamp"`
				} `json:"schedule"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Schedule.Start != "1642076400000" || input.Schedule.End != "1642083600000" {
				t.Errorf("schedule request: %#v, %v", input, err)
			}
			return contentResponse(http.StatusOK, `{"id":"item","start_timestamp":"1642076400000","end_timestamp":"1642083600000","remind_type":"0"}`), nil
		case r.Method == http.MethodGet && r.URL.Path == "/channels/123/schedules":
			if got := r.URL.Query().Get("since"); got != "1642076400000" {
				t.Errorf("since query = %q", got)
			}
			return contentResponse(http.StatusOK, `[]`), nil
		case r.Method == http.MethodGet && r.URL.Path == "/channels/123/threads/thread":
			return contentResponse(http.StatusOK, `{"thread":{"author_id":"user","thread_info":{"thread_id":"thread","content":"{\"paragraphs\":[]}"}}}`), nil
		case r.Method == http.MethodPost && r.URL.Path == "/channels/123/audio":
			var input struct {
				Status *int   `json:"status"`
				URL    string `json:"audio_url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Status == nil || *input.Status != 0 || input.URL != "https://example.test/audio.mp3" {
				t.Errorf("audio request: %#v, %v", input, err)
			}
			return contentResponse(http.StatusOK, `{}`), nil
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/123/pins/all":
			return contentResponse(http.StatusNoContent, ""), nil
		default:
			t.Errorf("unexpected content request %s %s", r.Method, r.URL.EscapedPath())
			return contentResponse(http.StatusNotFound, `{}`), nil
		}
	})
	client := channel.NewClient(api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: transport})))
	ctx := t.Context()
	schedule, err := client.CreateSchedule(ctx, "schedule/id", channel.ScheduleInput{
		Name: "活动", StartTimestamp: 1642076400000, EndTimestamp: 1642083600000,
	})
	if err != nil || schedule.ID != "item" || schedule.StartTimestamp != 1642076400000 || schedule.Reminder != channel.RemindNever {
		t.Fatalf("schedule = %#v, %v", schedule, err)
	}
	if _, err := client.ListSchedules(ctx, "123", 1642076400000); err != nil {
		t.Fatal(err)
	}
	thread, err := client.GetThread(ctx, "123", "thread")
	if err != nil || thread.AuthorID != "user" || thread.Info.ID != "thread" || thread.Info.Content != `{"paragraphs":[]}` {
		t.Fatalf("thread = %#v, %v", thread, err)
	}
	if err := client.ControlAudio(ctx, "123", channel.AudioControl{Status: channel.AudioStart, AudioURL: "https://example.test/audio.mp3"}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeletePin(ctx, "123", "all"); err != nil {
		t.Fatal(err)
	}
}
