package notificationstyle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

const (
	SetupKey       = "NOTIFICATION_STYLE_CONFIG"
	CurrentVersion = 1

	TypeTaskArrived            = "task_arrived"
	TypeTaskReminder           = "task_reminder"
	TypeApprovalResultApproved = "approval_result_approved"
	TypeApprovalResultRejected = "approval_result_rejected"
	TypeApprovalResultReturned = "approval_result_returned"
	TypeNodeCC                 = "node_cc"
	TypeNodeNotify             = "node_notify"
	TypeInstanceCommented      = "instance_commented"
	TypeInstanceFormRevised    = "instance_form_revised"
	TypeWorkflow               = "workflow"
	TypeAdminManual            = "admin_manual"
	TypeScheduledTask          = "scheduled_task"
	TypeSurveyStat             = "survey_stat"

	DingTalkMessageTypeAuto       = "auto"
	DingTalkMessageTypeText       = "text"
	DingTalkMessageTypeImage      = "image"
	DingTalkMessageTypeVoice      = "voice"
	DingTalkMessageTypeFile       = "file"
	DingTalkMessageTypeLink       = "link"
	DingTalkMessageTypeOA         = "oa"
	DingTalkMessageTypeMarkdown   = "markdown"
	DingTalkMessageTypeActionCard = "action_card"
)

type Tone string

const (
	TonePrimary Tone = "primary"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
	ToneInfo    Tone = "info"
)

var (
	ErrUnknownType             = errors.New("notification style type is unsupported")
	ErrDuplicateType           = errors.New("notification style type is duplicated")
	ErrInvalidLabel            = errors.New("notification style label is invalid")
	ErrInvalidIcon             = errors.New("notification style icon is invalid")
	ErrInvalidTone             = errors.New("notification style tone is invalid")
	ErrInvalidDingTalkTemplate = errors.New("notification DingTalk template is invalid")

	iconPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
	templateToken    = regexp.MustCompile(`\{\{[^{}]+\}\}`)
	dingTalkColor    = regexp.MustCompile(`^[0-9A-F]{8}$`)
	allowedVariables = map[string]struct{}{
		"{{title}}": {}, "{{content}}": {}, "{{url}}": {}, "{{sourceName}}": {},
		"{{picUrl}}": {}, "{{mediaId}}": {}, "{{duration}}": {},
	}
	defaults = []Style{
		newDefaultStyle(TypeTaskArrived, "待处理", "clock", ToneWarning),
		newDefaultStyle(TypeTaskReminder, "处理提醒", "bell", ToneWarning),
		newDefaultStyle(TypeApprovalResultApproved, "审批通过", "checkmark-circle", ToneSuccess),
		newDefaultStyle(TypeApprovalResultRejected, "审批驳回", "error-circle", ToneDanger),
		newDefaultStyle(TypeApprovalResultReturned, "审批退回", "reload", ToneWarning),
		newDefaultStyle(TypeNodeCC, "流程抄送", "share", ToneInfo),
		newDefaultStyle(TypeNodeNotify, "流程通知", "email", TonePrimary),
		newDefaultStyle(TypeInstanceCommented, "流程评论", "chat", TonePrimary),
		newDefaultStyle(TypeInstanceFormRevised, "表单修改", "edit-pen", TonePrimary),
		newDefaultStyle(TypeWorkflow, "流程消息", "file-text", ToneInfo),
		newDefaultStyle(TypeAdminManual, "系统通知", "email", TonePrimary),
		newDefaultStyle(TypeScheduledTask, "定时通知", "clock", ToneInfo),
		newDefaultStyle(TypeSurveyStat, "问卷统计", "file-text", ToneInfo),
	}
)

type Style struct {
	Type     string           `json:"type"`
	Label    string           `json:"label"`
	Icon     string           `json:"icon"`
	Tone     Tone             `json:"tone"`
	DingTalk DingTalkTemplate `json:"dingTalk"`
}

type DingTalkTemplate struct {
	MessageType string `json:"messageType"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	PicURL      string `json:"picUrl"`
	SourceName  string `json:"sourceName"`
	MediaID     string `json:"mediaId"`
	Duration    int    `json:"duration"`
	ButtonTitle string `json:"buttonTitle"`
	HeadColor   string `json:"headColor"`
}

type DingTalkTemplateData struct {
	Title      string
	Content    string
	URL        string
	PicURL     string
	SourceName string
	MediaID    string
	Duration   int
}

type RenderedDingTalkTemplate struct {
	MessageType string
	Title       string
	Content     string
	URL         string
	PicURL      string
	SourceName  string
	MediaID     string
	Duration    int
	ButtonTitle string
	HeadColor   string
}

type Config struct {
	Version int     `json:"version"`
	Styles  []Style `json:"styles"`
}

func SupportedTypes() []string {
	result := make([]string, 0, len(defaults))
	for _, style := range defaults {
		result = append(result, style.Type)
	}
	return result
}

func IsSupportedType(value string) bool {
	_, ok := defaultByType(strings.TrimSpace(value))
	return ok
}

func DefaultConfig() Config {
	styles := make([]Style, len(defaults))
	copy(styles, defaults)
	return Config{Version: CurrentVersion, Styles: styles}
}

func Normalize(input Config) (Config, error) {
	overrides := make(map[string]Style, len(input.Styles))
	for _, raw := range input.Styles {
		style := Style{
			Type:     strings.TrimSpace(raw.Type),
			Label:    strings.TrimSpace(raw.Label),
			Icon:     strings.TrimSpace(raw.Icon),
			Tone:     Tone(strings.TrimSpace(string(raw.Tone))),
			DingTalk: raw.DingTalk,
		}
		if !IsSupportedType(style.Type) {
			return Config{}, fmt.Errorf("%w: %s", ErrUnknownType, style.Type)
		}
		if _, exists := overrides[style.Type]; exists {
			return Config{}, fmt.Errorf("%w: %s", ErrDuplicateType, style.Type)
		}
		if style.Label == "" || utf8.RuneCountInString(style.Label) > 20 {
			return Config{}, fmt.Errorf("%w: %s", ErrInvalidLabel, style.Type)
		}
		if !iconPattern.MatchString(style.Icon) {
			return Config{}, fmt.Errorf("%w: %s", ErrInvalidIcon, style.Type)
		}
		if !validTone(style.Tone) {
			return Config{}, fmt.Errorf("%w: %s", ErrInvalidTone, style.Type)
		}
		if strings.TrimSpace(style.DingTalk.MessageType) == "" {
			style.DingTalk = defaultDingTalkTemplate()
		} else {
			style.DingTalk = normalizeDingTalkTemplate(style.DingTalk)
			if err := validateDingTalkTemplate(style.DingTalk); err != nil {
				return Config{}, fmt.Errorf("%w: %s: %v", ErrInvalidDingTalkTemplate, style.Type, err)
			}
		}
		overrides[style.Type] = style
	}

	result := DefaultConfig()
	for index := range result.Styles {
		if style, ok := overrides[result.Styles[index].Type]; ok {
			result.Styles[index] = style
		}
	}
	return result, nil
}

func RenderDingTalkTemplate(template DingTalkTemplate, data DingTalkTemplateData) RenderedDingTalkTemplate {
	template = normalizeDingTalkTemplate(template)
	replacements := map[string]string{
		"{{title}}": data.Title, "{{content}}": data.Content, "{{url}}": data.URL,
		"{{sourceName}}": data.SourceName, "{{picUrl}}": data.PicURL,
		"{{mediaId}}": data.MediaID, "{{duration}}": fmt.Sprintf("%d", data.Duration),
	}
	render := func(value string) string {
		for token, replacement := range replacements {
			value = strings.ReplaceAll(value, token, replacement)
		}
		return strings.TrimSpace(value)
	}
	duration := template.Duration
	if duration <= 0 {
		duration = data.Duration
	}
	return RenderedDingTalkTemplate{
		MessageType: template.MessageType, Title: render(template.Title), Content: render(template.Content),
		URL: render(template.URL), PicURL: render(template.PicURL), SourceName: render(template.SourceName),
		MediaID: render(template.MediaID), Duration: duration, ButtonTitle: render(template.ButtonTitle), HeadColor: template.HeadColor,
	}
}

func StyleFor(config Config, notificationType string) Style {
	notificationType = strings.TrimSpace(notificationType)
	for _, style := range config.Styles {
		if strings.TrimSpace(style.Type) == notificationType {
			if strings.TrimSpace(style.DingTalk.MessageType) == "" {
				style.DingTalk = defaultDingTalkTemplate()
			}
			return style
		}
	}
	if style, ok := defaultByType(notificationType); ok {
		return style
	}
	return newDefaultStyle(notificationType, "系统消息", "email", ToneInfo)
}

func Decode(value string) (Config, error) {
	if strings.TrimSpace(value) == "" {
		return DefaultConfig(), nil
	}
	var config Config
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return Config{}, fmt.Errorf("decode notification style config: %w", err)
	}
	return Normalize(config)
}

func Encode(config Config) (string, Config, error) {
	normalized, err := Normalize(config)
	if err != nil {
		return "", Config{}, err
	}
	value, err := json.Marshal(normalized)
	if err != nil {
		return "", Config{}, fmt.Errorf("encode notification style config: %w", err)
	}
	return string(value), normalized, nil
}

func Load(ctx context.Context, db *gorm.DB) (Config, error) {
	if db == nil {
		return Config{}, errors.New("notification style database is not initialized")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	var setup model.Setup
	result := db.WithContext(queryCtx).Where("setup_key = ?", SetupKey).Limit(1).Find(&setup)
	if result.Error != nil {
		return Config{}, result.Error
	}
	if result.RowsAffected == 0 {
		return DefaultConfig(), nil
	}
	return Decode(setup.Value)
}

func Save(ctx context.Context, db *gorm.DB, config Config) (Config, error) {
	if db == nil {
		return Config{}, errors.New("notification style database is not initialized")
	}
	value, normalized, err := Encode(config)
	if err != nil {
		return Config{}, err
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	now := database.Now()
	row := model.Setup{Key: SetupKey, Value: value, Type: "json", AddTime: now, EditTime: now}
	err = db.WithContext(queryCtx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "setup_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"setup_value", "setup_type", "setup_edit_time", "updated_at",
		}),
	}).Create(&row).Error
	if err != nil {
		return Config{}, err
	}
	return normalized, nil
}

func defaultByType(notificationType string) (Style, bool) {
	for _, style := range defaults {
		if style.Type == notificationType {
			return style, true
		}
	}
	return Style{}, false
}

func validTone(tone Tone) bool {
	switch tone {
	case TonePrimary, ToneSuccess, ToneWarning, ToneDanger, ToneInfo:
		return true
	default:
		return false
	}
}

func newDefaultStyle(notificationType, label, icon string, tone Tone) Style {
	return Style{Type: notificationType, Label: label, Icon: icon, Tone: tone, DingTalk: defaultDingTalkTemplate()}
}

func defaultDingTalkTemplate() DingTalkTemplate {
	return DingTalkTemplate{
		MessageType: DingTalkMessageTypeAuto,
		Title:       "{{title}}", Content: "{{content}}", URL: "{{url}}", PicURL: "{{picUrl}}",
		SourceName: "{{sourceName}}", MediaID: "{{mediaId}}", ButtonTitle: "查看流程", HeadColor: "FF1677FF",
	}
}

func normalizeDingTalkTemplate(template DingTalkTemplate) DingTalkTemplate {
	template.MessageType = strings.ToLower(strings.TrimSpace(template.MessageType))
	template.Title = strings.TrimSpace(template.Title)
	template.Content = strings.TrimSpace(template.Content)
	template.URL = strings.TrimSpace(template.URL)
	template.PicURL = strings.TrimSpace(template.PicURL)
	template.SourceName = strings.TrimSpace(template.SourceName)
	template.MediaID = strings.TrimSpace(template.MediaID)
	template.ButtonTitle = strings.TrimSpace(template.ButtonTitle)
	template.HeadColor = strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(template.HeadColor), "#"))
	if len(template.HeadColor) == 6 {
		template.HeadColor = "FF" + template.HeadColor
	}
	return template
}

func validateDingTalkTemplate(template DingTalkTemplate) error {
	switch template.MessageType {
	case DingTalkMessageTypeAuto, DingTalkMessageTypeText, DingTalkMessageTypeImage,
		DingTalkMessageTypeVoice, DingTalkMessageTypeFile, DingTalkMessageTypeLink,
		DingTalkMessageTypeOA, DingTalkMessageTypeMarkdown, DingTalkMessageTypeActionCard:
	default:
		return fmt.Errorf("unsupported message type %q", template.MessageType)
	}
	for _, value := range []string{template.Title, template.Content, template.URL, template.PicURL, template.SourceName, template.MediaID, template.ButtonTitle} {
		if utf8.RuneCountInString(value) > 5000 {
			return errors.New("template field is too long")
		}
		for _, token := range templateToken.FindAllString(value, -1) {
			if _, ok := allowedVariables[token]; !ok {
				return fmt.Errorf("unsupported variable %s", token)
			}
		}
	}
	if template.HeadColor != "" && !dingTalkColor.MatchString(template.HeadColor) {
		return errors.New("OA head color must contain 8 hexadecimal characters")
	}
	switch template.MessageType {
	case DingTalkMessageTypeText, DingTalkMessageTypeAuto:
		if template.Content == "" {
			return errors.New("content template is required")
		}
	case DingTalkMessageTypeImage, DingTalkMessageTypeFile:
		if template.MediaID == "" {
			return errors.New("media ID is required")
		}
	case DingTalkMessageTypeVoice:
		if template.MediaID == "" || template.Duration <= 0 {
			return errors.New("voice media ID and duration are required")
		}
	case DingTalkMessageTypeLink, DingTalkMessageTypeOA, DingTalkMessageTypeActionCard:
		if template.Title == "" || template.Content == "" || template.URL == "" {
			return errors.New("title, content and URL templates are required")
		}
	case DingTalkMessageTypeMarkdown:
		if template.Title == "" || template.Content == "" {
			return errors.New("title and content templates are required")
		}
	}
	return nil
}
