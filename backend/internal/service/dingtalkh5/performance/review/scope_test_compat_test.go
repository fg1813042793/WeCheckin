package review

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	reviewscope "wecheckin/backend/internal/service/dingtalkh5/performance/review/scope"
)

const (
	reviewScopeDashboard = reviewscope.ScopeDashboard
	reviewScopeMine      = reviewscope.ScopeMine
	reviewScopeManager   = reviewscope.ScopeManager
	reviewScopeHRBP      = reviewscope.ScopeHRBP
	reviewScopeSummary   = reviewscope.ScopeSummary
)

type reviewWhereClause struct {
	sql  string
	args []interface{}
}

func reviewVisibilityWhere(user *model.DingTalkH5PerfUser, scope string) (string, []interface{}) {
	return reviewscope.VisibilityWhere(user, scope)
}

func normalizeReviewScope(scope string) string {
	return reviewscope.NormalizeReviewScope(scope)
}

func applyReviewVisibilityScopeContext(ctx context.Context, db *gorm.DB, query *gorm.DB, user *model.DingTalkH5PerfUser, scope string) (*gorm.DB, error) {
	return reviewscope.ApplyVisibilityScopeContext(ctx, db, query, user, scope)
}

func reviewDataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (reviewWhereClause, error) {
	clause, err := reviewscope.DataScopeWhereContext(ctx, db, user)
	return reviewWhereClause{sql: clause.SQL, args: clause.Args}, err
}

func reviewInDataScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) (bool, error) {
	return reviewscope.InDataScopeContext(ctx, db, user, review)
}

func reviewExtraDataScopeWhereContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (reviewWhereClause, error) {
	clause, err := reviewscope.ExtraDataScopeWhereContext(ctx, db, user)
	return reviewWhereClause{sql: clause.SQL, args: clause.Args}, err
}

func reviewDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	return reviewscope.ReviewDepartmentScopeMatches(user, review)
}

func reviewResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	return reviewscope.ReviewResponsibleDepartmentScopeMatches(user, review)
}

func perfUserDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	return reviewscope.PerfUserDepartmentScopeMatches(user, target)
}
