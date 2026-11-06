package exam

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

var ErrDisplayTitleSettings = errors.New("考试设置无效")

// NormalizeDisplayTitleSettings keeps unrelated settings and numeric identifiers intact.
func NormalizeDisplayTitleSettings(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "{}", nil
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &settings) != nil || settings == nil {
		return "", ErrDisplayTitleSettings
	}
	if value, exists := settings["displayTitle"]; exists {
		var title string
		if json.Unmarshal(value, &title) != nil {
			return "", ErrDisplayTitleSettings
		}
		title = strings.TrimSpace(title)
		if utf8.RuneCountInString(title) > 100 {
			return "", ErrDisplayTitleSettings
		}
		settings["displayTitle"], _ = json.Marshal(title)
	}
	if value, exists := settings["answerDownloadFormat"]; exists {
		var format string
		if json.Unmarshal(value, &format) != nil {
			return "", ErrDisplayTitleSettings
		}
		format = strings.ToLower(strings.TrimSpace(format))
		if format != "pdf" && format != "docx" && format != "xlsx" {
			return "", ErrDisplayTitleSettings
		}
		settings["answerDownloadFormat"], _ = json.Marshal(format)
	}
	if value, exists := settings["answerDownloadExpireDays"]; exists {
		var days int
		if json.Unmarshal(value, &days) != nil || days < 1 || days > 30 {
			return "", ErrDisplayTitleSettings
		}
		settings["answerDownloadExpireDays"], _ = json.Marshal(days)
	}
	if value, exists := settings["answerDownloadBaseUrl"]; exists {
		var baseURL string
		if json.Unmarshal(value, &baseURL) != nil {
			return "", ErrDisplayTitleSettings
		}
		baseURL = strings.TrimSpace(baseURL)
		if baseURL != "" {
			parsed, err := url.Parse(baseURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" || utf8.RuneCountInString(baseURL) > 512 {
				return "", ErrDisplayTitleSettings
			}
			baseURL = parsed.Scheme + "://" + parsed.Host
		}
		settings["answerDownloadBaseUrl"], _ = json.Marshal(baseURL)
	}
	normalized, err := json.Marshal(settings)
	return string(normalized), err
}
