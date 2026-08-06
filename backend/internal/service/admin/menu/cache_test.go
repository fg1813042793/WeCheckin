package menu

import "testing"

func TestAdminPermCacheCopiesValuesAndInvalidatesRole(t *testing.T) {
	InvalidateAdminPermCache()
	setAdminPermCache("role:7", []string{"survey:list"})

	perms, ok := getAdminPermCache("role:7")
	if !ok {
		t.Fatalf("expected cached permissions")
	}
	perms[0] = "mutated"

	perms, ok = getAdminPermCache("role:7")
	if !ok || perms[0] != "survey:list" {
		t.Fatalf("cache should return a defensive copy, got %#v", perms)
	}

	InvalidateAdminPermCacheForRole(7)
	if _, ok := getAdminPermCache("role:7"); ok {
		t.Fatalf("role cache should be invalidated")
	}
}

func TestAdminPermCacheInvalidationBumpsVersion(t *testing.T) {
	InvalidateAdminPermCache()
	start := AdminPermCacheVersion()

	InvalidateAdminPermCacheForRole(7)
	afterRole := AdminPermCacheVersion()
	if afterRole <= start {
		t.Fatalf("role invalidation should bump cache version, start=%d after=%d", start, afterRole)
	}

	InvalidateAdminPermCache()
	afterAll := AdminPermCacheVersion()
	if afterAll <= afterRole {
		t.Fatalf("full invalidation should bump cache version, afterRole=%d afterAll=%d", afterRole, afterAll)
	}
}

func TestAdminPermCacheKeySeparatesUserOverrides(t *testing.T) {
	roleOnly := adminPermCacheKeyForRole(7)
	userA := adminPermCacheKeyForUserRole(11, 7)
	userB := adminPermCacheKeyForUserRole(12, 7)
	if roleOnly == userA {
		t.Fatalf("user override cache key must not reuse role-only key")
	}
	if userA == userB {
		t.Fatalf("different users with the same role must have different permission cache keys")
	}
}
