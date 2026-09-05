package notificationstyle

import (
	"errors"
	"testing"
)

func TestDefaultConfigCoversSupportedNotificationTypes(t *testing.T) {
	config := DefaultConfig()
	if config.Version != CurrentVersion {
		t.Fatalf("default version = %d, want %d", config.Version, CurrentVersion)
	}
	if len(config.Styles) != len(SupportedTypes()) {
		t.Fatalf("default styles = %d, supported types = %d", len(config.Styles), len(SupportedTypes()))
	}
	for _, notificationType := range SupportedTypes() {
		style := StyleFor(config, notificationType)
		if style.Type != notificationType || style.Label == "" || style.Icon == "" || style.Tone == "" {
			t.Fatalf("style for %q = %#v", notificationType, style)
		}
	}
}

func TestDefaultConfigIncludesInstanceFormRevisedStyle(t *testing.T) {
	style := StyleFor(DefaultConfig(), TypeInstanceFormRevised)
	if style.Type != TypeInstanceFormRevised || style.Label != "表单修改" || style.Icon == "" {
		t.Fatalf("instance form revised style = %#v", style)
	}
}

func TestNormalizeMergesPartialOverridesWithDefaults(t *testing.T) {
	config, err := Normalize(Config{Styles: []Style{{
		Type: TypeTaskArrived, Label: "新待办", Icon: "bell", Tone: ToneDanger,
	}}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	style := StyleFor(config, TypeTaskArrived)
	if style.Label != "新待办" || style.Icon != "bell" || style.Tone != ToneDanger {
		t.Fatalf("overridden style = %#v", style)
	}
	if fallback := StyleFor(config, TypeAdminManual); fallback.Type != TypeAdminManual || fallback.Label == "" {
		t.Fatalf("default style was not merged: %#v", fallback)
	}
}

func TestNormalizeRejectsInvalidStyleValues(t *testing.T) {
	tests := []struct {
		name  string
		style Style
		want  error
	}{
		{name: "unknown type", style: Style{Type: "unknown", Label: "x", Icon: "email", Tone: ToneInfo}, want: ErrUnknownType},
		{name: "empty label", style: Style{Type: TypeAdminManual, Icon: "email", Tone: ToneInfo}, want: ErrInvalidLabel},
		{name: "invalid icon", style: Style{Type: TypeAdminManual, Label: "x", Icon: "<script>", Tone: ToneInfo}, want: ErrInvalidIcon},
		{name: "invalid tone", style: Style{Type: TypeAdminManual, Label: "x", Icon: "email", Tone: "rainbow"}, want: ErrInvalidTone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Normalize(Config{Styles: []Style{tt.style}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("Normalize() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDecodeFallsBackToDefaultsWhenSetupIsEmpty(t *testing.T) {
	config, err := Decode("")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if style := StyleFor(config, TypeApprovalResultReturned); style.Label != "审批退回" {
		t.Fatalf("returned style = %#v", style)
	}
}

func TestNormalizeMergesDingTalkTemplateWithDefaults(t *testing.T) {
	config, err := Normalize(Config{Styles: []Style{{
		Type: TypeTaskArrived, Label: "新待办", Icon: "bell", Tone: ToneDanger,
		DingTalk: DingTalkTemplate{
			MessageType: DingTalkMessageTypeMarkdown,
			Title:       "【待办】{{title}}",
			Content:     "## {{title}}\n{{content}}",
		},
	}}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	template := StyleFor(config, TypeTaskArrived).DingTalk
	if template.MessageType != DingTalkMessageTypeMarkdown || template.Title != "【待办】{{title}}" || template.Content == "" {
		t.Fatalf("DingTalk template = %#v", template)
	}
	if fallback := StyleFor(config, TypeAdminManual).DingTalk; fallback.MessageType != DingTalkMessageTypeAuto {
		t.Fatalf("default DingTalk template = %#v", fallback)
	}
}

func TestNormalizeRejectsUnsupportedDingTalkTemplate(t *testing.T) {
	_, err := Normalize(Config{Styles: []Style{{
		Type: TypeAdminManual, Label: "系统通知", Icon: "email", Tone: TonePrimary,
		DingTalk: DingTalkTemplate{MessageType: "video", Content: "{{content}}"},
	}}})
	if !errors.Is(err, ErrInvalidDingTalkTemplate) {
		t.Fatalf("Normalize() error = %v, want ErrInvalidDingTalkTemplate", err)
	}

	_, err = Normalize(Config{Styles: []Style{{
		Type: TypeAdminManual, Label: "系统通知", Icon: "email", Tone: TonePrimary,
		DingTalk: DingTalkTemplate{MessageType: DingTalkMessageTypeText, Content: "{{password}}"},
	}}})
	if !errors.Is(err, ErrInvalidDingTalkTemplate) {
		t.Fatalf("unknown template variable error = %v, want ErrInvalidDingTalkTemplate", err)
	}
}

func TestRenderDingTalkTemplateReplacesOnlySupportedVariables(t *testing.T) {
	template := DingTalkTemplate{
		MessageType: DingTalkMessageTypeActionCard,
		Title:       "【{{sourceName}}】{{title}}",
		Content:     "## {{title}}\n{{content}}",
		URL:         "{{url}}",
		PicURL:      "{{picUrl}}",
		ButtonTitle: "立即处理",
	}
	rendered := RenderDingTalkTemplate(template, DingTalkTemplateData{
		Title: "审批待办", Content: "请及时处理", URL: "https://example.test/task/1",
		SourceName: "WeCheckin", PicURL: "https://example.test/logo.png",
	})
	if rendered.MessageType != DingTalkMessageTypeActionCard || rendered.Title != "【WeCheckin】审批待办" {
		t.Fatalf("rendered header = %#v", rendered)
	}
	if rendered.Content != "## 审批待办\n请及时处理" || rendered.URL != "https://example.test/task/1" {
		t.Fatalf("rendered content = %#v", rendered)
	}
	if rendered.PicURL != "https://example.test/logo.png" || rendered.ButtonTitle != "立即处理" {
		t.Fatalf("rendered appearance = %#v", rendered)
	}
}

func TestStyleForLegacyTypeKeepsDefaultDingTalkBehavior(t *testing.T) {
	style := StyleFor(Config{}, "legacy_workflow_event")
	if style.Type != "legacy_workflow_event" || style.DingTalk.MessageType != DingTalkMessageTypeAuto {
		t.Fatalf("legacy style = %#v", style)
	}
	if style.DingTalk.Content != "{{content}}" || style.DingTalk.URL != "{{url}}" {
		t.Fatalf("legacy DingTalk template = %#v", style.DingTalk)
	}
}
