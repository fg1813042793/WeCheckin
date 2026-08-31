package review

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"wecheckin/backend/internal/model"
	reviewnotification "wecheckin/backend/internal/service/dingtalkh5/performance/review/notification"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

const (
	nextObjectiveEditButtonKey   = "dingtalk_h5:button:review:next_objective_edit"
	nextObjectiveAddButtonKey    = "dingtalk_h5:button:review:next_objective_add"
	nextObjectiveDeleteButtonKey = "dingtalk_h5:button:review:next_objective_delete"
)

func SaveSelfContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能修改员工自评")
		}
		if err := validateSelfObjectiveNumbers(payload); err != nil {
			return err
		}
		if err := ensureNextObjectiveMutationPermissionsContext(ctx, db, user, review, payload); err != nil {
			return err
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
	var submittedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能提交员工自评")
		}
		if err := validateSelfObjectiveNumbers(payload); err != nil {
			return err
		}
		if err := validateSelfSubmitPayload(payload); err != nil {
			return err
		}
		if err := ensureNextObjectiveMutationPermissionsContext(ctx, db, user, review, payload); err != nil {
			return err
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
		submittedReview = *review
		return addHistoryWithDB(db, review, user, action)
	})
	if err != nil {
		return nil, err
	}
	reviewnotification.TransitionAsync(ctx, submittedReview, user, reviewnotification.EventSelfSubmitted)
	return result, nil
}

type nextObjectiveMutationSet struct {
	add    bool
	edit   bool
	delete bool
}

func ensureNextObjectiveMutationPermissionsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review *model.DingTalkH5PerfReview, payload ReviewPayload) error {
	changes := nextObjectiveMutations(
		sanitizeNextObjectives(decodeNextObjectives(review.NextObjectivesJSON)),
		sanitizeNextObjectives(payload.NextObjectives),
	)
	if !changes.add && !changes.edit && !changes.delete {
		return nil
	}
	if changes.edit {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveEditButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限编辑下月目标")
		}
	}
	if changes.add {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveAddButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限新增下月目标")
		}
	}
	if changes.delete {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveDeleteButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限删除下月目标")
		}
	}
	return nil
}

func nextObjectiveMutations(existing []NextObjective, incoming []NextObjective) nextObjectiveMutationSet {
	changes := nextObjectiveMutationSet{}
	existingByID := make(map[string]NextObjective, len(existing))
	for _, item := range existing {
		if key := strings.TrimSpace(item.ID); key != "" {
			existingByID[key] = item
		}
	}
	incomingByID := make(map[string]NextObjective, len(incoming))
	for _, item := range incoming {
		key := strings.TrimSpace(item.ID)
		if key == "" {
			changes.add = true
			continue
		}
		incomingByID[key] = item
		old, ok := existingByID[key]
		if !ok {
			changes.add = true
			continue
		}
		if strings.TrimSpace(old.Target) != strings.TrimSpace(item.Target) || old.Weight != item.Weight {
			changes.edit = true
		}
	}
	for _, item := range existing {
		key := strings.TrimSpace(item.ID)
		if key == "" {
			continue
		}
		if _, ok := incomingByID[key]; !ok {
			changes.delete = true
		}
	}
	return changes
}

func subjectHasDingTalkH5ButtonPermissionContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, key string) (bool, error) {
	if user == nil || db == nil {
		return false, nil
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return false, err
	}
	return permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, user.ID, roleIDs, key)
}
