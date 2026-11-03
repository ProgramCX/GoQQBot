package message_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func TestTypedIncomingNestedContent(t *testing.T) {
	for _, event := range []message.EventId{message.C2C_MESSAGE_CREATE, message.GROUP_MESSAGE_CREATE, message.GROUP_AT_MESSAGE_CREATE} {
		t.Run(string(event), func(t *testing.T) {
			input := []byte(`{"id":"incoming","message_type":103,"message_scene":{"source":"default","ext":["broken","msg_idx=REFIDX_in","ref_msg_idx=REFIDX_old","auth_token=a=b"]},"ark_data":{"ark_type":"future","fields":{"title":"hello","tags":["one","two"],"nested":{"enabled":true}}},"msg_elements":[{"msg_idx":"quote","message_type":102,"msg_elements":[{"message_type":3,"ark_data":{"ark_type":"miniapp","fields":{"title":"nested"}},"attachments":[{"filename":"voice","size":123,"voice_wav_url":"https://example.test/a.wav","asr_refer_text":"words"}],"future":true}]}]}`)
			m, err := message.DecodeReceived(event, input)
			if err != nil {
				t.Fatal(err)
			}
			var tags []string
			if ok, err := m.ARK.Field("tags", &tags); err != nil || !ok || len(tags) != 2 {
				t.Fatalf("tags: %v %v", tags, err)
			}
			if _, err := m.ARK.Field("tags", new(string)); err == nil {
				t.Fatal("wrong field type accepted")
			}
			if ok, err := m.ARK.Field("missing", new(string)); ok || err != nil {
				t.Fatal("missing field behavior")
			}
			if m.MessageScene.MessageIndex() != "REFIDX_in" || m.MessageScene.ReferencedIndex() != "REFIDX_old" {
				t.Fatal("reference indices lost")
			}
			if value, ok := m.MessageScene.Lookup("auth_token"); !ok || value != "a=b" {
				t.Fatal("token split incorrectly")
			}
			nested := m.MsgElements[0].MsgElements[0]
			if nested.ARK.Type != "miniapp" || nested.Attachments[0].ASRReferText != "words" || !strings.Contains(string(nested.Raw), `"future":true`) {
				t.Fatal("recursive content lost")
			}
			m.ARK.Fields["title"][0] = '!'
			m.MsgElements[0].Raw[0] = '!'
			if !json.Valid(m.ArkData) || !json.Valid(m.Elements[0]) {
				t.Fatal("typed and compatibility views share storage")
			}
			input[0] = '!'
			if !json.Valid(m.Raw) {
				t.Fatal("raw borrowed input")
			}
		})
	}
	for _, raw := range []string{`{"ark_data":[]}`, `{"msg_elements":[{"msg_elements":"invalid"}]}`} {
		if _, err := message.DecodeReceived(message.C2C_MESSAGE_CREATE, []byte(raw)); !errors.Is(err, message.ErrInvalidReceived) {
			t.Fatalf("invalid typed input: %v", err)
		}
	}
}

func TestCardTypingAndBodyConstraints(t *testing.T) {
	card := message.ImageTextCard{Title: "标题", Description: "说明", URL: "https://example.test/page", PictureURL: "https://example.test/image"}
	for _, scene := range []message.Scene{message.SceneGroup, message.SceneC2C} {
		body, err := message.Card(card).Build(scene)
		if err != nil || string(body["msg_type"]) != "8" {
			t.Fatalf("card: %v", err)
		}
	}
	if _, err := message.Typing(60).Build(message.SceneC2C); err != nil {
		t.Fatal(err)
	}
	bad := card
	bad.URL = "http://example.test"
	long := card
	long.Title = strings.Repeat("字", 51)
	for _, tc := range []struct {
		out   message.Outgoing
		scene message.Scene
	}{
		{message.Card(bad), message.SceneC2C}, {message.Card(long), message.SceneGroup},
		{message.Card(card), message.SceneChannel}, {message.Typing(10), message.SceneGroup},
		{message.Typing(61), message.SceneC2C}, {message.Typing(0), message.SceneC2C},
		{message.Media("info").WithField("card", card), message.SceneGroup},
		{message.Text("hello").WithField("media", message.MediaInfo{FileInfo: "info"}), message.SceneC2C},
	} {
		if _, err := tc.out.Build(tc.scene); !errors.Is(err, message.ErrInvalidMessage) {
			t.Fatalf("invalid body: %v", err)
		}
	}
	if _, err := message.New(message.CARD).WithField("card", message.MessageCard{Type: "future", Content: map[string]bool{"enabled": true}}).Build(message.SceneGroup); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyReferenceWireObjectPreservesWakeup(t *testing.T) {
	for _, value := range []any{
		message.GroupSendMessage{CommonSendMessage: message.CommonSendMessage{MessageReference: "REFIDX_old"}},
		message.PrivateSendMessage{CommonSendMessage: message.CommonSendMessage{Reference: &message.MessageReference{MessageID: "REFIDX_old"}}, IsWakeup: true},
	} {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var body struct {
			Reference message.MessageReference `json:"message_reference"`
			Wakeup    bool                     `json:"is_wakeup"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		if body.Reference.MessageID != "REFIDX_old" {
			t.Fatalf("legacy reference not converted: %s", raw)
		}
		if _, ok := value.(message.PrivateSendMessage); ok && !body.Wakeup {
			t.Fatal("reference marshaler lost wakeup")
		}
	}
}
