package setup

import (
	"os"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/model"
)

func TestSetupCacheCopiesExpiresAndInvalidates(t *testing.T) {
	invalidateSetupServiceCache()
	now := time.Now()
	if _, ok := getSetupServiceCache("A", now); ok {
		t.Fatalf("empty setup cache should miss")
	}

	setSetupServiceCache(model.Setup{Key: "A", Value: "one"}, now)
	got, ok := getSetupServiceCache("A", now.Add(setupServiceCacheTTL/2))
	if !ok || got.Value != "one" {
		t.Fatalf("expected cached setup value before ttl, got setup=%#v ok=%v", got, ok)
	}
	got.Value = "changed"
	gotAgain, ok := getSetupServiceCache("A", now.Add(setupServiceCacheTTL/2))
	if !ok || gotAgain.Value != "one" {
		t.Fatalf("setup cache should return a defensive copy, got setup=%#v ok=%v", gotAgain, ok)
	}

	if _, ok := getSetupServiceCache("A", now.Add(setupServiceCacheTTL+time.Second)); ok {
		t.Fatalf("expired setup cache should miss")
	}

	setSetupServiceCache(model.Setup{Key: "A", Value: "two"}, now)
	invalidateSetupServiceCache()
	if _, ok := getSetupServiceCache("A", now); ok {
		t.Fatalf("invalidated setup cache should miss")
	}
}

func TestSetupCacheKeepsHotConfigLongEnough(t *testing.T) {
	if setupServiceCacheTTL < 5*time.Minute {
		t.Fatalf("setup cache ttl is too short for hot config reads: %s", setupServiceCacheTTL)
	}
}

func TestGetSetupUsesUnorderedLookup(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "Take(&setup)") {
		t.Fatalf("GetSetup hot lookup should use Take to avoid unnecessary ORDER BY id")
	}
	if strings.Contains(text, "GetSetupContext(ctx context.Context, key string) (*model.Setup, error) {\n") &&
		strings.Contains(text, ".First(&setup).Error") {
		t.Fatalf("GetSetup hot lookup should not use First because it emits ORDER BY id")
	}
}

func TestSetSetupInvalidatesStaticDomainCache(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`"wecheckin/backend/internal/app/support/media"`,
		`key == "STATIC_DOMAIN"`,
		"media.InvalidateStaticDomainCache()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("STATIC_DOMAIN updates must invalidate media cache with %s", snippet)
		}
	}
}
