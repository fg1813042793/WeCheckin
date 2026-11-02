package auditlog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPublicSnapshotValuesAreTypedAndExact(t *testing.T) {
	tests := []struct {
		field string
		value interface{}
		want  interface{}
		ok    bool
	}{
		{"settings.ADMIN_LOGIN_CAPTCHA_ENABLED", "1", true, true},
		{"settings.DINGTALK_H5_LOGIN_CAPTCHA_ENABLED", "false", false, true},
		{"settings.ADMIN_SINGLE_LOGIN", 0, false, true},
		{"settings.TOKEN_ADMIN_EXPIRE", "24h", "24h0m0s", true},
		{"settings.DINGTALK_H5_NOTIFY_MODE", "agent_fallback", "agent_fallback", true},
		{"settings.corp.17.enabled", true, true, true},
		{"settings.corp.17.notify_mode", "robot", "robot", true},
		{"settings.ADMIN_LOGIN_CAPTCHA_ENABLED", nil, nil, true},
		{"settings.ADMIN_LOGIN_CAPTCHA_ENABLED", "never-store-secret", nil, false},
		{"settings.ADMIN_LOGIN_CAPTCHA_ENABLED", map[string]string{"token": "secret"}, nil, false},
		{"settings.ADMIN_LOGIN_CAPTCHA_ENABLED", 2, nil, false},
		{"settings.TOKEN_ADMIN_EXPIRE", "access-token-value", nil, false},
		{"settings.TOKEN_ADMIN_EXPIRE", "-1h", nil, false},
		{"settings.corp.0.enabled", true, nil, false},
		{"settings.corp.17.app_secret", "secret", nil, false},
		{"settings.UNKNOWN_ENABLED", true, nil, false},
		{"settings.admin_login_captcha_enabled", true, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, ok := PublicSnapshotValue(tt.field, tt.value)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("got %v (%v), want %v (%v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestChangesExposeOnlyValidatedPublicSettings(t *testing.T) {
	raw, err := EncodeChanges([]Change{
		{Field: "settings.ADMIN_LOGIN_CAPTCHA_ENABLED", Before: "0", After: "1"},
		{Field: "settings.TOKEN_ADMIN_EXPIRE", Before: "24h", After: "168h"},
		{Field: "settings.DINGTALK_H5_APP_SECRET", Before: "old-private", After: "new-private"},
		{Field: "settings.UNKNOWN_ENABLED", Before: "old-private", After: "new-private"},
		{Field: "settings.corp.2.enabled", Before: "private-disguised-as-flag", After: true},
		{Field: "settings", Before: nil, After: map[string]interface{}{"USER_SINGLE_LOGIN": "root-private-value", "UNKNOWN_ENABLED": "root-private-value"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var changes []Change
	if err = json.Unmarshal(raw, &changes); err != nil {
		t.Fatal(err)
	}
	if changes[0].Before != false || changes[0].After != true {
		t.Fatalf("flag: %#v", changes[0])
	}
	if changes[1].Before != "24h0m0s" || changes[1].After != "168h0m0s" {
		t.Fatalf("duration: %#v", changes[1])
	}
	if strings.Contains(string(raw), "private") {
		t.Fatalf("private values leaked: %s", raw)
	}
	for _, change := range changes[2:] {
		if change.Before != "[已隐藏]" || change.After != "[已变更]" {
			t.Fatalf("not redacted: %#v", change)
		}
	}
}
