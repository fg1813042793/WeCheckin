package passport

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appmenuperm"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type BootstrapResponse struct {
	User                model.User      `json:"user"`
	Domain              string          `json:"domain"`
	Menus               []ClientMenuDTO `json:"menus"`
	MenuPermissionKeys  []string        `json:"menuPermissionKeys"`
	MenuPermissionReady bool            `json:"menuPermissionReady"`
	APIPermissionKeys   []string        `json:"apiPermissionKeys"`
	APIPermissionReady  bool            `json:"apiPermissionReady"`
	PermissionVersion   int64           `json:"permissionVersion"`
}

type ClientMenuDTO struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Path          string `json:"path"`
	PermissionKey string `json:"permissionKey"`
	Sort          int    `json:"sort"`
}

type clientPermissionCatalogRow struct {
	Key      string `gorm:"column:permission_key"`
	Name     string `gorm:"column:permission_name"`
	EditTime int64  `gorm:"column:permission_edit_time"`
}

func BootstrapContext(ctx context.Context, userID string) (*BootstrapResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	if err := db.WithContext(ctx).Where("`user_mini_openid` = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	fillClientUserDepartmentContext(ctx, db, &user)
	return bootstrapForUserDB(ctx, db, &user), nil
}

func BootstrapByIDContext(ctx context.Context, id uint) (*BootstrapResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	if err := db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	fillClientUserDepartmentContext(ctx, db, &user)
	return bootstrapForUserDB(ctx, db, &user), nil
}

func bootstrapForUserDB(ctx context.Context, db *gorm.DB, user *model.User) *BootstrapResponse {
	if user == nil {
		return &BootstrapResponse{Domain: media.StaticDomain()}
	}
	fillUserRoleIDsContext(ctx, db, user)
	menuKeys, menuReady, _ := permissionsupport.SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, user.ID, user.RoleIDs, permissionsupport.PlatformClient)
	menuKeys = orderedClientMenuKeys(menuKeys)
	apiKeys, apiReady, _ := permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, user.ID, user.RoleIDs, permissionsupport.PlatformClient)
	labels, catalogVersion := clientPermissionLabelsByKeysContext(ctx, db, append(append([]string{}, menuKeys...), apiKeys...))
	return &BootstrapResponse{
		User:                *user,
		Domain:              media.StaticDomain(),
		Menus:               clientMenusByKeysWithLabels(menuKeys, labels),
		MenuPermissionKeys:  menuKeys,
		MenuPermissionReady: menuReady,
		APIPermissionKeys:   apiKeys,
		APIPermissionReady:  apiReady,
		PermissionVersion:   clientPermissionVersionContext(ctx, db, user, catalogVersion),
	}
}

func fillClientUserDepartmentContext(ctx context.Context, db *gorm.DB, user *model.User) {
	if db == nil || user == nil || user.ID == 0 {
		return
	}
	var ud model.UserDept
	if err := db.WithContext(ctx).Where("`user_dept_user_id` = ?", user.ID).First(&ud).Error; err != nil || ud.DeptID == 0 {
		return
	}
	var d model.Department
	if err := db.WithContext(ctx).First(&d, ud.DeptID).Error; err == nil {
		user.DeptName = d.Name
		user.TopDeptName = dept.TopDeptName(ud.DeptID)
	}
}

func orderedClientMenuKeys(keys []string) []string {
	selected := map[string]bool{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			selected[key] = true
		}
	}
	result := make([]string, 0, len(selected))
	for _, declaration := range appmenuperm.ClientMenuDeclarations() {
		if selected[declaration.Key] {
			result = append(result, declaration.Key)
		}
	}
	return result
}

func clientMenusByKeysWithLabels(menuKeys []string, labels map[string]string) []ClientMenuDTO {
	selected := map[string]bool{}
	for _, key := range menuKeys {
		selected[strings.TrimSpace(key)] = true
	}
	menus := make([]ClientMenuDTO, 0, len(menuKeys))
	for _, declaration := range appmenuperm.ClientMenuDeclarations() {
		if !selected[declaration.Key] {
			continue
		}
		label := declaration.Name
		if labels != nil && strings.TrimSpace(labels[declaration.Key]) != "" {
			label = strings.TrimSpace(labels[declaration.Key])
		}
		menus = append(menus, ClientMenuDTO{
			Key:           strings.TrimPrefix(declaration.Key, "client:menu:"),
			Label:         label,
			Path:          declaration.Path,
			PermissionKey: declaration.Key,
			Sort:          declaration.Sort,
		})
	}
	return menus
}

func clientPermissionLabelsByKeysContext(ctx context.Context, db *gorm.DB, keys []string) (map[string]string, int64) {
	if db == nil || len(keys) == 0 {
		return map[string]string{}, 0
	}
	var rows []clientPermissionCatalogRow
	if err := db.WithContext(ctx).
		Model(&model.Permission{}).
		Select("`permission_key`, `permission_name`, `permission_edit_time`").
		Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_status` = 1", keys, permissionsupport.PlatformClient).
		Find(&rows).Error; err != nil {
		return map[string]string{}, 0
	}
	labels := make(map[string]string, len(rows))
	var version int64
	for _, row := range rows {
		if strings.TrimSpace(row.Name) != "" {
			labels[row.Key] = row.Name
		}
		if row.EditTime > version {
			version = row.EditTime
		}
	}
	return labels, version
}

func clientPermissionVersionContext(ctx context.Context, db *gorm.DB, user *model.User, fallback int64) int64 {
	version := fallback
	if user != nil {
		versionInputs := []int64{user.AddTime, user.EditTime, int64(user.ID), int64(user.RoleID)}
		for _, roleID := range user.RoleIDs {
			versionInputs = append(versionInputs, int64(roleID))
		}
		for _, value := range versionInputs {
			if value > version {
				version = value
			}
		}
	}
	if db == nil || user == nil || (user.ID == 0 && len(user.RoleIDs) == 0 && user.RoleID == 0) {
		return version
	}
	roleIDs := uniqueClientRoleIDs(append([]uint{user.RoleID}, user.RoleIDs...))
	var grantVersion int64
	query := db.WithContext(ctx).
		Model(&model.PermissionGrant{}).
		Select("COALESCE(MAX(`grant_edit_time`), 0)").
		Where("`grant_status` = 1").
		Where("(`grant_permission_key` LIKE ? OR `grant_permission_key` LIKE ?)", "client:menu:%", "client:api:%")
	if user.ID > 0 && len(roleIDs) > 0 {
		query = query.Where("(`grant_subject_type` = ? AND `grant_subject_id` = ?) OR (`grant_subject_type` = ? AND `grant_subject_id` IN ?)", permissionsupport.SubjectUser, user.ID, permissionsupport.SubjectRole, roleIDs)
	} else if user.ID > 0 {
		query = query.Where("`grant_subject_type` = ? AND `grant_subject_id` = ?", permissionsupport.SubjectUser, user.ID)
	} else {
		query = query.Where("`grant_subject_type` = ? AND `grant_subject_id` IN ?", permissionsupport.SubjectRole, roleIDs)
	}
	if err := query.Scan(&grantVersion).Error; err == nil && grantVersion > version {
		version = grantVersion
	}
	return version
}

func uniqueClientRoleIDs(values []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
