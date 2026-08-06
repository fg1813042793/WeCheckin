package adminaccess

import (
	"testing"
	"time"

	"wecheckin/backend/internal/model"
)

func TestRoleAccessCacheCopiesExpiresAndInvalidates(t *testing.T) {
	InvalidateAdminAccessCache()
	now := time.Now()
	if _, ok := getRoleAccessCache(7, now); ok {
		t.Fatalf("empty role access cache should miss")
	}

	setRoleAccessCache(model.Role{ID: 7, Name: ReservedSuperAdminRoleName, Status: 1}, now)
	got, ok := getRoleAccessCache(7, now.Add(roleAccessCacheTTL/2))
	if !ok || got.Name != ReservedSuperAdminRoleName {
		t.Fatalf("expected cached role before ttl, got role=%#v ok=%v", got, ok)
	}
	got.Name = "changed"
	gotAgain, ok := getRoleAccessCache(7, now.Add(roleAccessCacheTTL/2))
	if !ok || gotAgain.Name != ReservedSuperAdminRoleName {
		t.Fatalf("role access cache should return a defensive copy, got role=%#v ok=%v", gotAgain, ok)
	}

	if _, ok := getRoleAccessCache(7, now.Add(roleAccessCacheTTL+time.Second)); ok {
		t.Fatalf("expired role access cache should miss")
	}

	setRoleAccessCache(model.Role{ID: 7, Name: ReservedSuperAdminRoleName, Status: 1}, now)
	InvalidateAdminAccessCacheForRole(7)
	if _, ok := getRoleAccessCache(7, now); ok {
		t.Fatalf("role cache should miss after role invalidation")
	}
}
