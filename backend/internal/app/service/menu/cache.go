package menu

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const adminPermCacheTTL = time.Minute

type adminPermCacheEntry struct {
	perms     []string
	expiresAt time.Time
}

var adminPermCache = struct {
	sync.RWMutex
	values  map[string]adminPermCacheEntry
	version uint64
}{
	values: map[string]adminPermCacheEntry{},
}

func adminPermCacheKeyForRole(roleID uint) string {
	return fmt.Sprintf("role:%d", roleID)
}

func adminPermCacheKeyForUserRole(userID, roleID uint) string {
	return fmt.Sprintf("user:%d:role:%d", userID, roleID)
}

func adminPermCacheKeyForSuperAdmin() string {
	return "super"
}

func getAdminPermCache(key string) ([]string, bool) {
	adminPermCache.RLock()
	entry, ok := adminPermCache.values[key]
	adminPermCache.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			adminPermCache.Lock()
			delete(adminPermCache.values, key)
			adminPermCache.Unlock()
		}
		return nil, false
	}
	return append([]string(nil), entry.perms...), true
}

func setAdminPermCache(key string, perms []string) {
	adminPermCache.Lock()
	adminPermCache.values[key] = adminPermCacheEntry{
		perms:     append([]string(nil), perms...),
		expiresAt: time.Now().Add(adminPermCacheTTL),
	}
	adminPermCache.Unlock()
}

func AdminPermCacheVersion() uint64 {
	adminPermCache.RLock()
	version := adminPermCache.version
	adminPermCache.RUnlock()
	return version
}

func InvalidateAdminPermCache() {
	adminPermCache.Lock()
	adminPermCache.values = map[string]adminPermCacheEntry{}
	adminPermCache.version++
	adminPermCache.Unlock()
}

func InvalidateAdminPermCacheForRole(roleID uint) {
	adminPermCache.Lock()
	delete(adminPermCache.values, adminPermCacheKeyForRole(roleID))
	for key := range adminPermCache.values {
		if strings.HasSuffix(key, fmt.Sprintf(":role:%d", roleID)) {
			delete(adminPermCache.values, key)
		}
	}
	adminPermCache.version++
	adminPermCache.Unlock()
}
