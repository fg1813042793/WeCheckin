package dingtalkh5

import "wecheckin-backend/backend/internal/model"

var (
	peopleLeaderRoles = map[string]struct{}{
		"director":   {},
		"manager":    {},
		"supervisor": {},
	}
	reviewCreatorRoles = map[string]struct{}{
		"admin":        {},
		"hrbp":         {},
		"hrbp_manager": {},
	}
	editableRoles = map[string]struct{}{
		"director":     {},
		"manager":      {},
		"supervisor":   {},
		"employee":     {},
		"hrbp":         {},
		"hrbp_manager": {},
		"admin":        {},
	}
)

func canCreateReview(user *model.DingTalkH5PerfUser) bool {
	_, ok := reviewCreatorRoles[user.Role]
	return ok
}

func isAdmin(user *model.DingTalkH5PerfUser) bool {
	return user != nil && user.Role == "admin"
}

func canBeReviewed(user model.DingTalkH5PerfUser) bool {
	_, ok := editableRoles[user.Role]
	return ok && user.Status == 1
}

func canViewReview(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview, employee *model.DingTalkH5PerfUser) bool {
	if user == nil {
		return false
	}
	if user.Role == "admin" || user.Role == "hrbp_manager" {
		return true
	}
	if user.Role == "hrbp" {
		return review.EmployeeAccount == user.Account ||
			review.HRBPAccount == user.Account ||
			review.HRBPReviewerAccount == user.Account ||
			reviewResponsibleDepartmentScopeMatches(*user, review)
	}
	if _, ok := peopleLeaderRoles[user.Role]; ok {
		return review.EmployeeAccount == user.Account ||
			review.ManagerAccount == user.Account ||
			review.HRBPAccount == user.Account ||
			departmentScopeMatches(*user, employee)
	}
	return review.EmployeeAccount == user.Account || review.ManagerAccount == user.Account || review.HRBPAccount == user.Account
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
	if leader.Role == "manager" {
		return employee.Role == "supervisor" || employee.Role == "employee"
	}
	if leader.Role == "supervisor" {
		return employee.Role == "employee" && employee.ManagerAccount == leader.Account
	}
	return true
}

func isHrbpReviewer(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return user.Role == "admin" || user.Role == "hrbp_manager" || (user.Role == "hrbp" && review.HRBPAccount == user.Account)
}

func canHandleHrbpFinal(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return review.HRBPAccount == user.Account || user.Role == "admin" || user.Role == "hrbp_manager"
}
