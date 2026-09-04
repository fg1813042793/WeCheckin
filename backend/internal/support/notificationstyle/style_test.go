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
