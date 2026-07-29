package dict

import (
	"sync"
	"time"

	"wecheckin-backend/backend/internal/model"
)

const dictServiceCacheTTL = 30 * time.Second

type dictTypesCacheEntry struct {
	items     []TypeSummary
	expiresAt time.Time
}

type dictItemsCacheEntry struct {
	items     []model.SysDict
	expiresAt time.Time
}

var (
	dictServiceCacheMu sync.RWMutex
	dictTypesCache     dictTypesCacheEntry
	dictItemsCache     = map[string]dictItemsCacheEntry{}
)

func getDictTypesCache(now time.Time) ([]TypeSummary, bool) {
	dictServiceCacheMu.RLock()
	entry := dictTypesCache
	dictServiceCacheMu.RUnlock()
	if entry.expiresAt.IsZero() || now.After(entry.expiresAt) {
		return nil, false
	}
	return append([]TypeSummary(nil), entry.items...), true
}

func setDictTypesCache(items []TypeSummary, now time.Time) {
	dictServiceCacheMu.Lock()
	dictTypesCache = dictTypesCacheEntry{
		items:     append([]TypeSummary(nil), items...),
		expiresAt: now.Add(dictServiceCacheTTL),
	}
	dictServiceCacheMu.Unlock()
}

func getDictItemsCache(typeCode string, now time.Time) ([]model.SysDict, bool) {
	dictServiceCacheMu.RLock()
	entry, ok := dictItemsCache[typeCode]
	dictServiceCacheMu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, false
	}
	return append([]model.SysDict(nil), entry.items...), true
}

func setDictItemsCache(typeCode string, items []model.SysDict, now time.Time) {
	dictServiceCacheMu.Lock()
	dictItemsCache[typeCode] = dictItemsCacheEntry{
		items:     append([]model.SysDict(nil), items...),
		expiresAt: now.Add(dictServiceCacheTTL),
	}
	dictServiceCacheMu.Unlock()
}

func invalidateDictServiceCache() {
	dictServiceCacheMu.Lock()
	dictTypesCache = dictTypesCacheEntry{}
	dictItemsCache = map[string]dictItemsCacheEntry{}
	dictServiceCacheMu.Unlock()
}
