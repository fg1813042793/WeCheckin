package review

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"wecheckin/backend/internal/model"
	reviewnotification "wecheckin/backend/internal/service/dingtalkh5/performance/review/notification"
	reviewscope "wecheckin/backend/internal/service/dingtalkh5/performance/review/scope"
	"wecheckin/backend/pkg/database"
)

func SubmitManagerContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
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
		review.Status = ReviewStatusHRBPReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, "提交上级评价")
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
}

func SubmitHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能提交 HRBP 评价")
		}
		copyHrbpFields(review, payload)
		if review.HRBPGrade == "" {
			return fmt.Errorf("请先选择 HRBP绩效分档")
		}
		if review.ManagerGrade == "" {
			return fmt.Errorf("上级绩效分档为空，不能提交 HRBP 评价")
		}
		if review.HRBPGrade != review.ManagerGrade {
			return fmt.Errorf("HRBP绩效分档与上级绩效分档不一致，不能提交给员工确认")
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
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, "提交 HRBP 评价，进入员工确认")
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
}

func ConfirmResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
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
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, action)
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
}

func DisputeResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
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
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, "员工提出异议："+review.EmployeeConfirmComment)
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
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

func WithdrawContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		action := ""
		switch {
		case review.Status == ReviewStatusManagerReview && review.EmployeeAccount == user.Account:
			if managerReviewStarted(*review) {
				return fmt.Errorf("上级已评价，不能撤回")
			}
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
		reason := normalizeReviewReason(payload.ReturnReason)
		if reason == "" {
			return fmt.Errorf("请填写撤回原因")
		}
		action += "：" + reason
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
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能退回上级修改")
		}
		copyHrbpFields(review, payload)
		review.Status = ReviewStatusManagerReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, "退回上级修改："+returnReason(payload))
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
}

func ReturnHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var transitionedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
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
		transitionedReview = *review
		return addHistoryWithDB(db, review, user, "退回 HRBP 修改："+returnReason(payload))
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)
	return result, nil
}

func DeleteReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) error {
	if user == nil {
		return fmt.Errorf("未登录")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var review model.DingTalkH5PerfReview
	if err := notDeletedReviewQuery(db).Where("review_no = ?", reviewNo).First(&review).Error; err != nil {
		return fmt.Errorf("没有找到这张考评单")
	}
	visible, err := reviewscope.InDataScopeContext(ctx, db, user, review)
	if err != nil {
		return err
	}
	if !visible {
		return fmt.Errorf("没有找到这张考评单")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		now := database.Now()
		audit := dingtalkH5AuditMetaForUserContext(ctx, tx, user, now)
		applyDingTalkH5DeleteAudit(&review.DingTalkH5AuditFields, audit)
		review.EditTime = now
		updates := dingtalkH5DeleteAuditUpdateValues(audit)
		updates["edit_time"] = now
		if err := tx.Model(&model.DingTalkH5PerfReview{}).
			Where("`id` = ? AND `deleted_at` = 0", review.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		return addHistoryWithDB(tx, &review, user, "删除考评单")
	})
}
