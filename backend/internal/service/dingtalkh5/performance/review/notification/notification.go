package notification

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

const (
	notifyEnabledKey = "DINGTALK_H5_NOTIFY_ENABLED"
	notifyTimeout    = 5 * time.Second

	EventSelfSubmitted = "self_submitted"
	EventFlowMoved     = "flow_moved"

	reviewStatusDraft           = "draft"
	reviewStatusManagerReview   = "manager_review"
	reviewStatusHRBPReview      = "hrbp_review"
	reviewStatusEmployeeConfirm = "employee_confirm"
	reviewStatusHRFinal         = "hr_final"
	reviewStatusCompleted       = "completed"

	notifyResultSuccessPrefix = "[DingTalkH5Notify] result=success"
	notifyResultFailedPrefix  = "[DingTalkH5Notify] result=failed"
	notifyResultSkippedPrefix = "[DingTalkH5Notify] result=skipped"
	notifySuccessMessage      = "钉钉通知发送成功"
	notifyFailedMessage       = "钉钉通知发送失败"
	notifySkippedMessage      = "钉钉通知跳过发送"
)

type recipient struct {
	Config         configsvc.DingTalkH5CorpConfig
	DingTalkUserID string
}

func EnabledContext(ctx context.Context) bool {
	value := strings.ToLower(strings.TrimSpace(configsvc.SetupValueContext(ctx, notifyEnabledKey)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func TransitionAsync(ctx context.Context, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, event string) {
	nextAccount := NextHandlerAccount(review)
	logger.Logger.Printf("[DingTalkH5Notify] schedule event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()
		if err := sendReviewTransitionContext(bg, configsvc.DefaultDingTalkWorkNotificationClient(), review, actor, event); err != nil {
			logTransitionResult(notifyResultFailedPrefix, notifyFailedMessage, event, review, nextAccount, "send_failed", err)
		}
	}()
}

func sendReviewTransitionContext(ctx context.Context, client configsvc.DingTalkWorkNotificationClient, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, event string) error {
	nextAccount := NextHandlerAccount(review)
	if client == nil {
		return fmt.Errorf("钉钉通知客户端不能为空")
	}
	switch event {
	case EventSelfSubmitted:
	case EventFlowMoved:
	default:
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=unsupported_event event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
		logTransitionResult(notifyResultSkippedPrefix, notifySkippedMessage, event, review, nextAccount, "unsupported_event", nil)
		return nil
	}
	if nextAccount == "" {
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=no_next_handler event=%s reviewNo=%s status=%s", event, review.ReviewNo, review.Status)
		logTransitionResult(notifyResultSkippedPrefix, notifySkippedMessage, event, review, nextAccount, "no_next_handler", nil)
		return nil
	}
	recipient, employeeName, err := resolveRecipientContext(ctx, review.EmployeeAccount, nextAccount)
	if err != nil {
		return err
	}
	if recipient.Config.NotifyEnabled == 0 {
		logger.Logger.Printf("[DingTalkH5Notify] skip reason=disabled event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
		logTransitionResult(notifyResultSkippedPrefix, notifySkippedMessage, event, review, nextAccount, "disabled", nil)
		return nil
	}
	notification := buildTransitionPayloadContext(ctx, recipient.Config, review, actor, employeeName)
	logSendConfig(event, review, nextAccount, recipient, notification)
	if err := client.SendWorkNotificationContext(ctx, recipient.Config, []string{recipient.DingTalkUserID}, notification); err != nil {
		return err
	}
	logger.Logger.Printf("[DingTalkH5Notify] sent event=%s reviewNo=%s status=%s next=%s", event, review.ReviewNo, review.Status, nextAccount)
	logTransitionResult(notifyResultSuccessPrefix, notifySuccessMessage, event, review, nextAccount, "sent", nil)
	return nil
}

func logTransitionResult(prefix, message, event string, review model.DingTalkH5PerfReview, nextAccount, reason string, err error) {
	if err != nil {
		logger.Logger.Printf(
			"%s message=%s event=%s reviewNo=%s status=%s next=%s reason=%s err=%v",
			prefix,
			message,
			event,
			review.ReviewNo,
			review.Status,
			logValue(nextAccount),
			logValue(reason),
			err,
		)
		return
	}
	logger.Logger.Printf(
		"%s message=%s event=%s reviewNo=%s status=%s next=%s reason=%s",
		prefix,
		message,
		event,
		review.ReviewNo,
		review.Status,
		logValue(nextAccount),
		logValue(reason),
	)
}

func logSendConfig(event string, review model.DingTalkH5PerfReview, nextAccount string, recipient recipient, notification configsvc.DingTalkWorkNotificationPayload) {
	config := configsvc.NormalizeDingTalkH5CorpConfig(recipient.Config)
	logger.Logger.Printf(
		"[DingTalkH5Notify] send config event=%s reviewNo=%s status=%s employee=%s next=%s dingtalkUserId=%s corpId=%s corpName=%s notifyMode=%s notifyEnabled=%d appKey=%s appSecretSet=%t agentId=%s unifiedAppId=%s openAppId=%s robotCode=%s appURL=%s notificationURLSet=%t title=%s sourceName=%s hasPic=%t",
		event,
		review.ReviewNo,
		review.Status,
		usersvc.NormalizeUserID(review.EmployeeAccount),
		usersvc.NormalizeUserID(nextAccount),
		logValue(recipient.DingTalkUserID),
		logValue(config.CorpID),
		logValue(config.CorpName),
		config.NotifyMode,
		config.NotifyEnabled,
		maskLogValue(config.AppKey),
		config.AppSecretSet,
		logValue(config.AgentID),
		logValue(config.UnifiedAppID),
		logValue(openAppID(config)),
		maskLogValue(config.RobotCode),
		logValue(config.AppURL),
		strings.TrimSpace(notification.URL) != "",
		logValue(notification.Title),
		logValue(notification.SourceName),
		strings.TrimSpace(notification.PicURL) != "",
	)
}

func logValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func maskLogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func NextHandlerAccount(review model.DingTalkH5PerfReview) string {
	switch review.Status {
	case reviewStatusDraft:
		return usersvc.NormalizeUserID(review.EmployeeAccount)
	case reviewStatusManagerReview:
		return usersvc.NormalizeUserID(review.ManagerAccount)
	case reviewStatusHRBPReview, reviewStatusHRFinal:
		return usersvc.NormalizeUserID(fallback(review.HRBPReviewerAccount, review.HRBPAccount))
	case reviewStatusEmployeeConfirm:
		return usersvc.NormalizeUserID(review.EmployeeAccount)
	default:
		return ""
	}
}

func resolveRecipientContext(ctx context.Context, employeeAccount, handlerAccount string) (recipient, string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return recipient{}, "", fmt.Errorf("database is not initialized")
	}
	handler, err := perfUserIdentityByAccountDB(db, handlerAccount)
	if err != nil {
		return recipient{}, "", fmt.Errorf("下一步处理人未绑定有效用户：%w", err)
	}
	corpID := ""
	employeeName := ""
	if employee, err := perfUserIdentityByAccountDB(db, employeeAccount); err == nil {
		employeeName = strings.TrimSpace(employee.Name)
		corpID = firstEnabledCorpIDByUserIDDB(db, employee.ID)
	}
	binding, err := firstEnabledBindingByUserIDDB(db, handler.ID, corpID)
	if err != nil {
		return recipient{}, "", fmt.Errorf("下一步处理人未绑定钉钉账号：%w", err)
	}
	config, err := configsvc.LoadDingTalkH5CorpConfigContext(ctx, binding.CorpID)
	if err != nil {
		return recipient{}, "", err
	}
	if configsvc.NormalizeDingTalkH5NotifyMode(config.NotifyMode, config.AgentID, config.RobotCode) != "robot" && strings.TrimSpace(config.AgentID) == "" {
		return recipient{}, "", fmt.Errorf("请先配置钉钉内部应用 AgentId")
	}
	return recipient{Config: config, DingTalkUserID: binding.DingTalkUserID}, employeeName, nil
}

func buildTransitionContent(review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, resolvedEmployeeName string) string {
	employeeName := strings.TrimSpace(resolvedEmployeeName)
	if employeeName == "" {
		employeeName = strings.TrimSpace(review.EmployeeAccount)
	}
	actorName := ""
	if actor != nil && strings.TrimSpace(actor.Name) != "" {
		actorName = strings.TrimSpace(actor.Name)
		if usersvc.NormalizeUserID(actor.Account) == usersvc.NormalizeUserID(review.EmployeeAccount) {
			employeeName = actorName
		}
	} else if actor != nil && strings.TrimSpace(actor.Account) != "" {
		actorName = strings.TrimSpace(actor.Account)
	}
	period := strings.TrimSpace(review.Period)
	stage := notifyStageName(review.Status)
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

func buildTransitionPayloadContext(ctx context.Context, config configsvc.DingTalkH5CorpConfig, review model.DingTalkH5PerfReview, actor *model.DingTalkH5PerfUser, employeeName string) configsvc.DingTalkWorkNotificationPayload {
	return configsvc.DingTalkWorkNotificationPayload{
		Title:      buildTransitionTitle(review),
		Content:    buildTransitionContent(review, actor, employeeName),
		URL:        buildNotificationURL(appURLContext(ctx, config), config, review),
		SourceName: appNameContext(ctx),
		PicURL:     logoURLContext(ctx),
	}
}

func buildTransitionTitle(review model.DingTalkH5PerfReview) string {
	period := strings.TrimSpace(review.Period)
	if period == "" {
		return "绩效流程待办"
	}
	return period + " 月度考评待办"
}

func appURLContext(ctx context.Context, config configsvc.DingTalkH5CorpConfig) string {
	if appURL := strings.TrimSpace(config.AppURL); appURL != "" {
		return appURL
	}
	return strings.TrimSpace(configsvc.SetupValueContext(ctx, "DINGTALK_H5_APP_URL"))
}

func appNameContext(ctx context.Context) string {
	if database.GetDB() == nil {
		return configsvc.DefaultDingTalkH5AppName
	}
	appName := strings.TrimSpace(configsvc.SetupValueContext(ctx, "DINGTALK_H5_APP_NAME"))
	if appName == "" {
		return configsvc.DefaultDingTalkH5AppName
	}
	return appName
}

func logoURLContext(ctx context.Context) string {
	if database.GetDB() == nil {
		return ""
	}
	return strings.TrimSpace(configsvc.SetupValueContext(ctx, "DINGTALK_H5_LOGO_URL"))
}

func buildOperationURL(baseURL string, review model.DingTalkH5PerfReview) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	operationURL, err := url.Parse(baseURL)
	if err != nil {
		return appendOperationQuery(baseURL, review)
	}
	query := operationURL.Query()
	query.Set("view", operationView(review.Status))
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

func buildNotificationURL(baseURL string, config configsvc.DingTalkH5CorpConfig, review model.DingTalkH5PerfReview) string {
	operationURL := buildOperationURL(baseURL, review)
	if operationURL == "" {
		return ""
	}
	return wrapOpenAppURL(operationURL, config)
}

func wrapOpenAppURL(operationURL string, config configsvc.DingTalkH5CorpConfig) string {
	operationURL = strings.TrimSpace(operationURL)
	if operationURL == "" {
		return ""
	}
	corpID := strings.TrimSpace(config.CorpID)
	appID := openAppID(config)
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

func openAppID(config configsvc.DingTalkH5CorpConfig) string {
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

func appendOperationQuery(baseURL string, review model.DingTalkH5PerfReview) string {
	values := url.Values{}
	values.Set("view", operationView(review.Status))
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

func operationView(status string) string {
	switch status {
	case reviewStatusManagerReview:
		return "performance:manager"
	case reviewStatusHRBPReview, reviewStatusHRFinal:
		return "performance:hrbp"
	default:
		return "performance:mine"
	}
}

func notifyStageName(status string) string {
	switch status {
	case reviewStatusDraft:
		return "员工填写"
	case reviewStatusManagerReview:
		return "上级评价"
	case reviewStatusHRBPReview:
		return "HRBP评价"
	case reviewStatusEmployeeConfirm:
		return "员工确认"
	case reviewStatusHRFinal:
		return "HRBP归档"
	default:
		return "绩效流程"
	}
}

func perfUserIdentityByAccountDB(db *gorm.DB, account string) (*model.DingTalkH5PerfUser, error) {
	account = usersvc.NormalizeUserID(account)
	if account == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var user model.DingTalkH5PerfUser
	if err := db.Select("`id`, `user_name`").Where("`user_mini_openid` = ? AND `user_status` = 1", account).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func firstEnabledCorpIDByUserIDDB(db *gorm.DB, userID uint) string {
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

func firstEnabledBindingByUserIDDB(db *gorm.DB, userID uint, preferredCorpID string) (model.DingTalkH5UserBinding, error) {
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

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallbackValue
}
