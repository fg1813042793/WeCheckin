package auditlog

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var publicSnapshotSettings = map[string]string{
	"settings.ADMIN_LOGIN_CAPTCHA_ENABLED":       "boolean",
	"settings.DINGTALK_H5_LOGIN_CAPTCHA_ENABLED": "boolean",
	"settings.USER_SINGLE_LOGIN":                 "boolean",
	"settings.ADMIN_SINGLE_LOGIN":                "boolean",
	"settings.DINGTALK_H5_SINGLE_LOGIN":          "boolean",
	"settings.DINGTALK_H5_SELF_BIND_ENABLED":     "boolean",
	"settings.DINGTALK_H5_NOTIFY_ENABLED":        "boolean",
	"settings.TOKEN_ADMIN_EXPIRE":                "duration",
	"settings.TOKEN_USER_EXPIRE":                 "duration",
	"settings.TOKEN_DINGTALK_H5_EXPIRE":          "duration",
	"settings.DINGTALK_H5_NOTIFY_MODE":           "notify_mode",
}
var publicCorpSnapshotField = regexp.MustCompile(`^settings\.corp\.[1-9][0-9]*\.(enabled|notify_enabled|notify_mode)$`)

// PublicSnapshotValue only accepts known settings and validated scalar values.
// A matching name alone never makes arbitrary stored content safe to log.
func PublicSnapshotValue(field string, value interface{}) (interface{}, bool) {
	kind := publicSnapshotSettings[field]
	if match := publicCorpSnapshotField.FindStringSubmatch(field); len(match) > 0 {
		kind = "boolean"
		if match[1] == "notify_mode" {
			kind = "notify_mode"
		}
	}
	if kind == "" {
		return nil, false
	}
	if value == nil {
		return nil, true
	}
	if kind == "boolean" {
		return publicSnapshotBoolean(value)
	}
	text, ok := value.(string)
	if !ok || len(text) > 64 {
		return nil, false
	}
	text = strings.TrimSpace(text)
	switch kind {
	case "duration":
		duration, err := time.ParseDuration(text)
		if err != nil || duration <= 0 {
			return nil, false
		}
		return duration.String(), true
	case "notify_mode":
		text = strings.ToLower(text)
		if text == "agent" || text == "agent_fallback" || text == "robot" {
			return text, true
		}
	}
	return nil, false
}

func publicSnapshotBoolean(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			return true, true
		case "false", "0":
			return false, true
		}
	case json.Number:
		n, err := strconv.ParseFloat(v.String(), 64)
		if err == nil {
			return publicSnapshotBoolean(n)
		}
	case int:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	case int64:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	case uint:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	case uint64:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	case float64:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	}
	return nil, false
}
