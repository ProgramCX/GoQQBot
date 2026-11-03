package message_test

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/pkg/message/text"
)

func TestKeyboardOverflowReportsCountsAndRemedy(t *testing.T) {
	buttons := make([]button.Button, 26)
	for i := range buttons {
		buttons[i] = button.Command("操作", "/action")
	}
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		_, err := message.Markdown("menu").Buttons(buttons[:14]...).Build(scene)
		if !errors.Is(err, message.ErrInvalidMessage) || !errors.Is(err, button.ErrInvalidButton) {
			t.Fatalf("unmatchable overflow error: %v", err)
		}
		for _, hint := range []string{"14 buttons", "2 columns", "7 rows", "Columns(3)"} {
			if !strings.Contains(err.Error(), hint) {
				t.Fatalf("missing %q in %v", hint, err)
			}
		}
		for columns := 1; columns <= 5; columns++ {
			if _, err := message.Markdown("menu").Columns(columns).Buttons(buttons[:columns*5]...).Build(scene); err != nil {
				t.Fatal(err)
			}
			if _, err := message.Markdown("menu").Columns(columns).Buttons(buttons[:columns*5+1]...).Build(scene); !errors.Is(err, button.ErrInvalidButton) {
				t.Fatalf("overflow accepted for %d columns: %v", columns, err)
			}
		}
		_, err = message.Markdown("menu").Columns(5).Buttons(buttons...).Build(scene)
		if err == nil || !strings.Contains(err.Error(), "paginate") {
			t.Fatalf("missing pagination advice: %v", err)
		}
	}
}

func TestMarkdownButtonsProtocolAndDefaults(t *testing.T) {
	outgoing := message.Markdown("## hello").VerifyImages().Reference("REFIDX_source").Buttons(
		button.Command("签到", "/checkin").ID("button_1").Primary(),
		button.Command("绑定", "/bind "), button.Link("官网", "https://example.com"),
	)
	body, err := outgoing.Build(message.SceneGroup)
	if err != nil {
		t.Fatal(err)
	}
	if string(body["msg_type"]) != "2" || body["content"] != nil {
		t.Fatal("invalid Markdown wire fields")
	}
	var markdown struct {
		Content string `json:"content"`
		Verify  bool   `json:"force_verify_image_resource"`
	}
	if err := json.Unmarshal(body["markdown"], &markdown); err != nil {
		t.Fatal(err)
	}
	if markdown.Content != "## hello" || !markdown.Verify {
		t.Fatal("missing Markdown settings")
	}
	if string(body["message_reference"]) != `{"message_id":"REFIDX_source"}` {
		t.Fatal("reference must be an object")
	}
	var kb interaction.Keyboard
	if err := json.Unmarshal(body["keyboard"], &kb); err != nil {
		t.Fatal(err)
	}
	if len(kb.Content.Rows) != 2 || len(kb.Content.Rows[0].Buttons) != 2 {
		t.Fatal("default two-column layout")
	}
	first, second := kb.Content.Rows[0].Buttons[0], kb.Content.Rows[0].Buttons[1]
	if first.Action.Type != 2 || first.Action.Permission.Type != 2 || first.Action.Enter || second.Action.Data != "/bind " {
		t.Fatal("unexpected command defaults")
	}
	if first.ID == second.ID || first.RenderData.Style != 1 {
		t.Fatal("generated IDs must avoid explicit IDs")
	}
}

func TestBuildersSnapshotAndBranchWithoutMutation(t *testing.T) {
	users := []string{"user"}
	raw := interaction.KeyboardButton{ID: "original", RenderData: interaction.ButtonRenderData{Label: "按钮"}, Action: interaction.ButtonAction{Type: 2, Data: "/hello", Permission: interaction.ButtonPermission{Type: 0, SpecifyUserIDs: users}, Modal: &interaction.ButtonModal{Content: "确认操作"}}}
	b := button.Raw(raw)
	users[0] = "changed"
	raw.Action.Modal.Content = "changed"
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if built.Action.Permission.SpecifyUserIDs[0] != "user" || built.Action.Modal.Content != "确认操作" {
		t.Fatal("button borrowed input state")
	}
	base := message.Markdown("body")
	left := base.Buttons(b)
	right := base.Buttons(button.Command("帮助", "/help"))
	var wg sync.WaitGroup
	for _, out := range []message.Outgoing{base, left, right} {
		wg.Add(1)
		go func(m message.Outgoing) {
			defer wg.Done()
			body, err := m.Build(message.SceneGroup)
			if err != nil {
				t.Error(err)
				return
			}
			body["markdown"][0] = '!'
		}(out)
	}
	wg.Wait()
	body, err := base.Build(message.SceneGroup)
	if err != nil {
		t.Fatal(err)
	}
	if body["keyboard"] != nil || !json.Valid(body["markdown"]) {
		t.Fatal("builder or payload snapshots share mutable state")
	}
	values := map[string]any{"enabled": true}
	extension := message.New(99).WithField("card", values)
	values["enabled"] = false
	fields, err := extension.Build(message.SceneGroup)
	if err != nil {
		t.Fatal(err)
	}
	if string(fields["card"]) != `{"enabled":true}` {
		t.Fatal("extension fields were not snapshotted")
	}
}

func TestBuilderRejectsInvalidCombinations(t *testing.T) {
	for name, build := range map[string]message.Outgoing{
		"empty":                {},
		"markdown-content":     message.Markdown("body").WithField("content", "text"),
		"raw-and-buttons":      message.Markdown("body").KeyboardTemplate("template").Buttons(button.Command("help", "/help")),
		"duplicate-id":         message.Markdown("body").Buttons(button.Command("one", "/one").ID("same"), button.Command("two", "/two").ID("same")),
		"too-many-rows":        message.Markdown("body").Columns(1).Buttons(button.Command("a", "a"), button.Command("b", "b"), button.Command("c", "c"), button.Command("d", "d"), button.Command("e", "e"), button.Command("f", "f")),
		"auto-and-explicit":    message.Markdown("body").Buttons(button.Command("a", "a")).Row(button.Command("b", "b")),
		"invalid-columns":      message.Markdown("body").Columns(0),
		"reserved":             message.Markdown("body").WithField("event_id", "event-name"),
		"marshal-error":        message.Markdown("body").WithField("future", func() {}),
		"empty-template":       message.Markdown("body").KeyboardTemplate(""),
		"template-and-content": message.Markdown("body").Keyboard(&interaction.Keyboard{ID: "id", Content: &interaction.KeyboardContent{}}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := build.Build(message.SceneGroup); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}

func TestSceneRestrictionsAndForwardCompatibility(t *testing.T) {
	auto := message.Markdown("body").Buttons(button.Command("hello", "/hello").AutoSend())
	if _, err := auto.Strict().Build(message.SceneGroup); !errors.Is(err, message.ErrInvalidMessage) {
		t.Fatalf("group AutoSend: %v", err)
	}
	if _, err := auto.Build(message.SceneC2C); err != nil {
		t.Fatal(err)
	}
	roles := message.Markdown("body").Buttons(button.Callback("refresh", "refresh").Roles("role"))
	if _, err := roles.Build(message.SceneGroup); err == nil {
		t.Fatal("roles accepted in group scene")
	}
	body, err := roles.Build(message.SceneChannel)
	if err != nil {
		t.Fatal(err)
	}
	if body["msg_type"] != nil {
		t.Fatal("channel does not use msg_type")
	}
	raw := map[string]any{"content": map[string]any{"rows": []any{map[string]any{"buttons": []any{map[string]any{"id": "callback", "future_button_field": true, "render_data": map[string]any{"label": "refresh", "style": 100}, "action": map[string]any{"type": 1, "data": "refresh", "permission": map[string]any{"type": 2}}}}}}}}
	body, err = message.Markdown("body").WithField("keyboard", raw).Build(message.SceneC2C)
	if err != nil {
		t.Fatal(err)
	}
	var kb map[string]any
	if err := json.Unmarshal(body["keyboard"], &kb); err != nil {
		t.Fatal(err)
	}
	rows := kb["content"].(map[string]any)["rows"].([]any)
	if !rows[0].(map[string]any)["buttons"].([]any)[0].(map[string]any)["future_button_field"].(bool) {
		t.Fatal("extension field was lost")
	}
}

func TestCallbackModalAndPermissions(t *testing.T) {
	b := button.Callback("确认", "confirm").ID("confirm").Group("choice").Users("user").Confirm("确定执行？").Visited("已完成")
	v, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if v.GroupID != "choice" || v.Action.Modal == nil || v.Action.Permission.SpecifyUserIDs[0] != "user" {
		t.Fatal("callback configuration missing")
	}
	for _, invalid := range []button.Button{button.Callback("a", "data").AutoSend(), button.Command("a", "/a").Group("g"), button.Command("a", "/a").Users(), button.Callback("a", "data").Confirm("https://example.com"), button.Command("12345678901", "/a")} {
		if _, err := invalid.Build(); !errors.Is(err, button.ErrInvalidButton) {
			t.Fatalf("want invalid button, got %v", err)
		}
	}
}
func TestExplicitRowsTemplatesAndRawFieldExtensions(t *testing.T) {
	m := message.MarkdownTemplate("template").Row(button.Command("help", "/help")).Row(button.Command("menu", "/menu"), button.Link("site", "https://example.com"))
	body, err := m.Build(message.SceneGroup)
	if err != nil {
		t.Fatal(err)
	}
	var kb interaction.Keyboard
	if err := json.Unmarshal(body["keyboard"], &kb); err != nil {
		t.Fatal(err)
	}
	if len(kb.Content.Rows) != 2 || len(kb.Content.Rows[0].Buttons) != 1 || len(kb.Content.Rows[1].Buttons) != 2 {
		t.Fatal("explicit row layout was changed")
	}
	body, err = message.Markdown("body").KeyboardTemplate("keyboard-template").Build(message.SceneGroup)
	if err != nil {
		t.Fatal(err)
	}
	if string(body["keyboard"]) != `{"id":"keyboard-template"}` {
		t.Fatal("keyboard template must use short form")
	}
	body, err = message.Text("body").Buttons(button.Command("help", "/help")).Build(message.SceneC2C)
	if err != nil {
		t.Fatal(err)
	}
	if string(body["msg_type"]) != "0" || body["keyboard"] == nil {
		t.Fatal("keyboard extension must preserve the text message type")
	}
	if _, err := message.Media("uploaded-file-info").Build(message.SceneGroup); err != nil {
		t.Fatal(err)
	}
	if _, err := message.Media("uploaded-file-info").Build(message.SceneChannel); err == nil {
		t.Fatal("file_info must not be sent to channel API")
	}
}
func TestCommandHintsFallBackWithoutMutatingSharedMessage(t *testing.T) {
	raw := map[string]any{
		"content": map[string]any{"future_content": true, "rows": []any{
			map[string]any{"future_row": true, "buttons": []any{
				map[string]any{"id": "bind", "future_button": json.RawMessage(`9007199254740993`),
					"render_data": map[string]any{"label": "绑定", "style": 1},
					"action": map[string]any{"type": 2, "data": "/bind ", "enter": true, "anchor": 1, "reply": true, "future_action": true,
						"permission": map[string]any{"type": 0, "specify_user_ids": []string{"user"}}}},
			}},
		}},
	}
	outgoing := message.Markdown("body").WithField("keyboard", raw)
	for _, scene := range []message.Scene{message.SceneGroup, message.SceneChannel, message.SceneDirectMessage, message.SceneC2C} {
		body, err := outgoing.Build(scene)
		if err != nil {
			t.Fatalf("%s: %v", scene, err)
		}
		var kb interaction.Keyboard
		if err := json.Unmarshal(body["keyboard"], &kb); err != nil {
			t.Fatal(err)
		}
		b := kb.Content.Rows[0].Buttons[0]
		if scene == message.SceneC2C {
			if !b.Action.Enter || b.Action.Anchor != 1 {
				t.Fatal("fallback mutated the shared C2C message")
			}
		} else if b.Action.Enter || b.Action.Anchor != 0 {
			t.Fatal("unsupported hints were sent")
		}
		if b.Action.Data != "/bind " || !b.Action.Reply || b.Action.Permission.Type != 0 || b.Action.Permission.SpecifyUserIDs[0] != "user" {
			t.Fatal("fallback changed command or permissions")
		}
		if !strings.Contains(string(body["keyboard"]), `9007199254740993`) || !strings.Contains(string(body["keyboard"]), `"future_action":true`) || !strings.Contains(string(body["keyboard"]), `"future_row":true`) {
			t.Fatal("fallback lost extension fields")
		}
	}
	if _, err := outgoing.Strict().Build(message.SceneGroup); !errors.Is(err, message.ErrInvalidMessage) {
		t.Fatal("strict mode must report unsupported hints")
	}
	if _, err := outgoing.Strict().Build(message.SceneC2C); err != nil {
		t.Fatal(err)
	}
	// The high-level helper follows exactly the same default fallback.
	// 高级便捷方法遵循完全相同的默认降级行为。
	if _, err := message.Markdown("body").Buttons(button.Command("帮助", "/help").AutoSend()).Build(message.SceneGroup); err != nil {
		t.Fatal(err)
	}
}
func TestTextFragmentsComposeAndRespectSceneWithoutMutation(t *testing.T) {
	shared := message.Markdown("## 菜单\n").VerifyImages().Append(text.Input("/bind 名字 ").Label("绑定账号").Reference(), text.Plain("\n"), text.Enter("/help")).Buttons(button.Command("帮助", "/help").AutoSend())
	for _, scene := range []message.Scene{message.SceneGroup, message.SceneC2C} {
		body, err := shared.Build(scene)
		if err != nil {
			t.Fatal(err)
		}
		var markdown struct {
			Content string `json:"content"`
			Verify  bool   `json:"force_verify_image_resource"`
		}
		if err := json.Unmarshal(body["markdown"], &markdown); err != nil {
			t.Fatal(err)
		}
		if !markdown.Verify || !strings.Contains(markdown.Content, `text="%2Fbind%20%E5%90%8D%E5%AD%97%20"`) || !strings.Contains(markdown.Content, `reference="true"`) {
			t.Fatal("fragments lost content, encoding or settings")
		}
		if strings.Contains(markdown.Content, "<qqbot-cmd-enter ") != (scene == message.SceneC2C) {
			t.Fatal("Enter scene fallback mutated original message")
		}
	}
	if _, err := shared.Strict().Build(message.SceneGroup); !errors.Is(err, text.ErrUnsupportedFragment) {
		t.Fatal("strict text Enter did not reject")
	}
	if _, err := message.MarkdownParts(text.AtUser("member"), text.Plain(" 你好")).Build(message.SceneGroup); err != nil {
		t.Fatal(err)
	}
	if _, err := message.TextParts(text.AtEveryone(), text.Plain(" 公告")).Build(message.SceneGroup); err != nil {
		t.Fatal(err)
	}
	if _, err := message.TextParts(text.AtUser("member")).Build(message.SceneGroup); !errors.Is(err, text.ErrUnsupportedFragment) {
		t.Fatal("group plain-text user mention was accepted")
	}
	if _, err := message.MarkdownTemplate("template").Append(text.Plain("body")).Build(message.SceneGroup); err == nil {
		t.Fatal("fragments must not silently replace template content")
	}
}

func TestButtonAdvancedFieldsAndChooserFallback(t *testing.T) {
	b := button.Command("发送图片", "/image ").ID("image").AutoSend().PickImage().ChooseChannel().ClickLimit(2).Confirmation(interaction.ButtonModal{Content: "确定？", ConfirmText: "发送", CancelText: "取消"})
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup, message.SceneChannel} {
		body, err := message.Markdown("body").Buttons(b).Build(scene)
		if err != nil {
			t.Fatal(err)
		}
		var kb interaction.Keyboard
		if err := json.Unmarshal(body["keyboard"], &kb); err != nil {
			t.Fatal(err)
		}
		action := kb.Content.Rows[0].Buttons[0].Action
		if action.AtBotShowChannelList != (scene == message.SceneChannel) {
			t.Fatal("channel chooser fallback incorrect")
		}
		if (action.Anchor == 1) != (scene == message.SceneC2C) || action.Enter != (scene == message.SceneC2C) {
			t.Fatal("picker/auto-send fallback incorrect")
		}
		if action.ClickLimit != 2 || action.Modal.ConfirmText != "发送" || action.Modal.CancelText != "取消" {
			t.Fatal("advanced fields lost")
		}
	}
	if _, err := message.Markdown("body").Strict().Buttons(button.Command("选择频道", "/help").ChooseChannel()).Build(message.SceneC2C); err == nil {
		t.Fatal("strict chooser must reject C2C")
	}
}
