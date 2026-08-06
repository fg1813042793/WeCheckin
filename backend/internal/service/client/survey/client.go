package survey

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"

	"gorm.io/gorm"
)

type ClientLimitInfo struct {
	DeviceFull bool `json:"deviceFull"`
	IPFull     bool `json:"ipFull"`
}

var publishedSurveyListColumns = []string{
	"survey_id",
	"survey_title",
	"survey_desc",
	"survey_category",
	"survey_tags",
	"survey_cover",
	"survey_visibility",
	"survey_allow_multi",
	"survey_start_time",
	"survey_end_time",
	"survey_max_response",
	"survey_show_result",
	"survey_anonymous",
	"survey_status",
	"survey_mode",
	"survey_order",
	"add_time",
	"edit_time",
	"survey_settings",
}

func (s *SurveyService) PublishedListWithLimitsContext(ctx context.Context, keyword, category string, page, pageSize int, deviceID, clientIP string) ([]model.Survey, int64, map[uint]ClientLimitInfo, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Survey{}).Where("`survey_status` = 1")
	if keyword != "" {
		q = q.Where("`survey_title` LIKE ?", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`survey_category` = ?", category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}
	var list []model.Survey
	if err := q.Select(publishedSurveyListColumns).Order("`survey_order` DESC, `survey_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, nil, err
	}
	limitsMap, err := s.loadSurveyListLimitInfoContext(ctx, db, list, deviceID, clientIP)
	if err != nil {
		return nil, 0, nil, err
	}
	return list, total, limitsMap, nil
}

func (s *SurveyService) loadSurveyListLimitInfoContext(ctx context.Context, db *gorm.DB, list []model.Survey, deviceID, clientIP string) (map[uint]ClientLimitInfo, error) {
	limitsMap := make(map[uint]ClientLimitInfo)
	if len(list) == 0 {
		return limitsMap, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	deviceLimits := make(map[uint]int)
	ipLimits := make(map[uint]int)
	deviceSurveyIDs := make([]uint, 0, len(list))
	ipSurveyIDs := make([]uint, 0, len(list))
	for _, item := range list {
		deviceLimit, ipLimit := surveySubmissionLimits(item.Settings)
		if deviceLimit > 0 && deviceID != "" {
			deviceLimits[item.ID] = deviceLimit
			deviceSurveyIDs = append(deviceSurveyIDs, item.ID)
		}
		if ipLimit > 0 && clientIP != "" {
			ipLimits[item.ID] = ipLimit
			ipSurveyIDs = append(ipSurveyIDs, item.ID)
		}
	}

	if len(deviceSurveyIDs) > 0 {
		countBySurveyID, err := loadSurveyResponseCountsByFilter(db, deviceSurveyIDs, "`survey_resp_device_id` = ?", deviceID)
		if err != nil {
			return nil, err
		}
		for surveyID, limit := range deviceLimits {
			if countBySurveyID[surveyID] >= int64(limit) {
				info := limitsMap[surveyID]
				info.DeviceFull = true
				limitsMap[surveyID] = info
			}
		}
	}
	if len(ipSurveyIDs) > 0 {
		countBySurveyID, err := loadSurveyResponseCountsByFilter(db, ipSurveyIDs, "`survey_resp_ip` = ?", clientIP)
		if err != nil {
			return nil, err
		}
		for surveyID, limit := range ipLimits {
			if countBySurveyID[surveyID] >= int64(limit) {
				info := limitsMap[surveyID]
				info.IPFull = true
				limitsMap[surveyID] = info
			}
		}
	}
	return limitsMap, nil
}

func loadSurveyResponseCountsByFilter(db *gorm.DB, surveyIDs []uint, filter string, arg interface{}) (map[uint]int64, error) {
	type countRow struct {
		SurveyID uint  `gorm:"column:survey_resp_survey_id"`
		Count    int64 `gorm:"column:cnt"`
	}
	var rows []countRow
	if err := db.Model(&model.SurveyResponse{}).
		Select("`survey_resp_survey_id`, COUNT(*) AS cnt").
		Where("`survey_resp_survey_id` IN ? AND `survey_resp_status` = 1", surveyIDs).
		Where(filter, arg).
		Group("`survey_resp_survey_id`").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	countBySurveyID := make(map[uint]int64, len(rows))
	for _, row := range rows {
		countBySurveyID[row.SurveyID] = row.Count
	}
	return countBySurveyID, nil
}

func (s *SurveyService) PublishedSurveyContext(ctx context.Context, id uint) (*model.Survey, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var survey model.Survey
	if err := db.Where("`survey_id` = ? AND `survey_status` = 1", id).First(&survey).Error; err != nil {
		return nil, err
	}
	return &survey, nil
}

func (s *SurveyService) UserDeptIDContext(ctx context.Context, userID uint) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var userDept model.UserDept
	if err := db.Where("`user_dept_user_id` = ?", userID).First(&userDept).Error; err != nil {
		return 0, err
	}
	return userDept.DeptID, nil
}

func (s *SurveyService) ValidatePublicSurveyContext(ctx context.Context, surveyID uint, deviceID, clientIP string) (string, string, error) {
	survey, err := s.PublishedSurveyContext(ctx, surveyID)
	if err != nil {
		return "", "", err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if survey.MaxResponse > 0 {
		var count int64
		if err := db.Model(&model.SurveyResponse{}).
			Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).
			Count(&count).Error; err != nil {
			return "", "", err
		}
		if count >= int64(survey.MaxResponse) {
			logger.Logger.Printf("[SurveyValidate] 回收上限已满 surveyId=%d max=%d current=%d", surveyID, survey.MaxResponse, count)
			return survey.Schema, "回收上限已满", nil
		}
	}
	deviceLimit, ipLimit := surveySubmissionLimits(survey.Settings)
	if deviceLimit > 0 && deviceID != "" {
		var deviceCount int64
		if err := db.Model(&model.SurveyResponse{}).
			Where("`survey_resp_survey_id` = ? AND `survey_resp_device_id` = ? AND `survey_resp_status` = 1", surveyID, deviceID).
			Count(&deviceCount).Error; err != nil {
			return "", "", err
		}
		if deviceCount >= int64(deviceLimit) {
			logger.Logger.Printf("[SurveyValidate] 设备次数上限 surveyId=%d limit=%d current=%d deviceId=%s", surveyID, deviceLimit, deviceCount, deviceID)
			return survey.Schema, "该设备答题次数已达上限", nil
		}
	}
	if ipLimit > 0 && clientIP != "" {
		var ipCount int64
		if err := db.Model(&model.SurveyResponse{}).
			Where("`survey_resp_survey_id` = ? AND `survey_resp_ip` = ? AND `survey_resp_status` = 1", surveyID, clientIP).
			Count(&ipCount).Error; err != nil {
			return "", "", err
		}
		if ipCount >= int64(ipLimit) {
			logger.Logger.Printf("[SurveyValidate] IP次数上限 surveyId=%d limit=%d current=%d ip=%s", surveyID, ipLimit, ipCount, clientIP)
			return survey.Schema, "该IP答题次数已达上限", nil
		}
	}
	return survey.Schema, "", nil
}

func surveySubmissionLimits(settingsJSON string) (int, int) {
	var deviceLimit, ipLimit int
	if settingsJSON == "" {
		return deviceLimit, ipLimit
	}
	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJSON), &settingsMap); err != nil {
		return deviceLimit, ipLimit
	}
	deviceLimit = positiveLimitValue(settingsMap["deviceLimit"])
	ipLimit = positiveLimitValue(settingsMap["ipLimit"])
	return deviceLimit, ipLimit
}

func positiveLimitValue(value interface{}) int {
	switch v := value.(type) {
	case float64:
		if v <= 0 {
			return 0
		}
		return int(v)
	case string:
		limit, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil || limit <= 0 {
			return 0
		}
		return int(limit)
	default:
		return 0
	}
}

func (r *ResponseService) MyResponsesContext(ctx context.Context, userID uint, limit int) ([]model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	uidStr := strconv.FormatUint(uint64(userID), 10)
	var list []model.SurveyResponse
	err := db.Where("`survey_resp_user_id` = ? AND `survey_resp_status` = 1", uidStr).
		Order("`survey_resp_id` DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
