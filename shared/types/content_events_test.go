package types

import (
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func TestForumEventKeepsExactIDsAndRawRichText(t *testing.T) {
	ctx := &Context{Payload: &Payload{
		T: message.FORUM_THREAD_CREATE,
		D: []byte(`{"guild_id":47129941624960822,"channel_id":1661124,"author_id":144115218182563108,"thread_info":{"thread_id":"thread","title":[{"type":1,"text_info":{"text":"主题"}}],"content":"内容"}}`),
	}}
	event, err := ctx.ForumEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.GuildID != "47129941624960822" || event.ChannelID != "1661124" || event.AuthorID != "144115218182563108" {
		t.Fatalf("forum identifiers: %q %q %q", event.GuildID, event.ChannelID, event.AuthorID)
	}
	if event.Thread == nil || event.Thread.ThreadID != "thread" || string(event.Thread.Title) != `[{"type":1,"text_info":{"text":"主题"}}]` || string(event.Thread.Content) != `"内容"` {
		t.Fatalf("forum thread: %#v", event.Thread)
	}
	ctx.Payload.T = message.OPEN_FORUM_POST_CREATE
	ctx.Payload.D = []byte(`{"guild_id":"guild","channel_id":"channel","author_id":"author"}`)
	event, err = ctx.ForumEvent()
	if err != nil || event == nil || event.AuthorID != "author" || event.Thread != nil || event.Post != nil {
		t.Fatalf("open forum event: %#v, %v", event, err)
	}
	if _, err := ctx.AudioAction(); err == nil {
		t.Fatal("audio decoder accepted open forum event")
	}
}
