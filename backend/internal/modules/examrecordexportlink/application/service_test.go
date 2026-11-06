package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

type storeStub struct {
	digest string
	grant  Grant
	ttl    time.Duration
	getErr error
}

func (store *storeStub) Put(_ context.Context, digest string, grant Grant, ttl time.Duration) error {
	store.digest, store.grant, store.ttl = digest, grant, ttl
	return nil
}

func (store *storeStub) Get(_ context.Context, digest string) (Grant, error) {
	if store.getErr != nil {
		return Grant{}, store.getErr
	}
	if digest != store.digest {
		return Grant{}, ErrGrantNotFound
	}
	return store.grant, nil
}

func TestServiceIssuesAndResolvesBoundGrant(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	token := strings.Repeat("ab", 32)
	store := &storeStub{}
	service := NewServiceWithDependencies(store, func() string { return token }, func() time.Time { return now })

	link, err := service.Issue(context.Background(), IssueInput{
		ExamID: 5, RecordID: 19, Format: "DOCX", ExpireDays: 7,
		BaseURL: "https://exam.example.com/ef1/5",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantLink := "https://exam.example.com/api/v2/public/exam-record-exports/download?token=" + token
	if link != wantLink {
		t.Fatalf("link = %q, want %q", link, wantLink)
	}
	sum := sha256.Sum256([]byte(token))
	if store.digest != hex.EncodeToString(sum[:]) {
		t.Fatalf("digest = %q", store.digest)
	}
	if store.ttl != 7*24*time.Hour || store.grant.ExamID != 5 || store.grant.RecordID != 19 || store.grant.Format != "docx" || !store.grant.ExpiresAt.Equal(now.Add(7*24*time.Hour)) {
		t.Fatalf("stored grant = %#v ttl=%v", store.grant, store.ttl)
	}
	resolved, err := service.Resolve(context.Background(), token)
	if err != nil || resolved != store.grant {
		t.Fatalf("resolved = %#v, err=%v", resolved, err)
	}
}

func TestServiceRejectsInvalidIssueInputs(t *testing.T) {
	service := NewServiceWithDependencies(&storeStub{}, func() string { return strings.Repeat("ab", 32) }, time.Now)
	base := IssueInput{ExamID: 1, RecordID: 2, Format: "pdf", ExpireDays: 7, BaseURL: "https://exam.example.com/ef1/1"}
	tests := []IssueInput{
		{RecordID: 2, Format: "pdf", ExpireDays: 7, BaseURL: base.BaseURL},
		{ExamID: 1, Format: "pdf", ExpireDays: 7, BaseURL: base.BaseURL},
		{ExamID: 1, RecordID: 2, Format: "zip", ExpireDays: 7, BaseURL: base.BaseURL},
		{ExamID: 1, RecordID: 2, Format: "pdf", ExpireDays: 0, BaseURL: base.BaseURL},
		{ExamID: 1, RecordID: 2, Format: "pdf", ExpireDays: 31, BaseURL: base.BaseURL},
		{ExamID: 1, RecordID: 2, Format: "pdf", ExpireDays: 7, BaseURL: "/ef1/1"},
		{ExamID: 1, RecordID: 2, Format: "pdf", ExpireDays: 7, BaseURL: "https://user:pass@example.com/ef1/1"},
	}
	for _, input := range tests {
		if _, err := service.Issue(context.Background(), input); !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("input %#v error = %v", input, err)
		}
	}
}

func TestServiceRejectsMissingAndExpiredTokens(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	token := strings.Repeat("cd", 32)
	store := &storeStub{getErr: ErrGrantNotFound}
	service := NewServiceWithDependencies(store, func() string { return token }, func() time.Time { return now })
	if _, err := service.Resolve(context.Background(), token); !errors.Is(err, ErrGrantNotFound) {
		t.Fatalf("missing error = %v", err)
	}
	sum := sha256.Sum256([]byte(token))
	store.digest = hex.EncodeToString(sum[:])
	store.grant = Grant{ExamID: 1, RecordID: 2, Format: "pdf", ExpiresAt: now.Add(-time.Second)}
	store.getErr = nil
	if _, err := service.Resolve(context.Background(), token); !errors.Is(err, ErrGrantExpired) {
		t.Fatalf("expired error = %v", err)
	}
}
