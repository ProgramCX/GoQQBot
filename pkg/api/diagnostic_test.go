package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrorRetainsRedactedReasonWithoutChangingDefaultError(t *testing.T) {
	for _, field := range []string{"message", "msg", "err_msg"} {
		t.Run(field, func(t *testing.T) {
			client := NewClient("example-app", "short-key", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/app/getAppAccessToken" {
					return response(`{"access_token":"short-token","expires_in":7200}`), nil
				}
				body, _ := json.Marshal(map[string]any{"code": 40034003, "trace_id": "trace", field: "file size limit exceeded; short-key short-token example-app https://storage.test/a?token=hidden auth_token=abc 0123456789abcdef0123456789abcdef"})
				res := response(string(body))
				res.StatusCode = 400
				return res, nil
			})}))
			_, err := client.Get(context.Background(), "/test")
			var failure *APIError
			if !errors.As(err, &failure) || failure.Code != 40034003 || failure.TraceID != "trace" {
				t.Fatalf("lost error fields: %v", err)
			}
			diagnostic := failure.DiagnosticMessage()
			if !strings.Contains(diagnostic, "file size limit exceeded") {
				t.Fatalf("missing useful diagnostic: %s", diagnostic)
			}
			for _, secret := range []string{"short-key", "short-token", "example-app", "storage.test", "hidden", "abc", "0123456789abcdef"} {
				if strings.Contains(diagnostic, secret) {
					t.Fatalf("diagnostic leaked %q", secret)
				}
			}
			if strings.Contains(failure.Error(), "file size") {
				t.Fatal("default error exposed remote response")
			}
		})
	}
}

func TestAPIDiagnosticIsBoundedAndNilSafe(t *testing.T) {
	if got := redactAPIDiagnostic(strings.Repeat("中文", 600)); len([]rune(got)) != 513 {
		t.Fatal("unbounded diagnostic")
	}
	var failure *APIError
	if failure.DiagnosticMessage() != "" {
		t.Fatal("nil diagnostic")
	}
}
