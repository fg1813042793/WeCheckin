package dict

import (
	"sync"
	"time"

	"wecheckin/backend/internal/model"
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
	dictTypesCache     = map[bool]dictTypesCacheEntry{}
	dictItemsCache     = map[string]dictItemsCacheEntry{}
)

func getDictTypesCache(now time.Time) ([]TypeSummary, bool) {
	return getScopedDictTypesCache(false, now)
}

func getScopedDictTypesCache(activeOnly bool, now time.Time) ([]TypeSummary, bool) {
	dictServiceCacheMu.RLock()
	entry, ok := dictTypesCache[activeOnly]
	dictServiceCacheMu.RUnlock()
	if !ok || entry.expiresAt.IsZero() || now.After(entry.expiresAt) {
		return nil, false
	}
	return append([]TypeSummary(nil), entry.items...), true
}

func setDictTypesCache(items []TypeSummary, now time.Time) {
	setScopedDictTypesCache(false, items, now)
}

func setScopedDictTypesCache(activeOnly bool, items []TypeSummary, now time.Time) {
	dictServiceCacheMu.Lock()
	dictTypesCache[activeOnly] = dictTypesCacheEntry{
		items:     append([]TypeSummary(nil), items...),
		expiresAt: now.Add(dictServiceCacheTTL),
	}
	dictServiceCacheMu.Unlock()
}

func getDictItemsCache(typeCode string, now time.Time) ([]model.SysDict, bool) {
	return getScopedDictItemsCache(typeCode, false, now)
}

func dictItemsCacheKey(typeCode string, activeOnly bool) string {
	if activeOnly {
		return "active:" + typeCode
	}
	return "all:" + typeCode
}

func getScopedDictItemsCache(typeCode string, activeOnly bool, now time.Time) ([]model.SysDict, bool) {
	dictServiceCacheMu.RLock()
	entry, ok := dictItemsCache[dictItemsCacheKey(typeCode, activeOnly)]
	dictServiceCacheMu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, false
	}
	return append([]model.SysDict(nil), entry.items...), true
}

func setDictItemsCache(typeCode string, items []model.SysDict, now time.Time) {
	setScopedDictItemsCache(typeCode, false, items, now)
}

func setScopedDictItemsCache(typeCode string, activeOnly bool, items []model.SysDict, now time.Time) {
	dictServiceCacheMu.Lock()
	dictItemsCache[dictItemsCacheKey(typeCode, activeOnly)] = dictItemsCacheEntry{
		items:     append([]model.SysDict(nil), items...),
		expiresAt: now.Add(dictServiceCacheTTL),
	}
	dictServiceCacheMu.Unlock()
}

func invalidateDictServiceCache() {
	dictServiceCacheMu.Lock()
	dictTypesCache = map[bool]dictTypesCacheEntry{}
	dictItemsCache = map[string]dictItemsCacheEntry{}
	dictServiceCacheMu.Unlock()
}
