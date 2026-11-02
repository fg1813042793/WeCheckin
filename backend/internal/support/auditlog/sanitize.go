package auditlog

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"
)

func CleanText(value string, limit int) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > limit {
		runes = runes[:limit]
	}
	return string(runes)
}

func SensitiveField(key string) bool {
	key = strings.ToLower(key)
	for _, marker := range []string{"password", "pwd", "token", "secret", "captcha", "authorization", "authcode", "bindticket", "privatekey", "appkey", "accesskey", "credential", "signature", "answer", "schema", "content", "description", "remark", "mobile", "phone", "email"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func SafeChanges(changes []Change) json.RawMessage {
	data, err := EncodeChanges(changes)
	if err != nil {
		return json.RawMessage(`[]`)
	}
	return data
}

var ErrChangeLimit = errors.New("audit changes exceed safe collection limits")

func EncodeChanges(changes []Change) (json.RawMessage, error) {
	if len(changes) > 100 {
		return nil, ErrChangeLimit
	}
	safe := make([]Change, 0, len(changes))
	for _, change := range changes {
		if strings.TrimSpace(change.Field) == "" || len([]rune(change.Field)) > 100 {
			return nil, ErrChangeLimit
		}
		change.Field = CleanText(change.Field, 100)
		if strings.EqualFold(change.Field, "settings") || strings.HasPrefix(strings.ToLower(change.Field), "settings.") {
			before, beforeOK := PublicSnapshotValue(change.Field, change.Before)
			after, afterOK := PublicSnapshotValue(change.Field, change.After)
			if beforeOK && afterOK {
				change.Before, change.After = before, after
			} else {
				change.Before, change.After = "[已隐藏]", "[已变更]"
			}
		} else if SensitiveField(change.Field) {
			change.Before, change.After = "[已隐藏]", "[已变更]"
		} else {
			var err error
			change.Before, err = safeValue(change.Before, 0)
			if err != nil {
				return nil, err
			}
			change.After, err = safeValue(change.After, 0)
			if err != nil {
				return nil, err
			}
		}
		safe = append(safe, change)
	}
	data, err := json.Marshal(safe)
	if err != nil || len(data) > 65535 {
		return nil, ErrChangeLimit
	}
	return data, nil
}

func safeValue(value interface{}, depth int) (interface{}, error) {
	if depth > 4 {
		return nil, ErrChangeLimit
	}
	switch v := value.(type) {
	case nil, bool, int, int64, uint, uint64, float64, json.Number:
		return v, nil
	case string:
		if len([]rune(v)) > 512 {
			return nil, ErrChangeLimit
		}
		return CleanText(v, 512), nil
	case []string:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		result := make([]string, len(v))
		for i := range v {
			if len([]rune(v[i])) > 512 {
				return nil, ErrChangeLimit
			}
			result[i] = CleanText(v[i], 512)
		}
		return result, nil
	case []interface{}:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		result := make([]interface{}, len(v))
		for i := range v {
			var err error
			result[i], err = safeValue(v[i], depth+1)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case []map[string]interface{}:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		result := make([]interface{}, len(v))
		for i := range v {
			var err error
			result[i], err = safeValue(v[i], depth+1)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case []uint64:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		return v, nil
	case []uint:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		return v, nil
	case []int:
		if len(v) > 1024 {
			return nil, ErrChangeLimit
		}
		return v, nil
	case map[string]interface{}:
		if len(v) > 30 {
			return nil, ErrChangeLimit
		}
		result := make(map[string]interface{})
		for key, item := range v {
			if len([]rune(key)) > 100 {
				return nil, ErrChangeLimit
			}
			if SensitiveField(key) {
				result[CleanText(key, 100)] = "[已隐藏]"
			} else {
				var err error
				result[CleanText(key, 100)], err = safeValue(item, depth+1)
				if err != nil {
					return nil, err
				}
			}
		}
		return result, nil
	default:
		return nil, ErrChangeLimit
	}
}
