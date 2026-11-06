package accountsecurity

import (
	"testing"
	"time"
)

func TestEvaluatePasswordState(t *testing.T) {
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	policy := DefaultPolicy()
	if state := EvaluatePasswordState(policy, 0, true, now); state.PasswordChangeRequired {
		t.Fatalf("disabled policy must not restrict: %#v", state)
	}

	policy.Enabled = true
	policy.EffectiveAt = now.Add(-85 * 24 * time.Hour).UnixMilli()
	state := EvaluatePasswordState(policy, 0, false, now)
	if state.PasswordChangeRequired || state.PasswordExpired || !state.PasswordExpiryWarning {
		t.Fatalf("warning state = %#v", state)
	}

	state = EvaluatePasswordState(policy, now.Add(-91*24*time.Hour).UnixMilli(), false, now)
	if !state.PasswordChangeRequired || !state.PasswordExpired || state.PasswordExpiresAt == 0 {
		t.Fatalf("expired state = %#v", state)
	}

	state = EvaluatePasswordState(policy, now.UnixMilli(), true, now)
	if !state.PasswordChangeRequired || state.PasswordExpired {
		t.Fatalf("must-change state = %#v", state)
	}
}
