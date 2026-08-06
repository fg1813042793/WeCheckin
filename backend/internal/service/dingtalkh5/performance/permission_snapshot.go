package performance

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

type dingTalkH5PermissionSnapshot struct {
	menuKeys    []string
	buttonKeys  []string
	apiKeys     []string
	labels      map[string]string
	icons       map[string]string
	menuReady   bool
	buttonReady bool
	apiReady    bool
	version     int64
}

type dingTalkH5PermissionGrantRow struct {
	SubjectType   string `gorm:"column:grant_subject_type"`
	SubjectID     uint   `gorm:"column:grant_subject_id"`
	PermissionKey string `gorm:"column:grant_permission_key"`
	Effect        string `gorm:"column:grant_effect"`
	EditTime      int64  `gorm:"column:grant_edit_time"`
}

type dingTalkH5PermissionCatalogRow struct {
	Key      string `gorm:"column:permission_key"`
	Name     string `gorm:"column:permission_name"`
	Type     string `gorm:"column:permission_type"`
	Icon     string `gorm:"column:permission_icon"`
	EditTime int64  `gorm:"column:permission_edit_time"`
}

func dingTalkH5PermissionSnapshotForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (dingTalkH5PermissionSnapshot, error) {
	ctx = normalizedDingTalkH5PermissionContext(ctx)
	snapshot := dingTalkH5PermissionSnapshot{
		labels:  map[string]string{},
		icons:   map[string]string{},
		version: permissionVersionFallback(user),
	}
	if err := ctx.Err(); err != nil {
		return snapshot, err
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return snapshot, err
	}
	if user == nil || db == nil || (user.ID == 0 && len(roleIDs) == 0) {
		return snapshot, nil
	}
	if !permissionsupport.TablesReady(db) {
		return snapshot, nil
	}

	rows, err := dingTalkH5PermissionGrantRowsForUserContext(ctx, db, user)
	if err != nil {
		return snapshot, err
	}
	roleSet := make(map[uint]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		roleSet[roleID] = true
	}
	roleAllowed := make(map[string]bool, len(rows))
	userAllowed := make(map[string]bool, len(rows))
	denied := make(map[string]bool, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.PermissionKey)
		if key == "" {
			continue
		}
		if row.EditTime > snapshot.version {
			snapshot.version = row.EditTime
		}
		if strings.HasPrefix(key, "dingtalk_h5:api:") && row.Effect == permissionsupport.EffectAllow {
			snapshot.apiReady = true
		}
		switch {
		case row.SubjectType == permissionsupport.SubjectRole && roleSet[row.SubjectID]:
			if row.Effect == permissionsupport.EffectAllow {
				roleAllowed[key] = true
			}
		case row.SubjectType == permissionsupport.SubjectUser && row.SubjectID == user.ID:
			if row.Effect == permissionsupport.EffectDeny {
				denied[key] = true
				continue
			}
			if row.Effect == permissionsupport.EffectAllow {
				userAllowed[key] = true
			}
		}
	}

	selected := make(map[string]bool, len(roleAllowed)+len(userAllowed))
	for key := range roleAllowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	for key := range userAllowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	snapshot.menuKeys = orderedDingTalkH5MenuKeys(selected)
	snapshot.buttonKeys = orderedDingTalkH5ButtonKeys(selected)
	snapshot.apiKeys = orderedDingTalkH5APIKeys(selected)
	snapshot.menuReady = true
	snapshot.buttonReady = true

	labels, icons, permissionVersion, err := dingTalkH5PermissionCatalogContext(ctx, db, snapshot.menuKeys, snapshot.buttonKeys)
	if err != nil {
		return snapshot, err
	}
	snapshot.labels = labels
	snapshot.icons = icons
	if permissionVersion > snapshot.version {
		snapshot.version = permissionVersion
	}
	return snapshot, nil
}

func dingTalkH5PermissionGrantRowsForUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]dingTalkH5PermissionGrantRow, error) {
	subjectWhere, subjectArgs := dingTalkH5PermissionSubjectClause(user)
	if subjectWhere == "" {
		return nil, nil
	}
	permissionWhere, permissionArgs := dingTalkH5PermissionGrantLikeClause()
	var rows []dingTalkH5PermissionGrantRow
	err := db.WithContext(ctx).
		Model(&model.PermissionGrant{}).
		Select("`grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_effect`, `grant_edit_time`").
		Where("`grant_status` = 1").
		Where("("+subjectWhere+")", subjectArgs...).
		Where("("+permissionWhere+")", permissionArgs...).
		Find(&rows).Error
	return rows, err
}

func normalizedDingTalkH5PermissionContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func dingTalkH5PermissionSubjectClause(user *model.DingTalkH5PerfUser) (string, []interface{}) {
	if user == nil {
		return "", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)
	if user.ID > 0 {
		clauses = append(clauses, "(`grant_subject_type` = ? AND `grant_subject_id` = ?)")
		args = append(args, permissionsupport.SubjectUser, user.ID)
	}
	roleIDs := uniqueUintIDs(append([]uint{user.RoleID}, user.RoleIDs...))
	if len(roleIDs) > 0 {
		clauses = append(clauses, "(`grant_subject_type` = ? AND `grant_subject_id` IN ?)")
		args = append(args, permissionsupport.SubjectRole, roleIDs)
	}
	return strings.Join(clauses, " OR "), args
}

func dingTalkH5PermissionGrantLikeClause() (string, []interface{}) {
	prefixes := []string{
		"dingtalk_h5:menu:%",
		"dingtalk_h5:button:%",
		"dingtalk_h5:api:%",
		"data:%",
	}
	clauses := make([]string, 0, len(prefixes))
	args := make([]interface{}, 0, len(prefixes))
	for _, prefix := range prefixes {
		clauses = append(clauses, "`grant_permission_key` LIKE ?")
		args = append(args, prefix)
	}
	return strings.Join(clauses, " OR "), args
}

func orderedDingTalkH5MenuKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appmenuperm.DingTalkH5MenuDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func orderedDingTalkH5ButtonKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appmenuperm.DingTalkH5ButtonDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func orderedDingTalkH5APIKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appapiperm.DingTalkH5APIDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func dingTalkH5PermissionCatalogContext(ctx context.Context, db *gorm.DB, menuKeys, buttonKeys []string) (map[string]string, map[string]string, int64, error) {
	queryKeys := dingTalkH5PermissionCatalogKeys(menuKeys, buttonKeys)
	if len(queryKeys) == 0 {
		return map[string]string{}, map[string]string{}, 0, nil
	}
	var rows []dingTalkH5PermissionCatalogRow
	if err := db.WithContext(ctx).
		Model(&model.Permission{}).
		Select("`permission_key`, `permission_name`, `permission_type`, `permission_icon`, `permission_edit_time`").
		Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_status` = 1", queryKeys, permissionsupport.PlatformDingTalkH5).
		Find(&rows).Error; err != nil {
		return nil, nil, 0, err
	}
	labels := make(map[string]string, len(rows))
	icons := make(map[string]string, len(rows))
	var version int64
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		name := strings.TrimSpace(row.Name)
		icon := strings.TrimSpace(row.Icon)
		if row.EditTime > version {
			version = row.EditTime
		}
		if key != "" && name != "" && dingTalkH5PermissionCatalogTypeHasMenuLabel(row.Type) {
			labels[key] = name
		}
		if key != "" && icon != "" && dingTalkH5PermissionCatalogTypeHasMenuLabel(row.Type) {
			icons[key] = icon
		}
	}
	return labels, icons, version, nil
}

func dingTalkH5PermissionCatalogKeys(menuKeys, buttonKeys []string) []string {
	allowed := dingTalkH5AllowedMenuKeySet(menuKeys)
	for _, key := range buttonKeys {
		key = strings.TrimSpace(key)
		if key != "" {
			allowed[key] = true
		}
	}
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		keys = append(keys, key)
	}
	return keys
}

func dingTalkH5PermissionCatalogTypeHasMenuLabel(permissionType string) bool {
	return permissionType == permissionsupport.TypeDirectory || permissionType == permissionsupport.TypeMenu
}
