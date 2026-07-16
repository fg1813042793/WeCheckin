package exam

import (
	"context"
	"encoding/json"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
)

func checkExamLimit(e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	return checkExamLimitContext(context.Background(), e, uidStr, device, deviceId, ip)
}

func checkExamLimitContext(ctx context.Context, e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	if e.MaxResponse > 0 {
		var cnt int64
		db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", e.ID).Count(&cnt)
		if int(cnt) >= e.MaxResponse {
			logger.Logger.Printf("[ExamCheckLimit] 答卷上限 examId=%d max=%d current=%d", e.ID, e.MaxResponse, cnt)
			return "已达最大答卷数"
		}
	}
	if e.MaxAttempts > 0 && uidStr != "" {
		var cnt int64
		db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` >= 1", e.ID, uidStr).Count(&cnt)
		if int(cnt) >= e.MaxAttempts {
			logger.Logger.Printf("[ExamCheckLimit] 个人答题次数上限 examId=%d uid=%s max=%d current=%d", e.ID, uidStr, e.MaxAttempts, cnt)
			return "已达最大答题次数"
		}
	}
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
	if deviceLimit, _ := settingsMap["deviceLimit"].(float64); deviceLimit > 0 && deviceId != "" {
		var cnt int64
		db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_device_id` = ? AND `exam_r_status` >= 1", e.ID, deviceId).Count(&cnt)
		if int(cnt) >= int(deviceLimit) {
			logger.Logger.Printf("[ExamCheckLimit] 设备次数上限 examId=%d limit=%d current=%d deviceId=%s", e.ID, int(deviceLimit), cnt, deviceId)
			return "已达每台设备最大答题次数"
		}
	}
	if ipLimit, _ := settingsMap["ipLimit"].(float64); ipLimit > 0 && ip != "" {
		var cnt int64
		db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_add_ip` = ? AND `exam_r_status` >= 1", e.ID, ip).Count(&cnt)
		if int(cnt) >= int(ipLimit) {
			logger.Logger.Printf("[ExamCheckLimit] IP次数上限 examId=%d limit=%d current=%d ip=%s", e.ID, int(ipLimit), cnt, ip)
			return "已达每个IP最大答题次数"
		}
	}
	return ""
}
