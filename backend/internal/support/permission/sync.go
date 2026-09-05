package permission

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminmenuperm"
	"wecheckin/backend/internal/support/adminrouteperm"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
	"wecheckin/backend/pkg/database"
)

func ensureBuiltinPermissions(db *gorm.DB) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	items := []model.Permission{
		{Key: AdminLoginPermissionKey, Name: "后台入口", Platform: PlatformAdmin, Type: TypeLogin, Status: 1, Sort: 0, AddTime: now, EditTime: now},
		{Key: DataAllPermissionKey, Name: "全部数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 1, AddTime: now, EditTime: now},
		{Key: DataDeptPermissionKey, Name: "本部门数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 2, AddTime: now, EditTime: now},
		{Key: DataSelfPermissionKey, Name: "本人数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 3, AddTime: now, EditTime: now},
		{Key: DataCustomPermissionKey, Name: "自定义部门数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 4, AddTime: now, EditTime: now},
		{Key: DataExtraPermissionKey, Name: "用户额外数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 5, AddTime: now, EditTime: now},
	}
	for _, item := range items {
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func SyncAdminMenuPermissionsContext(ctx context.Context, db *gorm.DB, enableExam bool) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if err := EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	return syncAdminMenuPermissions(db.WithContext(ctx), enableExam)
}

func syncAdminMenuPermissions(db *gorm.DB, enableExam bool) error {
	declarations := adminmenuperm.Declarations(enableExam)
	now := database.Now()
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         adminMenuDeclarationType(declaration.Type),
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertAdminMenuPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func adminMenuDeclarationType(value string) string {
	switch value {
	case adminmenuperm.TypeDirectory:
		return TypeDirectory
	case adminmenuperm.TypeButton:
		return TypeButton
	default:
		return TypeMenu
	}
}

func firstBool(values []bool, fallback bool) bool {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func upsertAdminMenuPermission(db *gorm.DB, item model.Permission) error {
	return upsertPermission(db, item)
}

func syncAdminAPIPermissions(db *gorm.DB) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, category := range adminrouteperm.Categories() {
		item := model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: PlatformAdmin,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	for index, declaration := range adminrouteperm.Declarations() {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         (index + 1) * 10,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncClientMenuPermissions(db *gorm.DB) error {
	return syncApplicationMenuPermissions(db, appmenuperm.ClientMenuDeclarations())
}

func syncDingTalkH5MenuPermissions(db *gorm.DB) error {
	return syncApplicationMenuPermissions(db, appmenuperm.DingTalkH5MenuDeclarations())
}

func syncDingTalkH5ButtonPermissions(db *gorm.DB) error {
	return syncApplicationButtonPermissions(db, appmenuperm.DingTalkH5ButtonDeclarations())
}

func syncClientAPIPermissions(db *gorm.DB) error {
	return syncApplicationAPIPermissions(db, appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations())
}

func syncDingTalkH5APIPermissions(db *gorm.DB) error {
	return syncApplicationAPIPermissions(db, appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations())
}

func syncApplicationMenuPermissions(db *gorm.DB, declarations []appmenuperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, declaration := range declarations {
		permissionType := declaration.Type
		if permissionType == "" {
			permissionType = TypeMenu
		}
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         permissionType,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncApplicationButtonPermissions(db *gorm.DB, declarations []appmenuperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeButton,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncApplicationAPIPermissions(db *gorm.DB, categories []appapiperm.Category, declarations []appapiperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, category := range categories {
		item := model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: category.Platform,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func ensureMissingApplicationMenuPermissionsContext(ctx context.Context, db *gorm.DB, declarations []appmenuperm.Declaration) error {
	items := make([]model.Permission, 0, len(declarations))
	for _, declaration := range declarations {
		permissionType := declaration.Type
		if permissionType == "" {
			permissionType = TypeMenu
		}
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         permissionType,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingApplicationButtonPermissionsContext(ctx context.Context, db *gorm.DB, declarations []appmenuperm.Declaration) error {
	items := make([]model.Permission, 0, len(declarations))
	for _, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeButton,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingAdminAPIPermissionsContext(ctx context.Context, db *gorm.DB) error {
	now := database.Now()
	categories := adminrouteperm.Categories()
	declarations := adminrouteperm.Declarations()
	items := make([]model.Permission, 0, len(categories)+len(declarations))
	for _, category := range categories {
		items = append(items, model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: PlatformAdmin,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		})
	}
	for index, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         (index + 1) * 10,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingApplicationAPIPermissionsContext(ctx context.Context, db *gorm.DB, categories []appapiperm.Category, declarations []appapiperm.Declaration) error {
	items := make([]model.Permission, 0, len(categories)+len(declarations))
	for _, category := range categories {
		items = append(items, model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: category.Platform,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
		})
	}
	for _, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func createMissingPermissionsContext(ctx context.Context, db *gorm.DB, items []model.Permission) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	keys := make([]string, 0, len(items))
	byKey := make(map[string]model.Permission, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		item.Key = key
		keys = append(keys, key)
		byKey[key] = item
	}
	keys = normalizePermissionKeys(keys)
	if len(keys) == 0 {
		return nil
	}
	var existingKeys []string
	if err := db.Model(&model.Permission{}).Where("`permission_key` IN ?", keys).Pluck("permission_key", &existingKeys).Error; err != nil {
		return err
	}
	existing := make(map[string]bool, len(existingKeys))
	for _, key := range existingKeys {
		existing[key] = true
	}
	now := database.Now()
	createItems := make([]model.Permission, 0)
	for _, key := range keys {
		if existing[key] {
			continue
		}
		item := byKey[key]
		item.AddTime = now
		item.EditTime = now
		createItems = append(createItems, item)
	}
	if len(createItems) == 0 {
		return ctxErr(ctx)
	}
	if err := db.CreateInBatches(createItems, 100).Error; err != nil {
		return err
	}
	return ctxErr(ctx)
}

func upsertPermission(db *gorm.DB, item model.Permission) error {
	var current model.Permission
	err := db.Where("`permission_key` = ?", item.Key).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&item).Error
	}
	if err != nil {
		return err
	}
	icon := strings.TrimSpace(item.Icon)
	if currentIcon := strings.TrimSpace(current.Icon); currentIcon != "" {
		icon = currentIcon
	}
	return db.Model(&current).Updates(map[string]interface{}{
		"permission_name":          item.Name,
		"permission_platform":      item.Platform,
		"permission_type":          item.Type,
		"permission_parent_key":    item.ParentKey,
		"permission_resource_id":   item.ResourceID,
		"permission_resource_path": item.ResourcePath,
		"permission_icon":          icon,
		"permission_perms":         item.Perms,
		"permission_sort":          item.Sort,
		"permission_status":        item.Status,
		"permission_edit_time":     database.Now(),
	}).Error
}
