package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// FileQuestion 文件/图片上传（type=file）
type FileQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&FileQuestion{BaseQuestion: question.BaseQuestion{T: "file", Dis: "文件/图片", Cat: "media"}})
}

func (q *FileQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请上传文件"}
		}
		return nil
	}
	// 接受 string URL 或 []string URLs
	switch v := value.(type) {
	case string:
		if v == "" && sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请上传文件"}
		}
		return nil
	case []interface{}:
		if len(v) == 0 && sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请上传文件"}
		}
		return nil
	}
	return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "文件题答案格式错误"}
}

func (q *FileQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		out := ""
		for i, item := range v {
			if i > 0 {
				out += ","
			}
			if s, ok := item.(string); ok {
				out += s
			}
		}
		return out
	}
	return ""
}

func (q *FileQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"maxSize": 5 * 1024 * 1024, "accept": "image/*"}
}
