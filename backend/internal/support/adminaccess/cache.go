package adminaccess

import (
	"sync"
	"time"

	"wecheckin/backend/internal/model"
)

const roleAccessCacheTTL = time.Minute

type roleAccessCacheEntry struct {
	role      model.Role
	expiresAt time.Time
}

var roleAccessCache = struct {
	sync.RWMutex
	items map[uint]roleAccessCacheEntry
}{
	items: map[uint]roleAccessCacheEntry{},
}

func getRoleAccessCache(roleID uint, now time.Time) (model.Role, bool) {
	if roleID == 0 {
		return model.Role{}, false
	}
	roleAccessCache.RLock()
	entry, ok := roleAccessCache.items[roleID]
	roleAccessCache.RUnlock()
	if !ok || !now.Before(entry.expiresAt) {
		if ok {
			roleAccessCache.Lock()
			delete(roleAccessCache.items, roleID)
			roleAccessCache.Unlock()
		}
		return model.Role{}, false
	}
	return entry.role, true
}

func setRoleAccessCache(role model.Role, now time.Time) {
	if role.ID == 0 {
		return
	}
	roleAccessCache.Lock()
	roleAccessCache.items[role.ID] = roleAccessCacheEntry{
		role:      role,
		expiresAt: now.Add(roleAccessCacheTTL),
	}
	roleAccessCache.Unlock()
}

func InvalidateAdminAccessCache() {
	roleAccessCache.Lock()
	roleAccessCache.items = map[uint]roleAccessCacheEntry{}
	roleAccessCache.Unlock()
}

func InvalidateAdminAccessCacheForRole(roleID uint) {
	if roleID == 0 {
		return
	}
	roleAccessCache.Lock()
	delete(roleAccessCache.items, roleID)
	roleAccessCache.Unlock()
}
