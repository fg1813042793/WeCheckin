package adminauth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type captchaGeneratorStub struct {
	id, image, answer string
	err               error
	calls             int
}

func (stub *captchaGeneratorStub) Generate() (string, string, string, error) {
	stub.calls++
	return stub.id, stub.image, stub.answer, stub.err
}

type captchaStoreStub struct {
	values     map[string]string
	putErr     error
	consumeErr error
	putTTL     time.Duration
}

func (stub *captchaStoreStub) Put(_ context.Context, id, value string, ttl time.Duration) error {
	if stub.putErr != nil {
		return stub.putErr
	}
	if stub.values == nil {
		stub.values = map[string]string{}
	}
	stub.values[id] = value
	stub.putTTL = ttl
	return nil
}

func (stub *captchaStoreStub) Consume(_ context.Context, id string) (string, error) {
	if stub.consumeErr != nil {
		return "", stub.consumeErr
	}
	value := stub.values[id]
	delete(stub.values, id)
	return value, nil
}

func TestCaptchaServiceSkipsChallengeWhenDisabled(t *testing.T) {
	generator := &captchaGeneratorStub{}
	service := newCaptchaService(func(context.Context) (bool, error) { return false, nil }, generator, &captchaStoreStub{}, 2*time.Minute)
	challenge, err := service.Challenge(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if challenge.Enabled || challenge.ID != "" || challenge.ImageData != "" || generator.calls != 0 {
		t.Fatalf("disabled challenge = %#v calls=%d", challenge, generator.calls)
	}
	if err := service.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("disabled captcha verification = %v", err)
	}
}

func TestCaptchaServiceGeneratesAndConsumesCaseInsensitiveAnswer(t *testing.T) {
	store := &captchaStoreStub{}
	service := newCaptchaService(
		func(context.Context) (bool, error) { return true, nil },
		&captchaGeneratorStub{id: "captcha-1", image: "data:image/png;base64,abc", answer: "A2BcD"},
		store,
		2*time.Minute,
	)
	challenge, err := service.Challenge(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !challenge.Enabled || challenge.ID != "captcha-1" || challenge.ExpiresIn != 120 || store.putTTL != 2*time.Minute {
		t.Fatalf("challenge = %#v ttl=%v", challenge, store.putTTL)
	}
	if err := service.Verify(context.Background(), "captcha-1", "a2bCd"); err != nil {
		t.Fatalf("verify = %v", err)
	}
	if err := service.Verify(context.Background(), "captcha-1", "a2bCd"); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("replayed captcha error = %v", err)
	}
}

func TestCaptchaServiceConsumesIncorrectAnswerAndFailsClosed(t *testing.T) {
	store := &captchaStoreStub{}
	service := newCaptchaService(
		func(context.Context) (bool, error) { return true, nil },
		&captchaGeneratorStub{id: "captcha-2", image: "image", answer: "ABCDE"},
		store,
		2*time.Minute,
	)
	if _, err := service.Challenge(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Verify(context.Background(), "captcha-2", "WRONG"); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("wrong answer error = %v", err)
	}
	if err := service.Verify(context.Background(), "captcha-2", "ABCDE"); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("wrong answer must consume challenge, got %v", err)
	}

	store.putErr = errors.New("redis unavailable")
	if _, err := service.Challenge(context.Background()); !errors.Is(err, ErrCaptchaUnavailable) {
		t.Fatalf("store failure error = %v", err)
	}
}

