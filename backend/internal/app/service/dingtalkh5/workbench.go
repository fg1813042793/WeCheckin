package dingtalkh5

import (
	"context"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func WorkbenchStatsContext(ctx context.Context, user *model.DingTalkH5PerfUser) (*WorkbenchStatsDTO, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var reviews []model.DingTalkH5PerfReview
	query := db.Select("employee_account, manager_account, hrbp_account, hrbp_reviewer_account, status")
	query = applyReviewVisibilityScope(query, user, reviewScopeDashboard)
	if err := query.Find(&reviews).Error; err != nil {
		return nil, err
	}
	stats := workbenchStatsFromReviews(user, reviews)
	return &stats, nil
}

func workbenchStatsFromReviews(user *model.DingTalkH5PerfUser, reviews []model.DingTalkH5PerfReview) WorkbenchStatsDTO {
	draft := 0
	reviewing := 0
	completed := 0
	queue := 0
	for _, review := range reviews {
		switch review.Status {
		case ReviewStatusDraft:
			draft++
		case ReviewStatusManagerReview, ReviewStatusHRBPReview, ReviewStatusEmployeeConfirm, ReviewStatusHRFinal:
			reviewing++
		case ReviewStatusCompleted:
			completed++
		}
		if isWorkbenchQueueReview(user, review) {
			queue++
		}
	}
	return WorkbenchStatsDTO{Cards: []WorkbenchStatCardDTO{
		{Key: "queue", Label: "我的待办", Value: queue},
		{Key: "all", Label: "全部考评单", Value: len(reviews)},
		{Key: "draft", Label: "员工填写", Value: draft},
		{Key: "reviewing", Label: "流转中", Value: reviewing},
		{Key: "completed", Label: "已完成", Value: completed},
	}}
}

func isWorkbenchQueueReview(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	switch review.Status {
	case ReviewStatusDraft:
		return review.EmployeeAccount == user.Account
	case ReviewStatusManagerReview:
		return review.ManagerAccount == user.Account
	case ReviewStatusHRBPReview:
		return isHrbpReviewer(user, review)
	case ReviewStatusEmployeeConfirm:
		return review.EmployeeAccount == user.Account
	case ReviewStatusHRFinal:
		return canHandleHrbpFinal(user, review)
	default:
		return false
	}
}
