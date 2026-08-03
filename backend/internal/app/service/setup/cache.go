package setup

import (
	"sync"
	"time"

	"wecheckin/backend/internal/model"
)

const setupServiceCacheTTL = 5 * time.Minute

type setupServiceCacheEntry struct {
	setup     model.Setup
	expiresAt time.Time
}

var (
	setupServiceCacheMu sync.RWMutex
	setupServiceCache   = map[string]setupServiceCacheEntry{}
)

func getSetupServiceCache(key string, now time.Time) (*model.Setup, bool) {
	setupServiceCacheMu.RLock()
	entry, ok := setupServiceCache[key]
	setupServiceCacheMu.RUnlock()
	if !ok || !now.Before(entry.expiresAt) {
		return nil, false
	}
	setupCopy := entry.setup
	return &setupCopy, true
}

func setSetupServiceCache(setup model.Setup, now time.Time) {
	setupServiceCacheMu.Lock()
	setupServiceCache[setup.Key] = setupServiceCacheEntry{
		setup:     setup,
		expiresAt: now.Add(setupServiceCacheTTL),
	}
	setupServiceCacheMu.Unlock()
}

func invalidateSetupServiceCache() {
	setupServiceCacheMu.Lock()
	setupServiceCache = map[string]setupServiceCacheEntry{}
	setupServiceCacheMu.Unlock()
}
