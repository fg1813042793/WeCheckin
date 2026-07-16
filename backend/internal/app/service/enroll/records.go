package enroll

import (
	"encoding/json"
	"strconv"
	"strings"

	"wecheckin-backend/backend/internal/app/formkit/schema"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetEnrollJoinByDay(enrollID, day string) ([]map[string]interface{}, error) {
	var joins []model.EnrollJoin
	query := database.DB.Where("`enroll_join_enroll_id` = ? AND `enroll_join_day` = ?", enrollID, day)
	query.Order("`enroll_join_add_time` DESC").Find(&joins)

	// Get enroll form definitions for type mapping
	var enrollModel model.Enroll
	var typeMap map[string]string // label -> type
	if len(joins) > 0 {
		database.DB.Where("`id` = ?", enrollID).First(&enrollModel)
		if enrollModel.Forms != "" {
			// 兼容老/新 schema 格式
			for _, fv := range schema.ExtractFieldValues("", enrollModel.Forms) {
				if fv.Label != "" && fv.Type != "" {
					if typeMap == nil {
						typeMap = make(map[string]string)
					}
					typeMap[fv.Label] = fv.Type
				}
			}
		}
	}

	// Get user info
	userMap := map[string]model.User{}
	var allUsers []model.User
	database.DB.Find(&allUsers)
	for _, u := range allUsers {
		userMap[u.MiniOpenID] = u
	}

	var result []map[string]interface{}
	for _, j := range joins {
		u, _ := userMap[j.UserID]
		item := map[string]interface{}{
			"id":         j.ID,
			"userId":     j.UserID,
			"userName":   u.Name,
			"userAvatar": media.FullURLWithStaticDomain(u.Pic),
			"forms":      j.Forms,
			"day":        j.Day,
			"addTime":    j.AddTime,
		}
		// Parse forms JSON (兼容老/新格式)
		formsArr := []map[string]interface{}{}
		if j.Forms != "" {
			var fvs []schema.FieldValue
			if err := json.Unmarshal([]byte(j.Forms), &fvs); err != nil {
				fvs = schema.ExtractFieldValues(j.Forms, enrollModel.Forms)
			}
			for _, fv := range fvs {
				entry := map[string]interface{}{"label": fv.Label, "type": fv.Type, "value": fv.Value}
				// 兼容老逻辑：保留 typeMap 合并
				if typeMap != nil {
					if t, ok := typeMap[fv.Label]; ok && entry["type"] == "" {
						entry["type"] = t
					}
				}
				formsArr = append(formsArr, entry)
			}
		}
		// Merge type from form definitions
		if typeMap != nil {
			for _, f := range formsArr {
				label, _ := f["label"].(string)
				if typ, ok := typeMap[label]; ok {
					f["type"] = typ
				} else {
					// Location subtypes: "打卡位置-地址", "打卡位置-纬度", "打卡位置-经度"
					for defLabel, defType := range typeMap {
						if strings.HasPrefix(label, defLabel+"-") {
							suffix := strings.TrimPrefix(label, defLabel+"-")
							f["type"] = defType
							f["locField"] = suffix
							break
						}
					}
				}
			}
		}
		item["formsArr"] = formsArr

		// Check for images in forms (fields named img/image/pic)
		var images []string
		for _, f := range formsArr {
			label, _ := f["label"].(string)
			val, _ := f["value"].(string)
			if val != "" {
				lower := strings.ToLower(label)
				if strings.Contains(lower, "图") || strings.Contains(lower, "照片") || strings.Contains(lower, "img") || strings.Contains(lower, "pic") || strings.Contains(lower, "image") {
					images = append(images, val)
				}
			}
		}
		if images == nil {
			images = []string{}
		}
		item["images"] = images
		result = append(result, item)
	}
	return result, nil
}

func GetMyDayRecords(userID, day string) ([]map[string]interface{}, error) {
	var joins []model.EnrollJoin
	err := database.DB.Where("`enroll_join_user_id` = ? AND `enroll_join_day` = ?", userID, day).
		Order("`enroll_join_add_time` ASC").Find(&joins).Error
	if err != nil {
		return nil, err
	}

	// Get enroll titles + forms schema (兼容老/新格式)
	enrollCache := map[string]model.Enroll{}

	var result []map[string]interface{}
	for _, j := range joins {
		e, ok := enrollCache[j.EnrollID]
		if !ok {
			if err := database.DB.Where("`id` = ?", j.EnrollID).First(&e).Error; err == nil {
				enrollCache[j.EnrollID] = e
			}
		}

		item := map[string]interface{}{
			"enrollTitle": e.Title,
			"addTime":     j.AddTime,
			"day":         j.Day,
		}
		// Parse forms (兼容老/新格式)
		images, location := schema.ExtractImagesLocation(j.Forms, e.Forms)
		item["images"] = images
		item["location"] = location
		result = append(result, item)
	}
	return result, nil
}

func GetMyJoinRecords(userID string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	var list []model.EnrollJoin
	var total int64
	database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_user_id` = ?", userID).Count(&total)
	err := database.DB.Where("`enroll_join_user_id` = ?", userID).
		Order("`enroll_join_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	// Populate enroll title + form fields (兼容老/新格式)
	for i := range list {
		var enroll model.Enroll
		if err := database.DB.Where("`id` = ?", list[i].EnrollID).First(&enroll).Error; err == nil {
			list[i].EnrollTitle = enroll.Title
		}
		images, location := schema.ExtractImagesLocation(list[i].Forms, enroll.Forms)
		list[i].Images = images
		list[i].Location = location
	}
	return list, total, nil
}

func GetMyCalendarDays(userID, yearMonth string) (map[string][]string, error) {
	var joins []struct {
		EnrollID string `gorm:"column:enroll_join_enroll_id"`
		Day      string `gorm:"column:enroll_join_day"`
	}
	query := database.DB.Model(&model.EnrollJoin{}).
		Where("`enroll_join_user_id` = ?", userID)
	if yearMonth != "" {
		query = query.Where("`enroll_join_day` LIKE ?", yearMonth+"%")
	}
	query.Group("`enroll_join_enroll_id`, `enroll_join_day`").
		Select("`enroll_join_enroll_id`, `enroll_join_day`").Find(&joins)

	result := map[string][]string{}
	for _, j := range joins {
		result[j.EnrollID] = append(result[j.EnrollID], j.Day)
	}
	return result, nil
}

func GetMyEnrollJoinList(userID, enrollID string, page, pageSize int) (interface{}, int64, error) {
	if enrollID != "" {
		var total int64
		query := database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_user_id` = ?", userID).Where("`enroll_join_enroll_id` = ?", enrollID)
		query.Count(&total)
		var list []model.EnrollJoin
		query.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
		return list, total, nil
	}

	enrollIDSet := map[uint]bool{}

	var joins []model.EnrollJoin
	database.DB.Where("`enroll_join_user_id` = ?", userID).Order("`enroll_join_add_time` DESC").Find(&joins)
	for _, j := range joins {
		id, err := strconv.ParseUint(j.EnrollID, 10, 64)
		if err == nil {
			enrollIDSet[uint(id)] = true
		}
	}

	var enrollUsers []model.EnrollUser
	database.DB.Where("`enroll_user_mini_openid` = ?", userID).Find(&enrollUsers)
	for _, eu := range enrollUsers {
		id, err := strconv.ParseUint(eu.EnrollID, 10, 64)
		if err == nil {
			enrollIDSet[uint(id)] = true
		}
	}

	var enrollIDs []string
	for id := range enrollIDSet {
		enrollIDs = append(enrollIDs, strconv.Itoa(int(id)))
	}

	var list []model.Enroll
	if len(enrollIDs) > 0 {
		database.DB.Where("`id` IN ? AND `enroll_status` = 1", enrollIDs).Find(&list)
	}
	list = populateEnrollFields(list)
	for i := range list {
		idStr := strconv.Itoa(int(list[i].ID))
		list[i].IsJoin = true
		_ = idStr
	}
	return list, 0, nil
}
