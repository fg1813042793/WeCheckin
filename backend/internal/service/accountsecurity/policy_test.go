package accountsecurity

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultPolicyIsDisabledStandardProfile(t *testing.T) {
	policy := DefaultPolicy()
	if policy.Enabled || policy.Version != 1 || policy.MinLength != 8 || !policy.RequireLetter || !policy.RequireNumber {
		t.Fatalf("default policy = %#v", policy)
	}
	if policy.MaxAgeDays != 90 || policy.ExpiryWarningDays != 7 || policy.HistoryCount != 5 {
		t.Fatalf("default lifecycle = %#v", policy)
	}
	if !policy.ForceChangeOnCreate || !policy.ForceChangeOnReset || !policy.AllowWeakTemporaryPassword {
		t.Fatalf("default first-login behavior = %#v", policy)
	}
}

func TestDecodePolicyFallsBackToDisabledDefaults(t *testing.T) {
	for _, raw := range []string{"", "{bad", `{"enabled":true,"minLength":2}`} {
		policy := DecodePolicy(raw)
		if policy.Enabled {
			t.Fatalf("invalid policy must fail closed to disabled, raw=%q policy=%#v", raw, policy)
		}
		if policy.MinLength != 8 {
			t.Fatalf("fallback defaults missing for %q: %#v", raw, policy)
		}
	}
}

func TestPreparePolicyForSaveValidatesAndSetsEffectiveAt(t *testing.T) {
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, time.Local)
	previous := DefaultPolicy()
	next := DefaultPolicy()
	next.Enabled = true
	prepared, err := PreparePolicyForSave(previous, next, now)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.EffectiveAt != now.UnixMilli() {
		t.Fatalf("effectiveAt=%d", prepared.EffectiveAt)
	}
	preparedAgain, err := PreparePolicyForSave(prepared, prepared, now.Add(time.Hour))
	if err != nil || preparedAgain.EffectiveAt != prepared.EffectiveAt {
		t.Fatalf("enabled policy must preserve effectiveAt: %#v err=%v", preparedAgain, err)
	}
	prepared.Enabled = false
	disabled, err := PreparePolicyForSave(preparedAgain, prepared, now.Add(2*time.Hour))
	if err != nil || disabled.EffectiveAt != 0 {
		t.Fatalf("disabled policy must clear effectiveAt: %#v err=%v", disabled, err)
	}
}

func TestValidatePolicyRejectsInvalidBounds(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy)
		want   string
	}{
		{"min length", func(p *Policy) { p.MinLength = 5 }, "6 至 64"},
		{"max age", func(p *Policy) { p.MaxAgeDays = 3651 }, "0 至 3650"},
		{"warning", func(p *Policy) { p.MaxAgeDays = 30; p.ExpiryWarningDays = 31 }, "不能大于"},
		{"history", func(p *Policy) { p.HistoryCount = 21 }, "0 至 20"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := DefaultPolicy()
			tc.mutate(&policy)
			if err := ValidatePolicy(policy); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidatePolicy() err=%v", err)
			}
		})
	}
}

func TestValidatePermanentPassword(t *testing.T) {
	policy := DefaultPolicy()
	policy.Enabled = true
	for _, tc := range []struct {
		password string
		ok       bool
	}{
		{"abc12345", true},
		{"abcdefgh", false},
		{"12345678", false},
		{"abc123", false},
	} {
		err := ValidatePermanentPassword(policy, tc.password)
		if (err == nil) != tc.ok {
			t.Fatalf("password %q err=%v", tc.password, err)
		}
	}
	policy.RequireUppercase = true
	policy.RequireLowercase = true
	policy.RequireSpecial = true
	if err := ValidatePermanentPassword(policy, "Abc123!x"); err != nil {
		t.Fatalf("strong password rejected: %v", err)
	}
	if err := ValidatePermanentPassword(policy, "abc12345"); err == nil {
		t.Fatal("missing uppercase/special password accepted")
	}
}
