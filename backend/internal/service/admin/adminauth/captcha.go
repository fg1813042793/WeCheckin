package adminauth

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/internal/support/logincaptcha"
)

const (
	AdminLoginCaptchaSetupKey = "ADMIN_LOGIN_CAPTCHA_ENABLED"
	adminCaptchaRedisPrefix   = "admin:login:captcha:"
	adminCaptchaTTL           = 2 * time.Minute
)

var (
	ErrCaptchaInvalid     = logincaptcha.ErrInvalid
	ErrCaptchaUnavailable = logincaptcha.ErrUnavailable
)

type CaptchaChallenge = logincaptcha.Challenge

func adminCaptchaEnabledContext(ctx context.Context) (bool, error) {
	setup, err := setupservice.GetSetupContext(ctx, AdminLoginCaptchaSetupKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if setup == nil {
		return false, nil
	}
	value := strings.TrimSpace(strings.ToLower(setup.Value))
	return value == "1" || value == "true" || value == "on", nil
}

var defaultCaptchaService = logincaptcha.NewService(
	adminCaptchaEnabledContext,
	logincaptcha.NewImageGenerator(),
	logincaptcha.NewRedisStore(adminCaptchaRedisPrefix),
	adminCaptchaTTL,
)

func GenerateCaptchaContext(ctx context.Context) (CaptchaChallenge, error) {
	return defaultCaptchaService.Challenge(ctx)
}

func VerifyCaptchaContext(ctx context.Context, id, answer string) error {
	return defaultCaptchaService.Verify(ctx, id, answer)
}

func LoginWithCaptchaContext(ctx context.Context, name, password, captchaID, captchaCode, addIP, device string) (*LoginResponse, error) {
	if err := VerifyCaptchaContext(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	return LoginContext(ctx, name, password, addIP, device)
}
