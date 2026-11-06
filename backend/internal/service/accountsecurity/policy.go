package accountsecurity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"

	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/pkg/database"
)

const PasswordPolicySetupKey = "PASSWORD_POLICY_CONFIG"

type Policy struct {
	Version                    int   `json:"version"`
	Enabled                    bool  `json:"enabled"`
	MinLength                  int   `json:"minLength"`
	RequireLetter              bool  `json:"requireLetter"`
	RequireNumber              bool  `json:"requireNumber"`
	RequireUppercase           bool  `json:"requireUppercase"`
	RequireLowercase           bool  `json:"requireLowercase"`
	RequireSpecial             bool  `json:"requireSpecial"`
	MaxAgeDays                 int   `json:"maxAgeDays"`
	ExpiryWarningDays          int   `json:"expiryWarningDays"`
	HistoryCount               int   `json:"historyCount"`
	ForceChangeOnCreate        bool  `json:"forceChangeOnCreate"`
	ForceChangeOnReset         bool  `json:"forceChangeOnReset"`
	AllowWeakTemporaryPassword bool  `json:"allowWeakTemporaryPassword"`
	EffectiveAt                int64 `json:"effectiveAt"`
}

func DefaultPolicy() Policy {
	return Policy{
		Version: 1, MinLength: 8, RequireLetter: true, RequireNumber: true,
		MaxAgeDays: 90, ExpiryWarningDays: 7, HistoryCount: 5,
		ForceChangeOnCreate: true, ForceChangeOnReset: true, AllowWeakTemporaryPassword: true,
	}
}

func DecodePolicy(raw string) Policy {
	policy := DefaultPolicy()
	if raw == "" {
		return policy
	}
	if err := json.Unmarshal([]byte(raw), &policy); err != nil || ValidatePolicy(policy) != nil {
		return DefaultPolicy()
	}
	return policy
}

func ValidatePolicy(policy Policy) error {
	if policy.Version != 1 {
		return validationError("密码策略版本无效")
	}
	if policy.MinLength < 6 || policy.MinLength > 64 {
		return validationError("密码最小长度必须为 6 至 64")
	}
	if policy.MaxAgeDays < 0 || policy.MaxAgeDays > 3650 {
		return validationError("密码有效期必须为 0 至 3650 天")
	}
	if policy.ExpiryWarningDays < 0 || policy.ExpiryWarningDays > policy.MaxAgeDays {
		return validationError("密码到期提醒天数不能大于有效期")
	}
	if policy.HistoryCount < 0 || policy.HistoryCount > 20 {
		return validationError("历史密码数量必须为 0 至 20")
	}
	return nil
}

func PreparePolicyForSave(previous, next Policy, now time.Time) (Policy, error) {
	next.Version = 1
	if err := ValidatePolicy(next); err != nil {
		return Policy{}, err
	}
	if next.Enabled && !previous.Enabled {
		next.EffectiveAt = now.UnixMilli()
	} else if !next.Enabled {
		next.EffectiveAt = 0
	} else if next.EffectiveAt == 0 {
		next.EffectiveAt = previous.EffectiveAt
	}
	return next, nil
}

func ValidatePermanentPassword(policy Policy, password string) error {
	minLength := 6
	if policy.Enabled {
		minLength = policy.MinLength
	}
	if utf8.RuneCountInString(password) < minLength {
		return validationError(fmt.Sprintf("密码至少需要 %d 位", minLength))
	}
	if !policy.Enabled {
		return nil
	}
	var hasLetter, hasNumber, hasUpper, hasLower, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
			hasUpper = hasUpper || unicode.IsUpper(r)
			hasLower = hasLower || unicode.IsLower(r)
		case unicode.IsNumber(r):
			hasNumber = true
		default:
			hasSpecial = true
		}
	}
	if policy.RequireLetter && !hasLetter {
		return validationError("密码必须包含字母")
	}
	if policy.RequireNumber && !hasNumber {
		return validationError("密码必须包含数字")
	}
	if policy.RequireUppercase && !hasUpper {
		return validationError("密码必须包含大写字母")
	}
	if policy.RequireLowercase && !hasLower {
		return validationError("密码必须包含小写字母")
	}
	if policy.RequireSpecial && !hasSpecial {
		return validationError("密码必须包含特殊字符")
	}
	return nil
}

func LoadPolicyContext(ctx context.Context) (Policy, error) {
	if database.GetDB() == nil {
		return DefaultPolicy(), nil
	}
	setup, err := setupservice.GetSetupContext(ctx, PasswordPolicySetupKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultPolicy(), nil
	}
	if err != nil {
		return Policy{}, err
	}
	return DecodePolicy(setup.Value), nil
}

func SavePolicyContext(ctx context.Context, next Policy, addIP string) (Policy, error) {
	previous, err := LoadPolicyContext(ctx)
	if err != nil {
		return Policy{}, err
	}
	prepared, err := PreparePolicyForSave(previous, next, time.Now())
	if err != nil {
		return Policy{}, err
	}
	encoded, err := json.Marshal(prepared)
	if err != nil {
		return Policy{}, err
	}
	if err := setupservice.SetSetupContext(ctx, PasswordPolicySetupKey, string(encoded), "json", addIP); err != nil {
		return Policy{}, err
	}
	return prepared, nil
}
