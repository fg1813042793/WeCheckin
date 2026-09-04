package dict

import (
	"testing"
	"time"

	"wecheckin/backend/internal/model"
)

func TestDictServiceCacheReturnsDefensiveCopiesAndExpires(t *testing.T) {
	now := time.Now()
	invalidateDictServiceCache()

	setDictTypesCache([]TypeSummary{{TypeCode: "gender", TypeName: "性别", ItemCnt: 2}}, now)
	types, ok := getDictTypesCache(now.Add(dictServiceCacheTTL / 2))
	if !ok || len(types) != 1 || types[0].TypeCode != "gender" {
		t.Fatalf("expected cached dict types, got %#v ok=%v", types, ok)
	}
	types[0].TypeCode = "mutated"
	typesAgain, ok := getDictTypesCache(now.Add(dictServiceCacheTTL / 2))
	if !ok || typesAgain[0].TypeCode != "gender" {
		t.Fatalf("dict type cache should return defensive copies, got %#v ok=%v", typesAgain, ok)
	}
	if _, ok := getDictTypesCache(now.Add(dictServiceCacheTTL + time.Second)); ok {
		t.Fatalf("expired dict types cache should miss")
	}
	setScopedDictTypesCache(true, []TypeSummary{{TypeCode: "active", Status: 1}}, now)
	activeTypes, ok := getScopedDictTypesCache(true, now.Add(time.Second))
	if !ok || len(activeTypes) != 1 || activeTypes[0].TypeCode != "active" {
		t.Fatalf("active dictionary type cache = %#v, %v", activeTypes, ok)
	}
	adminTypes, ok := getDictTypesCache(now.Add(time.Second))
	if !ok || len(adminTypes) != 1 || adminTypes[0].TypeCode != "gender" {
		t.Fatalf("admin and public dictionary type caches must be isolated: %#v, %v", adminTypes, ok)
	}

	setDictItemsCache("gender", []model.SysDict{{TypeCode: "gender", Label: "男"}}, now)
	items, ok := getDictItemsCache("gender", now.Add(dictServiceCacheTTL/2))
	if !ok || len(items) != 1 || items[0].Label != "男" {
		t.Fatalf("expected cached dict items, got %#v ok=%v", items, ok)
	}
	items[0].Label = "mutated"
	itemsAgain, ok := getDictItemsCache("gender", now.Add(dictServiceCacheTTL/2))
	if !ok || itemsAgain[0].Label != "男" {
		t.Fatalf("dict item cache should return defensive copies, got %#v ok=%v", itemsAgain, ok)
	}
	setScopedDictItemsCache("gender", true, []model.SysDict{{TypeCode: "gender", Label: "启用项"}}, now)
	activeItems, ok := getScopedDictItemsCache("gender", true, now.Add(time.Second))
	if !ok || len(activeItems) != 1 || activeItems[0].Label != "启用项" {
		t.Fatalf("active dictionary item cache = %#v, %v", activeItems, ok)
	}

	invalidateDictServiceCache()
	if _, ok := getDictItemsCache("gender", now.Add(dictServiceCacheTTL/2)); ok {
		t.Fatalf("invalidated dict item cache should miss")
	}
}
