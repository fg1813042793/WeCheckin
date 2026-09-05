package permission

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
)

func EnsureUnifiedPermissionsContext(ctx context.Context, db *gorm.DB, enableExam ...bool) error {
	if err := EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	if err := ensureBuiltinPermissions(db); err != nil {
		return err
	}
	if err := SyncAdminMenuPermissionsContext(ctx, db, firstBool(enableExam, true)); err != nil {
		return err
	}
	if err := syncAdminAPIPermissions(db); err != nil {
		return err
	}
	if err := syncClientMenuPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5MenuPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5ButtonPermissions(db); err != nil {
		return err
	}
	if err := syncClientAPIPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5APIPermissions(db); err != nil {
		return err
	}
	return ctxErr(ctx)
}

func EnsureApplicationPermissionCatalogContext(ctx context.Context, db *gorm.DB, platform string, types []string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	platform = strings.TrimSpace(platform)
	typeSet := permissionTypeSet(types)
	shouldInclude := func(values ...string) bool {
		if len(typeSet) == 0 {
			return true
		}
		for _, value := range values {
			if typeSet[value] {
				return true
			}
		}
		return false
	}
	if platform == "" || platform == PlatformAdmin {
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingAdminAPIPermissionsContext(ctx, db); err != nil {
				return err
			}
		}
	}
	if platform == "" || platform == PlatformClient {
		if shouldInclude(TypeDirectory, TypeMenu, TypeButton) {
			if err := ensureMissingApplicationMenuPermissionsContext(ctx, db, appmenuperm.ClientMenuDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingApplicationAPIPermissionsContext(ctx, db, appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations()); err != nil {
				return err
			}
		}
	}
	if platform == "" || platform == PlatformDingTalkH5 {
		if shouldInclude(TypeDirectory, TypeMenu) {
			if err := ensureMissingApplicationMenuPermissionsContext(ctx, db, appmenuperm.DingTalkH5MenuDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeButton) {
			if err := ensureMissingApplicationButtonPermissionsContext(ctx, db, appmenuperm.DingTalkH5ButtonDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingApplicationAPIPermissionsContext(ctx, db, appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations()); err != nil {
				return err
			}
		}
	}
	return ctxErr(ctx)
}

func permissionTypeSet(types []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range types {
		for _, part := range strings.Split(value, ",") {
			item := strings.TrimSpace(part)
			if item != "" {
				result[item] = true
			}
		}
	}
	return result
}

func EnsurePermissionSchemaContext(ctx context.Context, db *gorm.DB) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if permissionTablesReadyCached() {
		return nil
	}
	if !db.Migrator().HasTable(&model.Permission{}) {
		markPermissionTablesReady(false)
		return ErrPermissionSchemaNotReady
	}
	if !db.Migrator().HasColumn(&model.Permission{}, "Icon") {
		markPermissionTablesReady(false)
		return ErrPermissionSchemaNotReady
	}
	if !db.Migrator().HasTable(&model.PermissionGrant{}) {
		markPermissionTablesReady(false)
		return ErrPermissionSchemaNotReady
	}
	markPermissionSchemaReady(true)
	return ctxErr(ctx)
}

func ResetPermissionTablesReadyCache() {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = false
	permissionTablesReadyCache.ready = false
	permissionTablesReadyCache.schemaReady = false
	permissionTablesReadyCache.checkedAt = time.Time{}
	permissionTablesReadyCache.Unlock()
	userRolesTableReadyCache.Lock()
	userRolesTableReadyCache.checked = false
	userRolesTableReadyCache.ready = false
	userRolesTableReadyCache.checkedAt = time.Time{}
	userRolesTableReadyCache.Unlock()
}

func markPermissionTablesReady(ready bool) {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = ready
	if !ready {
		permissionTablesReadyCache.schemaReady = false
	}
	permissionTablesReadyCache.checkedAt = time.Now()
	permissionTablesReadyCache.Unlock()
}

func permissionTablesReadyCached() bool {
	permissionTablesReadyCache.RLock()
	ready := permissionTablesReadyCache.checked && permissionTablesReadyCache.ready && permissionTablesReadyCache.schemaReady
	permissionTablesReadyCache.RUnlock()
	return ready
}

func markPermissionSchemaReady(ready bool) {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = ready
	permissionTablesReadyCache.schemaReady = ready
	permissionTablesReadyCache.checkedAt = time.Now()
	permissionTablesReadyCache.Unlock()
}

func TablesReady(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	permissionTablesReadyCache.RLock()
	if permissionTablesReadyCache.checked {
		ready := permissionTablesReadyCache.ready
		checkedAt := permissionTablesReadyCache.checkedAt
		permissionTablesReadyCache.RUnlock()
		if ready || time.Since(checkedAt) < permissionTablesReadyNegativeCacheTTL {
			return ready
		}
	} else {
		permissionTablesReadyCache.RUnlock()
	}

	permissionTablesReadyCache.Lock()
	defer permissionTablesReadyCache.Unlock()
	if permissionTablesReadyCache.checked {
		if permissionTablesReadyCache.ready || time.Since(permissionTablesReadyCache.checkedAt) < permissionTablesReadyNegativeCacheTTL {
			return permissionTablesReadyCache.ready
		}
	}
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = db.Migrator().HasTable(&model.Permission{}) && db.Migrator().HasTable(&model.PermissionGrant{})
	permissionTablesReadyCache.checkedAt = time.Now()
	return permissionTablesReadyCache.ready
}

func UserRolesTableReady(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	userRolesTableReadyCache.RLock()
	if userRolesTableReadyCache.checked {
		ready := userRolesTableReadyCache.ready
		checkedAt := userRolesTableReadyCache.checkedAt
		userRolesTableReadyCache.RUnlock()
		if ready || time.Since(checkedAt) < userRolesTableReadyNegativeCacheTTL {
			return ready
		}
	} else {
		userRolesTableReadyCache.RUnlock()
	}

	userRolesTableReadyCache.Lock()
	defer userRolesTableReadyCache.Unlock()
	if userRolesTableReadyCache.checked {
		if userRolesTableReadyCache.ready || time.Since(userRolesTableReadyCache.checkedAt) < userRolesTableReadyNegativeCacheTTL {
			return userRolesTableReadyCache.ready
		}
	}
	userRolesTableReadyCache.checked = true
	userRolesTableReadyCache.ready = db.Migrator().HasTable(&model.UserRole{})
	userRolesTableReadyCache.checkedAt = time.Now()
	return userRolesTableReadyCache.ready
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
