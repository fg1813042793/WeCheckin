package access

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

func AdminDeptIDs(adminID uint) []uint {
	return AdminDeptIDsContext(context.Background(), adminID)
}

func AdminDeptIDsContext(ctx context.Context, adminID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return adminDeptIDsWithDBContext(ctx, db, adminID)
}

func adminDeptIDsWithDBContext(ctx context.Context, db *gorm.DB, adminID uint) []uint {
	var list []model.UserDept
	if db == nil || (ctx != nil && ctx.Err() != nil) {
		return nil
	}
	db.Where("`user_dept_user_id` = ?", adminID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func SaveAdminDepts(adminID uint, deptIDs []uint) {
	_ = SaveAdminDeptsContext(context.Background(), adminID, deptIDs)
}

func SaveAdminDeptsContext(ctx context.Context, adminID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return SaveAdminDeptsTx(db, adminID, deptIDs)
}

func SaveAdminDeptsTx(tx *gorm.DB, adminID uint, deptIDs []uint) error {
	if err := tx.Where("`user_dept_user_id` = ?", adminID).Delete(&model.UserDept{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if deptID > 0 {
			if err := tx.Create(&model.UserDept{UserID: adminID, DeptID: deptID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func RoleDeptIDs(roleID uint) []uint {
	return RoleDeptIDsContext(context.Background(), roleID)
}

func RoleDeptIDsContext(ctx context.Context, roleID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	deptIDs, _ := permissionsupport.RoleCustomDeptIDsContext(ctx, db, roleID)
	return deptIDs
}

func DeptDescendantIDs(all []*model.Department, parentIDs []uint) []uint {
	ids := make([]uint, 0)
	idSet := make(map[uint]bool)
	for _, id := range parentIDs {
		idSet[id] = true
	}
	queue := make([]uint, len(parentIDs))
	copy(queue, parentIDs)
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, d := range all {
			if d.ParentID == pid && !idSet[d.ID] {
				idSet[d.ID] = true
				queue = append(queue, d.ID)
			}
		}
	}
	for id := range idSet {
		ids = append(ids, id)
	}
	return ids
}

func VisibleDeptIDs(admin *model.Admin) []uint {
	return VisibleDeptIDsContext(context.Background(), admin)
}

func VisibleDeptIDsContext(ctx context.Context, admin *model.Admin) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return VisibleDeptIDsWithDBContext(ctx, db, admin)
}

func VisibleDeptIDsWithDBContext(ctx context.Context, db *gorm.DB, admin *model.Admin) []uint {
	if admin == nil {
		return []uint{}
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return nil
	}
	visibleDeptIDs := []uint{}
	if scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err == nil && scope.Ready {
		visibleDeptIDs = visibleDeptIDsByScope(ctx, db, admin, scope.Mode, scope.DeptIDs)
		if visibleDeptIDs == nil {
			return nil
		}
	}
	extraDeptIDs, err := extraDataScopeDeptIDs(ctx, db, admin.ID)
	if err != nil || len(extraDeptIDs) == 0 {
		return visibleDeptIDs
	}
	return unionUintIDs(visibleDeptIDs, extraDeptIDs)
}

func DataScopeFilter(admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	return DataScopeFilterContext(context.Background(), admin, deptField, createByField)
}

func DataScopeFilterContext(ctx context.Context, admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return DataScopeFilterWithDBContext(ctx, db, admin, deptField, createByField)
}

func DataScopeFilterWithDBContext(ctx context.Context, db *gorm.DB, admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	if admin == nil {
		return "1 = 0", nil
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return "", nil
	}
	baseWhere := "1 = 0"
	baseArgs := []interface{}(nil)
	if scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err == nil && scope.Ready {
		baseWhere, baseArgs = dataScopeFilterByMode(ctx, db, admin, deptField, createByField, scope.Mode, scope.DeptIDs)
		if baseWhere == "" {
			return "", nil
		}
	}
	extraWhere, extraArgs := dataScopeExtraFilter(ctx, db, admin.ID, deptField, createByField)
	return orFilters(baseWhere, baseArgs, extraWhere, extraArgs)
}

func DataScopeFilterForResourceContext(ctx context.Context, admin *model.Admin, fields ResourceAuditFields) (string, []interface{}) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return DataScopeFilterForResourceWithDBContext(ctx, db, admin, fields)
}

func DataScopeFilterForResourceWithDBContext(ctx context.Context, db *gorm.DB, admin *model.Admin, fields ResourceAuditFields) (string, []interface{}) {
	return DataScopeFilterWithDBContext(ctx, db, admin, fields.DataScopeDeptField(), fields.DataScopeCreateByField())
}

func ScopedResourceQueryContext(ctx context.Context, db *gorm.DB, adminID uint, resource interface{}, deptField, createByField string) (*gorm.DB, error) {
	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	query := db.Model(resource)
	where, args := DataScopeFilterWithDBContext(ctx, db, &admin, deptField, createByField)
	if where != "" {
		query = query.Where(where, args...)
	}
	return query, nil
}

func ScopedResourceQueryByFieldsContext(ctx context.Context, db *gorm.DB, adminID uint, resource interface{}, fields ResourceAuditFields) (*gorm.DB, error) {
	return ScopedResourceQueryContext(ctx, db, adminID, resource, fields.DataScopeDeptField(), fields.DataScopeCreateByField())
}

func UserDataScopeFilterContext(ctx context.Context, admin *model.Admin) (string, []interface{}) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return UserDataScopeFilterWithDBContext(ctx, db, admin)
}

func UserDataScopeFilterWithDBContext(ctx context.Context, db *gorm.DB, admin *model.Admin) (string, []interface{}) {
	return UserIDDataScopeFilterWithDBContext(ctx, db, admin, UserAuditFields.IDField())
}

func UserIDDataScopeFilterWithDBContext(ctx context.Context, db *gorm.DB, admin *model.Admin, userIDField string) (string, []interface{}) {
	if admin == nil {
		return "1 = 0", nil
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return "", nil
	}
	baseWhere := "1 = 0"
	baseArgs := []interface{}(nil)
	if scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err == nil && scope.Ready {
		switch scope.Mode {
		case 1:
			return "", nil
		case 2, 4:
			deptIDs := visibleDeptIDsByScope(ctx, db, admin, scope.Mode, scope.DeptIDs)
			baseWhere, baseArgs = userDeptScopeFilter(userIDField, deptIDs)
		case 3:
			baseWhere, baseArgs = userIDField+" = ?", []interface{}{admin.ID}
		default:
			baseWhere = "1 = 0"
		}
	}
	extraWhere, extraArgs := userDataScopeExtraFilter(ctx, db, admin.ID, userIDField)
	return orFilters(baseWhere, baseArgs, extraWhere, extraArgs)
}

func RequireRowsAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func visibleDeptIDsByScope(ctx context.Context, db *gorm.DB, admin *model.Admin, mode int, customDeptIDs []uint) []uint {
	var all []*model.Department
	db.Find(&all)
	switch mode {
	case 1:
		return nil
	case 2:
		deptIDs := adminDeptIDsWithDBContext(ctx, db, admin.ID)
		if len(deptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, deptIDs)
	case 3:
		return adminDeptIDsWithDBContext(ctx, db, admin.ID)
	case 4:
		if len(customDeptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, customDeptIDs)
	}
	return nil
}

func dataScopeFilterByMode(ctx context.Context, db *gorm.DB, admin *model.Admin, deptField, createByField string, mode int, customDeptIDs []uint) (string, []interface{}) {
	switch mode {
	case 1:
		return "", nil
	case 2:
		deptIDs := visibleDeptIDsByScope(ctx, db, admin, mode, customDeptIDs)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(deptIDs)}
	case 3:
		return createByField + " = ?", []interface{}{admin.ID}
	case 4:
		deptIDs := visibleDeptIDsByScope(ctx, db, admin, mode, customDeptIDs)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(deptIDs)}
	}
	return "", nil
}

func dataScopeExtraFilter(ctx context.Context, db *gorm.DB, userID uint, deptField, createByField string) (string, []interface{}) {
	extras, err := permissionsupport.UserDataScopeExtrasContext(ctx, db, userID)
	if err != nil || !extras.Ready {
		return "", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if len(extras.DeptIDs) > 0 && deptField != "" {
		deptIDs, err := extraDataScopeDeptIDsByBase(ctx, db, extras.DeptIDs)
		if err == nil && len(deptIDs) > 0 {
			clauses = append(clauses, deptField+" IN ?")
			args = append(args, toInterfaceSlice(deptIDs))
		}
	}
	if len(extras.UserIDs) > 0 && createByField != "" {
		clauses = append(clauses, createByField+" IN ?")
		args = append(args, toInterfaceSlice(extras.UserIDs))
	}
	return joinORClauses(clauses, args)
}

func userDataScopeExtraFilter(ctx context.Context, db *gorm.DB, userID uint, userIDField string) (string, []interface{}) {
	extras, err := permissionsupport.UserDataScopeExtrasContext(ctx, db, userID)
	if err != nil || !extras.Ready {
		return "", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if len(extras.UserIDs) > 0 {
		clauses = append(clauses, userIDField+" IN ?")
		args = append(args, toInterfaceSlice(extras.UserIDs))
	}
	if len(extras.DeptIDs) > 0 {
		deptIDs, err := extraDataScopeDeptIDsByBase(ctx, db, extras.DeptIDs)
		if err == nil && len(deptIDs) > 0 {
			clauses = append(clauses, userIDField+" IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)")
			args = append(args, deptIDs)
		}
	}
	return joinORClauses(clauses, args)
}

func extraDataScopeDeptIDs(ctx context.Context, db *gorm.DB, userID uint) ([]uint, error) {
	extras, err := permissionsupport.UserDataScopeExtrasContext(ctx, db, userID)
	if err != nil || !extras.Ready || len(extras.DeptIDs) == 0 {
		return nil, err
	}
	return extraDataScopeDeptIDsByBase(ctx, db, extras.DeptIDs)
}

func extraDataScopeDeptIDsByBase(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]uint, error) {
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var all []*model.Department
	if err := db.WithContext(ctx).Find(&all).Error; err != nil {
		return nil, err
	}
	return DeptDescendantIDs(all, deptIDs), nil
}

func userDeptScopeFilter(userIDField string, deptIDs []uint) (string, []interface{}) {
	if deptIDs == nil {
		return "", nil
	}
	if len(deptIDs) == 0 {
		return "1 = 0", nil
	}
	return userIDField + " IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", []interface{}{deptIDs}
}

func orFilters(leftWhere string, leftArgs []interface{}, rightWhere string, rightArgs []interface{}) (string, []interface{}) {
	if leftWhere == "" || rightWhere == "" {
		if leftWhere == "" {
			return rightWhere, rightArgs
		}
		return leftWhere, leftArgs
	}
	if leftWhere == "1 = 0" {
		return rightWhere, rightArgs
	}
	if rightWhere == "1 = 0" {
		return leftWhere, leftArgs
	}
	args := make([]interface{}, 0, len(leftArgs)+len(rightArgs))
	args = append(args, leftArgs...)
	args = append(args, rightArgs...)
	return "(" + leftWhere + " OR " + rightWhere + ")", args
}

func joinORClauses(clauses []string, args []interface{}) (string, []interface{}) {
	switch len(clauses) {
	case 0:
		return "", nil
	case 1:
		return clauses[0], args
	default:
		where := "(" + clauses[0]
		for _, clause := range clauses[1:] {
			where += " OR " + clause
		}
		return where + ")", args
	}
}

func unionUintIDs(groups ...[]uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0)
	for _, group := range groups {
		for _, id := range group {
			if id == 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func toInterfaceSlice(ids []uint) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
