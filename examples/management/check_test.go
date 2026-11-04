package main

import (
	"context"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

func TestCheckSeparatesWriteMissingFixtureAndPermission(t *testing.T) {
	op := operation{write: true, run: func(context.Context, []string, []byte) (any, error) {
		t.Fatal("batch check attempted a write")
		return nil, nil
	}}
	if got := checkOperation(t.Context(), "channel-delete", op, checkInputs{}); got.Status != "manual" {
		t.Fatalf("write result = %+v", got)
	}
	op.write = false
	op.ids = "channel-id"
	if got := checkOperation(t.Context(), "channel-online", op, checkInputs{ChannelID: "text"}); got.Status != "missing_fixture" {
		t.Fatalf("missing audio fixture result = %+v", got)
	}
	op.run = func(_ context.Context, ids []string, _ []byte) (any, error) {
		if len(ids) != 1 || ids[0] != "audio" {
			t.Fatalf("unexpected channel IDs: %v", ids)
		}
		return nil, &api.APIError{StatusCode: 400, Code: 11253, TraceID: "trace"}
	}
	if got := checkOperation(t.Context(), "channel-online", op, checkInputs{AudioChannelID: "audio"}); got.Status != "permission_blocked" || got.Code != 11253 || got.TraceID != "trace" {
		t.Fatalf("permission result = %+v", got)
	}
}
