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

	invalidateDictServiceCache()
	if _, ok := getDictItemsCache("gender", now.Add(dictServiceCacheTTL/2)); ok {
		t.Fatalf("invalidated dict item cache should miss")
	}
}
