package message_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func TestMaterialIntegerWireCompatibilityAndBounds(t *testing.T) {
	for _, body := range []string{
		`{"images":[{"image_id":"i","created_at":1793746502,"expires_at":1801522502}],"total":1}`,
		`{"images":[{"image_id":"i","created_at":"1793746502","expires_at":"1801522502"}],"total":"1"}`,
	} {
		var page message.ImageMaterialPage
		if err := json.Unmarshal([]byte(body), &page); err != nil {
			t.Fatal(err)
		}
		if page.Total != 1 || page.Images[0].CreatedAt != 1793746502 || page.Images[0].ExpiresAt != 1801522502 {
			t.Fatal("integer conversion changed values")
		}
		if !json.Valid(page.Raw) || !json.Valid(page.Images[0].Raw) {
			t.Fatal("wire decoder lost raw snapshots")
		}
	}
	var page message.ImageMaterialPage
	if err := json.Unmarshal([]byte(`{"total":"18446744073709551615"}`), &page); err != nil || page.Total != ^uint64(0) {
		t.Fatalf("uint64 precision: %d %v", page.Total, err)
	}
	for _, total := range []string{`"18446744073709551616"`, `-1`, `1.5`, `true`, `null`, `""`, `"oops"`, `"1e3"`} {
		if err := json.Unmarshal([]byte(`{"total":`+total+`}`), &page); err == nil {
			t.Fatalf("invalid total accepted: %s", total)
		}
	}
	var image message.ImageMaterial
	if err := json.Unmarshal([]byte(`{"created_at":"9223372036854775807","expires_at":"-1"}`), &image); err != nil || image.CreatedAt != 9223372036854775807 || image.ExpiresAt != -1 {
		t.Fatalf("int64 precision: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"expires_at":"9223372036854775808"}`), &image); err == nil {
		t.Fatal("timestamp overflow accepted")
	}
}

func TestMaterialDimensionsPreserveReferenceAndOriginal(t *testing.T) {
	image := message.ImageMaterial{URL: "qqbot-media://image/v1/exact-token"}
	sized := image.WithDimensions(640, 480)
	text, err := sized.MarkdownText("图片")
	if err != nil || text != `![图片 #640px #480px](qqbot-media://image/v1/exact-token)` {
		t.Fatalf("sized Markdown: %s %v", text, err)
	}
	if image.Width != 0 || image.Height != 0 || sized.URL != image.URL {
		t.Fatal("dimensions mutated original or reference")
	}
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		if _, err := sized.Markdown("图片").Build(scene); err != nil {
			t.Fatal(err)
		}
	}
	for _, dims := range [][2]int{{0, 10}, {10, 0}, {-1, 10}} {
		if _, err := image.WithDimensions(dims[0], dims[1]).MarkdownText("说明"); !errors.Is(err, message.ErrInvalidMessage) {
			t.Fatal("invalid dimensions accepted")
		}
	}
}

func TestImageMaterialReferencesAreEscapedAndPreserved(t *testing.T) {
	material := message.ImageMaterial{ID: "image", URL: "qqbot-media://image/v1/exact-token_1"}
	inline, err := material.MarkdownText("说明]\n[图片")
	if err != nil || inline != `![说明\] \[图片](qqbot-media://image/v1/exact-token_1)` {
		t.Fatalf("reference: %q %v", inline, err)
	}
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		body, err := material.Markdown("说明").Build(scene)
		if err != nil || !strings.Contains(string(body["markdown"]), material.URL) {
			t.Fatalf("outgoing reference: %v", err)
		}
	}
	for _, link := range []string{"", "https://example.test/image", "qqbot-media://image/v1/", "qqbot-media://image/v1/x) injected"} {
		if _, err := (message.ImageMaterial{URL: link}).Markdown("alt").Build(message.SceneC2C); !errors.Is(err, message.ErrInvalidMessage) {
			t.Fatalf("invalid reference: %v", err)
		}
	}
}

func TestReceivedImagesIncludeNestedStickersWithoutGuessingText(t *testing.T) {
	m, err := message.DecodeReceived(message.GROUP_AT_MESSAGE_CREATE, []byte(`{"content":"😀 <emoji:307> [表情]","attachments":[{"content_type":"image/jpeg","url":"image"},{"content_type":"audio/ogg","url":"voice"}],"msg_elements":[{"attachments":[{"content_type":"image/gif","url":"sticker"}],"msg_elements":[{"attachments":[{"content_type":"image/webp","url":"nested"},{"content_type":"application/octet-stream","filename":"fake.jpg","url":"file"}]}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	images := m.Images()
	if len(images) != 3 || images[0].URL != "image" || images[1].URL != "sticker" || images[2].URL != "nested" || len(m.AllAttachments()) != 5 {
		t.Fatal("image or recursive attachment classification failed")
	}
	if m.Content != "😀 <emoji:307> [表情]" {
		t.Fatal("emoji text was modified")
	}
	images[0].URL = "changed"
	if m.Attachments[0].URL != "image" {
		t.Fatal("image list borrowed attachment storage")
	}
	var empty *message.Received
	if len(empty.Images()) != 0 || len(empty.AllAttachments()) != 0 {
		t.Fatal("nil message helpers")
	}
}
