package auth

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
	DingTalkH5LoginCaptchaSetupKey = "DINGTALK_H5_LOGIN_CAPTCHA_ENABLED"
	dingTalkH5CaptchaRedisPrefix   = "dingtalk_h5:login:captcha:"
	dingTalkH5CaptchaTTL           = 2 * time.Minute
)

var (
	ErrCaptchaInvalid     = logincaptcha.ErrInvalid
	ErrCaptchaUnavailable = logincaptcha.ErrUnavailable
)

type CaptchaChallenge = logincaptcha.Challenge

type captchaVerifier func(context.Context, string, string) error
type passwordLogin func(context.Context, string, string, string, string) (*LoginResponse, error)

func LoginCaptchaEnabledContext(ctx context.Context) (bool, error) {
	setup, err := setupservice.GetSetupContext(ctx, DingTalkH5LoginCaptchaSetupKey)
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
	LoginCaptchaEnabledContext,
	logincaptcha.NewImageGenerator(),
	logincaptcha.NewRedisStore(dingTalkH5CaptchaRedisPrefix),
	dingTalkH5CaptchaTTL,
)

func GenerateCaptchaContext(ctx context.Context) (CaptchaChallenge, error) {
	return defaultCaptchaService.Challenge(ctx)
}

func VerifyCaptchaContext(ctx context.Context, id, answer string) error {
	return defaultCaptchaService.Verify(ctx, id, answer)
}

func LoginWithCaptchaContext(ctx context.Context, name, password, captchaID, captchaCode, addIP, device string) (*LoginResponse, error) {
	return loginWithCaptchaContext(ctx, VerifyCaptchaContext, LoginContext, name, password, captchaID, captchaCode, addIP, device)
}

func loginWithCaptchaContext(
	ctx context.Context,
	verify captchaVerifier,
	login passwordLogin,
	name, password, captchaID, captchaCode, addIP, device string,
) (*LoginResponse, error) {
	if err := verify(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	return login(ctx, name, password, addIP, device)
}
