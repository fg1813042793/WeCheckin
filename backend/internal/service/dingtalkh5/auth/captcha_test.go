package auth

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/internal/support/logincaptcha"
)

func TestDingTalkH5CaptchaUsesIndependentConfiguration(t *testing.T) {
	if DingTalkH5LoginCaptchaSetupKey != "DINGTALK_H5_LOGIN_CAPTCHA_ENABLED" {
		t.Fatalf("setup key = %q", DingTalkH5LoginCaptchaSetupKey)
	}
	if dingTalkH5CaptchaRedisPrefix != "dingtalk_h5:login:captcha:" {
		t.Fatalf("redis prefix = %q", dingTalkH5CaptchaRedisPrefix)
	}
}

func TestLoginWithCaptchaStopsBeforePasswordLoginOnInvalidChallenge(t *testing.T) {
	verifyCalls := 0
	loginCalls := 0
	_, err := loginWithCaptchaContext(
		context.Background(),
		func(context.Context, string, string) error {
			verifyCalls++
			return logincaptcha.ErrInvalid
		},
		func(context.Context, string, string, string, string) (*LoginResponse, error) {
			loginCalls++
			return &LoginResponse{}, nil
		},
		"user", "password", "captcha-id", "ABCDE", "127.0.0.1", "UA",
	)
	if !errors.Is(err, logincaptcha.ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
	if verifyCalls != 1 || loginCalls != 0 {
		t.Fatalf("verify calls=%d login calls=%d", verifyCalls, loginCalls)
	}
}

func TestLoginWithCaptchaRunsPasswordLoginAfterVerification(t *testing.T) {
	order := make([]string, 0, 2)
	want := &LoginResponse{Token: "token-1"}
	got, err := loginWithCaptchaContext(
		context.Background(),
		func(context.Context, string, string) error {
			order = append(order, "verify")
			return nil
		},
		func(_ context.Context, name, password, addIP, device string) (*LoginResponse, error) {
			order = append(order, "login")
			if name != "user" || password != "password" || addIP != "127.0.0.1" || device != "UA" {
				t.Fatalf("login args = %q %q %q %q", name, password, addIP, device)
			}
			return want, nil
		},
		"user", "password", "captcha-id", "ABCDE", "127.0.0.1", "UA",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || len(order) != 2 || order[0] != "verify" || order[1] != "login" {
		t.Fatalf("result=%#v order=%#v", got, order)
	}
}
