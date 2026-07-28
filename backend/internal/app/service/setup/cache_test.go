package setup

import (
	"testing"
	"time"

	"wecheckin-backend/backend/internal/model"
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
