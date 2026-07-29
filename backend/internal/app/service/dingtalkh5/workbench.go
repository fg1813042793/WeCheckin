package dingtalkh5

import (
	"context"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func WorkbenchStatsContext(ctx context.Context, user *model.DingTalkH5PerfUser) (*WorkbenchStatsDTO, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	statusCounts, total, err := workbenchStatusCountsContext(ctx, db, user)
	if err != nil {
		return nil, err
	}
	queue, err := workbenchQueueCountContext(ctx, db, user)
	if err != nil {
		return nil, err
	}
	stats := workbenchStatsFromCounts(statusCounts, total, queue)
	return &stats, nil
}

type workbenchStatusCountRow struct {
	Status string `gorm:"column:status"`
	Count  int64  `gorm:"column:cnt"`
}

func workbenchStatusCountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (map[string]int, int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
	}
	var rows []workbenchStatusCountRow
	query := db.Model(&model.DingTalkH5PerfReview{}).Select("status, COUNT(*) AS cnt")
	query = applyReviewVisibilityScope(query, user, reviewScopeDashboard)
	if err := query.Group("status").Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	counts := make(map[string]int, len(rows))
	total := 0
	for _, row := range rows {
		value := int(row.Count)
		counts[row.Status] = value
		total += value
	}
	return counts, total, nil
}

func workbenchQueueCountContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
	}
	where, args := workbenchQueueWhere(user)
	var total int64
	query := db.Model(&model.DingTalkH5PerfReview{})
	query = applyReviewVisibilityScope(query, user, reviewScopeDashboard)
	if err := query.Where(where, args...).Count(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

func workbenchQueueWhere(user *model.DingTalkH5PerfUser) (string, []interface{}) {
	if user == nil || user.Account == "" {
		return "1 = 0", nil
	}
	account := NormalizeUserID(user.Account)
	clauses := []reviewWhereClause{
		{sql: "status = ? AND employee_account = ?", args: []interface{}{ReviewStatusDraft, account}},
		{sql: "status = ? AND manager_account = ?", args: []interface{}{ReviewStatusManagerReview, account}},
		{sql: "status = ? AND employee_account = ?", args: []interface{}{ReviewStatusEmployeeConfirm, account}},
	}
	if user.Role == "admin" || user.Role == "hrbp_manager" {
		clauses = append(clauses,
			reviewWhereClause{sql: "status = ? AND (hrbp_reviewer_account = ? OR hrbp_reviewer_account = '')", args: []interface{}{ReviewStatusHRBPReview, account}},
			reviewWhereClause{sql: "status = ? AND (hrbp_reviewer_account = ? OR hrbp_reviewer_account = '' OR hrbp_account = ?)", args: []interface{}{ReviewStatusHRFinal, account, account}},
		)
	} else {
		clauses = append(clauses,
			reviewWhereClause{sql: "status = ? AND (hrbp_reviewer_account = ? OR (hrbp_reviewer_account = '' AND hrbp_account = ?))", args: []interface{}{ReviewStatusHRBPReview, account, account}},
			reviewWhereClause{sql: "status = ? AND (hrbp_reviewer_account = ? OR hrbp_account = ?)", args: []interface{}{ReviewStatusHRFinal, account, account}},
		)
	}
	return orReviewWhere(clauses)
}

func workbenchStatsFromCounts(statusCounts map[string]int, total, queue int) WorkbenchStatsDTO {
	reviewing := statusCounts[ReviewStatusManagerReview] + statusCounts[ReviewStatusHRBPReview] + statusCounts[ReviewStatusEmployeeConfirm] + statusCounts[ReviewStatusHRFinal]
	return WorkbenchStatsDTO{Cards: []WorkbenchStatCardDTO{
		{Key: "queue", Label: "我的待办", Value: queue},
		{Key: "all", Label: "全部考评单", Value: total},
		{Key: "draft", Label: "员工填写", Value: statusCounts[ReviewStatusDraft]},
		{Key: "reviewing", Label: "流转中", Value: reviewing},
		{Key: "completed", Label: "已完成", Value: statusCounts[ReviewStatusCompleted]},
	}}
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
