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
	TypeWorkflow               = "workflow"
	TypeAdminManual            = "admin_manual"
	TypeScheduledTask          = "scheduled_task"
	TypeSurveyStat             = "survey_stat"
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
	ErrUnknownType   = errors.New("notification style type is unsupported")
	ErrDuplicateType = errors.New("notification style type is duplicated")
	ErrInvalidLabel  = errors.New("notification style label is invalid")
	ErrInvalidIcon   = errors.New("notification style icon is invalid")
	ErrInvalidTone   = errors.New("notification style tone is invalid")

	iconPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
	defaults    = []Style{
		{Type: TypeTaskArrived, Label: "待处理", Icon: "clock", Tone: ToneWarning},
		{Type: TypeTaskReminder, Label: "处理提醒", Icon: "bell", Tone: ToneWarning},
		{Type: TypeApprovalResultApproved, Label: "审批通过", Icon: "checkmark-circle", Tone: ToneSuccess},
		{Type: TypeApprovalResultRejected, Label: "审批驳回", Icon: "error-circle", Tone: ToneDanger},
		{Type: TypeApprovalResultReturned, Label: "审批退回", Icon: "reload", Tone: ToneWarning},
		{Type: TypeNodeCC, Label: "流程抄送", Icon: "share", Tone: ToneInfo},
		{Type: TypeNodeNotify, Label: "流程通知", Icon: "email", Tone: TonePrimary},
		{Type: TypeInstanceCommented, Label: "流程评论", Icon: "chat", Tone: TonePrimary},
		{Type: TypeWorkflow, Label: "流程消息", Icon: "file-text", Tone: ToneInfo},
		{Type: TypeAdminManual, Label: "系统通知", Icon: "email", Tone: TonePrimary},
		{Type: TypeScheduledTask, Label: "定时通知", Icon: "clock", Tone: ToneInfo},
		{Type: TypeSurveyStat, Label: "问卷统计", Icon: "file-text", Tone: ToneInfo},
	}
)

type Style struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Tone  Tone   `json:"tone"`
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
			Type:  strings.TrimSpace(raw.Type),
			Label: strings.TrimSpace(raw.Label),
			Icon:  strings.TrimSpace(raw.Icon),
			Tone:  Tone(strings.TrimSpace(string(raw.Tone))),
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

func StyleFor(config Config, notificationType string) Style {
	notificationType = strings.TrimSpace(notificationType)
	for _, style := range config.Styles {
		if strings.TrimSpace(style.Type) == notificationType {
			return style
		}
	}
	if style, ok := defaultByType(notificationType); ok {
		return style
	}
	return Style{Type: notificationType, Label: "系统消息", Icon: "email", Tone: ToneInfo}
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
