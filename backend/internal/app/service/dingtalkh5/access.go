package dingtalkh5

import "wecheckin/backend/internal/model"

func canBeReviewed(user model.DingTalkH5PerfUser) bool {
	return user.Status == 1
}

func canViewReview(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview, employee *model.DingTalkH5PerfUser) bool {
	if user == nil {
		return false
	}
	account := NormalizeUserID(user.Account)
	if review.EmployeeAccount == account ||
		review.ManagerAccount == account ||
		review.HRBPAccount == account ||
		review.HRBPReviewerAccount == account {
		return true
	}
	return reviewResponsibleDepartmentScopeMatches(*user, review)
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

func isHrbpReviewer(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return review.HRBPAccount == user.Account
}

func canHandleHrbpFinal(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return review.HRBPAccount == user.Account
}
