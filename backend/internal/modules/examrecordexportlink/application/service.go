package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"wecheckin/backend/pkg/randutil"
)

const DownloadPath = "/api/v2/public/exam-record-exports/download"

var (
	ErrInvalidGrant  = errors.New("invalid exam record export grant")
	ErrGrantNotFound = errors.New("exam record export grant not found")
	ErrGrantExpired  = errors.New("exam record export grant expired")
	ErrStoreMissing  = errors.New("exam record export grant store is unavailable")
)

type Grant struct {
	ExamID    uint      `json:"examId"`
	RecordID  uint      `json:"recordId"`
	Format    string    `json:"format"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type IssueInput struct {
	ExamID     uint
	RecordID   uint
	Format     string
	ExpireDays int
	BaseURL    string
}

type Store interface {
	Put(context.Context, string, Grant, time.Duration) error
	Get(context.Context, string) (Grant, error)
}

type Service struct {
	store          Store
	tokenGenerator func() string
	now            func() time.Time
}

func NewService(store Store) *Service {
	return NewServiceWithDependencies(store, func() string { return randutil.HexString(64) }, time.Now)
}

func NewServiceWithDependencies(store Store, tokenGenerator func() string, now func() time.Time) *Service {
	return &Service{store: store, tokenGenerator: tokenGenerator, now: now}
}

func (service *Service) Issue(ctx context.Context, input IssueInput) (string, error) {
	format := strings.ToLower(strings.TrimSpace(input.Format))
	publicURL := strings.TrimSpace(input.BaseURL)
	parsed, err := url.Parse(publicURL)
	if input.ExamID == 0 || input.RecordID == 0 || !validFormat(format) || input.ExpireDays < 1 || input.ExpireDays > 30 || err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
		return "", ErrInvalidGrant
	}
	if service == nil || service.store == nil || service.tokenGenerator == nil || service.now == nil {
		return "", ErrStoreMissing
	}
	token := strings.TrimSpace(service.tokenGenerator())
	if len(token) < 64 {
		return "", ErrInvalidGrant
	}
	ttl := time.Duration(input.ExpireDays) * 24 * time.Hour
	grant := Grant{ExamID: input.ExamID, RecordID: input.RecordID, Format: format, ExpiresAt: service.now().Add(ttl)}
	if err := service.store.Put(ctx, tokenDigest(token), grant, ttl); err != nil {
		return "", err
	}
	query := url.Values{"token": []string{token}}
	return parsed.Scheme + "://" + parsed.Host + DownloadPath + "?" + query.Encode(), nil
}

func (service *Service) Resolve(ctx context.Context, token string) (Grant, error) {
	if service == nil || service.store == nil || service.now == nil {
		return Grant{}, ErrStoreMissing
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return Grant{}, ErrGrantNotFound
	}
	grant, err := service.store.Get(ctx, tokenDigest(token))
	if err != nil {
		return Grant{}, err
	}
	if grant.ExamID == 0 || grant.RecordID == 0 || !validFormat(grant.Format) || grant.ExpiresAt.IsZero() {
		return Grant{}, ErrGrantNotFound
	}
	if !service.now().Before(grant.ExpiresAt) {
		return Grant{}, ErrGrantExpired
	}
	return grant, nil
}

func validFormat(format string) bool {
	return format == "pdf" || format == "docx" || format == "xlsx"
}

func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
