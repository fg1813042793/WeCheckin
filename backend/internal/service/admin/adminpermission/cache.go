package adminpermission

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const permissionTreeCacheTTL = 30 * time.Second

type permissionTreeCacheEntry struct {
	items     []*PermissionNode
	expiresAt time.Time
}

var permissionTreeCache = struct {
	sync.RWMutex
	items map[string]permissionTreeCacheEntry
}{
	items: map[string]permissionTreeCacheEntry{},
}

func permissionTreeCacheKey(platform string, types []string) string {
	normalized := append([]string(nil), types...)
	sort.Strings(normalized)
	return strings.TrimSpace(platform) + "|" + strings.Join(normalized, ",")
}

func getPermissionTreeCache(platform string, types []string, now time.Time) ([]*PermissionNode, bool) {
	key := permissionTreeCacheKey(platform, types)
	permissionTreeCache.RLock()
	entry, ok := permissionTreeCache.items[key]
	permissionTreeCache.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, false
	}
	return clonePermissionNodes(entry.items), true
}

func setPermissionTreeCache(platform string, types []string, items []*PermissionNode, now time.Time) {
	key := permissionTreeCacheKey(platform, types)
	permissionTreeCache.Lock()
	permissionTreeCache.items[key] = permissionTreeCacheEntry{
		items:     clonePermissionNodes(items),
		expiresAt: now.Add(permissionTreeCacheTTL),
	}
	permissionTreeCache.Unlock()
}

func invalidatePermissionTreeCache() {
	permissionTreeCache.Lock()
	permissionTreeCache.items = map[string]permissionTreeCacheEntry{}
	permissionTreeCache.Unlock()
}

func clonePermissionNodes(items []*PermissionNode) []*PermissionNode {
	if len(items) == 0 {
		return nil
	}
	result := make([]*PermissionNode, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		copied := *item
		copied.Children = clonePermissionNodes(item.Children)
		result = append(result, &copied)
	}
	return result
}
