package event

import (
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func SaveEventScore(eventID, participantID, score, judgeID string) error {
	// Upsert: find existing or create new
	var existing model.EventScore
	result := database.DB.Where("`event_score_event_id` = ? AND `event_score_participant_id` = ?", eventID, participantID).First(&existing)
	if result.Error == nil {
		return database.DB.Model(&existing).Updates(map[string]interface{}{
			"event_score_score":     score,
			"event_score_judge_id":  judgeID,
			"event_score_edit_time": database.Now(),
		}).Error
	}
	es := model.EventScore{
		EventID:       uint(parseUint(eventID)),
		ParticipantID: participantID,
		Score:         score,
		JudgeID:       judgeID,
		AddTime:       database.Now(),
	}
	return database.DB.Create(&es).Error
}

func GetEventScores(eventID string, page, pageSize int) (map[string]interface{}, error) {
	var list []model.EventScore
	var total int64
	query := database.DB.Model(&model.EventScore{}).Where("`event_score_event_id` = ?", eventID)
	query.Count(&total)
	err := query.Order("`event_score_add_time` ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var user model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].ParticipantID).First(&user)
		list[i].ParticipantName = user.Name
		list[i].ParticipantAvatar = media.FullURLWithStaticDomain(user.Pic)
		if user.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", user.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].ParticipantDept = dept.Name
				list[i].ParticipantTopDept = getTopDeptName(ud.DeptID)
			}
		}
	}
	return map[string]interface{}{"list": list, "total": total}, nil
}
