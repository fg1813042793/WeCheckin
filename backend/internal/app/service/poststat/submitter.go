package poststat

import (
	"encoding/json"

	"wecheckin-backend/backend/internal/app/formkit/schema"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func resolveSubmitter(schemaJSON, currentAnswers, nickname string, userID uint) string {
	sch, err := schema.Parse(schemaJSON)
	if err == nil && currentAnswers != "" {
		var ans map[string]interface{}
		if json.Unmarshal([]byte(currentAnswers), &ans) == nil {
			personalTypes := []string{"name", "phone", "email", "studentId", "employeeId"}
			for _, pt := range personalTypes {
				for _, q := range sch.Questions {
					if q.Type == pt {
						if val, ok := ans[q.ID]; ok && val != nil {
							if s, ok := val.(string); ok && s != "" {
								return s
							}
						}
					}
				}
			}
		}
	}
	if nickname != "" {
		return nickname
	}
	if userID > 0 {
		var u model.User
		if database.DB.Where("`id` = ?", userID).First(&u).Error == nil && u.Name != "" {
			return u.Name
		}
	}
	return "匿名用户"
}
