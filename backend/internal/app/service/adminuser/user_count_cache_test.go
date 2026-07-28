package adminuser

import (
	"testing"
	"time"
)

func TestUserTotalCountCacheExpiresAndInvalidates(t *testing.T) {
	invalidateUserTotalCountCache()
	now := time.Now()
	if _, ok := getUserTotalCountCache(now); ok {
		t.Fatalf("empty cache should miss")
	}

	setUserTotalCountCache(42, now)
	if total, ok := getUserTotalCountCache(now.Add(userTotalCountCacheTTL / 2)); !ok || total != 42 {
		t.Fatalf("expected cached total 42 before ttl, got total=%d ok=%v", total, ok)
	}
	if _, ok := getUserTotalCountCache(now.Add(userTotalCountCacheTTL + time.Second)); ok {
		t.Fatalf("expired cache should miss")
	}

	setUserTotalCountCache(7, now)
	invalidateUserTotalCountCache()
	if _, ok := getUserTotalCountCache(now); ok {
		t.Fatalf("invalidated cache should miss")
	}
}
