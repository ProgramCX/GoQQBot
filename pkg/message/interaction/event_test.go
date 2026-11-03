package interaction_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

func TestDecodeDocumentedInteractionVariants(t *testing.T) {
	for _, kind := range []interaction.InteractionType{11, 12, 13, 14, 15, 16, 18, 19, 20, 99} {
		data, err := json.Marshal(map[string]any{"id": "interaction-id", "type": kind, "application_id": "app", "future_field": true, "data": map[string]any{"resolved": map[string]any{
			"button_data": "refresh", "button_id": "refresh", "feature_id": "menu", "user_id": "user", "message_id": "message", "feedback_opt": "LIKE", "checked": 1, "action": "ENTER_STORY",
			"message_scene":  map[string]any{"ext": []string{"disable_net_search=1"}},
			"authorize_data": map[string]string{"opt_scene": "setting", "scope": "group_push"},
		}}})
		if err != nil {
			t.Fatal(err)
		}
		event, err := interaction.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if event.Kind() != kind || event.ApplicationID != "app" || event.NeedsAcknowledgement() != (kind == 11 || kind == 12) {
			t.Fatal("wrong interaction kind or ACK requirement")
		}
		r := event.Data.Resolved
		if r.FeedbackOpt != "LIKE" || r.Checked != 1 || r.Action != "ENTER_STORY" || r.AuthorizeData.Scope != "group_push" || r.MessageScene.Ext[0] != "disable_net_search=1" {
			t.Fatal("documented fields lost")
		}
		data[0] = '!'
		if !json.Valid(event.Raw) {
			t.Fatal("raw data borrowed caller memory")
		}
		var raw map[string]any
		if err := json.Unmarshal(event.Raw, &raw); err != nil || raw["future_field"] != true {
			t.Fatal("future field lost")
		}
	}
	legacy, err := interaction.Decode([]byte(`{"id":"legacy","data":{"type":11}}`))
	if err != nil || legacy.Kind() != 11 || !legacy.NeedsAcknowledgement() {
		t.Fatal("legacy data.type fallback missing")
	}
	for _, bad := range []string{`null`, `{}`, `not-json`} {
		if _, err := interaction.Decode([]byte(bad)); !errors.Is(err, interaction.ErrInvalidEvent) {
			t.Fatalf("bad event accepted: %s", bad)
		}
	}
}
