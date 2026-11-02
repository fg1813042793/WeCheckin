package logincaptcha

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	base64Captcha "github.com/mojocn/base64Captcha"
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

func TestServiceSkipsChallengeWhenDisabled(t *testing.T) {
	generator := &captchaGeneratorStub{}
	service := NewService(func(context.Context) (bool, error) { return false, nil }, generator, &captchaStoreStub{}, 2*time.Minute)
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

func TestServiceGeneratesAndConsumesCaseInsensitiveAnswer(t *testing.T) {
	store := &captchaStoreStub{}
	service := NewService(
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
	if err := service.Verify(context.Background(), "captcha-1", "a2bCd"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("replayed captcha error = %v", err)
	}
}

func TestServiceConsumesIncorrectAnswerAndFailsClosed(t *testing.T) {
	store := &captchaStoreStub{}
	service := NewService(
		func(context.Context) (bool, error) { return true, nil },
		&captchaGeneratorStub{id: "captcha-2", image: "image", answer: "ABCDE"},
		store,
		2*time.Minute,
	)
	if _, err := service.Challenge(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Verify(context.Background(), "captcha-2", "WRONG"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong answer error = %v", err)
	}
	if err := service.Verify(context.Background(), "captcha-2", "ABCDE"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong answer must consume challenge, got %v", err)
	}

	store.putErr = errors.New("redis unavailable")
	if _, err := service.Challenge(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("store failure error = %v", err)
	}
}

func TestImageGeneratorProducesExpectedShape(t *testing.T) {
	id, imageData, answer, err := NewImageGenerator().Generate()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 32 {
		t.Fatalf("captcha id length = %d", len(id))
	}
	if len(answer) != 5 {
		t.Fatalf("captcha answer = %q", answer)
	}
	for _, char := range answer {
		if !strings.ContainsRune("23456789ABCDEFGHJKLMNPQRSTUVWXYZ", char) {
			t.Fatalf("unexpected captcha character %q in %q", char, answer)
		}
	}
	if !strings.HasPrefix(imageData, "data:image/png;base64,") {
		t.Fatalf("captcha image prefix = %q", imageData[:min(len(imageData), 32)])
	}
}

func TestImageGeneratorUsesReadableCanvasWithoutReducingInterference(t *testing.T) {
	generator, ok := NewImageGenerator().(*imageGenerator)
	if !ok {
		t.Fatalf("image generator type = %T", NewImageGenerator())
	}
	if generator.driver.Height != 56 || generator.driver.Width != 180 {
		t.Fatalf("captcha canvas = %dx%d, want 180x56", generator.driver.Width, generator.driver.Height)
	}
	if generator.driver.NoiseCount != 2 {
		t.Fatalf("captcha noise count = %d, want 2", generator.driver.NoiseCount)
	}
	wantLines := base64Captcha.OptionShowSineLine | base64Captcha.OptionShowSlimeLine
	if generator.driver.ShowLineOptions != wantLines {
		t.Fatalf("captcha line options = %d, want %d", generator.driver.ShowLineOptions, wantLines)
	}
	if len(generator.driver.Fonts) != 1 || generator.driver.Fonts[0] != "RitaSmith.ttf" {
		t.Fatalf("captcha fonts = %#v, want RitaSmith.ttf", generator.driver.Fonts)
	}
}

func TestRedisStoreKeepsConfiguredNamespace(t *testing.T) {
	store, ok := NewRedisStore("dingtalk_h5:login:captcha:").(*redisStore)
	if !ok {
		t.Fatalf("redis store type = %T", NewRedisStore("prefix:"))
	}
	if store.prefix != "dingtalk_h5:login:captcha:" {
		t.Fatalf("redis prefix = %q", store.prefix)
	}
}
