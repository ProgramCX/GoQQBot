package group_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func managementClient(t *testing.T, handle func(*http.Request) (int, string)) *group.Client {
	t.Helper()
	client := api.NewClient("test", "secret", api.WithHTTPClient(&http.Client{
		Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			status, body := http.StatusOK, `{"access_token":"test-token","expires_in":7200}`
			if r.URL.Path != "/app/getAppAccessToken" {
				if got := r.Header.Get("Authorization"); got != "QQBot test-token" {
					t.Errorf("unexpected authorization: %q", got)
				}
				status, body = handle(r)
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}))
	return group.NewClient(client)
}

func TestListMembersEscapesCursorAndGroup(t *testing.T) {
	const cursor = "opaque+=&/中文"
	client := managementClient(t, func(r *http.Request) (int, string) {
		if got := r.URL.EscapedPath(); got != "/v2/groups/group%2F%3Fx/members" || r.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", r.Method, got)
		}
		if query := r.URL.Query(); len(query) != 1 || query.Get("cursor") != cursor {
			t.Fatalf("unexpected pagination: %v", query)
		}
		return http.StatusOK, `{"members":[{"member_openid":"member","joined_at":"2026-08-05T11:23:05+08:00"}],"next_cursor":"next"}`
	})
	page, err := client.ListMembers(context.Background(), "group/?x", cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Members) != 1 || page.Members[0].MemberOpenID != "member" || page.NextCursor != "next" {
		t.Fatalf("unexpected members page: %+v", page)
	}
}

func TestRemoveMembersPreservesPartialBlacklistFailure(t *testing.T) {
	client := managementClient(t, func(r *http.Request) (int, string) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/groups/group/batch_remove_members" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			IDs []string `json:"member_openids"`
			Ban bool     `json:"add_to_member_blacklist"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !body.Ban || len(body.IDs) != 1 || body.IDs[0] != "member" {
			t.Fatalf("unexpected request body: %+v, %v", body, err)
		}
		return http.StatusOK, `{"remove_members_result":"success","add_to_member_blacklist_fail_openids":["member"]}`
	})
	result, err := client.RemoveMembers(context.Background(), "group", group.RemoveMembersRequest{
		MemberOpenIDs: []string{"member"}, AddToMemberBlacklist: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Result != "success" || len(result.BlacklistFailIDs) != 1 || result.BlacklistFailIDs[0] != "member" {
		t.Fatalf("unexpected removal result: %+v", result)
	}
}

func TestListStrategiesPreservesMaskedAndNumericGroupNumbers(t *testing.T) {
	client := managementClient(t, func(r *http.Request) (int, string) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/groups/join_approval_strategy" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var query struct {
			Cursor string `json:"cursor"`
			Limit  int    `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil || query.Cursor != "next" || query.Limit != 50 {
			t.Fatalf("unexpected pagination: %+v, %v", query, err)
		}
		return http.StatusOK, `{"strategies":[{"strategy_id":"strategy","group_ids":["10****499",9007199254740993]}]}`
	})
	page, err := client.ListStrategies(context.Background(), group.PageOptions{Cursor: "next", Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Strategies) != 1 {
		t.Fatalf("unexpected strategies: %+v", page)
	}
	ids := page.Strategies[0].GroupIDs
	if len(ids) != 2 || ids[0] != "10****499" || ids[1] != "9007199254740993" {
		t.Fatalf("unexpected group numbers: %v", ids)
	}
}

func TestDecideJoinPreservesPlatformError(t *testing.T) {
	client := managementClient(t, func(r *http.Request) (int, string) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/groups/group/approval_join_request/member" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		return http.StatusForbidden, `{"code":11253,"trace_id":"trace"}`
	})
	err := client.DecideJoin(context.Background(), "group", "member", group.DecideJoinRequest{
		Decision: group.JoinDecline, JoinRequestID: "request",
	})
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 11253 || apiErr.TraceID != "trace" {
		t.Fatalf("unexpected error: %v", err)
	}
}
