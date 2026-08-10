package review

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"wecheckin/backend/internal/model"
	reviewnotification "wecheckin/backend/internal/service/dingtalkh5/performance/review/notification"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

func CreateReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload) (*ReviewDTO, error) {
	resp, err := CreateReviewsContext(ctx, user, payload)
	if err != nil {
		return nil, err
	}
	if len(resp.List) == 0 {
		if len(resp.Failed) > 0 {
			return nil, fmt.Errorf("%s", resp.Failed[0].Message)
		}
		return nil, fmt.Errorf("考评单创建失败")
	}
	return &resp.List[0], nil
}

func CreateReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload) (*CreateReviewBatchResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("未登录")
	}
	period := strings.TrimSpace(payload.Period)
	nextPeriod := strings.TrimSpace(payload.NextPeriod)
	if !validMonth(period) || !validMonth(nextPeriod) {
		return nil, fmt.Errorf("月份格式应为 YYYY-MM")
	}
	employeeAccounts := reviewPayloadEmployeeIDs(payload)
	if len(employeeAccounts) == 0 {
		employeeAccounts = []string{NormalizeUserID(user.Account)}
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	tpl, err := LoadTemplateContext(ctx)
	if err != nil {
		return nil, err
	}
	accessScope, err := createReviewAccessScopeContext(ctx, db, user)
	if err != nil {
		return nil, err
	}

	result := &CreateReviewBatchResponse{
		List:   make([]ReviewDTO, 0, len(employeeAccounts)),
		Failed: make([]CreateReviewFailure, 0),
	}
	for _, employeeAccount := range employeeAccounts {
		dto, err := createReviewForEmployeeContext(ctx, db, user, accessScope, employeeAccount, period, nextPeriod, tpl)
		if err != nil {
			result.Failed = append(result.Failed, CreateReviewFailure{EmployeeID: employeeAccount, Message: err.Error()})
			continue
		}
		result.List = append(result.List, *dto)
	}
	result.Total = len(result.List)
	if result.Total == 0 {
		messages := make([]string, 0, len(result.Failed))
		for _, item := range result.Failed {
			messages = append(messages, item.Message)
		}
		if len(messages) == 0 {
			return nil, fmt.Errorf("请选择被考评人")
		}
		return nil, fmt.Errorf("%s", strings.Join(messages, "；"))
	}
	return result, nil
}

type createReviewAccessScope struct {
	allowed map[string]struct{}
	all     bool
}

func createReviewAccessScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (createReviewAccessScope, error) {
	if user == nil {
		return createReviewAccessScope{allowed: map[string]struct{}{}}, nil
	}
	scope, err := permissionsupport.DataScopeContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return createReviewAccessScope{}, err
	}
	allowed, all, err := usersvc.DataScopeUserAccountsContext(ctx, db, user, scope)
	if err != nil {
		return createReviewAccessScope{}, err
	}
	if allowed == nil {
		allowed = map[string]struct{}{}
	}
	return createReviewAccessScope{allowed: allowed, all: all}, nil
}

func (scope createReviewAccessScope) canAccess(account string) bool {
	if scope.all {
		return true
	}
	_, ok := scope.allowed[NormalizeUserID(account)]
	return ok
}

func reviewPayloadEmployeeIDs(payload ReviewPayload) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(payload.EmployeeIDs)+1)
	add := func(value string) {
		account := NormalizeUserID(value)
		if account == "" || seen[account] {
			return
		}
		seen[account] = true
		result = append(result, account)
	}
	for _, id := range payload.EmployeeIDs {
		add(id)
	}
	add(payload.EmployeeID)
	return result
}

func createReviewForEmployeeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, accessScope createReviewAccessScope, employeeAccount, period, nextPeriod string, tpl TemplateDTO) (*ReviewDTO, error) {
	employee, err := usersvc.LoadPerfUserByAccountDB(db, employeeAccount)
	if err != nil || !canBeReviewed(*employee) {
		return nil, fmt.Errorf("请选择有效被考评人")
	}
	if !accessScope.canAccess(employee.Account) {
		return nil, fmt.Errorf("请选择有效被考评人")
	}
	reviewNo := employee.Account + "-" + period
	var count int64
	if err := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{})).Where("review_no = ?", reviewNo).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("该员工这个月份已经有考评单")
	}
	now := database.Now()
	ownerAudit := dingtalkH5AuditMetaForUserContext(ctx, db, employee, now)
	operatorAudit := dingtalkH5AuditMetaForUserContext(ctx, db, user, now)
	previousReview, err := loadPreviousNextObjectivesForCreate(ctx, db, employee.Account, period)
	if err != nil {
		return nil, err
	}
	objectives, objectiveSource := currentObjectivesForNewReview(reviewNo, tpl.ObjectiveDefaults, previousReview)
	review := model.DingTalkH5PerfReview{
		ReviewNo:                reviewNo,
		EmployeeAccount:         employee.Account,
		ManagerAccount:          employee.ManagerAccount,
		HRBPAccount:             fallback(employee.HRBPAccount, "hrbp"),
		Department:              usersvc.DepartmentFromUser(*employee),
		DepartmentLevel1:        employee.DepartmentLevel1,
		DepartmentLevel2:        employee.DepartmentLevel2,
		DepartmentLevel3:        employee.DepartmentLevel3,
		Period:                  period,
		NextPeriod:              nextPeriod,
		Status:                  ReviewStatusDraft,
		ObjectiveSourceReviewNo: objectiveSource.reviewNo,
		ObjectiveSourcePeriod:   objectiveSource.period,
		ObjectivesJSON:          encodeJSON(objectives),
		NextObjectivesJSON:      encodeJSON(defaultNextObjectives(tpl.NextObjectiveDefaults, reviewNo)),
		ValuesJSON:              encodeJSON(defaultValues(tpl.Values)),
		AddTime:                 now,
		EditTime:                now,
	}
	applyDingTalkH5CreateAudit(&review.DingTalkH5AuditFields, ownerAudit)
	applyDingTalkH5UpdateAudit(&review.DingTalkH5AuditFields, operatorAudit)
	if err := db.Create(&review).Error; err != nil {
		return nil, err
	}
	if err := addHistoryWithDB(db, &review, user, "创建考评单"); err != nil {
		return nil, err
	}
	histories, _ := historiesForReview(ctx, review.ID)
	dto := reviewDTO(review, histories)
	hydrateReviewDTOValues(&dto, tpl.Values)
	reviewnotification.TransitionAsync(ctx, review, user, reviewnotification.EventFlowMoved)
	return &dto, nil
}

func loadPreviousNextObjectivesForCreate(ctx context.Context, db *gorm.DB, employeeAccount, period string) (*model.DingTalkH5PerfReview, error) {
	if db == nil {
		return nil, nil
	}
	var review model.DingTalkH5PerfReview
	err := db.WithContext(ctx).
		Scopes(func(tx *gorm.DB) *gorm.DB { return notDeletedReviewQuery(tx) }).
		Where("`employee_account` = ? AND `next_period` = ?", strings.TrimSpace(employeeAccount), strings.TrimSpace(period)).
		Order("`period` DESC, `id` DESC").
		First(&review).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(sanitizeNextObjectives(decodeNextObjectives(review.NextObjectivesJSON))) == 0 {
		return nil, nil
	}
	return &review, nil
}
