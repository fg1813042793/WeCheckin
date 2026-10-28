package permissiondelegation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

var (
	ErrPermissionExceeded = errors.New("不能授予超出当前管理员范围的权限")
	ErrDataScopeExceeded  = errors.New("不能授予超出当前管理员范围的数据权限")
	ErrDepartmentExceeded = errors.New("不能选择当前管理员范围外的部门")
	ErrUserExceeded       = errors.New("不能选择当前管理员范围外的用户")
	ErrRoleExceeded       = errors.New("不能分配当前管理员无权委派的角色")
	ErrReservedRole       = errors.New("只有超级管理员可以管理超级管理员角色")
)

type Ceiling struct {
	ActorID        uint
	ActorRoleID    uint
	RoleIDs        []uint
	All            bool
	PermissionKeys map[string]struct{}
	TreeKeys       map[string]struct{}
	DataScope      permissionsupport.DataScope
	DataExtras     permissionsupport.DataScopeExtras
	DepartmentIDs  []uint
}

type RoleGrant struct {
	Name            string
	AllowAdminLogin int
	DataScope       int
	DeptIDs         []uint
	PermissionKeys  []string
}

func ResolveByAdminIDContext(ctx context.Context, db *gorm.DB, adminID uint) (Ceiling, error) {
	if db == nil || adminID == 0 {
		return Ceiling{}, fmt.Errorf("管理员无效")
	}
	var admin model.Admin
	if err := db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return Ceiling{}, err
	}
	return ResolveContext(ctx, db, &admin)
}

func ResolveContext(ctx context.Context, db *gorm.DB, admin *model.Admin) (Ceiling, error) {
	if db == nil || admin == nil || admin.ID == 0 {
		return Ceiling{}, fmt.Errorf("管理员无效")
	}
	roleIDs := append([]uint(nil), admin.RoleIDs...)
	if len(roleIDs) == 0 {
		var err error
		roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
		if err != nil {
			return Ceiling{}, err
		}
	}
	result := Ceiling{
		ActorID:        admin.ID,
		ActorRoleID:    admin.RoleID,
		RoleIDs:        roleIDs,
		PermissionKeys: map[string]struct{}{},
		TreeKeys:       map[string]struct{}{},
	}
	if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, db, roleIDs) {
		result.All = true
		result.DataScope = permissionsupport.DataScope{Ready: true, Mode: 1}
		return result, nil
	}

	if err := result.loadPermissionKeys(ctx, db); err != nil {
		return Ceiling{}, err
	}
	for key := range result.PermissionKeys {
		result.TreeKeys[key] = struct{}{}
	}
	if err := result.expandPermissionAncestors(ctx, db); err != nil {
		return Ceiling{}, err
	}
	scope, extras, err := permissionsupport.DataScopeBundleWithRoleIDsContext(ctx, db, admin.ID, roleIDs)
	if err != nil {
		return Ceiling{}, err
	}
	result.DataScope = scope
	result.DataExtras = extras
	result.DepartmentIDs, err = delegableDepartmentIDs(ctx, db, admin.ID, scope, extras)
	if err != nil {
		return Ceiling{}, err
	}
	return result, nil
}

func (ceiling *Ceiling) expandPermissionAncestors(ctx context.Context, db *gorm.DB) error {
	if len(ceiling.PermissionKeys) == 0 {
		return nil
	}
	var rows []model.Permission
	if err := db.WithContext(ctx).Select("permission_key", "permission_parent_key").Find(&rows).Error; err != nil {
		return err
	}
	parentByKey := make(map[string]string, len(rows))
	for _, row := range rows {
		parentByKey[row.Key] = row.ParentKey
	}
	for key := range ceiling.TreeKeys {
		visited := map[string]bool{}
		for parent := strings.TrimSpace(parentByKey[key]); parent != "" && !visited[parent]; parent = strings.TrimSpace(parentByKey[parent]) {
			visited[parent] = true
			ceiling.TreeKeys[parent] = struct{}{}
		}
	}
	return nil
}

func (ceiling *Ceiling) loadPermissionKeys(ctx context.Context, db *gorm.DB) error {
	load := func(keys []string, _ bool, err error) error {
		if err != nil {
			return err
		}
		for _, key := range keys {
			key = strings.TrimSpace(key)
			if key != "" {
				ceiling.PermissionKeys[key] = struct{}{}
			}
		}
		return nil
	}
	if allowed, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.AdminLoginPermissionKey); err != nil {
		return err
	} else if allowed {
		ceiling.PermissionKeys[permissionsupport.AdminLoginPermissionKey] = struct{}{}
	}
	if err := load(permissionsupport.SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.PlatformAdmin)); err != nil {
		return err
	}
	if err := load(permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.PlatformAdmin)); err != nil {
		return err
	}
	if err := load(permissionsupport.SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.PlatformClient)); err != nil {
		return err
	}
	if err := load(permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.PlatformClient)); err != nil {
		return err
	}
	if err := load(permissionsupport.DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs)); err != nil {
		return err
	}
	if err := load(permissionsupport.DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs)); err != nil {
		return err
	}
	return load(permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, ceiling.ActorID, ceiling.RoleIDs, permissionsupport.PlatformDingTalkH5))
}

func delegableDepartmentIDs(ctx context.Context, db *gorm.DB, adminID uint, scope permissionsupport.DataScope, extras permissionsupport.DataScopeExtras) ([]uint, error) {
	if scope.Ready && scope.Mode == 1 {
		return nil, nil
	}
	baseIDs := make([]uint, 0)
	if scope.Ready {
		switch scope.Mode {
		case 2:
			var relations []model.UserDept
			if err := db.WithContext(ctx).Where("user_dept_user_id = ?", adminID).Find(&relations).Error; err != nil {
				return nil, err
			}
			for _, relation := range relations {
				baseIDs = append(baseIDs, relation.DeptID)
			}
		case 4:
			baseIDs = append(baseIDs, scope.DeptIDs...)
		}
	}
	baseIDs = append(baseIDs, extras.DeptIDs...)
	if len(baseIDs) == 0 {
		return []uint{}, nil
	}
	var departments []*model.Department
	if err := db.WithContext(ctx).Find(&departments).Error; err != nil {
		return nil, err
	}
	return sortedUintIDs(access.DeptDescendantIDs(departments, baseIDs)), nil
}

func (ceiling Ceiling) AllowsPermission(key string) bool {
	if ceiling.All {
		return true
	}
	_, ok := ceiling.PermissionKeys[strings.TrimSpace(key)]
	return ok
}

func (ceiling Ceiling) AllowsTreeKey(key string) bool {
	if ceiling.All {
		return true
	}
	_, ok := ceiling.TreeKeys[strings.TrimSpace(key)]
	return ok
}

func (ceiling Ceiling) RequireSuperAdmin() error {
	if ceiling.All {
		return nil
	}
	return ErrReservedRole
}

func (ceiling Ceiling) AllowedDataScopes() []int {
	if ceiling.All || (ceiling.DataScope.Ready && ceiling.DataScope.Mode == 1) {
		return []int{1, 2, 3, 4}
	}
	if !ceiling.DataScope.Ready {
		return []int{}
	}
	switch ceiling.DataScope.Mode {
	case 2:
		return []int{2, 3, 4}
	case 3:
		if len(ceiling.DepartmentIDs) > 0 {
			return []int{3, 4}
		}
		return []int{3}
	case 4:
		if len(ceiling.DepartmentIDs) > 0 {
			return []int{3, 4}
		}
		return []int{3}
	default:
		return []int{}
	}
}

func (ceiling Ceiling) ValidatePermissionKeys(keys []string) error {
	if ceiling.All {
		return nil
	}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || strings.HasPrefix(key, "data:") {
			continue
		}
		if !ceiling.AllowsPermission(key) {
			return fmt.Errorf("%w: %s", ErrPermissionExceeded, key)
		}
	}
	return nil
}

func (ceiling Ceiling) ValidateDepartmentIDs(departmentIDs []uint) error {
	if ceiling.All || (ceiling.DataScope.Ready && ceiling.DataScope.Mode == 1) {
		return nil
	}
	allowed := uintSet(ceiling.DepartmentIDs)
	for _, departmentID := range normalizeUintIDs(departmentIDs) {
		if !allowed[departmentID] {
			return fmt.Errorf("%w: %d", ErrDepartmentExceeded, departmentID)
		}
	}
	return nil
}

func (ceiling Ceiling) ValidateDepartmentIDsContext(ctx context.Context, db *gorm.DB, departmentIDs []uint) error {
	departmentIDs = normalizeUintIDs(departmentIDs)
	if err := ceiling.ValidateDepartmentIDs(departmentIDs); err != nil || len(departmentIDs) == 0 {
		return err
	}
	var count int64
	if err := db.WithContext(ctx).Model(&model.Department{}).Where("id IN ?", departmentIDs).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(departmentIDs)) {
		return ErrDepartmentExceeded
	}
	return nil
}

func (ceiling Ceiling) ValidateUserIDsContext(ctx context.Context, db *gorm.DB, userIDs []uint) error {
	userIDs = normalizeUintIDs(userIDs)
	if len(userIDs) == 0 {
		return nil
	}
	query := db.WithContext(ctx).Model(&model.User{}).Where("id IN ? AND user_status = 1", userIDs)
	if !ceiling.All {
		var admin model.Admin
		if err := db.WithContext(ctx).First(&admin, ceiling.ActorID).Error; err != nil {
			return err
		}
		if where, args := access.UserDataScopeFilterWithDBContext(ctx, db, &admin); where != "" {
			query = query.Where(where, args...)
		}
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(userIDs)) {
		return ErrUserExceeded
	}
	return nil
}

func (ceiling Ceiling) ValidateRoleGrant(grant RoleGrant) error {
	if ceiling.All {
		return nil
	}
	if strings.TrimSpace(grant.Name) == adminaccess.ReservedSuperAdminRoleName {
		return ErrReservedRole
	}
	if grant.AllowAdminLogin != 0 && !ceiling.AllowsPermission(permissionsupport.AdminLoginPermissionKey) {
		return ErrPermissionExceeded
	}
	if err := ceiling.ValidatePermissionKeys(grant.PermissionKeys); err != nil {
		return err
	}
	if !ceiling.DataScope.Ready {
		return ErrDataScopeExceeded
	}
	switch grant.DataScope {
	case 1:
		if ceiling.DataScope.Mode != 1 {
			return ErrDataScopeExceeded
		}
	case 2:
		if ceiling.DataScope.Mode != 1 && ceiling.DataScope.Mode != 2 {
			return ErrDataScopeExceeded
		}
	case 3:
		return nil
	case 4:
		if len(normalizeUintIDs(grant.DeptIDs)) == 0 {
			return ErrDataScopeExceeded
		}
		return ceiling.ValidateDepartmentIDs(grant.DeptIDs)
	default:
		return ErrDataScopeExceeded
	}
	return nil
}

func (ceiling Ceiling) ValidateRoleIDsContext(ctx context.Context, db *gorm.DB, roleIDs []uint) error {
	roleIDs = normalizeUintIDs(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	var roles []model.Role
	if err := db.WithContext(ctx).Where("id IN ? AND role_status = 1", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	if len(roles) != len(roleIDs) {
		return ErrRoleExceeded
	}
	assignments, err := permissionsupport.RoleAssignmentMapsContext(ctx, db, roleIDs)
	if err != nil {
		return err
	}
	for _, role := range roles {
		if err := ceiling.ValidateRoleGrant(RoleGrantFromAssignments(role, assignments)); err != nil {
			return fmt.Errorf("%w: %s", ErrRoleExceeded, err)
		}
	}
	return nil
}

func (ceiling Ceiling) ValidateUserAdminAccessContext(ctx context.Context, db *gorm.DB, userID uint) error {
	if ceiling.All || userID == 0 {
		return nil
	}
	var user model.User
	if err := db.WithContext(ctx).Select("id", "user_role_id").First(&user, userID).Error; err != nil {
		return err
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, userID, user.RoleID)
	if err != nil {
		return err
	}
	if err := ceiling.ValidateRoleIDsContext(ctx, db, roleIDs); err != nil {
		return err
	}
	allowKeys, denyKeys, err := permissionsupport.UserPermissionKeySetsContext(ctx, db, userID)
	if err != nil {
		return err
	}
	if err := ceiling.ValidatePermissionKeys(append(allowKeys, denyKeys...)); err != nil {
		return err
	}
	scope, extras, err := permissionsupport.DataScopeBundleWithRoleIDsContext(ctx, db, userID, roleIDs)
	if err != nil {
		return err
	}
	if scope.Ready {
		if err := ceiling.ValidateRoleGrant(RoleGrant{DataScope: scope.Mode, DeptIDs: scope.DeptIDs}); err != nil {
			return err
		}
	}
	if err := ceiling.ValidateDepartmentIDsContext(ctx, db, extras.DeptIDs); err != nil {
		return err
	}
	return ceiling.ValidateUserIDsContext(ctx, db, extras.UserIDs)
}

func RoleGrantFromAssignments(role model.Role, assignments permissionsupport.RoleAssignmentMaps) RoleGrant {
	keys := make([]string, 0)
	keys = append(keys, assignments.AdminPermissionKeys[role.ID]...)
	keys = append(keys, assignments.AdminAPIPermissionKeys[role.ID]...)
	keys = append(keys, assignments.ClientMenuKeys[role.ID]...)
	keys = append(keys, assignments.DingTalkH5MenuKeys[role.ID]...)
	keys = append(keys, assignments.ClientAPIPermissionKeys[role.ID]...)
	keys = append(keys, assignments.DingTalkH5APIPermissionKeys[role.ID]...)
	return RoleGrant{
		Name: role.Name, AllowAdminLogin: role.AllowAdminLogin, DataScope: role.DataScope,
		DeptIDs: assignments.DeptIDs[role.ID], PermissionKeys: keys,
	}
}

func stringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}

func normalizeUintIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
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
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sortedUintIDs(values []uint) []uint {
	return normalizeUintIDs(values)
}

func uintSet(values []uint) map[uint]bool {
	result := make(map[uint]bool, len(values))
	for _, value := range values {
		if value > 0 {
			result[value] = true
		}
	}
	return result
}
