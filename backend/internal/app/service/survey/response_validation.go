package survey

import (
	"strconv"

	"wecheckin/backend/internal/app/formkit/question"
	"wecheckin/backend/internal/app/formkit/schema"
	"wecheckin/backend/internal/model"
)

// userIDToStr 转字符串，匿名时为空
func userIDToStr(uid uint, anonymous bool) string {
	if anonymous {
		return ""
	}
	return strconv.FormatUint(uint64(uid), 10)
}

// ValidateAnswers 答案校验（必填 + regex）
// 失败返回详细错误
func ValidateAnswers(sv *model.Survey, answers map[string]interface{}) []ValidationError {
	sch, err := schema.Parse(sv.Schema)
	if err != nil {
		return []ValidationError{{QuestionID: "", Message: "schema 解析失败: " + err.Error()}}
	}
	var errs []ValidationError
	ans := answers
	for _, q := range sch.Questions {
		val, _ := ans[q.ID]
		if q.Required {
			if val == nil || isEmpty(val) {
				errs = append(errs, ValidationError{QuestionID: q.ID, Message: q.Title + " 必填"})
				continue
			}
		}
		if qt := question.Get(q.Type); qt != nil {
			if verr := qt.Validate(val, q); verr != nil {
				errs = append(errs, ValidationError{QuestionID: q.ID, Message: verr.Error()})
			}
		}
	}
	return errs
}

// ValidationError 校验错误
type ValidationError struct {
	QuestionID string `json:"questionId"`
	Message    string `json:"message"`
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	if a, ok := v.([]interface{}); ok {
		return len(a) == 0
	}
	return false
}
