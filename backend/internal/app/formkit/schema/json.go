package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Marshal 序列化 FormSchema
func Marshal(s *FormSchema) (string, error) {
	if s == nil {
		return "", errors.New("schema is nil")
	}
	if s.Version == "" {
		s.Version = CurrentVersion
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Parse 解析 + 校验 JSON
func Parse(s string) (*FormSchema, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return &FormSchema{Version: CurrentVersion, Questions: []Question{}}, nil
	}
	var out FormSchema
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, fmt.Errorf("invalid schema json: %w", err)
	}
	if out.Version == "" {
		return nil, errors.New("schema missing version field")
	}
	if out.Version != CurrentVersion {
		return nil, fmt.Errorf("unsupported schema version: %s", out.Version)
	}
	seen := make(map[string]bool)
	for i, q := range out.Questions {
		if q.ID == "" {
			return nil, fmt.Errorf("question[%d] missing id", i)
		}
		if seen[q.ID] {
			return nil, fmt.Errorf("duplicate question id: %s", q.ID)
		}
		seen[q.ID] = true
		if q.Type == "" {
			return nil, fmt.Errorf("question %s missing type", q.ID)
		}
		if q.Title == "" {
			return nil, fmt.Errorf("question %s missing title", q.ID)
		}
	}
	return &out, nil
}

// MustParse 解析失败时 panic（用于已知合法数据）
func MustParse(s string) *FormSchema {
	out, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return out
}

// Empty 返回空 schema
func Empty() *FormSchema {
	return &FormSchema{Version: CurrentVersion, Questions: []Question{}}
}
