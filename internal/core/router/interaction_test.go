package router

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestInteractionDispatchPriorityAndNoCommandParsing(t *testing.T) {
	for _, tc := range []struct {
		kind                  int
		button, feature, want string
	}{
		{11, "refresh", "same", "button"}, {11, "unknown", "same", "kind"}, {12, "refresh", "same", "menu"}, {13, "refresh", "same", "fallback"}, {99, "refresh", "same", "fallback"},
	} {
		var calls []string
		handler := func(name string) Handler {
			return func(c *types.Context) error {
				if c.Command != "" {
					t.Error("interaction was parsed as a command")
				}
				if c.RequestContext().Err() != nil {
					t.Error("handler context missing")
				}
				calls = append(calls, name)
				return nil
			}
		}
		router := New(handler("default"))
		router.OnCommand("menu", message.INTERACTION_CREATE, handler("command"))
		router.OnEvent(message.INTERACTION_CREATE, handler("fallback"))
		router.OnInteraction(interaction.InteractionTypeButton, handler("kind"))
		router.OnButton("refresh", handler("button"))
		router.OnMenu("same", handler("menu"))
		data, err := json.Marshal(map[string]any{"id": "interaction", "type": tc.kind, "content": "/menu", "data": map[string]any{"resolved": map[string]string{"button_id": tc.button, "feature_id": tc.feature}}})
		if err != nil {
			t.Fatal(err)
		}
		router.Handle(&types.Task{Ctx: context.Background(), Payload: &types.Payload{T: message.INTERACTION_CREATE, D: data}})
		if len(calls) != 1 || calls[0] != tc.want {
			t.Fatalf("kind %d: calls %v, want %s", tc.kind, calls, tc.want)
		}
	}
}

func TestCommandButtonStillUsesMessageRouter(t *testing.T) {
	calls := 0
	router := New(nil)
	router.OnCommand("help", message.GROUP_AT_MESSAGE_CREATE, func(c *types.Context) error {
		calls++
		if c.Command != "help" {
			t.Error("command missing")
		}
		return nil
	})
	router.Handle(&types.Task{Ctx: context.Background(), Payload: &types.Payload{T: message.GROUP_AT_MESSAGE_CREATE, D: json.RawMessage(`{"content":"/help"}`)}})
	if calls != 1 {
		t.Fatal("command button message did not reach ordinary command handler")
	}
}
