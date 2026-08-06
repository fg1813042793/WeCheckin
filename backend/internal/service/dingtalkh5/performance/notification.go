package performance

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	dingtalkH5NotifyEventFlowMoved     = "flow_moved"
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
	nextAccount := reviewNextHandlerAccount(review)
	logger.Logger.Printf("[DingTalkH5Notify] schedule event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
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
	case dingtalkH5NotifyEventFlowMoved:
	default:
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=unsupported_event event=%s reviewNo=%s status=%s", event, review.ReviewNo, review.Status)
		return nil
	}
	nextAccount := reviewNextHandlerAccount(review)
	if nextAccount == "" {
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=no_next_handler event=%s reviewNo=%s status=%s", event, review.ReviewNo, review.Status)
		return nil
	}
	recipient, err := resolveDingTalkH5NotifyRecipientContext(ctx, review.EmployeeAccount, nextAccount)
	if err != nil {
		return err
	}
	if recipient.Config.NotifyEnabled == 0 {
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=disabled event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
		return nil
	}
	notification := buildReviewTransitionNotificationPayloadContext(ctx, recipient.Config, review, actor)
	logDingTalkH5NotifySendConfig(event, review, nextAccount, recipient, notification)
	if err := client.SendWorkNotificationContext(ctx, recipient.Config, []string{recipient.DingTalkUserID}, notification); err != nil {
		return err
	}
	logger.Logger.Printf("[DingTalkH5Notify] sent event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
	return nil
}

func logDingTalkH5NotifySendConfig(event string, review model.DingTalkH5PerfReview, nextAccount string, recipient dingTalkH5NotifyRecipient, notification DingTalkWorkNotificationPayload) {
	config := normalizeDingTalkH5CorpConfig(recipient.Config)
	logger.Logger.Printf(
		"[DingTalkH5Notify] send config event=%s reviewNo=%s status=%s employee=%s next=%s dingtalkUserId=%s corpId=%s corpName=%s notifyMode=%s notifyEnabled=%d appKey=%s appSecretSet=%t agentId=%s unifiedAppId=%s openAppId=%s robotCode=%s appURL=%s notificationURLSet=%t title=%s sourceName=%s hasPic=%t",
		event,
		review.ReviewNo,
		review.Status,
		NormalizeUserID(review.EmployeeAccount),
		NormalizeUserID(nextAccount),
		dingTalkH5LogValue(recipient.DingTalkUserID),
		dingTalkH5LogValue(config.CorpID),
		dingTalkH5LogValue(config.CorpName),
		config.NotifyMode,
		config.NotifyEnabled,
		dingTalkH5MaskLogValue(config.AppKey),
		config.AppSecretSet,
		dingTalkH5LogValue(config.AgentID),
		dingTalkH5LogValue(config.UnifiedAppID),
		dingTalkH5LogValue(dingtalkH5OpenAppID(config)),
		dingTalkH5MaskLogValue(config.RobotCode),
		dingTalkH5LogValue(config.AppURL),
		strings.TrimSpace(notification.URL) != "",
		dingTalkH5LogValue(notification.Title),
		dingTalkH5LogValue(notification.SourceName),
		strings.TrimSpace(notification.PicURL) != "",
	)
}

func dingTalkH5LogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func dingTalkH5MaskLogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func reviewNextHandlerAccount(review model.DingTalkH5PerfReview) string {
	switch review.Status {
	case ReviewStatusDraft:
		return NormalizeUserID(review.EmployeeAccount)
	case ReviewStatusManagerReview:
		return NormalizeUserID(review.ManagerAccount)
	case ReviewStatusHRBPReview, ReviewStatusHRFinal:
		return NormalizeUserID(fallback(review.HRBPReviewerAccount, review.HRBPAccount))
	case ReviewStatusEmployeeConfirm:
		return NormalizeUserID(review.EmployeeAccount)
	default:
		return ""
	}
}

func resolveDingTalkH5NotifyRecipientContext(ctx context.Context, employeeAccount, handlerAccount string) (dingTalkH5NotifyRecipient, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("database is not initialized")
	}
	handlerID, err := perfUserIDByAccountDB(db, handlerAccount)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("下一步处理人未绑定有效用户：%w", err)
	}
	corpID := ""
	if employeeID, err := perfUserIDByAccountDB(db, employeeAccount); err == nil {
		corpID = firstEnabledDingTalkH5CorpIDByUserIDDB(db, employeeID)
	}
	binding, err := firstEnabledDingTalkH5BindingByUserIDDB(db, handlerID, corpID)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("下一步处理人未绑定钉钉账号：%w", err)
	}
	config, err := loadDingTalkH5CorpConfigContext(ctx, binding.CorpID)
	if err != nil {
		return dingTalkH5NotifyRecipient{}, err
	}
	if normalizeDingTalkH5NotifyMode(config.NotifyMode, config.AgentID, config.RobotCode) != "robot" && strings.TrimSpace(config.AgentID) == "" {
		return dingTalkH5NotifyRecipient{}, fmt.Errorf("请先配置钉钉内部应用 AgentId")
	}
	return dingTalkH5NotifyRecipient{Config: config, DingTalkUserID: binding.DingTalkUserID}, nil
}

func buildReviewTransitionNotificationContent(review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser) string {
	employeeName := strings.TrimSpace(review.EmployeeAccount)
	actorName := ""
	if actor != nil && strings.TrimSpace(actor.Name) != "" {
		actorName = strings.TrimSpace(actor.Name)
		if NormalizeUserID(actor.Account) == NormalizeUserID(review.EmployeeAccount) {
			employeeName = actorName
		}
	} else if actor != nil && strings.TrimSpace(actor.Account) != "" {
		actorName = strings.TrimSpace(actor.Account)
	}
	period := strings.TrimSpace(review.Period)
	stage := reviewNotifyStageName(review.Status)
	if period == "" {
		if actorName == "" {
			return fmt.Sprintf("%s 的月度考评已流转到「%s」，请及时处理。", employeeName, stage)
		}
		return fmt.Sprintf("%s 已将 %s 的月度考评流转到「%s」，请及时处理。", actorName, employeeName, stage)
	}
	if actorName == "" {
		return fmt.Sprintf("%s 的 %s 月度考评已流转到「%s」，请及时处理。", employeeName, period, stage)
	}
	return fmt.Sprintf("%s 已将 %s 的 %s 月度考评流转到「%s」，请及时处理。", actorName, employeeName, period, stage)
}

func buildReviewTransitionNotificationPayloadContext(ctx context.Context, config DingTalkH5CorpConfig, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser) DingTalkWorkNotificationPayload {
	return DingTalkWorkNotificationPayload{
		Title:      buildReviewTransitionNotificationTitle(review),
		Content:    buildReviewTransitionNotificationContent(review, actor),
		URL:        buildReviewNotificationURL(dingTalkH5NotificationAppURLContext(ctx, config), config, review),
		SourceName: dingTalkH5NotificationAppNameContext(ctx),
		PicURL:     dingTalkH5NotificationLogoURLContext(ctx),
	}
}

func buildReviewTransitionNotificationTitle(review model.DingTalkH5PerfReview) string {
	period := strings.TrimSpace(review.Period)
	if period == "" {
		return "绩效流程待办"
	}
	return period + " 月度考评待办"
}

func dingTalkH5AppURLContext(ctx context.Context) string {
	return strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_URL"))
}

func dingTalkH5NotificationAppURLContext(ctx context.Context, config DingTalkH5CorpConfig) string {
	if appURL := strings.TrimSpace(config.AppURL); appURL != "" {
		return appURL
	}
	return dingTalkH5AppURLContext(ctx)
}

func dingTalkH5NotificationAppNameContext(ctx context.Context) string {
	if database.GetDB() == nil {
		return defaultDingTalkH5AppName
	}
	appName := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_NAME"))
	if appName == "" {
		return defaultDingTalkH5AppName
	}
	return appName
}

func dingTalkH5NotificationLogoURLContext(ctx context.Context) string {
	if database.GetDB() == nil {
		return ""
	}
	return strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_LOGO_URL"))
}

func buildReviewOperationURL(baseURL string, review model.DingTalkH5PerfReview) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	operationURL, err := url.Parse(baseURL)
	if err != nil {
		return appendReviewOperationQuery(baseURL, review)
	}
	query := operationURL.Query()
	query.Set("view", reviewOperationView(review.Status))
	if reviewNo := strings.TrimSpace(review.ReviewNo); reviewNo != "" {
		query.Set("reviewNo", reviewNo)
	}
	if period := strings.TrimSpace(review.Period); period != "" {
		query.Set("period", period)
	}
	if status := strings.TrimSpace(review.Status); status != "" {
		query.Set("status", status)
	}
	operationURL.RawQuery = query.Encode()
	return operationURL.String()
}

func buildReviewNotificationURL(baseURL string, config DingTalkH5CorpConfig, review model.DingTalkH5PerfReview) string {
	operationURL := buildReviewOperationURL(baseURL, review)
	if operationURL == "" {
		return ""
	}
	return wrapDingTalkH5OpenAppURL(operationURL, config)
}

func wrapDingTalkH5OpenAppURL(operationURL string, config DingTalkH5CorpConfig) string {
	operationURL = strings.TrimSpace(operationURL)
	if operationURL == "" {
		return ""
	}
	corpID := strings.TrimSpace(config.CorpID)
	appID := dingtalkH5OpenAppID(config)
	if corpID == "" || appID == "" {
		return operationURL
	}
	query := url.Values{}
	query.Set("corpid", corpID)
	query.Set("container_type", "work_platform")
	query.Set("app_id", appID)
	query.Set("redirect_type", "jump")
	query.Set("redirect_url", operationURL)
	openAppURL := url.URL{
		Scheme:   "dingtalk",
		Host:     "dingtalkclient",
		Path:     "/action/openapp",
		RawQuery: query.Encode(),
	}
	return openAppURL.String()
}

func dingtalkH5OpenAppID(config DingTalkH5CorpConfig) string {
	unifiedAppID := strings.TrimSpace(config.UnifiedAppID)
	if unifiedAppID != "" {
		return unifiedAppID
	}
	agentID := strings.TrimSpace(config.AgentID)
	if agentID != "" {
		if strings.HasPrefix(agentID, "0_") {
			return agentID
		}
		return "0_" + agentID
	}
	return unifiedAppID
}

func appendReviewOperationQuery(baseURL string, review model.DingTalkH5PerfReview) string {
	values := url.Values{}
	values.Set("view", reviewOperationView(review.Status))
	if reviewNo := strings.TrimSpace(review.ReviewNo); reviewNo != "" {
		values.Set("reviewNo", reviewNo)
	}
	if period := strings.TrimSpace(review.Period); period != "" {
		values.Set("period", period)
	}
	if status := strings.TrimSpace(review.Status); status != "" {
		values.Set("status", status)
	}
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}
	return baseURL + separator + values.Encode()
}

func reviewOperationView(status string) string {
	switch status {
	case ReviewStatusManagerReview:
		return "performance:manager"
	case ReviewStatusHRBPReview, ReviewStatusHRFinal:
		return "performance:hrbp"
	default:
		return "performance:mine"
	}
}

func reviewNotifyStageName(status string) string {
	switch status {
	case ReviewStatusDraft:
		return "员工填写"
	case ReviewStatusManagerReview:
		return "上级评价"
	case ReviewStatusHRBPReview:
		return "HRBP评价"
	case ReviewStatusEmployeeConfirm:
		return "员工确认"
	case ReviewStatusHRFinal:
		return "HRBP归档"
	default:
		return "绩效流程"
	}
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
