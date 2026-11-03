package types_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestMessageGetterPreservesOriginalTextAndOwnsAttachments(t *testing.T) {
	event := &types.Context{RawMessage: "hello", Payload: &types.Payload{T: message.GROUP_AT_MESSAGE_CREATE, D: json.RawMessage(`{
 "id":"message","content":"  hello  ","group_openid":"group","message_type":103,
 "author":{"id":"id","member_openid":"member","user_openid":"different-user","member_role":"admin"},
 "attachments":[{"content_type":"voice","url":"https://example.invalid/silk","voice_wav_url":"https://example.invalid/wav","asr_refer_text":"hello"}],
 "message_scene":{"source":"default","ext":["msg_idx=ref"]},"msg_elements":[{"content":"quoted"}],"future_field":true}`)}}
	msg, err := event.Message()
	if err != nil {
		t.Fatal(err)
	}
	if event.Text() != "hello" || msg.Content != "  hello  " {
		t.Fatal("normalized and original text were conflated")
	}
	if msg.SenderID() != "member" || msg.ConversationID() != "group" || msg.Scene() != message.SceneGroup {
		t.Fatal("group and private IDs conflated")
	}
	if len(msg.Attachments) != 1 || msg.Attachments[0].ASRReferText != "hello" || msg.Attachments[0].VoiceWAVURL == "" || msg.MessageType != 103 {
		t.Fatal("message metadata missing")
	}
	msg.Attachments[0].URL = "changed"
	msg.MessageScene.Ext[0] = "changed"
	msg.Raw[0] = '!'
	again, err := event.Message()
	if err != nil {
		t.Fatal(err)
	}
	if again.Attachments[0].URL == "changed" || again.MessageScene.Ext[0] == "changed" || !json.Valid(again.Raw) {
		t.Fatal("message snapshot borrowed mutable memory")
	}
	event.Payload.D[0] = '!'
	if !json.Valid(again.Raw) {
		t.Fatal("raw message borrowed payload memory")
	}
}

func TestMessageScenesAndInvalidBody(t *testing.T) {
	for _, tc := range []struct {
		event                      message.EventId
		body, sender, conversation string
		scene                      message.Scene
	}{
		{message.C2C_MESSAGE_CREATE, `{"author":{"id":"fallback","user_openid":"user","member_openid":"member"}}`, "user", "user", message.SceneC2C},
		{message.AT_MESSAGE_CREATE, `{"channel_id":"channel","author":{"id":"author"}}`, "author", "channel", message.SceneChannel},
		{message.DIRECT_MESSAGE_CREATE, `{"guild_id":"dm","author":{"id":"author"}}`, "author", "dm", message.SceneDirectMessage},
	} {
		event := &types.Context{Payload: &types.Payload{T: tc.event, D: json.RawMessage(tc.body)}}
		msg, err := event.Message()
		if err != nil {
			t.Fatal(err)
		}
		if msg.Scene() != tc.scene || msg.SenderID() != tc.sender || msg.ConversationID() != tc.conversation {
			t.Fatal("wrong message scene metadata")
		}
	}
	for _, data := range []string{"null", "[]", "invalid"} {
		if _, err := message.DecodeReceived(message.C2C_MESSAGE_CREATE, []byte(data)); !errors.Is(err, message.ErrInvalidReceived) {
			t.Fatal("invalid message accepted")
		}
	}
	var absent *types.Context
	if absent.Text() != "" {
		t.Fatal("nil text")
	}
	if _, err := absent.Message(); !errors.Is(err, message.ErrInvalidReceived) {
		t.Fatal(err)
	}
	if _, err := (&types.Context{Payload: &types.Payload{T: message.INTERACTION_CREATE, D: json.RawMessage(`{"id":"interaction"}`)}}).Message(); !errors.Is(err, message.ErrInvalidReceived) {
		t.Fatal("interaction misclassified as message")
	}
}
