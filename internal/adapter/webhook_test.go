package adapter

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func TestWebhookACKAndOverload(t *testing.T) {
	secret := "example-secret"
	validator := NewValidator(secret)
	for _, overloaded := range []bool{false, true} {
		called := false
		web := New(*config.DefaultConfig(), *validator, func(p *types.Payload, done func()) error {
			called = true
			if p.ID != "event" || p.T != "FRIEND_ADD" {
				t.Error("envelope not preserved")
			}
			if overloaded {
				return errors.New("full")
			}
			done()
			return nil
		})
		body := `{"id":"event","op":0,"t":"FRIEND_ADD","d":{}}`
		request := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
		key := ed25519.NewKeyFromSeed([]byte(validator.generateSeed()))
		request.Header.Set("X-Signature-Timestamp", "123")
		request.Header.Set("X-Signature-Ed25519", hex.EncodeToString(ed25519.Sign(key, []byte("123"+body))))
		recorder := httptest.NewRecorder()
		web.handleFunc(recorder, request)
		expected := 200
		if overloaded {
			expected = 503
		}
		if recorder.Code != expected || !called {
			t.Fatalf("status %d, called %v", recorder.Code, called)
		}
		if !overloaded && strings.TrimSpace(recorder.Body.String()) != `{"op":12}` {
			t.Fatal("incorrect protocol ACK")
		}
	}
}
func TestWebhookRejectsUnsignedEvent(t *testing.T) {
	called := false
	web := New(*config.DefaultConfig(), *NewValidator("example-secret"), func(*types.Payload, func()) error { called = true; return nil })
	recorder := httptest.NewRecorder()
	web.handleFunc(recorder, httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"op":0,"d":{}}`)))
	if recorder.Code != 401 || called {
		t.Fatal("unsigned event accepted")
	}
}
