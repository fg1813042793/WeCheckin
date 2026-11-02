package logincaptcha

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"math/big"
	"regexp"
	"strings"
	"time"

	base64Captcha "github.com/mojocn/base64Captcha"
	goredis "github.com/redis/go-redis/v9"

	rd "wecheckin/backend/pkg/redis"
)

const (
	defaultLength = 5
	defaultSource = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
)

var (
	ErrInvalid     = errors.New("验证码错误或已过期")
	ErrUnavailable = errors.New("验证码服务暂不可用")
	validID        = regexp.MustCompile(`^[A-Za-z0-9_-]{8,128}$`)
	consumeScript  = goredis.NewScript(`
local value = redis.call("GET", KEYS[1])
if value then
  redis.call("DEL", KEYS[1])
end
return value
`)
)

type Challenge struct {
	Enabled   bool   `json:"enabled"`
	ID        string `json:"captchaId,omitempty"`
	ImageData string `json:"imageData,omitempty"`
	ExpiresIn int    `json:"expiresIn,omitempty"`
}

type Generator interface {
	Generate() (id, imageData, answer string, err error)
}

type Store interface {
	Put(context.Context, string, string, time.Duration) error
	Consume(context.Context, string) (string, error)
}

type Service struct {
	enabled   func(context.Context) (bool, error)
	generator Generator
	store     Store
	ttl       time.Duration
}

func NewService(enabled func(context.Context) (bool, error), generator Generator, store Store, ttl time.Duration) *Service {
	return &Service{enabled: enabled, generator: generator, store: store, ttl: ttl}
}

func (service *Service) Challenge(ctx context.Context) (Challenge, error) {
	if service == nil || service.enabled == nil {
		return Challenge{}, ErrUnavailable
	}
	enabled, err := service.enabled(ctx)
	if err != nil {
		return Challenge{}, err
	}
	if !enabled {
		return Challenge{Enabled: false}, nil
	}
	if service.generator == nil || service.store == nil || service.ttl <= 0 {
		return Challenge{}, ErrUnavailable
	}
	id, imageData, answer, err := service.generator.Generate()
	if err != nil {
		return Challenge{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if !validID.MatchString(id) || strings.TrimSpace(imageData) == "" || strings.TrimSpace(answer) == "" {
		return Challenge{}, ErrUnavailable
	}
	if err := service.store.Put(ctx, id, digest(answer), service.ttl); err != nil {
		return Challenge{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return Challenge{
		Enabled: true, ID: id, ImageData: imageData, ExpiresIn: int(service.ttl.Seconds()),
	}, nil
}

func (service *Service) Verify(ctx context.Context, id, answer string) error {
	if service == nil || service.enabled == nil {
		return ErrUnavailable
	}
	enabled, err := service.enabled(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if !validID.MatchString(id) || answer == "" || len(answer) > 16 || service.store == nil {
		return ErrInvalid
	}
	stored, err := service.store.Consume(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if stored == "" || subtle.ConstantTimeCompare([]byte(stored), []byte(digest(answer))) != 1 {
		return ErrInvalid
	}
	return nil
}

type imageGenerator struct {
	driver *base64Captcha.DriverString
}

func NewImageGenerator() Generator {
	// Keep glyphs solid at login-card sizes; bundled decorative fonts can render hollow or striped text.
	return &imageGenerator{driver: base64Captcha.NewDriverString(
		56,
		180,
		2,
		base64Captcha.OptionShowSineLine|base64Captcha.OptionShowSlimeLine,
		defaultLength,
		defaultSource,
		&color.RGBA{R: 246, G: 248, B: 252, A: 255},
		nil,
		[]string{"RitaSmith.ttf"},
	)}
}

func (generator *imageGenerator) Generate() (string, string, string, error) {
	if generator == nil || generator.driver == nil {
		return "", "", "", ErrUnavailable
	}
	idBytes := make([]byte, 16)
	if _, err := cryptorand.Read(idBytes); err != nil {
		return "", "", "", err
	}
	answer, err := secureText(defaultLength, defaultSource)
	if err != nil {
		return "", "", "", err
	}
	item, err := generator.driver.DrawCaptcha(answer)
	if err != nil {
		return "", "", "", err
	}
	return hex.EncodeToString(idBytes), item.EncodeB64string(), answer, nil
}

func secureText(length int, source string) (string, error) {
	characters := []rune(source)
	if length <= 0 || len(characters) == 0 {
		return "", ErrUnavailable
	}
	result := make([]rune, length)
	limit := big.NewInt(int64(len(characters)))
	for index := range result {
		value, err := cryptorand.Int(cryptorand.Reader, limit)
		if err != nil {
			return "", err
		}
		result[index] = characters[value.Int64()]
	}
	return string(result), nil
}

type redisStore struct {
	prefix string
}

func NewRedisStore(prefix string) Store {
	return &redisStore{prefix: prefix}
}

func (store *redisStore) Put(ctx context.Context, id, value string, ttl time.Duration) error {
	if rd.RDB == nil {
		return ErrUnavailable
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	return rd.RDB.Set(redisCtx, store.prefix+id, value, ttl).Err()
}

func (store *redisStore) Consume(ctx context.Context, id string) (string, error) {
	if rd.RDB == nil {
		return "", ErrUnavailable
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	value, err := consumeScript.Run(redisCtx, rd.RDB, []string{store.prefix + id}).Text()
	if errors.Is(err, goredis.Nil) {
		return "", nil
	}
	return value, err
}

func digest(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
