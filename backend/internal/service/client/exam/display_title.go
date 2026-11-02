package exam

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrDisplayTitleSettings = errors.New("考试设置无效，显示名称须为100字以内的文本")

// NormalizeDisplayTitleSettings keeps unrelated settings and numeric identifiers intact.
func NormalizeDisplayTitleSettings(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "{}", nil
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &settings) != nil || settings == nil {
		return "", ErrDisplayTitleSettings
	}
	value, exists := settings["displayTitle"]
	if !exists {
		return raw, nil
	}
	var title string
	if json.Unmarshal(value, &title) != nil {
		return "", ErrDisplayTitleSettings
	}
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > 100 {
		return "", ErrDisplayTitleSettings
	}
	settings["displayTitle"], _ = json.Marshal(title)
	normalized, err := json.Marshal(settings)
	return string(normalized), err
}
