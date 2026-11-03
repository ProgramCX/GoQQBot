package text_test

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message/text"
)

func TestCommandTagsEncodeExactlyOnceAndFallback(t *testing.T) {
	command := "/bind 名字 &value=10%+ "
	input, err := text.Input(command).Label("绑定 账号").Reference().Render(text.Options{Scene: "group", Markdown: true})
	if err != nil {
		t.Fatal(err)
	}
	want := `<qqbot-cmd-input text="` + strings.ReplaceAll(url.QueryEscape(command), "+", "%20") + `" show="` + strings.ReplaceAll(url.QueryEscape("绑定 账号"), "+", "%20") + `" reference="true" />`
	if input != want {
		t.Fatalf("tag %s want %s", input, want)
	}
	enter := text.Enter(command)
	c2c, err := enter.Render(text.Options{Scene: "c2c", Markdown: true})
	if err != nil {
		t.Fatal(err)
	}
	if c2c != `<qqbot-cmd-enter text="`+strings.ReplaceAll(url.QueryEscape(command), "+", "%20")+`" />` {
		t.Fatal("wrong Enter tag")
	}
	for _, scene := range []string{"group", "channel", "channel_dm"} {
		fallback, err := enter.Render(text.Options{Scene: scene, Markdown: true})
		if err != nil || !strings.HasPrefix(fallback, "<qqbot-cmd-input ") {
			t.Fatal("missing command fallback")
		}
		if _, err := enter.Render(text.Options{Scene: scene, Markdown: true, Strict: true}); !errors.Is(err, text.ErrUnsupportedFragment) {
			t.Fatal("strict Enter accepted")
		}
	}
	if _, err := text.Input(strings.Repeat("中", 101)).Render(text.Options{Scene: "c2c", Markdown: true}); !errors.Is(err, text.ErrInvalidFragment) {
		t.Fatal("oversized command accepted")
	}
}

func TestMentionsChannelAndEmojiScopes(t *testing.T) {
	for _, tc := range []struct {
		f        text.Fragment
		scene    string
		markdown bool
		want     string
	}{
		{text.AtUser("member"), "group", true, `<qqbot-at-user id="member" />`},
		{text.AtUser(`x"/><tag>`), "channel", false, `<qqbot-at-user id="x&#34;/&gt;&lt;tag&gt;" />`},
		{text.AtEveryone(), "group", false, `<qqbot-at-everyone />`},
		{text.Channel("123"), "channel", false, `<#123>`},
		{text.Emoji("307"), "channel", true, `<emoji:307>`},
	} {
		value, err := tc.f.Render(text.Options{Scene: tc.scene, Markdown: tc.markdown})
		if err != nil || value != tc.want {
			t.Fatalf("render %q want %q error %v", value, tc.want, err)
		}
	}
	for _, tc := range []struct {
		f     text.Fragment
		scene string
		md    bool
	}{
		{text.AtUser("member"), "group", false}, {text.AtEveryone(), "c2c", true}, {text.Channel("1"), "group", true}, {text.Emoji("1"), "c2c", true}, {text.Input("/help"), "group", false},
	} {
		if _, err := tc.f.Render(text.Options{Scene: tc.scene, Markdown: tc.md}); !errors.Is(err, text.ErrUnsupportedFragment) {
			t.Fatal("unsupported tag accepted")
		}
	}
}
