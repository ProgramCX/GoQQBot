package robot_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPanelAndMenuWire(t *testing.T) {
	client := robot.NewClient(api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		var response string
		switch {
		case r.URL.Path == "/app/getAppAccessToken":
			response = `{"access_token":"token","expires_in":7200}`
		case r.URL.Path == "/v2/panels" && r.Method == http.MethodGet:
			if r.URL.Query().Get("cursor") != "a+&=" || r.URL.Query().Get("scope") != "group" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			response = `{"records":[{"panel_id":"one","scope":"group"}],"next_cursor":"next"}`
		case r.URL.EscapedPath() == "/v2/panels/a%2Fb/target" && r.Method == http.MethodPut:
			var input robot.PanelTargetChange
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Action != robot.TargetAdd || len(input.GroupOpenIDs) != 1 {
				t.Errorf("target = %#v, %v", input, err)
			}
			response = `{}`
		case r.URL.Path == "/v2/menu" && r.Method == http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if string(body) != `{"menu":{"items":[]}}` {
				t.Errorf("menu = %s", body)
			}
			response = `{"version":7}`
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
			response = `{}`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response))}, nil
	})})))
	ctx := t.Context()
	page, err := client.ListPanels(ctx, robot.ScopeGroup, "a+&=", 20)
	if err != nil || len(page.Records) != 1 || page.Records[0].ID != "one" || page.NextCursor != "next" {
		t.Fatalf("page = %#v, %v", page, err)
	}
	if err := client.UpdatePanelTarget(ctx, "a/b", robot.PanelTargetChange{Action: robot.TargetAdd, GroupOpenIDs: []string{"group"}}); err != nil {
		t.Fatal(err)
	}
	state, err := client.UpdateMenu(ctx, robot.Menu{})
	if err != nil || state.Version != 7 {
		t.Fatalf("menu state = %#v, %v", state, err)
	}
}
