package dingtalkh5

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

const (
	dingTalkH5NotifyEnabledKey         = "DINGTALK_H5_NOTIFY_ENABLED"
	dingtalkH5NotifyEventSelfSubmitted = "self_submitted"
	dingTalkH5NotifyTimeout            = 5 * time.Second
)

type dingTalkH5NotifyRecipient struct {
	Config         DingTalkH5CorpConfig
	DingTalkUserID string
}

func DingTalkH5NotificationEnabledContext(ctx context.Context) bool {
	value := strings.ToLower(strings.TrimSpace(dingTalkH5SetupValueContext(ctx, dingTalkH5NotifyEnabledKey)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func notifyReviewTransitionAsync(ctx context.Context, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, event string) {
	if !DingTalkH5NotificationEnabledContext(ctx) {
		return
	}
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), dingTalkH5NotifyTimeout)
		defer cancel()
		if err := sendReviewTransitionNotificationContext(bg, defaultDingTalkIdentityClient{}, review, actor, event); err != nil {
			logger.Logger.Printf("[DingTalkH5Notify] event=%s reviewNo=%s err=%v", event, review.ReviewNo, err)
		}
	}()
}

func sendReviewTransitionNotificationContext(ctx context.Context, client DingTalkWorkNotificationClient, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, event string) error {
	if client == nil {
		return fmt.Errorf("钉钉通知客户端不能为空")
	}
	switch event {
	case dingtalkH5NotifyEventSelfSubmitted:
		if review.Status != ReviewStatusManagerReview {
			return nil
		}
		if strings.TrimSpace(review.ManagerAccount) == "" {
			return fmt.Errorf("manager_account 为空，无法发送上级通知")
		}
		recipient, err := resolveDingTalkH5NotifyRecipientContext(ctx, review.EmployeeAccount, review.ManagerAccount)
		if err != nil {
			return err
		}
		return client.SendWorkNotificationContext(ctx, recipient.Config, []string{recipient.DingTalkUserID}, buildSelfSubmitNotificationContent(review, actor))
	default:
		return nil
	}
}

func resolveDingTalkH5NotifyRecipientContext(ctx context.Context, employeeAccount, managerAccount string) (dingTalkH5NotifyRecipient, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("database is not initialized")
	}
	managerID, err := perfUserIDByAccountDB(db, managerAccount)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("直属上级未绑定有效用户：%w", err)
	}
	corpID := ""
	if employeeID, err := perfUserIDByAccountDB(db, employeeAccount); err == nil {
		corpID = firstEnabledDingTalkH5CorpIDByUserIDDB(db, employeeID)
	}
	binding, err := firstEnabledDingTalkH5BindingByUserIDDB(db, managerID, corpID)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("直属上级未绑定钉钉账号：%w", err)
	}
	config, err := loadDingTalkH5CorpConfigContext(ctx, binding.CorpID)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, err
	}
	if strings.TrimSpace(config.AgentID) == "" {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("请先配置钉钉内部应用 AgentId")
	}
	return dingTalkH5NotifyRecipient{Config: config, DingTalkUserID: binding.DingTalkUserID}, nil
}

func buildSelfSubmitNotificationContent(review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser) string {
	employeeName := strings.TrimSpace(review.EmployeeAccount)
	if actor != nil && strings.TrimSpace(actor.Name) != "" {
		employeeName = strings.TrimSpace(actor.Name)
	}
	period := strings.TrimSpace(review.Period)
	if period == "" {
		return fmt.Sprintf("%s 已提交月度考评，请处理上级评价。", employeeName)
	}
	return fmt.Sprintf("%s 已提交 %s 月度考评，请处理上级评价。", employeeName, period)
}

func perfUserIDByAccountDB(db *gorm.DB, account string) (uint, error) {
	account = NormalizeUserID(account)
	if account == "" {
		return 0, gorm.ErrRecordNotFound
	}
	var user model.DingTalkH5PerfUser
	if err := db.Select("`id`").Where("`user_mini_openid` = ? AND `user_status` = 1", account).Take(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func firstEnabledDingTalkH5CorpIDByUserIDDB(db *gorm.DB, userID uint) string {
	if userID == 0 {
		return ""
	}
	var binding model.DingTalkH5UserBinding
	err := db.Select("`corp_id`").
		Where("`user_id` = ? AND `enabled` = 1", userID).
		Order("`id` ASC").
		Take(&binding).Error
	if err != nil {
		return ""
	}
	return strings.TrimSpace(binding.CorpID)
}

func firstEnabledDingTalkH5BindingByUserIDDB(db *gorm.DB, userID uint, preferredCorpID string) (model.DingTalkH5UserBinding, error) {
	if userID == 0 {
		return model.DingTalkH5UserBinding{}, gorm.ErrRecordNotFound
	}
	preferredCorpID = strings.TrimSpace(preferredCorpID)
	if preferredCorpID != "" {
		var binding model.DingTalkH5UserBinding
		err := db.Where("`user_id` = ? AND `corp_id` = ? AND `enabled` = 1", userID, preferredCorpID).
			Order("`id` ASC").
			Take(&binding).Error
		if err == nil {
			return binding, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return model.DingTalkH5UserBinding{}, err
		}
	}
	var binding model.DingTalkH5UserBinding
	err := db.Where("`user_id` = ? AND `enabled` = 1", userID).
		Order("`id` ASC").
		Take(&binding).Error
	return binding, err
}
