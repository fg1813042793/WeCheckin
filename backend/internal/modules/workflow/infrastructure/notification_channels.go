package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
)

const dingTalkWorkflowNotificationTimeout = 5 * time.Second

type DingTalkNotificationTarget struct {
	Config         configsvc.DingTalkH5CorpConfig
	DingTalkUserID string
}

type DingTalkNotificationResolver interface {
	Resolve(ctx context.Context, notification application.NotificationRecord) (DingTalkNotificationTarget, error)
}

type DingTalkNotificationChannel struct {
	client   configsvc.DingTalkWorkNotificationClient
	resolver DingTalkNotificationResolver
}

func NewDingTalkNotificationChannel(db *gorm.DB, client configsvc.DingTalkWorkNotificationClient) *DingTalkNotificationChannel {
	if client == nil {
		client = configsvc.DefaultDingTalkWorkNotificationClient()
	}
	return newDingTalkNotificationChannel(client, &gormDingTalkNotificationResolver{db: db})
}

func newDingTalkNotificationChannel(client configsvc.DingTalkWorkNotificationClient, resolver DingTalkNotificationResolver) *DingTalkNotificationChannel {
	return &DingTalkNotificationChannel{client: client, resolver: resolver}
}

func (*DingTalkNotificationChannel) Name() string {
	return workflowcore.NotificationChannelDingTalkOA
}

func (channel *DingTalkNotificationChannel) Deliver(ctx context.Context, notifications []application.NotificationRecord) []application.NotificationDeliveryResult {
	results := make([]application.NotificationDeliveryResult, len(notifications))
	groups := make(map[string]*dingTalkNotificationGroup)
	groupKeys := make([]string, 0)
	for index, notification := range notifications {
		results[index].ID = notification.ID
		if channel == nil || channel.client == nil || channel.resolver == nil {
			results[index].Err = errors.New("钉钉工作通知渠道未初始化")
			continue
		}
		target, err := channel.resolver.Resolve(ctx, notification)
		if err != nil {
			results[index].Err = err
			continue
		}
		payload := configsvc.DingTalkWorkNotificationPayload{
			Title: notification.Payload.Title, Content: notification.Payload.Content,
			URL: target.Config.AppURL, SourceName: "WeCheckin 流程",
		}
		key := dingTalkNotificationGroupKey(target.Config.CorpID, payload)
		group, exists := groups[key]
		if !exists {
			group = &dingTalkNotificationGroup{config: target.Config, payload: payload}
			groups[key] = group
			groupKeys = append(groupKeys, key)
		}
		group.resultIndexes = append(group.resultIndexes, index)
		group.userIDs = appendUniqueString(group.userIDs, target.DingTalkUserID)
	}

	sort.Strings(groupKeys)
	for _, key := range groupKeys {
		group := groups[key]
		sendCtx, cancel := context.WithTimeout(ctx, dingTalkWorkflowNotificationTimeout)
		err := channel.client.SendWorkNotificationContext(sendCtx, group.config, group.userIDs, group.payload)
		cancel()
		for _, index := range group.resultIndexes {
			results[index].Err = err
		}
	}
	return results
}

type dingTalkNotificationGroup struct {
	config        configsvc.DingTalkH5CorpConfig
	payload       configsvc.DingTalkWorkNotificationPayload
	userIDs       []string
	resultIndexes []int
}

type gormDingTalkNotificationResolver struct {
	db *gorm.DB
}

func (resolver *gormDingTalkNotificationResolver) Resolve(ctx context.Context, notification application.NotificationRecord) (DingTalkNotificationTarget, error) {
	if resolver == nil || resolver.db == nil {
		return DingTalkNotificationTarget{}, errors.New("钉钉绑定数据库未初始化")
	}
	recipientID, err := parseNotificationUserID(notification.RecipientUserID)
	if err != nil {
		return DingTalkNotificationTarget{}, fmt.Errorf("通知接收人无效: %w", err)
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	db := resolver.db.WithContext(queryCtx)

	var recipientBindings []model.DingTalkH5UserBinding
	bindingQuery := db.Where("user_id = ? AND enabled = 1", recipientID)
	if corpID := strings.TrimSpace(notification.CorpID); corpID != "" {
		bindingQuery = bindingQuery.Where("corp_id = ?", corpID)
	}
	if err := bindingQuery.Order("corp_id ASC").Order("dingtalk_user_id ASC").Find(&recipientBindings).Error; err != nil {
		return DingTalkNotificationTarget{}, err
	}
	if len(recipientBindings) == 0 {
		return DingTalkNotificationTarget{}, errors.New("通知接收人未绑定可用的钉钉企业")
	}

	corpIDs := make([]string, 0, len(recipientBindings))
	for _, binding := range recipientBindings {
		corpIDs = appendUniqueString(corpIDs, binding.CorpID)
	}
	var configRows []model.DingTalkH5CorpConfig
	if err := db.Where("corp_id IN ? AND enabled = 1 AND notify_enabled = 1", corpIDs).Find(&configRows).Error; err != nil {
		return DingTalkNotificationTarget{}, err
	}
	configs := make(map[string]model.DingTalkH5CorpConfig, len(configRows))
	for _, config := range configRows {
		configs[strings.TrimSpace(config.CorpID)] = config
	}

	starterCorpIDs := make([]string, 0)
	if strings.TrimSpace(notification.CorpID) == "" {
		if starterID, parseErr := parseNotificationUserID(notification.Payload.StarterID); parseErr == nil {
			var starterBindings []model.DingTalkH5UserBinding
			if err := db.Select("corp_id").Where("user_id = ? AND enabled = 1", starterID).Order("corp_id ASC").Find(&starterBindings).Error; err != nil {
				return DingTalkNotificationTarget{}, err
			}
			for _, binding := range starterBindings {
				starterCorpIDs = appendUniqueString(starterCorpIDs, binding.CorpID)
			}
		}
	}
	binding, configRow, err := selectDingTalkNotificationTarget(starterCorpIDs, recipientBindings, configs)
	if err != nil {
		return DingTalkNotificationTarget{}, err
	}
	if strings.TrimSpace(notification.CorpID) == "" {
		result := db.Model(&workflowmodel.NotificationOutbox{}).
			Where("id = ? AND corp_id = ?", notification.ID, "").
			Update("corp_id", binding.CorpID)
		if result.Error != nil {
			return DingTalkNotificationTarget{}, result.Error
		}
		if result.RowsAffected == 0 {
			var persisted workflowmodel.NotificationOutbox
			if err := db.Select("id", "corp_id").First(&persisted, "id = ?", notification.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return DingTalkNotificationTarget{}, ErrNotificationNotFound
				}
				return DingTalkNotificationTarget{}, err
			}
			if persisted.CorpID != binding.CorpID {
				notification.CorpID = persisted.CorpID
				return resolver.Resolve(ctx, notification)
			}
		}
	}
	return DingTalkNotificationTarget{
		Config:         dingTalkNotificationConfigFromModel(configRow),
		DingTalkUserID: strings.TrimSpace(binding.DingTalkUserID),
	}, nil
}

func selectDingTalkNotificationTarget(
	starterCorpIDs []string,
	recipientBindings []model.DingTalkH5UserBinding,
	configs map[string]model.DingTalkH5CorpConfig,
) (model.DingTalkH5UserBinding, model.DingTalkH5CorpConfig, error) {
	starterCorps := make(map[string]struct{}, len(starterCorpIDs))
	for _, corpID := range starterCorpIDs {
		corpID = strings.TrimSpace(corpID)
		if corpID != "" {
			starterCorps[corpID] = struct{}{}
		}
	}
	bindings := append([]model.DingTalkH5UserBinding(nil), recipientBindings...)
	sort.SliceStable(bindings, func(i, j int) bool {
		leftCorp, rightCorp := strings.TrimSpace(bindings[i].CorpID), strings.TrimSpace(bindings[j].CorpID)
		if leftCorp == rightCorp {
			return strings.TrimSpace(bindings[i].DingTalkUserID) < strings.TrimSpace(bindings[j].DingTalkUserID)
		}
		return leftCorp < rightCorp
	})
	selectBinding := func(requireCommon bool) (model.DingTalkH5UserBinding, model.DingTalkH5CorpConfig, bool) {
		for _, binding := range bindings {
			corpID := strings.TrimSpace(binding.CorpID)
			if binding.Enabled != 1 || corpID == "" || strings.TrimSpace(binding.DingTalkUserID) == "" {
				continue
			}
			if requireCommon {
				if _, exists := starterCorps[corpID]; !exists {
					continue
				}
			}
			config, exists := configs[corpID]
			if !exists || config.Enabled != 1 || config.NotifyEnabled != 1 {
				continue
			}
			return binding, config, true
		}
		return model.DingTalkH5UserBinding{}, model.DingTalkH5CorpConfig{}, false
	}
	if len(starterCorps) > 0 {
		if binding, config, ok := selectBinding(true); ok {
			return binding, config, nil
		}
	}
	if binding, config, ok := selectBinding(false); ok {
		return binding, config, nil
	}
	return model.DingTalkH5UserBinding{}, model.DingTalkH5CorpConfig{}, errors.New("通知接收人无可用的钉钉通知企业")
}

func dingTalkNotificationConfigFromModel(row model.DingTalkH5CorpConfig) configsvc.DingTalkH5CorpConfig {
	return configsvc.NormalizeDingTalkH5CorpConfig(configsvc.DingTalkH5CorpConfig{
		CorpID: row.CorpID, CorpName: row.CorpName, AppKey: row.AppKey, AppSecret: row.AppSecret,
		AgentID: row.AgentID, UnifiedAppID: row.UnifiedAppID, AppURL: row.AppURL,
		NotifyEnabled: row.NotifyEnabled, NotifyMode: row.NotifyMode, RobotCode: row.RobotCode, Enabled: row.Enabled,
	})
}

func dingTalkNotificationGroupKey(corpID string, payload configsvc.DingTalkWorkNotificationPayload) string {
	return strings.Join([]string{
		strings.TrimSpace(corpID), strings.TrimSpace(payload.Title), strings.TrimSpace(payload.Content),
		strings.TrimSpace(payload.URL), strings.TrimSpace(payload.SourceName), strings.TrimSpace(payload.PicURL),
	}, "\x00")
}

func parseNotificationUserID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("本地用户 ID 无效")
	}
	return uint(parsed), nil
}

func appendUniqueString(values []string, value string) []string {
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
