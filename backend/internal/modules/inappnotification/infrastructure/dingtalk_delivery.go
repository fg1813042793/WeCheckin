package infrastructure

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/notificationstyle"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

const dingTalkAdminNotificationTimeout = 8 * time.Second

type dingTalkNotificationTarget struct {
	LocalUserID    uint
	Config         configsvc.DingTalkH5CorpConfig
	DingTalkUserID string
}

type dingTalkTargetResolution struct {
	Targets      []dingTalkNotificationTarget
	SkippedCount int
}

type dingTalkTargetResolver interface {
	ResolveTargets(context.Context, []uint) (dingTalkTargetResolution, error)
}

type DingTalkDelivery struct {
	client      configsvc.DingTalkWorkNotificationClient
	resolver    dingTalkTargetResolver
	styleLoader dingTalkStyleLoader
}

type dingTalkStyleLoader interface {
	NotificationStyles(context.Context) (notificationstyle.Config, error)
}

func NewDingTalkDelivery(db *gorm.DB, client configsvc.DingTalkWorkNotificationClient) *DingTalkDelivery {
	if client == nil {
		client = configsvc.DefaultDingTalkWorkNotificationClient()
	}
	delivery := newDingTalkDelivery(client, &gormDingTalkTargetResolver{db: db})
	delivery.styleLoader = NewGormStore(db)
	return delivery
}

func newDingTalkDelivery(client configsvc.DingTalkWorkNotificationClient, resolver dingTalkTargetResolver) *DingTalkDelivery {
	return &DingTalkDelivery{client: client, resolver: resolver}
}

func (delivery *DingTalkDelivery) DeliverDingTalk(ctx context.Context, batch application.DingTalkDeliveryBatch) (application.DingTalkDeliveryResult, error) {
	if delivery == nil || delivery.client == nil || delivery.resolver == nil {
		return application.DingTalkDeliveryResult{}, application.ErrDingTalkDeliveryUnavailable
	}
	resolution, err := delivery.resolver.ResolveTargets(ctx, normalizeIDs(batch.UserIDs))
	if err != nil {
		return application.DingTalkDeliveryResult{}, err
	}
	result := application.DingTalkDeliveryResult{SkippedCount: resolution.SkippedCount}
	styles := notificationstyle.DefaultConfig()
	if delivery.styleLoader != nil {
		styles, err = delivery.styleLoader.NotificationStyles(ctx)
		if err != nil {
			return application.DingTalkDeliveryResult{}, err
		}
	}
	notificationType := strings.TrimSpace(batch.NotificationType)
	if notificationType == "" {
		notificationType = notificationstyle.TypeAdminManual
	}
	template := notificationstyle.StyleFor(styles, notificationType).DingTalk
	groups := make(map[string]*dingTalkDeliveryGroup)
	groupKeys := make([]string, 0)
	for _, target := range resolution.Targets {
		corpID := strings.TrimSpace(target.Config.CorpID)
		if corpID == "" || strings.TrimSpace(target.DingTalkUserID) == "" {
			result.SkippedCount++
			continue
		}
		group, exists := groups[corpID]
		if !exists {
			group = &dingTalkDeliveryGroup{config: target.Config}
			groups[corpID] = group
			groupKeys = append(groupKeys, corpID)
		}
		group.userIDs = appendUniqueDingTalkUserID(group.userIDs, target.DingTalkUserID)
		group.localUserCount++
	}
	sort.Strings(groupKeys)
	for _, corpID := range groupKeys {
		group := groups[corpID]
		payload := configsvc.ApplyDingTalkTemplate(template, dingTalkManualNotificationPayload(group.config, batch.Title, batch.Content))
		sendCtx, cancel := context.WithTimeout(ctx, dingTalkAdminNotificationTimeout)
		err := delivery.client.SendWorkNotificationContext(sendCtx, group.config, group.userIDs, payload)
		cancel()
		if err != nil {
			result.FailedCount += group.localUserCount
			if logger.Logger != nil {
				logger.Logger.Printf("[AdminDingTalkNotification] delivery_failed corpId=%s recipients=%d err=%v", corpID, group.localUserCount, err)
			}
			continue
		}
		result.SentCount += group.localUserCount
	}
	return result, nil
}

func dingTalkManualNotificationPayload(config configsvc.DingTalkH5CorpConfig, title, content string) configsvc.DingTalkWorkNotificationPayload {
	payload := configsvc.DingTalkWorkNotificationPayload{
		Title: title, Content: content, URL: config.AppURL, SourceName: "WeCheckin 通知",
	}
	if strings.TrimSpace(payload.URL) == "" && strings.TrimSpace(title) != "" {
		payload.Content = strings.TrimSpace(title) + "\n" + strings.TrimSpace(content)
	}
	return payload
}

type dingTalkDeliveryGroup struct {
	config         configsvc.DingTalkH5CorpConfig
	userIDs        []string
	localUserCount int
}

type gormDingTalkTargetResolver struct {
	db *gorm.DB
}

func (resolver *gormDingTalkTargetResolver) ResolveTargets(ctx context.Context, userIDs []uint) (dingTalkTargetResolution, error) {
	userIDs = normalizeIDs(userIDs)
	if len(userIDs) == 0 {
		return dingTalkTargetResolution{}, nil
	}
	if resolver == nil || resolver.db == nil {
		return dingTalkTargetResolution{}, errors.New("DingTalk notification database is not initialized")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	db := resolver.db.WithContext(queryCtx)

	var bindings []model.DingTalkH5UserBinding
	if err := db.Where("user_id IN ? AND enabled = 1", userIDs).
		Order("user_id ASC, corp_id ASC, dingtalk_user_id ASC").Find(&bindings).Error; err != nil {
		return dingTalkTargetResolution{}, err
	}
	corpIDs := make([]string, 0, len(bindings))
	bindingsByUser := make(map[uint][]model.DingTalkH5UserBinding)
	for _, binding := range bindings {
		bindingsByUser[binding.UserID] = append(bindingsByUser[binding.UserID], binding)
		corpIDs = appendUniqueDingTalkUserID(corpIDs, binding.CorpID)
	}

	configs := make(map[string]model.DingTalkH5CorpConfig)
	if len(corpIDs) > 0 {
		var rows []model.DingTalkH5CorpConfig
		if err := db.Where("corp_id IN ? AND enabled = 1 AND notify_enabled = 1", corpIDs).
			Order("corp_id ASC").Find(&rows).Error; err != nil {
			return dingTalkTargetResolution{}, err
		}
		for _, row := range rows {
			configs[strings.TrimSpace(row.CorpID)] = row
		}
	}

	result := dingTalkTargetResolution{Targets: make([]dingTalkNotificationTarget, 0, len(userIDs))}
	for _, userID := range userIDs {
		var selected *dingTalkNotificationTarget
		for _, binding := range bindingsByUser[userID] {
			corpID := strings.TrimSpace(binding.CorpID)
			config, ok := configs[corpID]
			if !ok || strings.TrimSpace(binding.DingTalkUserID) == "" {
				continue
			}
			target := dingTalkNotificationTarget{
				LocalUserID: userID, Config: dingTalkConfigFromModel(config), DingTalkUserID: strings.TrimSpace(binding.DingTalkUserID),
			}
			selected = &target
			break
		}
		if selected == nil {
			result.SkippedCount++
			continue
		}
		result.Targets = append(result.Targets, *selected)
	}
	return result, nil
}

func dingTalkConfigFromModel(row model.DingTalkH5CorpConfig) configsvc.DingTalkH5CorpConfig {
	return configsvc.NormalizeDingTalkH5CorpConfig(configsvc.DingTalkH5CorpConfig{
		CorpID: row.CorpID, CorpName: row.CorpName, AppKey: row.AppKey, AppSecret: row.AppSecret,
		AgentID: row.AgentID, UnifiedAppID: row.UnifiedAppID, AppURL: row.AppURL,
		NotifyEnabled: row.NotifyEnabled, NotifyMode: row.NotifyMode, RobotCode: row.RobotCode, Enabled: row.Enabled,
	})
}

func appendUniqueDingTalkUserID(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

var _ application.DingTalkDelivery = (*DingTalkDelivery)(nil)
