package adminuser

import (
	"sync"
	"time"
)

const userTotalCountCacheTTL = 30 * time.Second

var userTotalCountCache = struct {
	sync.RWMutex
	total     int64
	expiresAt time.Time
}{}

func getUserTotalCountCache(now time.Time) (int64, bool) {
	userTotalCountCache.RLock()
	defer userTotalCountCache.RUnlock()
	if userTotalCountCache.expiresAt.IsZero() || !now.Before(userTotalCountCache.expiresAt) {
		return 0, false
	}
	return userTotalCountCache.total, true
}

func setUserTotalCountCache(total int64, now time.Time) {
	userTotalCountCache.Lock()
	defer userTotalCountCache.Unlock()
	userTotalCountCache.total = total
	userTotalCountCache.expiresAt = now.Add(userTotalCountCacheTTL)
}

func invalidateUserTotalCountCache() {
	userTotalCountCache.Lock()
	defer userTotalCountCache.Unlock()
	userTotalCountCache.total = 0
	userTotalCountCache.expiresAt = time.Time{}
}
