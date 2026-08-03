package menu

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type AdminMenu struct {
	ID        string       `json:"id"`
	Key       string       `json:"key"`
	Name      string       `json:"name"`
	ParentKey string       `json:"parentKey"`
	Path      string       `json:"path"`
	Perms     string       `json:"perms"`
	Icon      string       `json:"icon"`
	Sort      int          `json:"sort"`
	Status    int          `json:"status"`
	Type      int          `json:"type"`
	AddTime   int64        `json:"addTime"`
	EditTime  int64        `json:"editTime"`
	Children  []*AdminMenu `json:"children,omitempty"`
}

func GetAdminMenuTree(admin *model.Admin) ([]*AdminMenu, error) {
	return GetAdminMenuTreeContext(context.Background(), admin)
}

func GetAdminMenuTreeContext(ctx context.Context, admin *model.Admin) ([]*AdminMenu, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if !adminRoleAllowsAdminAccess(ctx, db, admin) {
		return []*AdminMenu{}, nil
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		rows, err := allAdminMenuPermissionsContext(ctx, db)
		if err != nil {
			return nil, err
		}
		return permissionsToMenuTree(rows), nil
	}
	if admin.RoleID == 0 {
		return []*AdminMenu{}, nil
	}
	rows, ready, err := permissionsupport.AdminMenuPermissionsContext(ctx, db, admin.ID, admin.RoleID)
	if err != nil {
		return nil, err
	}
	if ready {
		return permissionsToMenuTree(rows), nil
	}
	return []*AdminMenu{}, nil
}

func GetAdminPerms(admin *model.Admin) []string {
	return GetAdminPermsContext(context.Background(), admin)
}

func AdminHasReservedSuperAdminRoleContext(ctx context.Context, admin *model.Admin) bool {
	if admin == nil || admin.RoleID == 0 {
		return false
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID)
}

func GetAdminPermsContext(ctx context.Context, admin *model.Admin) []string {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if !adminRoleAllowsAdminAccess(ctx, db, admin) {
		return nil
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		cacheKey := adminPermCacheKeyForSuperAdmin()
		if perms, ok := getAdminPermCache(cacheKey); ok {
			return perms
		}
		rows, err := allAdminPermissionsContext(ctx, db)
		if err != nil {
			return nil
		}
		perms := permissionKeys(rows)
		setAdminPermCache(cacheKey, perms)
		return perms
	}
	if admin.RoleID == 0 {
		return nil
	}
	cacheKey := adminPermCacheKeyForUserRole(admin.ID, admin.RoleID)
	if perms, ok := getAdminPermCache(cacheKey); ok {
		return perms
	}
	if perms, ready, err := permissionsupport.AdminPermissionKeysContext(ctx, db, admin.ID, admin.RoleID); err == nil && ready {
		setAdminPermCache(cacheKey, perms)
		return perms
	}
	return nil
}

func adminRoleAllowsAdminAccess(ctx context.Context, db *gorm.DB, admin *model.Admin) bool {
	if admin == nil || admin.RoleID == 0 {
		return false
	}
	_, err := adminaccess.UserAllowsAdminAccessContext(ctx, db, admin.ID, admin.RoleID)
	return err == nil
}

func allAdminMenuPermissionsContext(ctx context.Context, db *gorm.DB) ([]model.Permission, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || !permissionsupport.TablesReady(db) {
		return nil, nil
	}
	var rows []model.Permission
	err := db.Where("`permission_platform` = ? AND `permission_type` IN ? AND `permission_status` = 1", permissionsupport.PlatformAdmin, []string{permissionsupport.TypeDirectory, permissionsupport.TypeMenu, permissionsupport.TypeButton}).
		Order("`permission_sort` ASC, `id` ASC").
		Find(&rows).Error
	return rows, err
}

func allAdminPermissionsContext(ctx context.Context, db *gorm.DB) ([]model.Permission, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || !permissionsupport.TablesReady(db) {
		return nil, nil
	}
	var rows []model.Permission
	err := db.Select("`permission_key`").
		Where("`permission_platform` = ? AND `permission_status` = 1", permissionsupport.PlatformAdmin).
		Find(&rows).Error
	return rows, err
}

func permissionsToMenuTree(rows []model.Permission) []*AdminMenu {
	return buildAdminMenuTree(permissionRowsToMenus(rows))
}

func permissionRowsToMenus(rows []model.Permission) []*AdminMenu {
	list := make([]*AdminMenu, 0, len(rows))
	for _, row := range rows {
		list = append(list, &AdminMenu{
			ID:        row.Key,
			Key:       row.Key,
			Name:      row.Name,
			ParentKey: row.ParentKey,
			Path:      row.ResourcePath,
			Perms:     row.Perms,
			Icon:      row.Icon,
			Sort:      row.Sort,
			Status:    row.Status,
			Type:      permissionTypeToMenuType(row.Type),
			AddTime:   row.AddTime,
			EditTime:  row.EditTime,
		})
	}
	return list
}

func buildAdminMenuTree(list []*AdminMenu) []*AdminMenu {
	byKey := make(map[string]*AdminMenu, len(list))
	for _, item := range list {
		item.Children = nil
		byKey[item.Key] = item
	}
	tree := make([]*AdminMenu, 0)
	for _, item := range list {
		parent := byKey[item.ParentKey]
		if parent == nil || parent.Key == item.Key {
			tree = append(tree, item)
			continue
		}
		parent.Children = append(parent.Children, item)
	}
	return tree
}

func permissionTypeToMenuType(permissionType string) int {
	switch permissionType {
	case permissionsupport.TypeDirectory:
		return 0
	case permissionsupport.TypeButton:
		return 2
	default:
		return 1
	}
}

func permissionKeys(rows []model.Permission) []string {
	seen := map[string]bool{}
	perms := make([]string, 0)
	for _, row := range rows {
		if row.Key != "" && !seen[row.Key] {
			seen[row.Key] = true
			perms = append(perms, row.Key)
		}
	}
	return perms
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
