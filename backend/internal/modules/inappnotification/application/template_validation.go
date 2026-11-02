package application

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxTemplateVersion = 2147483647
const MaxTemplateID = uint64(9007199254740991)

var templateColor = regexp.MustCompile(`^[0-9A-F]{8}$`)

func invalidTemplate(message string) error { return &TemplateValidationError{Message: message} }

func normalizeTemplateInput(input DingTalkTemplateInput) (DingTalkTemplateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Title = strings.TrimSpace(input.Title)
	input.ButtonTitle = strings.TrimSpace(input.ButtonTitle)
	input.HeadColor = strings.ToUpper(strings.TrimSpace(input.HeadColor))
	if input.ButtonTitle == "" {
		input.ButtonTitle = "打开页面"
	}
	if input.HeadColor == "" {
		input.HeadColor = "FF1677FF"
	}
	for _, field := range []struct {
		value, label string
		max          int
	}{
		{input.Name, "模板名称", 80}, {input.Title, "默认标题", 255}, {input.Content, "默认正文", 5000}, {input.ButtonTitle, "按钮文字", 20},
	} {
		if !utf8.ValidString(field.value) || strings.TrimSpace(field.value) == "" || utf8.RuneCountInString(field.value) > field.max {
			return input, invalidTemplate(fmt.Sprintf("%s不能为空且最多%d个字符", field.label, field.max))
		}
	}
	if !templateColor.MatchString(input.HeadColor) {
		return input, invalidTemplate("OA颜色必须为8位十六进制字符")
	}
	switch input.MessageType {
	case "oa", "link", "action_card":
	default:
		return input, invalidTemplate("消息类型仅支持OA、链接和卡片")
	}
	// Display names from clients are never trusted or persisted.
	input.Target.DefinitionName = ""
	switch input.Target.Kind {
	case "app_home":
		if input.Target.DefinitionID != 0 || input.Target.URL != "" {
			return input, invalidTemplate("应用首页目标不能包含流程或自定义地址")
		}
	case "workflow_start":
		if input.Target.DefinitionID == 0 || uint64(input.Target.DefinitionID) > MaxTemplateID || input.Target.URL != "" {
			return input, invalidTemplate("请选择有效的目标流程")
		}
	case "url":
		if input.Target.DefinitionID != 0 || !ValidTemplateHTTPSURL(input.Target.URL) {
			return input, invalidTemplate("目标地址必须为不含账号密码和控制字符的完整HTTPS地址，最多2048字符")
		}
	default:
		return input, invalidTemplate("跳转目标类型无效")
	}
	return input, nil
}

// ValidTemplateHTTPSURL validates a navigation target without requesting it.

func ValidTemplateHTTPSURL(value string) bool {
	if !utf8.ValidString(value) || len(value) == 0 || utf8.RuneCountInString(value) > 2048 || strings.TrimSpace(value) != value {
		return false
	}
	decoded, err := url.PathUnescape(value)
	if err != nil || strings.ContainsFunc(decoded, unicode.IsControl) {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Hostname() != "" && parsed.User == nil && parsed.Opaque == "" && !strings.ContainsAny(parsed.Host, "\\ \t\r\n")
}

func NormalizeDingTalkTemplateQuery(query DingTalkTemplateQuery) (DingTalkTemplateQuery, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.TargetKind = strings.TrimSpace(query.TargetKind)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.Page < 1 || query.Page > 10000 || query.PageSize < 1 || query.PageSize > 100 || !utf8.ValidString(query.Keyword) || utf8.RuneCountInString(query.Keyword) > 80 || uint64(query.ID) > MaxTemplateID {
		return query, invalidTemplate("筛选参数无效：关键词最多80字，页码1至10000，每页1至100条")
	}
	if query.TargetKind != "" && query.TargetKind != "app_home" && query.TargetKind != "workflow_start" && query.TargetKind != "url" {
		return query, invalidTemplate("模板目标类型筛选无效")
	}
	return query, nil
}

func validateTemplateVersion(version int) error {
	if version <= 0 || version > MaxTemplateVersion {
		return invalidTemplate("请提供有效的模板版本号并刷新后重试")
	}
	return nil
}
