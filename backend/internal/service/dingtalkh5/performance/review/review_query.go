package review

import (
	"context"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
	"wecheckin/backend/internal/model"
	reviewscope "wecheckin/backend/internal/service/dingtalkh5/performance/review/scope"
	"wecheckin/backend/pkg/database"
)

func ListReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters) (*ReviewListResponse, error) {
	normalizeReviewPagination(&filters)
	return listReviewsContext(ctx, user, filters, true)
}

func listReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters, paginate bool) (*ReviewListResponse, error) {
	enforceSummaryCompletedReviewFilter(&filters)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var reviews []model.DingTalkH5PerfReview
	query := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{}))
	if filters.Period != "" {
		query = query.Where("period = ?", filters.Period)
	} else {
		year := normalizeReviewYear(filters.Year)
		month := normalizeReviewMonth(filters.Month)
		if year != "" && month != "" {
			query = query.Where("period = ?", year+"-"+month)
		} else if year != "" {
			query = query.Where("period LIKE ?", year+"-%")
		} else if month != "" {
			query = query.Where("period LIKE ?", "%-"+month)
		}
	}
	if filters.NotPeriod != "" {
		query = query.Where("period <> ?", filters.NotPeriod)
	}
	if filters.NextPeriod != "" {
		query = query.Where("next_period = ?", filters.NextPeriod)
	}
	statuses := normalizeReviewStatuses(filters.Status, filters.Statuses)
	if len(statuses) == 1 {
		query = query.Where("status = ?", statuses[0])
	} else if len(statuses) > 1 {
		query = query.Where("status IN ?", statuses)
	}
	if filters.EmployeeName != "" {
		query = applyReviewEmployeeNameQuery(query, filters.EmployeeName)
	}
	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
	}
	if len(filters.DepartmentNames) > 0 {
		query = applyReviewDepartmentNamesQuery(query, filters.DepartmentNames)
	} else if filters.DepartmentName != "" {
		query = query.Where("department LIKE ?", "%"+filters.DepartmentName+"%")
	}
	if filters.ManagerID != "" {
		query = query.Where("manager_account = ?", filters.ManagerID)
	}
	if filters.HRBPID != "" {
		query = query.Where("hrbp_account = ?", filters.HRBPID)
	}
	if filters.Grade != "" {
		query = query.Where("(final_grade = ? OR (final_grade = '' AND hrbp_grade = ?))", filters.Grade, filters.Grade)
	}
	query = applyReviewKeywordQuery(query, filters.Keyword)
	query, err := reviewscope.ApplyVisibilityScopeContext(ctx, db, query, user, filters.Scope)
	if err != nil {
		return nil, err
	}
	var total int64
	query = query.Order("period DESC, id DESC")
	if score, ok := normalizeReviewObjectiveScoreFilter(filters.ObjectiveScore); ok {
		if err := query.Find(&reviews).Error; err != nil {
			return nil, err
		}
		reviews = filterReviewsByObjectiveScore(reviews, score)
		total = int64(len(reviews))
		reviews = paginateReviewRows(reviews, filters, paginate)
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, err
		}
		if paginate {
			query = query.Offset((filters.Page - 1) * filters.PageSize).Limit(filters.PageSize)
		}
		if err := query.Find(&reviews).Error; err != nil {
			return nil, err
		}
	}
	if len(reviews) == 0 {
		return &ReviewListResponse{List: []ReviewDTO{}, Total: total, Page: filters.Page, PageSize: filters.PageSize}, nil
	}
	participants, err := usersByAccounts(ctx, collectReviewParticipantAccounts(reviews))
	if err != nil {
		return nil, err
	}
	historiesByID := map[uint][]model.DingTalkH5PerfHistory{}
	latestHistoriesByID := map[uint]model.DingTalkH5PerfHistory{}
	reviewIDs := collectReviewIDs(reviews)
	if !filters.SkipHistory {
		var err error
		historiesByID, err = historiesByReviewIDs(ctx, reviewIDs)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		latestHistoriesByID, err = latestHistoriesByReviewIDs(ctx, reviewIDs)
		if err != nil {
			return nil, err
		}
	}
	valueTemplates := loadReviewValueTemplatesContext(ctx)
	result := make([]ReviewDTO, 0, len(reviews))
	for _, review := range reviews {
		dto := reviewDTOWithUsers(review, historiesByID[review.ID], participants)
		if latestHistory, ok := latestHistoriesByID[review.ID]; ok {
			dto.LatestAction = latestHistory.Action
		}
		hydrateReviewDTOValues(&dto, valueTemplates)
		result = append(result, dto)
	}
	return &ReviewListResponse{List: result, Total: total, Page: filters.Page, PageSize: filters.PageSize}, nil
}

func normalizeReviewStatuses(status string, statuses []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(statuses)+1)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		result = append(result, value)
	}
	add(status)
	for _, item := range statuses {
		add(item)
	}
	return result
}

func enforceSummaryCompletedReviewFilter(filters *ReviewFilters) {
	if filters == nil || reviewscope.NormalizeReviewScope(filters.Scope) != reviewscope.ScopeSummary {
		return
	}
	filters.Scope = reviewscope.ScopeSummary
	filters.Status = ReviewStatusCompleted
	filters.Statuses = nil
}

func normalizeReviewYear(year string) string {
	year = strings.TrimSpace(year)
	if len(year) != 4 {
		return ""
	}
	for _, ch := range year {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	return year
}

func normalizeReviewMonth(month string) string {
	month = strings.TrimSpace(month)
	if month == "" {
		return ""
	}
	if len(month) == 1 {
		month = "0" + month
	}
	if len(month) != 2 {
		return ""
	}
	if month < "01" || month > "12" {
		return ""
	}
	return month
}

func applyReviewDepartmentNamesQuery(query *gorm.DB, names []string) *gorm.DB {
	normalized := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		text := strings.TrimSpace(name)
		if text == "" || seen[text] {
			continue
		}
		seen[text] = true
		normalized = append(normalized, text)
	}
	if len(normalized) == 0 {
		return query
	}
	if len(normalized) == 1 {
		return query.Where("department LIKE ?", "%"+normalized[0]+"%")
	}
	parts := make([]string, 0, len(normalized))
	args := make([]interface{}, 0, len(normalized))
	for _, name := range normalized {
		parts = append(parts, "department LIKE ?")
		args = append(args, "%"+name+"%")
	}
	return query.Where("("+strings.Join(parts, " OR ")+")", args...)
}

func normalizeReviewPagination(filters *ReviewFilters) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
}

func normalizeReviewObjectiveScoreFilter(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	score, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return math.NaN(), true
	}
	return math.Round(score*10) / 10, true
}

func filterReviewsByObjectiveScore(reviews []model.DingTalkH5PerfReview, expected float64) []model.DingTalkH5PerfReview {
	if math.IsNaN(expected) || math.IsInf(expected, 0) {
		return []model.DingTalkH5PerfReview{}
	}
	result := make([]model.DingTalkH5PerfReview, 0, len(reviews))
	for _, review := range reviews {
		if math.Abs(objectiveTotal(decodeObjectives(review.ObjectivesJSON))-expected) <= 0.0001 {
			result = append(result, review)
		}
	}
	return result
}

func paginateReviewRows(reviews []model.DingTalkH5PerfReview, filters ReviewFilters, paginate bool) []model.DingTalkH5PerfReview {
	if !paginate || len(reviews) == 0 {
		return reviews
	}
	start := (filters.Page - 1) * filters.PageSize
	if start >= len(reviews) {
		return []model.DingTalkH5PerfReview{}
	}
	end := start + filters.PageSize
	if end > len(reviews) {
		end = len(reviews)
	}
	return reviews[start:end]
}

func applyReviewKeywordQuery(query *gorm.DB, keyword string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return query
	}
	likeKeyword := "%" + keyword + "%"
	return query.Where(
		"`review_no` LIKE ? OR `employee_account` LIKE ? OR `manager_account` LIKE ? OR `hrbp_account` LIKE ? OR `department` LIKE ? OR `period` LIKE ? OR `next_period` LIKE ? OR `status` LIKE ? OR `manager_grade` LIKE ? OR `hrbp_grade` LIKE ? OR `final_grade` LIKE ? OR `employee_account` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)",
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
	)
}

func applyReviewEmployeeNameQuery(query *gorm.DB, employeeName string) *gorm.DB {
	employeeName = strings.TrimSpace(employeeName)
	if employeeName == "" {
		return query
	}
	likeName := "%" + employeeName + "%"
	return query.Where(
		"`employee_account` LIKE ? OR `employee_account` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)",
		likeName,
		likeName,
	)
}

func GetReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) (*ReviewDTO, error) {
	review, err := findVisibleReview(ctx, user, reviewNo)
	if err != nil {
		return nil, err
	}
	histories, err := historiesForReview(ctx, review.ID)
	if err != nil {
		return nil, err
	}
	dto := reviewDTO(*review, histories)
	hydrateReviewDTOValues(&dto, loadReviewValueTemplatesContext(ctx))
	return &dto, nil
}
