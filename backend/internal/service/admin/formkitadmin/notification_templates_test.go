package formkitadmin

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeBuiltinNotificationTemplatesFallsBackToDefaults(t *testing.T) {
	for _, raw := range []string{"", "not-json", "[]"} {
		got := decodeBuiltinNotificationTemplates(raw)
		if !reflect.DeepEqual(got, defaultBuiltinNotificationTemplates()) {
			t.Fatalf("decode %q = %#v", raw, got)
		}
	}
}

func TestDecodeBuiltinNotificationTemplatesPreservesConfiguredValues(t *testing.T) {
	raw := `[{"label":"自定义内置","value":"{result}"}]`
	got := decodeBuiltinNotificationTemplates(raw)
	want := []TemplatePreset{{"label": "自定义内置", "value": "{result}"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decode configured templates = %#v", got)
	}
}

func TestValidateNotificationTemplateScope(t *testing.T) {
	for _, scope := range []string{NotificationTemplateScopeBuiltin, NotificationTemplateScopeCustom} {
		if err := validateNotificationTemplateScope(scope); err != nil {
			t.Fatalf("scope %s: %v", scope, err)
		}
	}
	if err := validateNotificationTemplateScope("unknown"); !errors.Is(err, ErrInvalidNotificationTemplateScope) {
		t.Fatalf("invalid scope error = %v", err)
	}
}

func TestDefaultExamNotificationTemplatesExposeScoreVariables(t *testing.T) {
	templates := defaultBuiltinExamNotificationTemplates()
	if len(templates) == 0 {
		t.Fatal("exam notification templates are required")
	}
	joined := ""
	for _, item := range templates {
		joined += item["value"]
	}
	for _, variable := range []string{"{score}", "{fullScore}", "{result}"} {
		if !strings.Contains(joined, variable) {
			t.Fatalf("exam templates missing %s: %q", variable, joined)
		}
	}
}
