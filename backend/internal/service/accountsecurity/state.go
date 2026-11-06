package accountsecurity

import "time"

import "context"

type PasswordState struct {
	PasswordChangeRequired bool  `json:"passwordChangeRequired"`
	PasswordExpired        bool  `json:"passwordExpired"`
	PasswordExpiresAt      int64 `json:"passwordExpiresAt"`
	PasswordExpiryWarning  bool  `json:"passwordExpiryWarning"`
}

func EvaluatePasswordState(policy Policy, changedAt int64, mustChange bool, now time.Time) PasswordState {
	if !policy.Enabled {
		return PasswordState{}
	}
	state := PasswordState{PasswordChangeRequired: mustChange}
	if policy.MaxAgeDays <= 0 {
		return state
	}
	baseline := changedAt
	if baseline <= 0 {
		baseline = policy.EffectiveAt
	}
	if baseline <= 0 {
		return state
	}
	expiresAt := time.UnixMilli(baseline).Add(time.Duration(policy.MaxAgeDays) * 24 * time.Hour)
	state.PasswordExpiresAt = expiresAt.UnixMilli()
	if !now.Before(expiresAt) {
		state.PasswordExpired = true
		state.PasswordChangeRequired = true
		return state
	}
	if policy.ExpiryWarningDays > 0 {
		warningAt := expiresAt.Add(-time.Duration(policy.ExpiryWarningDays) * 24 * time.Hour)
		state.PasswordExpiryWarning = !now.Before(warningAt)
	}
	return state
}

func EvaluatePasswordStateContext(ctx context.Context, changedAt int64, mustChange bool) (PasswordState, error) {
	policy, err := LoadPolicyContext(ctx)
	if err != nil {
		return PasswordState{}, err
	}
	return EvaluatePasswordState(policy, changedAt, mustChange, time.Now()), nil
}
