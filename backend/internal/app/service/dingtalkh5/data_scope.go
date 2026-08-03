package dingtalkh5

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/access"
	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
)

const (
	dingtalkH5DataScopeAll    = 1
	dingtalkH5DataScopeDept   = 2
	dingtalkH5DataScopeSelf   = 3
	dingtalkH5DataScopeCustom = 4
)

func reviewDataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (reviewWhereClause, error) {
	if user == nil || NormalizeUserID(user.Account) == "" {
		return reviewWhereClause{sql: "1 = 0"}, nil
	}
	scope, err := permissionsupport.DataScopeContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return reviewWhereClause{}, err
	}
	var base reviewWhereClause
	if !scope.Ready {
		base = personalReviewAuditScopeWhere(user)
	} else {
		switch scope.Mode {
		case dingtalkH5DataScopeAll:
			return reviewWhereClause{}, nil
		case dingtalkH5DataScopeSelf:
			base = personalReviewAuditScopeWhere(user)
		case dingtalkH5DataScopeDept:
			base, err = reviewDeptScopeWhereContext(ctx, db, user, scope, true)
			if err != nil {
				return reviewWhereClause{}, err
			}
		case dingtalkH5DataScopeCustom:
			base, err = reviewDeptScopeWhereContext(ctx, db, user, scope, false)
			if err != nil {
				return reviewWhereClause{}, err
			}
		default:
			base = reviewWhereClause{sql: "1 = 0"}
		}
	}
	extra, err := reviewExtraDataScopeWhereContext(ctx, db, user)
	if err != nil {
		return reviewWhereClause{}, err
	}
	return mergeReviewScopeWithExtra(base, extra), nil
}

func reviewInDataScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) (bool, error) {
	scope, err := reviewDataScopeWhereContext(ctx, db, user)
	if err != nil {
		return false, err
	}
	fields := access.DingTalkH5ReviewAuditFields
	query := notDeletedReviewQuery(db.WithContext(ctx).Model(&model.DingTalkH5PerfReview{})).Where(fields.IDField()+" = ?", review.ID)
	if scope.sql != "" {
		query = query.Where(scope.sql, scope.args...)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func reviewDeptScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope, includePersonal bool) (reviewWhereClause, error) {
	deptIDs, err := dataScopeDeptIDsContext(ctx, db, user, scope)
	if err != nil {
		return reviewWhereClause{}, err
	}
	fields := access.DingTalkH5ReviewAuditFields
	clauses := make([]reviewWhereClause, 0, 3)
	if includePersonal {
		clauses = append(clauses, personalReviewAuditScopeWhere(user))
	}
	if len(deptIDs) > 0 {
		clauses = append(clauses, reviewWhereClause{
			sql:  "employee_account IN (SELECT u.`user_mini_openid` FROM `users` u JOIN `user_depts` ud ON ud.`user_dept_user_id` = u.`id` WHERE u.`user_status` = 1 AND ud.`user_dept_dept_id` IN ?)",
			args: []interface{}{deptIDs},
		})
		clauses = append(clauses, reviewWhereClause{
			sql:  fields.CreateDeptField() + " IN ?",
			args: []interface{}{deptIDs},
		})
	}
	return clauseFromOR(clauses), nil
}

func personalReviewAuditScopeWhere(user *model.DingTalkH5PerfUser) reviewWhereClause {
	if user == nil {
		return reviewWhereClause{sql: "1 = 0"}
	}
	account := NormalizeUserID(user.Account)
	where, args := personalReviewVisibilityWhere(account)
	fields := access.DingTalkH5ReviewAuditFields
	clauses := []reviewWhereClause{
		{sql: where, args: args},
	}
	if user.ID > 0 {
		clauses = append(clauses, reviewWhereClause{sql: fields.CreateByField() + " = ?", args: []interface{}{user.ID}})
	}
	return clauseFromOR(clauses)
}

func dataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	accounts, all, err := baseDataScopeUserAccountsContext(ctx, db, user, scope)
	if err != nil || all {
		return accounts, all, err
	}
	extraAccounts, err := extraDataScopeUserAccountsContext(ctx, db, user)
	if err != nil {
		return nil, false, err
	}
	for account := range extraAccounts {
		accounts[account] = struct{}{}
	}
	return accounts, false, nil
}

func baseDataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	if user == nil {
		return map[string]struct{}{}, false, nil
	}
	if !scope.Ready {
		return map[string]struct{}{NormalizeUserID(user.Account): {}}, false, nil
	}
	switch scope.Mode {
	case dingtalkH5DataScopeAll:
		return nil, true, nil
	case dingtalkH5DataScopeSelf:
		return map[string]struct{}{NormalizeUserID(user.Account): {}}, false, nil
	case dingtalkH5DataScopeDept, dingtalkH5DataScopeCustom:
		accounts := map[string]struct{}{}
		if scope.Mode == dingtalkH5DataScopeDept {
			accounts[NormalizeUserID(user.Account)] = struct{}{}
		}
		deptIDs, err := dataScopeDeptIDsContext(ctx, db, user, scope)
		if err != nil {
			return nil, false, err
		}
		if len(deptIDs) == 0 {
			return accounts, false, nil
		}
		deptAccounts, err := userAccountsByDeptIDsContext(ctx, db, deptIDs)
		if err != nil {
			return nil, false, err
		}
		for _, account := range deptAccounts {
			accounts[account] = struct{}{}
		}
		return accounts, false, nil
	default:
		return map[string]struct{}{}, false, nil
	}
}

func reviewExtraDataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (reviewWhereClause, error) {
	if user == nil {
		return reviewWhereClause{}, nil
	}
	extras, err := permissionsupport.DataScopeExtrasContext(ctx, db, user.ID, user.RoleID)
	if err != nil || !extras.Ready {
		return reviewWhereClause{}, err
	}
	clauses := make([]reviewWhereClause, 0, 2)
	userIDs := uniqueUintIDs(extras.UserIDs)
	if len(userIDs) > 0 {
		accounts, err := accountsByUserIDsContext(ctx, db, userIDs)
		if err != nil {
			return reviewWhereClause{}, err
		}
		fields := access.DingTalkH5ReviewAuditFields
		userClauses := []reviewWhereClause{
			{sql: fields.CreateByField() + " IN ?", args: []interface{}{userIDs}},
		}
		if len(accounts) > 0 {
			userClauses = append(userClauses, reviewWhereClause{sql: "`employee_account` IN ?", args: []interface{}{accounts}})
		}
		clauses = append(clauses, clauseFromOR(userClauses))
	}
	if len(extras.DeptIDs) > 0 {
		deptIDs, err := dataScopeExtraDeptIDsContext(ctx, db, extras.DeptIDs)
		if err != nil {
			return reviewWhereClause{}, err
		}
		if len(deptIDs) > 0 {
			fields := access.DingTalkH5ReviewAuditFields
			clauses = append(clauses, reviewWhereClause{
				sql:  "(`employee_account` IN (SELECT u.`user_mini_openid` FROM `users` u JOIN `user_depts` ud ON ud.`user_dept_user_id` = u.`id` WHERE u.`user_status` = 1 AND ud.`user_dept_dept_id` IN ?) OR " + fields.CreateDeptField() + " IN ?)",
				args: []interface{}{deptIDs, deptIDs},
			})
		}
	}
	return mergeReviewClauses(clauses), nil
}

func extraDataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (map[string]struct{}, error) {
	result := map[string]struct{}{}
	if user == nil {
		return result, nil
	}
	extras, err := permissionsupport.DataScopeExtrasContext(ctx, db, user.ID, user.RoleID)
	if err != nil || !extras.Ready {
		return result, err
	}
	if len(extras.UserIDs) > 0 {
		accounts, err := accountsByUserIDsContext(ctx, db, extras.UserIDs)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			result[account] = struct{}{}
		}
	}
	if len(extras.DeptIDs) > 0 {
		deptIDs, err := dataScopeExtraDeptIDsContext(ctx, db, extras.DeptIDs)
		if err != nil {
			return nil, err
		}
		if len(deptIDs) > 0 {
			accounts, err := userAccountsByDeptIDsContext(ctx, db, deptIDs)
			if err != nil {
				return nil, err
			}
			for _, account := range accounts {
				result[account] = struct{}{}
			}
		}
	}
	return result, nil
}

func userAccountsByDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]string, error) {
	deptIDs = uniqueUintIDs(deptIDs)
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var rows []model.DingTalkH5PerfUser
	if err := db.WithContext(ctx).
		Model(&model.DingTalkH5PerfUser{}).
		Distinct("`users`.`user_mini_openid`").
		Joins("JOIN `user_depts` ON `user_depts`.`user_dept_user_id` = `users`.`id`").
		Where("`users`.`user_status` = 1").
		Where("`user_depts`.`user_dept_dept_id` IN ?", deptIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	accounts := make([]string, 0, len(rows))
	for _, row := range rows {
		if account := NormalizeUserID(row.Account); account != "" {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func accountsByUserIDsContext(ctx context.Context, db *gorm.DB, userIDs []uint) ([]string, error) {
	userIDs = uniqueUintIDs(userIDs)
	if len(userIDs) == 0 {
		return nil, nil
	}
	var rows []model.DingTalkH5PerfUser
	if err := db.WithContext(ctx).
		Select("`user_mini_openid`").
		Where("`id` IN ? AND `user_status` = 1", userIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	accounts := make([]string, 0, len(rows))
	for _, row := range rows {
		account := NormalizeUserID(row.Account)
		if account == "" {
			continue
		}
		if _, ok := seen[account]; ok {
			continue
		}
		seen[account] = struct{}{}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func dataScopeExtraDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]uint, error) {
	deptIDs = uniqueUintIDs(deptIDs)
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var departments []model.Department
	if err := db.WithContext(ctx).Find(&departments).Error; err != nil {
		return nil, err
	}
	return departmentDescendantIDs(departments, deptIDs), nil
}

func mergeReviewScopeWithExtra(base, extra reviewWhereClause) reviewWhereClause {
	if base.sql == "" {
		return reviewWhereClause{}
	}
	if extra.sql == "" {
		return base
	}
	return clauseFromOR([]reviewWhereClause{base, extra})
}

func mergeReviewClauses(clauses []reviewWhereClause) reviewWhereClause {
	filtered := make([]reviewWhereClause, 0, len(clauses))
	for _, clause := range clauses {
		if clause.sql != "" {
			filtered = append(filtered, clause)
		}
	}
	if len(filtered) == 0 {
		return reviewWhereClause{}
	}
	return clauseFromOR(filtered)
}

func dataScopeDeptIDsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) ([]uint, error) {
	baseIDs := make([]uint, 0)
	switch scope.Mode {
	case dingtalkH5DataScopeDept:
		var rows []model.UserDept
		if err := db.WithContext(ctx).Where("`user_dept_user_id` = ?", user.ID).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row.DeptID > 0 {
				baseIDs = append(baseIDs, row.DeptID)
			}
		}
	case dingtalkH5DataScopeCustom:
		baseIDs = append(baseIDs, scope.DeptIDs...)
	default:
		return nil, nil
	}
	baseIDs = uniqueUintIDs(baseIDs)
	if len(baseIDs) == 0 {
		return nil, nil
	}
	var departments []model.Department
	if err := db.WithContext(ctx).Find(&departments).Error; err != nil {
		return nil, err
	}
	return departmentDescendantIDs(departments, baseIDs), nil
}

func departmentDescendantIDs(all []model.Department, parentIDs []uint) []uint {
	seen := map[uint]struct{}{}
	queue := uniqueUintIDs(parentIDs)
	for _, id := range queue {
		seen[id] = struct{}{}
	}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		for _, dept := range all {
			if dept.ParentID != parentID || dept.ID == 0 {
				continue
			}
			if _, ok := seen[dept.ID]; ok {
				continue
			}
			seen[dept.ID] = struct{}{}
			queue = append(queue, dept.ID)
		}
	}
	result := make([]uint, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return uniqueUintIDs(result)
}

func uniqueUintIDs(items []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(items))
	for _, item := range items {
		if item == 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func clauseFromOR(clauses []reviewWhereClause) reviewWhereClause {
	sql, args := orReviewWhere(clauses)
	return reviewWhereClause{sql: sql, args: args}
}
