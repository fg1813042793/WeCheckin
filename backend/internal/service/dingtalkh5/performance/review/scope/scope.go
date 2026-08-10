package scope

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
	"wecheckin/backend/internal/support/access"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

const (
	ScopeDashboard = "dashboard"
	ScopeMine      = "mine"
	ScopeManager   = "manager"
	ScopeHRBP      = "hrbp"
	ScopeSummary   = "summary"

	dataScopeAll    = 1
	dataScopeDept   = 2
	dataScopeSelf   = 3
	dataScopeCustom = 4
)

type WhereClause struct {
	SQL  string
	Args []interface{}
}

func ApplyVisibilityScopeContext(ctx context.Context, db *gorm.DB, query *gorm.DB, user *model.DingTalkH5PerfUser, scope string) (*gorm.DB, error) {
	normalizedScope := NormalizeReviewScope(scope)
	where, args := VisibilityWhere(user, scope)
	if normalizedScope == ScopeMine || normalizedScope == ScopeDashboard {
		if where == "" {
			return query, nil
		}
		return query.Where(where, args...), nil
	}
	dataScope, err := DataScopeWhereContext(ctx, db, user)
	if err != nil {
		return nil, err
	}
	if dataScope.SQL != "" {
		query = query.Where(dataScope.SQL, dataScope.Args...)
	}
	if where == "" {
		return query, nil
	}
	return query.Where(where, args...), nil
}

func VisibilityWhere(user *model.DingTalkH5PerfUser, scope string) (string, []interface{}) {
	if user == nil || strings.TrimSpace(user.Account) == "" {
		return "1 = 0", nil
	}
	account := usersvc.NormalizeUserID(user.Account)
	switch NormalizeReviewScope(scope) {
	case ScopeMine:
		return "employee_account = ?", []interface{}{account}
	case ScopeManager:
		return "manager_account = ?", []interface{}{account}
	case ScopeHRBP:
		return hrbpReviewVisibilityWhere(user, account)
	case ScopeSummary:
		return "", nil
	default:
		return PersonalReviewVisibilityWhere(account)
	}
}

func NormalizeReviewScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case ScopeMine, ScopeManager, ScopeHRBP, ScopeSummary:
		return strings.TrimSpace(scope)
	default:
		return ScopeDashboard
	}
}

func PersonalReviewVisibilityWhere(account string) (string, []interface{}) {
	return "(employee_account = ? OR manager_account = ? OR hrbp_account = ? OR hrbp_reviewer_account = ?)",
		[]interface{}{account, account, account, account}
}

func hrbpReviewVisibilityWhere(user *model.DingTalkH5PerfUser, account string) (string, []interface{}) {
	return orWhere([]WhereClause{
		{SQL: "(hrbp_account = ? OR hrbp_reviewer_account = ?)", Args: []interface{}{account, account}},
		ResponsibleDepartmentScopeWhere(*user),
	})
}

func DepartmentScopeWhere(user model.DingTalkH5PerfUser) WhereClause {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return ResponsibleDepartmentScopeWhere(user)
	}
	parts := make([]string, 0, 3)
	args := make([]interface{}, 0, 3)
	for _, item := range []struct {
		column string
		value  string
	}{
		{column: "department_level1", value: user.DepartmentLevel1},
		{column: "department_level2", value: user.DepartmentLevel2},
		{column: "department_level3", value: user.DepartmentLevel3},
	} {
		value := strings.TrimSpace(item.value)
		if value == "" {
			continue
		}
		parts = append(parts, item.column+" = ?")
		args = append(args, value)
	}
	if len(parts) > 0 {
		return WhereClause{SQL: strings.Join(parts, " AND "), Args: args}
	}
	if department := strings.TrimSpace(user.Department); department != "" {
		return WhereClause{SQL: "department = ?", Args: []interface{}{department}}
	}
	return WhereClause{}
}

func ResponsibleDepartmentScopeWhere(user model.DingTalkH5PerfUser) WhereClause {
	departments := decodeStringList(user.ResponsibleDepartments)
	if len(departments) == 0 {
		return WhereClause{}
	}
	return WhereClause{
		SQL:  "(department_level1 IN ? OR department_level2 IN ? OR department_level3 IN ?)",
		Args: []interface{}{departments, departments, departments},
	}
}

func ReviewDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return ReviewResponsibleDepartmentScopeMatches(user, review)
	}
	userLevels := []string{user.DepartmentLevel1, user.DepartmentLevel2, user.DepartmentLevel3}
	reviewLevels := []string{review.DepartmentLevel1, review.DepartmentLevel2, review.DepartmentLevel3}
	hasScope := false
	for _, item := range userLevels {
		if strings.TrimSpace(item) != "" {
			hasScope = true
			break
		}
	}
	if !hasScope {
		return user.Department != "" && user.Department == review.Department
	}
	for index, item := range userLevels {
		if strings.TrimSpace(item) != "" && item != reviewLevels[index] {
			return false
		}
	}
	return true
}

func PerfUserDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return PerfUserResponsibleDepartmentScopeMatches(user, target)
	}
	return departmentScopeMatches(user, &target)
}

func ReviewResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	for _, department := range decodeStringList(user.ResponsibleDepartments) {
		if department == review.DepartmentLevel1 || department == review.DepartmentLevel2 || department == review.DepartmentLevel3 {
			return true
		}
		if review.Department != "" && strings.Contains(review.Department, department) {
			return true
		}
	}
	return false
}

func PerfUserResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	for _, department := range decodeStringList(user.ResponsibleDepartments) {
		if department == target.DepartmentLevel1 || department == target.DepartmentLevel2 || department == target.DepartmentLevel3 {
			return true
		}
		if target.Department != "" && strings.Contains(target.Department, department) {
			return true
		}
	}
	return false
}

func DataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (WhereClause, error) {
	if user == nil || usersvc.NormalizeUserID(user.Account) == "" {
		return WhereClause{SQL: "1 = 0"}, nil
	}
	scope, extras, err := permissionsupport.DataScopeBundleContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return WhereClause{}, err
	}
	var base WhereClause
	if !scope.Ready {
		base = personalReviewAuditScopeWhere(user)
	} else {
		switch scope.Mode {
		case dataScopeAll:
			return WhereClause{}, nil
		case dataScopeSelf:
			base = personalReviewAuditScopeWhere(user)
		case dataScopeDept:
			base, err = DeptScopeWhereContext(ctx, db, user, scope, true)
			if err != nil {
				return WhereClause{}, err
			}
		case dataScopeCustom:
			base, err = DeptScopeWhereContext(ctx, db, user, scope, false)
			if err != nil {
				return WhereClause{}, err
			}
		default:
			base = WhereClause{SQL: "1 = 0"}
		}
	}
	extra, err := ExtraDataScopeWhereFromExtrasContext(ctx, db, extras)
	if err != nil {
		return WhereClause{}, err
	}
	return mergeScopeWithExtra(base, extra), nil
}

func InDataScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) (bool, error) {
	scope, err := DataScopeWhereContext(ctx, db, user)
	if err != nil {
		return false, err
	}
	fields := access.DingTalkH5ReviewAuditFields
	query := db.WithContext(ctx).
		Model(&model.DingTalkH5PerfReview{}).
		Where("`deleted_at` = 0").
		Where(fields.IDField()+" = ?", review.ID)
	if scope.SQL != "" {
		query = query.Where(scope.SQL, scope.Args...)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeptScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope, includePersonal bool) (WhereClause, error) {
	deptIDs, err := usersvc.DataScopeDeptIDsContext(ctx, db, user, scope)
	if err != nil {
		return WhereClause{}, err
	}
	fields := access.DingTalkH5ReviewAuditFields
	clauses := make([]WhereClause, 0, 3)
	if includePersonal {
		clauses = append(clauses, personalReviewAuditScopeWhere(user))
	}
	if len(deptIDs) > 0 {
		clauses = append(clauses, WhereClause{
			SQL:  "employee_account IN (SELECT u.`user_mini_openid` FROM `users` u JOIN `user_depts` ud ON ud.`user_dept_user_id` = u.`id` WHERE u.`user_status` = 1 AND ud.`user_dept_dept_id` IN ?)",
			Args: []interface{}{deptIDs},
		})
		clauses = append(clauses, WhereClause{
			SQL:  fields.CreateDeptField() + " IN ?",
			Args: []interface{}{deptIDs},
		})
	}
	return clauseFromOR(clauses), nil
}

func ExtraDataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (WhereClause, error) {
	if user == nil {
		return WhereClause{}, nil
	}
	extras, err := permissionsupport.DataScopeExtrasContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return WhereClause{}, err
	}
	return ExtraDataScopeWhereFromExtrasContext(ctx, db, extras)
}

func ExtraDataScopeWhereFromExtrasContext(ctx context.Context, db *gorm.DB, extras permissionsupport.DataScopeExtras) (WhereClause, error) {
	if !extras.Ready {
		return WhereClause{}, nil
	}
	clauses := make([]WhereClause, 0, 2)
	userIDs := usersvc.UniqueUintIDs(extras.UserIDs)
	if len(userIDs) > 0 {
		accounts, err := usersvc.AccountsByUserIDsContext(ctx, db, userIDs)
		if err != nil {
			return WhereClause{}, err
		}
		fields := access.DingTalkH5ReviewAuditFields
		userClauses := []WhereClause{
			{SQL: fields.CreateByField() + " IN ?", Args: []interface{}{userIDs}},
		}
		if len(accounts) > 0 {
			userClauses = append(userClauses, WhereClause{SQL: "`employee_account` IN ?", Args: []interface{}{accounts}})
		}
		clauses = append(clauses, clauseFromOR(userClauses))
	}
	if len(extras.DeptIDs) > 0 {
		deptIDs, err := usersvc.DataScopeExtraDeptIDsContext(ctx, db, extras.DeptIDs)
		if err != nil {
			return WhereClause{}, err
		}
		if len(deptIDs) > 0 {
			fields := access.DingTalkH5ReviewAuditFields
			clauses = append(clauses, WhereClause{
				SQL:  "(`employee_account` IN (SELECT u.`user_mini_openid` FROM `users` u JOIN `user_depts` ud ON ud.`user_dept_user_id` = u.`id` WHERE u.`user_status` = 1 AND ud.`user_dept_dept_id` IN ?) OR " + fields.CreateDeptField() + " IN ?)",
				Args: []interface{}{deptIDs, deptIDs},
			})
		}
	}
	return mergeClauses(clauses), nil
}

func personalReviewAuditScopeWhere(user *model.DingTalkH5PerfUser) WhereClause {
	if user == nil {
		return WhereClause{SQL: "1 = 0"}
	}
	account := usersvc.NormalizeUserID(user.Account)
	where, args := PersonalReviewVisibilityWhere(account)
	fields := access.DingTalkH5ReviewAuditFields
	clauses := []WhereClause{
		{SQL: where, Args: args},
	}
	if user.ID > 0 {
		clauses = append(clauses, WhereClause{SQL: fields.CreateByField() + " = ?", Args: []interface{}{user.ID}})
	}
	return clauseFromOR(clauses)
}

func mergeScopeWithExtra(base, extra WhereClause) WhereClause {
	if base.SQL == "" {
		return WhereClause{}
	}
	if extra.SQL == "" {
		return base
	}
	return clauseFromOR([]WhereClause{base, extra})
}

func mergeClauses(clauses []WhereClause) WhereClause {
	filtered := make([]WhereClause, 0, len(clauses))
	for _, clause := range clauses {
		if clause.SQL != "" {
			filtered = append(filtered, clause)
		}
	}
	if len(filtered) == 0 {
		return WhereClause{}
	}
	return clauseFromOR(filtered)
}

func clauseFromOR(clauses []WhereClause) WhereClause {
	sql, args := orWhere(clauses)
	return WhereClause{SQL: sql, Args: args}
}

func orWhere(clauses []WhereClause) (string, []interface{}) {
	parts := make([]string, 0, len(clauses))
	args := make([]interface{}, 0)
	for _, clause := range clauses {
		if strings.TrimSpace(clause.SQL) == "" {
			continue
		}
		parts = append(parts, "("+clause.SQL+")")
		args = append(args, clause.Args...)
	}
	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return strings.Join(parts, " OR "), args
}

func departmentScopeMatches(leader model.DingTalkH5PerfUser, employee *model.DingTalkH5PerfUser) bool {
	if employee == nil {
		return false
	}
	leaderLevels := []string{leader.DepartmentLevel1, leader.DepartmentLevel2, leader.DepartmentLevel3}
	employeeLevels := []string{employee.DepartmentLevel1, employee.DepartmentLevel2, employee.DepartmentLevel3}
	hasScope := false
	for _, item := range leaderLevels {
		if item != "" {
			hasScope = true
			break
		}
	}
	if !hasScope {
		return leader.Department != "" && leader.Department == employee.Department
	}
	for index, item := range leaderLevels {
		if item != "" && item != employeeLevels[index] {
			return false
		}
	}
	return true
}

func decodeStringList(raw string) []string {
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}
