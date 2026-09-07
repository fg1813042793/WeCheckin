package workflowcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxInstanceTitleTemplateLength = 500
	maxInstanceTitleLength         = 200
	maxBusinessPeriodOffset        = 120
)

var instanceIdentityTokenPattern = regexp.MustCompile(`\{\{([^{}]+)\}\}`)

func validateInstanceIdentity(config *InstanceIdentityConfig, fields map[string]FormField, availability *StartAvailabilityConfig) []ValidationError {
	if config == nil {
		return nil
	}
	template := strings.TrimSpace(config.TitleTemplate)
	if template == "" || utf8.RuneCountInString(template) > maxInstanceTitleTemplateLength {
		return instanceIdentityValidationError("单据标题模板长度无效")
	}
	periodEnabled := config.BusinessPeriod != nil && config.BusinessPeriod.Enabled
	valid := true
	instanceIdentityTokenPattern.ReplaceAllStringFunc(template, func(token string) string {
		matches := instanceIdentityTokenPattern.FindStringSubmatch(token)
		if len(matches) != 2 || strings.TrimSpace(matches[1]) != matches[1] {
			valid = false
			return ""
		}
		name := matches[1]
		switch name {
		case "workflowName", "starterName":
		case "businessPeriod":
			valid = valid && periodEnabled
		default:
			if !strings.HasPrefix(name, "field.") {
				valid = false
				return ""
			}
			field, exists := fields[strings.TrimPrefix(name, "field.")]
			valid = valid && exists && instanceTitleFieldSupported(field.Type)
		}
		return ""
	})
	if !valid || strings.Contains(instanceIdentityTokenPattern.ReplaceAllString(template, ""), "{{") || strings.Contains(instanceIdentityTokenPattern.ReplaceAllString(template, ""), "}}") {
		return instanceIdentityValidationError("单据标题模板包含不支持的变量")
	}
	if !periodEnabled {
		return nil
	}
	period := config.BusinessPeriod
	if period.Offset < -maxBusinessPeriodOffset || period.Offset > maxBusinessPeriodOffset {
		return instanceIdentityValidationError("业务期间偏移量无效")
	}
	switch strings.TrimSpace(period.Granularity) {
	case BusinessPeriodGranularityDay, BusinessPeriodGranularityWeek, BusinessPeriodGranularityMonth,
		BusinessPeriodGranularityQuarter, BusinessPeriodGranularityYear:
	default:
		return instanceIdentityValidationError("业务期间粒度无效")
	}
	switch strings.TrimSpace(period.Source) {
	case BusinessPeriodSourceSubmitTime:
		if strings.TrimSpace(period.Field) != "" {
			return instanceIdentityValidationError("提交时间业务期间不能指定表单字段")
		}
	case BusinessPeriodSourceAvailabilityWindowStart, BusinessPeriodSourceAvailabilityWindowEnd:
		if strings.TrimSpace(period.Field) != "" {
			return instanceIdentityValidationError("开放窗口业务期间不能指定表单字段")
		}
		if !availabilitySupportsBusinessPeriod(availability) {
			return instanceIdentityValidationError("业务期间选择开放窗口时，流程必须配置指定时间段、每周、每月或跨月开放")
		}
	case BusinessPeriodSourceFormField:
		field, exists := fields[strings.TrimSpace(period.Field)]
		if !exists || (field.Type != FormFieldTypeDate && field.Type != FormFieldTypeDateTime) {
			return instanceIdentityValidationError("业务期间必须关联日期或日期时间字段")
		}
	default:
		return instanceIdentityValidationError("业务期间取值来源无效")
	}
	return nil
}

func availabilitySupportsBusinessPeriod(availability *StartAvailabilityConfig) bool {
	if availability == nil {
		return false
	}
	switch strings.TrimSpace(availability.Mode) {
	case StartAvailabilityFixed, StartAvailabilityWeekly, StartAvailabilityMonthly, StartAvailabilityMonthlyWindow:
		return true
	default:
		return false
	}
}

func instanceIdentityValidationError(message string) []ValidationError {
	return []ValidationError{{Code: ValidationInstanceIdentity, Message: message}}
}

func instanceTitleFieldSupported(fieldType string) bool {
	switch fieldType {
	case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypeNumber, FormFieldTypeAmount,
		FormFieldTypePhone, FormFieldTypeEmail, FormFieldTypeSelect, FormFieldTypeRadio,
		FormFieldTypeDate, FormFieldTypeDateTime, FormFieldTypeTime, FormFieldTypeBoolean,
		FormFieldTypeCalculation:
		return true
	default:
		return false
	}
}

func ResolveInstanceIdentity(
	definition Definition,
	submittedAt time.Time,
	starterName string,
	formData map[string]interface{},
) (InstanceIdentity, error) {
	config := definition.InstanceIdentity
	if config == nil {
		return InstanceIdentity{Title: defaultInstanceTitle(definition.EffectiveName(), starterName)}, nil
	}
	fields, _ := validateFormSchema(definition.Form)
	if errors := validateInstanceIdentity(config, fields, definitionStartAvailability(definition)); len(errors) > 0 {
		return InstanceIdentity{}, ValidationErrors(errors)
	}
	identity := InstanceIdentity{}
	if config.BusinessPeriod != nil && config.BusinessPeriod.Enabled {
		periodTime, err := resolveBusinessPeriodTime(definition, *config.BusinessPeriod, submittedAt, formData)
		if err != nil {
			return InstanceIdentity{}, err
		}
		periodTime = offsetBusinessPeriod(periodTime, config.BusinessPeriod.Granularity, config.BusinessPeriod.Offset)
		identity.BusinessPeriodType = config.BusinessPeriod.Granularity
		identity.BusinessPeriodKey, identity.BusinessPeriodLabel = formatBusinessPeriod(periodTime, config.BusinessPeriod.Granularity)
	}

	replacements := map[string]string{
		"workflowName":   definition.EffectiveName(),
		"starterName":    starterName,
		"businessPeriod": identity.BusinessPeriodLabel,
	}
	for key, field := range fields {
		replacements["field."+key] = formatInstanceTitleField(field, formData[key])
	}
	title := instanceIdentityTokenPattern.ReplaceAllStringFunc(config.TitleTemplate, func(token string) string {
		matches := instanceIdentityTokenPattern.FindStringSubmatch(token)
		if len(matches) != 2 {
			return ""
		}
		return replacements[matches[1]]
	})
	title = strings.Join(strings.Fields(title), " ")
	if title == "" {
		title = defaultInstanceTitle(definition.EffectiveName(), starterName)
	}
	identity.Title = truncateRunes(title, maxInstanceTitleLength)
	return identity, nil
}

func resolveBusinessPeriodTime(definition Definition, config BusinessPeriodConfig, submittedAt time.Time, formData map[string]interface{}) (time.Time, error) {
	location, err := time.LoadLocation(startLimitTimezone(definitionStartAvailability(definition)))
	if err != nil {
		return time.Time{}, err
	}
	localSubmittedAt := submittedAt.In(location)
	switch config.Source {
	case BusinessPeriodSourceSubmitTime:
		return localSubmittedAt, nil
	case BusinessPeriodSourceAvailabilityWindowStart, BusinessPeriodSourceAvailabilityWindowEnd:
		window, ok := resolveAvailabilityStartLimitWindow(definitionStartAvailability(definition), localSubmittedAt)
		if !ok {
			return time.Time{}, errors.New("当前提交时间无法解析业务期间开放窗口")
		}
		millis := window.StartsAt
		if config.Source == BusinessPeriodSourceAvailabilityWindowEnd {
			millis = window.EndsAt
		}
		return time.UnixMilli(millis).In(location), nil
	case BusinessPeriodSourceFormField:
		value, exists := formData[strings.TrimSpace(config.Field)]
		if !exists {
			return time.Time{}, fmt.Errorf("业务期间字段 %s 不能为空", config.Field)
		}
		return parseBusinessPeriodFormTime(value, location)
	default:
		return time.Time{}, errors.New("业务期间取值来源无效")
	}
}

func definitionStartAvailability(definition Definition) *StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == NodeTypeStart {
			return definition.Nodes[index].Availability
		}
	}
	return nil
}

func parseBusinessPeriodFormTime(value interface{}, location *time.Location) (time.Time, error) {
	if millis, ok := numericMillis(value); ok {
		return time.UnixMilli(millis).In(location), nil
	}
	raw := strings.TrimSpace(fmt.Sprint(value))
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04", "2006-01-02"} {
		var parsed time.Time
		var err error
		if layout == time.RFC3339Nano {
			parsed, err = time.Parse(layout, raw)
		} else {
			parsed, err = time.ParseInLocation(layout, raw, location)
		}
		if err == nil {
			return parsed.In(location), nil
		}
	}
	return time.Time{}, errors.New("业务期间字段必须是有效的日期")
}

func numericMillis(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func offsetBusinessPeriod(value time.Time, granularity string, offset int) time.Time {
	switch granularity {
	case BusinessPeriodGranularityDay:
		return value.AddDate(0, 0, offset)
	case BusinessPeriodGranularityWeek:
		return value.AddDate(0, 0, offset*7)
	case BusinessPeriodGranularityMonth:
		return value.AddDate(0, offset, 0)
	case BusinessPeriodGranularityQuarter:
		return value.AddDate(0, offset*3, 0)
	case BusinessPeriodGranularityYear:
		return value.AddDate(offset, 0, 0)
	default:
		return value
	}
}

func formatBusinessPeriod(value time.Time, granularity string) (string, string) {
	switch granularity {
	case BusinessPeriodGranularityDay:
		return value.Format("2006-01-02"), fmt.Sprintf("%d年%d月%d日", value.Year(), value.Month(), value.Day())
	case BusinessPeriodGranularityWeek:
		year, week := value.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week), fmt.Sprintf("%d年第%d周", year, week)
	case BusinessPeriodGranularityMonth:
		return value.Format("2006-01"), fmt.Sprintf("%d年%d月", value.Year(), value.Month())
	case BusinessPeriodGranularityQuarter:
		quarter := (int(value.Month())-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", value.Year(), quarter), fmt.Sprintf("%d年第%d季度", value.Year(), quarter)
	case BusinessPeriodGranularityYear:
		return fmt.Sprintf("%d", value.Year()), fmt.Sprintf("%d年", value.Year())
	default:
		return "", ""
	}
}

func formatInstanceTitleField(field FormField, value interface{}) string {
	if value == nil {
		return ""
	}
	if field.Type == FormFieldTypeBoolean {
		if typed, ok := value.(bool); ok {
			if typed {
				return "是"
			}
			return "否"
		}
	}
	if field.Type == FormFieldTypeSelect || field.Type == FormFieldTypeRadio {
		if label := findFormOptionLabel(field.Options, fmt.Sprint(value)); label != "" {
			return label
		}
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func findFormOptionLabel(options []FormOption, value string) string {
	for _, option := range options {
		if option.Value == value {
			return option.Label
		}
		if label := findFormOptionLabel(option.Children, value); label != "" {
			return label
		}
	}
	return ""
}

func defaultInstanceTitle(workflowName, starterName string) string {
	workflowName = strings.TrimSpace(workflowName)
	starterName = strings.TrimSpace(starterName)
	if starterName == "" {
		return truncateRunes(workflowName, maxInstanceTitleLength)
	}
	return truncateRunes(starterName+"提交的"+workflowName, maxInstanceTitleLength)
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	return string([]rune(value)[:limit])
}
