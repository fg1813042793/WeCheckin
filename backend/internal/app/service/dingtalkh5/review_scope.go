package dingtalkh5

import (
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
)

const (
	reviewScopeDashboard = "dashboard"
	reviewScopeMine      = "mine"
	reviewScopeManager   = "manager"
	reviewScopeHRBP      = "hrbp"
	reviewScopeSummary   = "summary"
)

type reviewWhereClause struct {
	sql  string
	args []interface{}
}

func applyReviewVisibilityScope(query *gorm.DB, user *model.DingTalkH5PerfUser, scope string) *gorm.DB {
	where, args := reviewVisibilityWhere(user, scope)
	if where == "" {
		return query
	}
	return query.Where(where, args...)
}

func reviewVisibilityWhere(user *model.DingTalkH5PerfUser, scope string) (string, []interface{}) {
	if user == nil || strings.TrimSpace(user.Account) == "" {
		return "1 = 0", nil
	}
	account := NormalizeUserID(user.Account)
	switch normalizeReviewScope(scope) {
	case reviewScopeMine:
		return "employee_account = ?", []interface{}{account}
	case reviewScopeManager:
		return "manager_account = ?", []interface{}{account}
	case reviewScopeHRBP:
		return hrbpReviewVisibilityWhere(user, account)
	case reviewScopeSummary:
		return summaryReviewVisibilityWhere(user, account)
	default:
		return personalReviewVisibilityWhere(account)
	}
}

func normalizeReviewScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case reviewScopeMine, reviewScopeManager, reviewScopeHRBP, reviewScopeSummary:
		return strings.TrimSpace(scope)
	default:
		return reviewScopeDashboard
	}
}

func personalReviewVisibilityWhere(account string) (string, []interface{}) {
	return "(employee_account = ? OR manager_account = ? OR hrbp_account = ? OR hrbp_reviewer_account = ?)",
		[]interface{}{account, account, account, account}
}

func hrbpReviewVisibilityWhere(user *model.DingTalkH5PerfUser, account string) (string, []interface{}) {
	if user.Role == "admin" || user.Role == "hrbp_manager" {
		return "", nil
	}
	return orReviewWhere([]reviewWhereClause{
		{sql: "(hrbp_account = ? OR hrbp_reviewer_account = ?)", args: []interface{}{account, account}},
		reviewResponsibleDepartmentScopeWhere(*user),
	})
}

func summaryReviewVisibilityWhere(user *model.DingTalkH5PerfUser, account string) (string, []interface{}) {
	if user.Role == "admin" || user.Role == "hrbp_manager" {
		return "", nil
	}
	departmentScope := reviewDepartmentScopeWhere(*user)
	if user.Role == "hrbp" {
		departmentScope = reviewResponsibleDepartmentScopeWhere(*user)
	}
	personalWhere, personalArgs := personalReviewVisibilityWhere(account)
	return orReviewWhere([]reviewWhereClause{
		{sql: personalWhere, args: personalArgs},
		departmentScope,
	})
}

func reviewDepartmentScopeWhere(user model.DingTalkH5PerfUser) reviewWhereClause {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return reviewResponsibleDepartmentScopeWhere(user)
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
		return reviewWhereClause{sql: strings.Join(parts, " AND "), args: args}
	}
	if department := strings.TrimSpace(user.Department); department != "" {
		return reviewWhereClause{sql: "department = ?", args: []interface{}{department}}
	}
	return reviewWhereClause{}
}

func reviewResponsibleDepartmentScopeWhere(user model.DingTalkH5PerfUser) reviewWhereClause {
	departments := decodeStringList(user.ResponsibleDepartments)
	if len(departments) == 0 {
		return reviewWhereClause{}
	}
	return reviewWhereClause{
		sql:  "(department_level1 IN ? OR department_level2 IN ? OR department_level3 IN ?)",
		args: []interface{}{departments, departments, departments},
	}
}

func orReviewWhere(clauses []reviewWhereClause) (string, []interface{}) {
	parts := make([]string, 0, len(clauses))
	args := make([]interface{}, 0)
	for _, clause := range clauses {
		if strings.TrimSpace(clause.sql) == "" {
			continue
		}
		parts = append(parts, "("+clause.sql+")")
		args = append(args, clause.args...)
	}
	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return strings.Join(parts, " OR "), args
}

func reviewDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return reviewResponsibleDepartmentScopeMatches(user, review)
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

func perfUserDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return perfUserResponsibleDepartmentScopeMatches(user, target)
	}
	return departmentScopeMatches(user, &target)
}

func reviewResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
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

func perfUserResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
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
