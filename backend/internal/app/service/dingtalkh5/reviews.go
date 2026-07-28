package dingtalkh5

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/app/support/appmenuperm"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func BootstrapContext(ctx context.Context, user *model.DingTalkH5PerfUser) (*BootstrapResponse, error) {
	return &BootstrapResponse{User: userDTO(*user), Menus: DingTalkH5MenusForUserContext(ctx, user)}, nil
}

func DingTalkH5MenusForUserContext(ctx context.Context, user *model.DingTalkH5PerfUser) []AppMenuDTO {
	if user == nil {
		return nil
	}
	if user.RoleID == 0 {
		return dingTalkH5DefaultMenusByRole(user.Role)
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if keys, ready, err := permissionsupport.DingTalkH5MenuPermissionKeysContext(ctx, db, user.ID, user.RoleID); err == nil && ready && len(keys) > 0 {
		return dingTalkH5MenusByKeys(keys)
	}
	return dingTalkH5DefaultMenusByRole(user.Role)
}

func dingTalkH5MenusByKeys(keys []string) []AppMenuDTO {
	allowed := map[string]bool{}
	for _, key := range keys {
		allowed[key] = true
	}
	expandLegacyDingTalkH5MenuKeys(allowed)
	menus := make([]AppMenuDTO, 0)
	for _, declaration := range appmenuperm.DingTalkH5MenuDeclarations() {
		if !allowed[declaration.Key] {
			continue
		}
		menus = append(menus, AppMenuDTO{
			Key:           declaration.Path,
			Label:         declaration.Name,
			Icon:          declaration.Path,
			PermissionKey: declaration.Key,
		})
	}
	return menus
}

func expandLegacyDingTalkH5MenuKeys(allowed map[string]bool) {
	legacyMap := map[string]string{
		"dingtalk_h5:menu:mine":     "dingtalk_h5:menu:performance:mine",
		"dingtalk_h5:menu:manager":  "dingtalk_h5:menu:performance:mine",
		"dingtalk_h5:menu:hrbp":     "dingtalk_h5:menu:performance:hrbp",
		"dingtalk_h5:menu:summary":  "dingtalk_h5:menu:performance:summary",
		"dingtalk_h5:menu:org":      "dingtalk_h5:menu:performance:org",
		"dingtalk_h5:menu:template": "dingtalk_h5:menu:performance:template",
	}
	for legacy, current := range legacyMap {
		if allowed[legacy] {
			allowed[current] = true
		}
	}
	for key := range allowed {
		if strings.HasPrefix(key, "dingtalk_h5:menu:performance:") {
			allowed["dingtalk_h5:menu:performance"] = true
			return
		}
	}
}

func dingTalkH5DefaultMenusByRole(role string) []AppMenuDTO {
	items := map[string][][2]string{
		"employee": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
		},
		"supervisor": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
			{"performance:summary", "HRBP汇总"},
		},
		"manager": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
			{"performance:summary", "HRBP汇总"},
		},
		"director": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
			{"performance:summary", "HRBP汇总"},
		},
		"hrbp": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
			{"performance:hrbp", "HRBP评价"},
			{"performance:summary", "HRBP汇总"},
		},
		"hrbp_manager": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:mine", "本月绩效"},
			{"performance:history", "历史绩效"},
			{"performance:hrbp", "HRBP评价"},
			{"performance:summary", "HRBP汇总"},
		},
		"admin": {
			{"dashboard", "工作台"},
			{"performance", "绩效管理"},
			{"performance:hrbp", "HRBP评价"},
			{"performance:summary", "HRBP汇总"},
			{"performance:org", "流程执行"},
			{"performance:template", "绩效模版"},
		},
	}
	raw, ok := items[role]
	if !ok {
		raw = items["employee"]
	}
	menus := make([]AppMenuDTO, 0, len(raw))
	for _, item := range raw {
		menus = append(menus, AppMenuDTO{
			Key:           item[0],
			Label:         item[1],
			Icon:          item[0],
			PermissionKey: "dingtalk_h5:menu:" + item[0],
		})
	}
	return menus
}

func TemplateContext(ctx context.Context) (TemplateDTO, error) {
	return LoadTemplateContext(ctx)
}

func ListReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters) ([]ReviewDTO, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var reviews []model.DingTalkH5PerfReview
	query := db.Order("period DESC, id DESC")
	if filters.Period != "" {
		query = query.Where("period = ?", filters.Period)
	}
	if filters.NextPeriod != "" {
		query = query.Where("next_period = ?", filters.NextPeriod)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
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
	query = applyReviewVisibilityScope(query, user, filters.Scope)
	if err := query.Find(&reviews).Error; err != nil {
		return nil, err
	}
	employees, err := usersByAccounts(ctx, collectEmployeeAccounts(reviews))
	if err != nil {
		return nil, err
	}
	result := make([]ReviewDTO, 0, len(reviews))
	for _, review := range reviews {
		employee := employees[review.EmployeeAccount]
		if !canViewReview(user, review, employee) || !matchesKeyword(review, employee, filters.Keyword) {
			continue
		}
		histories, err := historiesForReview(ctx, review.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, reviewDTO(review, histories))
	}
	return result, nil
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
	return &dto, nil
}

func CreateReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload) (*ReviewDTO, error) {
	if !canCreateReview(user) {
		return nil, fmt.Errorf("当前账号不能创建考评单")
	}
	period := strings.TrimSpace(payload.Period)
	nextPeriod := strings.TrimSpace(payload.NextPeriod)
	if !validMonth(period) || !validMonth(nextPeriod) {
		return nil, fmt.Errorf("月份格式应为 YYYY-MM")
	}
	employeeAccount := NormalizeUserID(payload.EmployeeID)
	if employeeAccount == "" {
		return nil, fmt.Errorf("请选择被考评人")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	employee, err := loadPerfUserByAccountDB(db, employeeAccount)
	if err != nil || !canBeReviewed(*employee) {
		return nil, fmt.Errorf("请选择有效被考评人")
	}
	reviewNo := employee.Account + "-" + period
	var count int64
	if err := db.Model(&model.DingTalkH5PerfReview{}).Where("review_no = ?", reviewNo).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("该员工这个月份已经有考评单")
	}
	tpl, err := LoadTemplateContext(ctx)
	if err != nil {
		return nil, err
	}
	now := database.Now()
	review := model.DingTalkH5PerfReview{
		ReviewNo:           reviewNo,
		EmployeeAccount:    employee.Account,
		ManagerAccount:     employee.ManagerAccount,
		HRBPAccount:        fallback(employee.HRBPAccount, "hrbp"),
		Department:         departmentFromUser(*employee),
		DepartmentLevel1:   employee.DepartmentLevel1,
		DepartmentLevel2:   employee.DepartmentLevel2,
		DepartmentLevel3:   employee.DepartmentLevel3,
		Period:             period,
		NextPeriod:         nextPeriod,
		Status:             ReviewStatusDraft,
		ObjectivesJSON:     encodeJSON(defaultObjectives(tpl.ObjectiveDefaults, reviewNo)),
		NextObjectivesJSON: encodeJSON(defaultNextObjectives(tpl.NextObjectiveDefaults, reviewNo)),
		ValuesJSON:         encodeJSON(defaultValues(tpl.Values)),
		AddTime:            now,
		EditTime:           now,
	}
	if err := db.Create(&review).Error; err != nil {
		return nil, err
	}
	if err := addHistoryWithDB(db, &review, user, "创建考评单"); err != nil {
		return nil, err
	}
	histories, _ := historiesForReview(ctx, review.ID)
	dto := reviewDTO(review, histories)
	return &dto, nil
}

func SaveSelfContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能修改员工自评")
		}
		copySelfFields(review, payload)
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "保存员工自评")
	})
}

func SubmitSelfContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能提交员工自评")
		}
		copySelfFields(review, payload)
		review.Status = nextStatusAfterSelfSubmit(DingTalkH5Review{EmployeeID: review.EmployeeAccount, ManagerID: review.ManagerAccount, HRBPID: review.HRBPAccount, Status: review.Status})
		if review.Status == ReviewStatusHRFinal {
			review.HRBPReviewerAccount = review.HRBPAccount
		}
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		action := "提交员工自评"
		if review.Status == ReviewStatusHRFinal {
			action = "提交员工自评，进入 HRBP 归档"
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func SubmitManagerContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.ManagerAccount != user.Account || review.Status != ReviewStatusManagerReview {
			return fmt.Errorf("当前阶段不能提交上级评价")
		}
		copyManagerFields(review, payload)
		if review.ManagerGrade == "" {
			return fmt.Errorf("请先选择绩效分档")
		}
		if strings.TrimSpace(review.ManagerComment) == "" {
			return fmt.Errorf("请先填写上级评价")
		}
		if !allStageScoresFilled(decodeValues(review.ValuesJSON), "manager") {
			return fmt.Errorf("请先填写上级价值观评分")
		}
		if shouldSkipHrbpStage(ctx, *review) {
			review.Status = ReviewStatusEmployeeConfirm
			review.HRBPReviewerAccount = fallback(review.HRBPAccount, user.Account)
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmComment = ""
			review.EmployeeConfirmedAt = 0
			if err := db.Save(review).Error; err != nil {
				return err
			}
			return addHistoryWithDB(db, review, user, "提交上级评价，进入员工确认")
		}
		review.Status = ReviewStatusHRBPReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "提交上级评价")
	})
}

func SubmitHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能提交 HRBP 评价")
		}
		copyHrbpFields(review, payload)
		if review.HRBPGrade == "" {
			return fmt.Errorf("请先选择 HRBP绩效分档")
		}
		if review.ManagerGrade != "" && review.HRBPGrade != review.ManagerGrade {
			return fmt.Errorf("HRBP绩效分档与上级绩效分档不一致，不能提交至归档，请先退回上级调整或沟通确认一致")
		}
		if strings.TrimSpace(review.HRBPComment) == "" {
			return fmt.Errorf("请先填写 HRBP 评价")
		}
		if !allStageScoresFilled(decodeValues(review.ValuesJSON), "hrbp") {
			return fmt.Errorf("请先填写 HRBP 价值观评分")
		}
		review.Status = ReviewStatusEmployeeConfirm
		review.HRBPReviewerAccount = user.Account
		review.EmployeeConfirmResult = ""
		review.EmployeeConfirmComment = ""
		review.EmployeeConfirmedAt = 0
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "提交 HRBP 评价，进入员工确认")
	})
}

func ConfirmResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusEmployeeConfirm {
			return fmt.Errorf("当前阶段不能确认绩效结果")
		}
		review.EmployeeConfirmComment = strings.TrimSpace(payload.EmployeeConfirmComment)
		review.EmployeeConfirmResult = "confirmed"
		review.EmployeeConfirmedAt = database.Now()
		review.Status = ReviewStatusHRFinal
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		action := "员工确认结果"
		if review.EmployeeConfirmComment != "" {
			action = "员工确认结果：" + review.EmployeeConfirmComment
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func DisputeResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusEmployeeConfirm {
			return fmt.Errorf("当前阶段不能提出异议")
		}
		review.EmployeeConfirmComment = strings.TrimSpace(payload.EmployeeConfirmComment)
		if review.EmployeeConfirmComment == "" {
			return fmt.Errorf("请填写异议原因")
		}
		review.EmployeeConfirmResult = "disputed"
		review.EmployeeConfirmedAt = database.Now()
		review.Status = ReviewStatusHRBPReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "员工提出异议："+review.EmployeeConfirmComment)
	})
}

func FinalizeContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !canHandleHrbpFinal(user, *review) || (review.Status != ReviewStatusHRFinal && review.Status != ReviewStatusCompleted) {
			return fmt.Errorf("当前阶段不能归档")
		}
		review.FinalGrade = strings.TrimSpace(fallback(payload.FinalGrade, review.HRBPGrade))
		review.FinalNote = strings.TrimSpace(payload.FinalNote)
		if review.FinalGrade == "" {
			return fmt.Errorf("请先选择最终分档")
		}
		review.Status = ReviewStatusCompleted
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "HRBP 归档")
	})
}

func WithdrawContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		action := ""
		switch {
		case review.Status == ReviewStatusManagerReview && review.EmployeeAccount == user.Account:
			review.Status = ReviewStatusDraft
			action = "撤销员工自评提交"
		case review.Status == ReviewStatusHRBPReview && review.ManagerAccount == user.Account:
			review.Status = ReviewStatusManagerReview
			review.HRBPReviewerAccount = ""
			action = "撤销上级评价提交"
		case review.Status == ReviewStatusEmployeeConfirm && isHrbpReviewer(user, *review):
			review.Status = ReviewStatusHRBPReview
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmComment = ""
			review.EmployeeConfirmedAt = 0
			action = "撤销 HRBP 评价提交"
		case review.Status == ReviewStatusHRFinal && review.EmployeeAccount == user.Account && review.FinalGrade == "":
			review.Status = ReviewStatusEmployeeConfirm
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmedAt = 0
			action = "撤销员工确认"
		case review.Status == ReviewStatusHRFinal && canHandleHrbpFinal(user, *review) && review.FinalGrade == "":
			review.Status = ReviewStatusHRBPReview
			action = "撤销 HRBP 评价提交"
		default:
			return fmt.Errorf("当前阶段不能撤销提交")
		}
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func ReturnEmployeeContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return returnReview(ctx, user, reviewNo, ReviewStatusManagerReview, ReviewStatusDraft, "退回员工修改："+returnReason(payload))
}

func ReturnManagerContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能退回上级修改")
		}
		copyHrbpFields(review, payload)
		review.Status = ReviewStatusManagerReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "退回上级修改："+returnReason(payload))
	})
}

func ReturnHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !canHandleHrbpFinal(user, *review) || review.Status != ReviewStatusHRFinal {
			return fmt.Errorf("当前阶段不能退回 HRBP 修改")
		}
		review.FinalGrade = strings.TrimSpace(payload.FinalGrade)
		review.FinalNote = strings.TrimSpace(payload.FinalNote)
		review.Status = ReviewStatusHRBPReview
		review.EmployeeConfirmResult = ""
		review.EmployeeConfirmComment = ""
		review.EmployeeConfirmedAt = 0
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "退回 HRBP 修改："+returnReason(payload))
	})
}

func DeleteReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) error {
	if !isAdmin(user) {
		return fmt.Errorf("当前账号不能删除考评单")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var review model.DingTalkH5PerfReview
	if err := db.Where("review_no = ?", reviewNo).First(&review).Error; err != nil {
		return fmt.Errorf("没有找到这张考评单")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("review_id = ?", review.ID).Delete(&model.DingTalkH5PerfHistory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&review).Error
	})
}
