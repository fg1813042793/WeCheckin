package formkitadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	setupservice "wecheckin/backend/internal/service/admin/setup"
)

const (
	NotificationTemplateScopeBuiltin = "builtin"
	NotificationTemplateScopeCustom  = "custom"
	builtinTemplateSetupKey          = "BUILTIN_TEMPLATE_PRESETS"
	builtinExamTemplateSetupKey      = "EXAM_BUILTIN_TEMPLATE_PRESETS"
)

var ErrInvalidNotificationTemplateScope = errors.New("invalid survey notification template scope")

type NotificationTemplateCollection struct {
	Builtin []TemplatePreset `json:"builtin"`
	Custom  []TemplatePreset `json:"custom"`
}

func defaultBuiltinNotificationTemplates() []TemplatePreset {
	return []TemplatePreset{
		{"label": "简洁", "value": "📋 问卷「{title}」收到新答卷\n新提交人：{submitter}\n时间：{date}\n共 {total} 份提交\n\n{result}"},
		{"label": "详细", "value": "📊 问卷统计报告\n━━━━━━━━━━━━━━━━━━\n📌 问卷：{title}\n👤 新提交人：{submitter}\n🕐 时间：{date}\n📈 总提交数：{total}\n\n{result}\n━━━━━━━━━━━━━━━━━━"},
		{"label": "仅统计结果", "value": "{result}"},
	}
}

func defaultBuiltinExamNotificationTemplates() []TemplatePreset {
	return []TemplatePreset{
		{"label": "简洁", "value": "📝 考试「{title}」收到新答卷\n提交人：{submitter}\n时间：{date}\n共 {total} 份提交\n\n{result}"},
		{"label": "成绩通知", "value": "📊 考试成绩通知\n考试：{title}\n提交人：{submitter}\n得分：{score} / {fullScore}\n提交时间：{date}"},
		{"label": "仅统计结果", "value": "{result}"},
		{"label": "答卷下载", "value": "📝 考试「{title}」收到新答卷\n提交人：{submitter}\n提交时间：{date}\n答卷下载：{answerDownloadUrl}"},
	}
}

func decodeBuiltinNotificationTemplates(raw string) []TemplatePreset {
	var presets []TemplatePreset
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &presets) != nil || len(presets) == 0 {
		return defaultBuiltinNotificationTemplates()
	}
	return presets
}

func decodeBuiltinExamNotificationTemplates(raw string) []TemplatePreset {
	var presets []TemplatePreset
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &presets) != nil || len(presets) == 0 {
		return defaultBuiltinExamNotificationTemplates()
	}
	return presets
}

func validateNotificationTemplateScope(scope string) error {
	switch strings.TrimSpace(scope) {
	case NotificationTemplateScopeBuiltin, NotificationTemplateScopeCustom:
		return nil
	default:
		return ErrInvalidNotificationTemplateScope
	}
}

func GetNotificationTemplatesContext(ctx context.Context, adminID uint) (*NotificationTemplateCollection, error) {
	builtin := defaultBuiltinNotificationTemplates()
	entry, err := setupservice.GetSetupContext(ctx, builtinTemplateSetupKey)
	if err == nil {
		builtin = decodeBuiltinNotificationTemplates(entry.Value)
	} else if !IsNotFound(err) {
		return nil, err
	}
	custom, err := GetTemplatePresetsContext(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &NotificationTemplateCollection{Builtin: builtin, Custom: custom}, nil
}

func SaveNotificationTemplatesContext(ctx context.Context, adminID uint, scope string, presets []TemplatePreset, addIP string) error {
	if err := validateNotificationTemplateScope(scope); err != nil {
		return err
	}
	if scope == NotificationTemplateScopeCustom {
		return SaveTemplatePresetsContext(ctx, adminID, presets)
	}
	value, err := json.Marshal(presets)
	if err != nil {
		return err
	}
	return setupservice.SetContentSetupContext(ctx, builtinTemplateSetupKey, string(value), addIP)
}

func GetExamNotificationTemplatesContext(ctx context.Context, adminID uint) (*NotificationTemplateCollection, error) {
	builtin := defaultBuiltinExamNotificationTemplates()
	entry, err := setupservice.GetSetupContext(ctx, builtinExamTemplateSetupKey)
	if err == nil {
		builtin = decodeBuiltinExamNotificationTemplates(entry.Value)
	} else if !IsNotFound(err) {
		return nil, err
	}
	custom, err := GetExamTemplatePresetsContext(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &NotificationTemplateCollection{Builtin: builtin, Custom: custom}, nil
}

func SaveExamNotificationTemplatesContext(ctx context.Context, adminID uint, scope string, presets []TemplatePreset, addIP string) error {
	if err := validateNotificationTemplateScope(scope); err != nil {
		return err
	}
	if scope == NotificationTemplateScopeCustom {
		return SaveExamTemplatePresetsContext(ctx, adminID, presets)
	}
	value, err := json.Marshal(presets)
	if err != nil {
		return err
	}
	return setupservice.SetContentSetupContext(ctx, builtinExamTemplateSetupKey, string(value), addIP)
}
