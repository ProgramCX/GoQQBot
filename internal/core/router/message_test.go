package router

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func task(event message.EventId, body string) *types.Task {
	return &types.Task{Ctx: context.Background(), Payload: &types.Payload{T: event, D: json.RawMessage(body)}}
}

func TestMessageRoutingCommandPriorityScopesAndAttachments(t *testing.T) {
	var got []string
	h := func(name string) Handler {
		return func(c *types.Context) error { got = append(got, name+":"+c.Text()); return nil }
	}
	r := New(h("legacy"))
	r.OnAnyMessage(h("all"))
	r.OnMessage(message.GROUP_AT_MESSAGE_CREATE, h("group"))
	r.OnMessage(message.GROUP_MESSAGE_CREATE, h("group"))
	r.OnCommand("hello", message.GROUP_AT_MESSAGE_CREATE, h("command"))
	r.Handle(task(message.GROUP_AT_MESSAGE_CREATE, `{"content":" /hello name "}`))
	r.Handle(task(message.GROUP_AT_MESSAGE_CREATE, `{"content":"  ordinary  "}`))
	r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"private"}`))
	r.Handle(task(message.GROUP_MESSAGE_CREATE, `{"attachments":[{"url":"https://example.invalid/image"}]}`))
	r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"/unknown arg"}`))
	want := []string{"command:/hello name", "group:ordinary", "all:private", "group:", "all:/unknown arg"}
	if len(got) != len(want) {
		t.Fatalf("routes %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("route %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestLegacyDefaultsAndEventsRemainAvailable(t *testing.T) {
	defaults, events, messages := 0, 0, 0
	r := New(func(*types.Context) error { defaults++; return nil })
	r.OnEvent(message.FRIEND_ADD, func(*types.Context) error { events++; return nil })
	r.OnMessage(message.GROUP_AT_MESSAGE_CREATE, func(*types.Context) error { messages++; return nil })
	r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"hello"}`))
	r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"/unknown"}`))
	r.Handle(task(message.FRIEND_ADD, `{"openid":"user"}`))
	if defaults != 1 || events != 1 || messages != 0 {
		t.Fatal("legacy default or event behavior changed")
	}
	r.OnAnyMessage(func(*types.Context) error { messages++; return nil })
	r.Handle(task(message.FRIEND_ADD, `{"openid":"user"}`))
	if events != 2 || messages != 0 {
		t.Fatal("lifecycle event reached global OnMessage")
	}
	r.OnAnyMessage(nil)
	r.Handle(task(message.AT_MESSAGE_CREATE, `{"content":"channel"}`))
	if messages != 1 {
		t.Fatal("nil registration removed message handler")
	}
}

func TestMessageHandlerDoesNotInterceptInteractionsAndSupportsCustomEvents(t *testing.T) {
	r := New(nil)
	messages, buttons := 0, 0
	r.OnMessage(message.EventId("FUTURE_MESSAGE"), func(c *types.Context) error {
		messages++
		if c.Command != "missing" || len(c.Args) != 1 {
			t.Error("unknown command metadata missing")
		}
		return nil
	})
	r.OnAnyMessage(func(*types.Context) error { t.Error("global message handler intercepted an interaction"); return nil })
	r.OnButton("refresh", func(*types.Context) error { buttons++; return nil })
	r.Handle(task(message.INTERACTION_CREATE, `{"id":"interaction","type":11,"content":"hello","data":{"resolved":{"button_id":"refresh"}}}`))
	r.Handle(task("FUTURE_MESSAGE", `{"content":"/missing arg"}`))
	if buttons != 1 || messages != 1 {
		t.Fatal("wrong event classification")
	}
}

func TestMessageRegistrationDoesNotHoldLockDuringCallbacks(t *testing.T) {
	r := New(nil)
	r.OnAnyMessage(func(*types.Context) error { r.OnAnyMessage(func(*types.Context) error { return nil }); return nil })
	// Deadlocks if routing retains the registry lock while invoking application code.
	// 若路由调用业务代码时仍持有注册表锁，此测试将发生死锁。
	r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"hello"}`))
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.OnAnyMessage(func(*types.Context) error { return nil })
			r.Handle(task(message.C2C_MESSAGE_CREATE, `{"content":"hello"}`))
		}()
	}
	wg.Wait()
}
