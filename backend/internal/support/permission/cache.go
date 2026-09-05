package permission

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"sort"
	"strconv"
	"strings"
	"time"
	"wecheckin/backend/internal/support/appmenuperm"
)

func ClientMenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return SubjectMenuPermissionKeysContext(ctx, db, userID, roleID, PlatformClient)
}

func DingTalkH5MenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID})
}

func DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) ([]string, bool, error) {
	roleIDs = normalizeRoleIDs(roleIDs...)
	if keys, ready, ok := getDingTalkH5MenuPermissionCache(userID, roleIDs); ok {
		return keys, ready, nil
	}
	keys, ready, err := SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, userID, roleIDs, PlatformDingTalkH5)
	if err == nil {
		setDingTalkH5MenuPermissionCache(userID, roleIDs, keys, ready)
	}
	return keys, ready, err
}

func DingTalkH5ButtonPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID})
}

func DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) ([]string, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{"dingtalk_h5:button:%"})
	if err != nil {
		return nil, true, err
	}
	selected := make(map[string]bool, len(allowed))
	for key := range allowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	return orderedApplicationMenuKeys(selected, appmenuperm.DingTalkH5ButtonDeclarations()), true, nil
}

func dingtalkH5MenuPermissionCacheKey(userID uint, roleIDs []uint) string {
	roleIDs = normalizeRoleIDs(roleIDs...)
	roleParts := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleParts = append(roleParts, strconv.FormatUint(uint64(roleID), 10))
	}
	return strconv.FormatUint(uint64(userID), 10) + ":" + strings.Join(roleParts, ",")
}

func getDingTalkH5MenuPermissionCache(userID uint, roleIDs []uint) ([]string, bool, bool) {
	key := dingtalkH5MenuPermissionCacheKey(userID, roleIDs)
	now := time.Now()
	dingtalkH5MenuPermissionCache.RLock()
	entry, ok := dingtalkH5MenuPermissionCache.items[key]
	dingtalkH5MenuPermissionCache.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, false, false
	}
	return append([]string(nil), entry.keys...), entry.ready, true
}

func setDingTalkH5MenuPermissionCache(userID uint, roleIDs []uint, keys []string, ready bool) {
	key := dingtalkH5MenuPermissionCacheKey(userID, roleIDs)
	dingtalkH5MenuPermissionCache.Lock()
	dingtalkH5MenuPermissionCache.items[key] = dingtalkH5MenuPermissionCacheEntry{
		keys:      append([]string(nil), keys...),
		ready:     ready,
		expiresAt: time.Now().Add(dingtalkH5MenuPermissionCacheTTL),
	}
	dingtalkH5MenuPermissionCache.Unlock()
}

func invalidateDingTalkH5MenuPermissionCache() {
	dingtalkH5MenuPermissionCache.Lock()
	dingtalkH5MenuPermissionCache.items = map[string]dingtalkH5MenuPermissionCacheEntry{}
	dingtalkH5MenuPermissionCache.Unlock()
}

func InvalidateRuntimePermissionCaches() {
	invalidateSubjectPermissionSetCache()
	invalidateDingTalkH5MenuPermissionCache()
}

func subjectPermissionSetCacheKey(userID uint, roleIDs []uint, prefixes []string) string {
	roleIDs = normalizeRoleIDs(roleIDs...)
	roleParts := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleParts = append(roleParts, strconv.FormatUint(uint64(roleID), 10))
	}
	normalized := normalizePermissionKeys(prefixes)
	sort.Strings(normalized)
	return strconv.FormatUint(uint64(userID), 10) + ":" +
		strings.Join(roleParts, ",") + ":" +
		strings.Join(normalized, "\x1f")
}

func getSubjectPermissionSetCache(userID uint, roleIDs []uint, prefixes []string) (map[string]bool, map[string]bool, bool) {
	key := subjectPermissionSetCacheKey(userID, roleIDs, prefixes)
	now := time.Now()
	subjectPermissionSetCache.RLock()
	entry, ok := subjectPermissionSetCache.items[key]
	subjectPermissionSetCache.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, nil, false
	}
	return copyBoolMap(entry.allowed), copyBoolMap(entry.denied), true
}

func setSubjectPermissionSetCache(userID uint, roleIDs []uint, prefixes []string, allowed, denied map[string]bool) {
	key := subjectPermissionSetCacheKey(userID, roleIDs, prefixes)
	subjectPermissionSetCache.Lock()
	subjectPermissionSetCache.items[key] = subjectPermissionSetCacheEntry{
		allowed:   copyBoolMap(allowed),
		denied:    copyBoolMap(denied),
		expiresAt: time.Now().Add(subjectPermissionSetCacheTTL),
	}
	subjectPermissionSetCache.Unlock()
}

func invalidateSubjectPermissionSetCache() {
	subjectPermissionSetCache.Lock()
	subjectPermissionSetCache.items = map[string]subjectPermissionSetCacheEntry{}
	subjectPermissionSetCache.Unlock()
}

func copyBoolMap(source map[string]bool) map[string]bool {
	if len(source) == 0 {
		return map[string]bool{}
	}
	copied := make(map[string]bool, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
