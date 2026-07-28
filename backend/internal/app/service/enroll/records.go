package enroll

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"wecheckin-backend/backend/internal/app/formkit/schema"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"

	"gorm.io/gorm"
)

type RecordFormField struct {
	Label    string      `json:"label"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
	LocField string      `json:"locField,omitempty"`
}

type EnrollJoinDayRecord struct {
	ID         uint              `json:"id"`
	UserID     string            `json:"userId"`
	UserName   string            `json:"userName"`
	UserAvatar string            `json:"userAvatar"`
	Forms      string            `json:"forms"`
	FormsArr   []RecordFormField `json:"formsArr"`
	Day        string            `json:"day"`
	AddTime    int64             `json:"addTime"`
	Images     []string          `json:"images"`
}

type MyDayRecord struct {
	EnrollTitle string   `json:"enrollTitle"`
	AddTime     int64    `json:"addTime"`
	Day         string   `json:"day"`
	Images      []string `json:"images"`
	Location    string   `json:"location"`
}

type MyEnrollJoinListResult struct {
	JoinRecords []model.EnrollJoin
	Enrolls     []model.Enroll
	Total       int64
}

func GetEnrollJoinByDay(enrollID, day string) ([]EnrollJoinDayRecord, error) {
	return GetEnrollJoinByDayContext(context.Background(), enrollID, day)
}

func GetEnrollJoinByDayContext(ctx context.Context, enrollID, day string) ([]EnrollJoinDayRecord, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var joins []model.EnrollJoin
	query := db.Where("`enroll_join_enroll_id` = ? AND `enroll_join_day` = ?", enrollID, day)
	if err := query.Order("`enroll_join_add_time` DESC").Find(&joins).Error; err != nil {
		return nil, err
	}

	// Get enroll form definitions for type mapping
	var enrollModel model.Enroll
	var typeMap map[string]string // label -> type
	if len(joins) > 0 {
		_ = db.Where("`id` = ?", enrollID).First(&enrollModel).Error
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
	userIDs := make([]string, 0, len(joins))
	for _, join := range joins {
		userIDs = append(userIDs, join.UserID)
	}
	userMap, err := loadUsersByOpenID(db, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]EnrollJoinDayRecord, 0, len(joins))
	for _, j := range joins {
		u, _ := userMap[j.UserID]
		item := EnrollJoinDayRecord{
			ID:         j.ID,
			UserID:     j.UserID,
			UserName:   u.Name,
			UserAvatar: media.FullURLWithStaticDomain(u.Pic),
			Forms:      j.Forms,
			Day:        j.Day,
			AddTime:    j.AddTime,
		}
		// Parse forms JSON (兼容老/新格式)
		formsArr := []RecordFormField{}
		if j.Forms != "" {
			var fvs []schema.FieldValue
			if err := json.Unmarshal([]byte(j.Forms), &fvs); err != nil {
				fvs = schema.ExtractFieldValues(j.Forms, enrollModel.Forms)
			}
			for _, fv := range fvs {
				entry := RecordFormField{Label: fv.Label, Type: fv.Type, Value: fv.Value}
				// 兼容老逻辑：保留 typeMap 合并
				if typeMap != nil {
					if t, ok := typeMap[fv.Label]; ok && entry.Type == "" {
						entry.Type = t
					}
				}
				formsArr = append(formsArr, entry)
			}
		}
		// Merge type from form definitions
		if typeMap != nil {
			for idx := range formsArr {
				label := formsArr[idx].Label
				if typ, ok := typeMap[label]; ok {
					formsArr[idx].Type = typ
				} else {
					// Location subtypes: "打卡位置-地址", "打卡位置-纬度", "打卡位置-经度"
					for defLabel, defType := range typeMap {
						if strings.HasPrefix(label, defLabel+"-") {
							suffix := strings.TrimPrefix(label, defLabel+"-")
							formsArr[idx].Type = defType
							formsArr[idx].LocField = suffix
							break
						}
					}
				}
			}
		}
		item.FormsArr = formsArr

		// Check for images in forms (fields named img/image/pic)
		var images []string
		for _, f := range formsArr {
			label := f.Label
			val, _ := f.Value.(string)
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
		item.Images = images
		result = append(result, item)
	}
	return result, nil
}

func GetMyDayRecords(userID, day string) ([]MyDayRecord, error) {
	return GetMyDayRecordsContext(context.Background(), userID, day)
}

func GetMyDayRecordsContext(ctx context.Context, userID, day string) ([]MyDayRecord, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var joins []model.EnrollJoin
	err := db.Where("`enroll_join_user_id` = ? AND `enroll_join_day` = ?", userID, day).
		Order("`enroll_join_add_time` ASC").Find(&joins).Error
	if err != nil {
		return nil, err
	}

	enrollMap, err := loadEnrollsByStringID(db, collectJoinEnrollIDs(joins))
	if err != nil {
		return nil, err
	}

	result := make([]MyDayRecord, 0, len(joins))
	for _, j := range joins {
		e := enrollMap[j.EnrollID]

		// Parse forms (兼容老/新格式)
		images, location := schema.ExtractImagesLocation(j.Forms, e.Forms)
		item := MyDayRecord{
			EnrollTitle: e.Title,
			AddTime:     j.AddTime,
			Day:         j.Day,
			Images:      images,
			Location:    location,
		}
		result = append(result, item)
	}
	return result, nil
}

func GetMyJoinRecords(userID string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	return GetMyJoinRecordsContext(context.Background(), userID, page, pageSize)
}

func GetMyJoinRecordsContext(ctx context.Context, userID string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EnrollJoin
	var total int64
	if err := db.Model(&model.EnrollJoin{}).Where("`enroll_join_user_id` = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Where("`enroll_join_user_id` = ?", userID).
		Order("`enroll_join_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	enrollMap, err := loadEnrollsByStringID(db, collectJoinEnrollIDs(list))
	if err != nil {
		return nil, 0, err
	}
	for i := range list {
		enroll := enrollMap[list[i].EnrollID]
		list[i].EnrollTitle = enroll.Title
		images, location := schema.ExtractImagesLocation(list[i].Forms, enroll.Forms)
		list[i].Images = images
		list[i].Location = location
	}
	return list, total, nil
}

func GetMyCalendarDays(userID, yearMonth string) (map[string][]string, error) {
	return GetMyCalendarDaysContext(context.Background(), userID, yearMonth)
}

func GetMyCalendarDaysContext(ctx context.Context, userID, yearMonth string) (map[string][]string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var joins []struct {
		EnrollID string `gorm:"column:enroll_join_enroll_id"`
		Day      string `gorm:"column:enroll_join_day"`
	}
	query := db.Model(&model.EnrollJoin{}).
		Where("`enroll_join_user_id` = ?", userID)
	if yearMonth != "" {
		query = query.Where("`enroll_join_day` LIKE ?", yearMonth+"%")
	}
	if err := query.Group("`enroll_join_enroll_id`, `enroll_join_day`").
		Select("`enroll_join_enroll_id`, `enroll_join_day`").Find(&joins).Error; err != nil {
		return nil, err
	}

	result := map[string][]string{}
	for _, j := range joins {
		result[j.EnrollID] = append(result[j.EnrollID], j.Day)
	}
	return result, nil
}

func GetMyEnrollJoinList(userID, enrollID string, page, pageSize int) (MyEnrollJoinListResult, error) {
	return GetMyEnrollJoinListContext(context.Background(), userID, enrollID, page, pageSize)
}

func GetMyEnrollJoinListContext(ctx context.Context, userID, enrollID string, page, pageSize int) (MyEnrollJoinListResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if enrollID != "" {
		var total int64
		query := db.Model(&model.EnrollJoin{}).Where("`enroll_join_user_id` = ?", userID).Where("`enroll_join_enroll_id` = ?", enrollID)
		if err := query.Count(&total).Error; err != nil {
			return MyEnrollJoinListResult{}, err
		}
		var list []model.EnrollJoin
		if err := query.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
			return MyEnrollJoinListResult{}, err
		}
		return MyEnrollJoinListResult{JoinRecords: list, Total: total}, nil
	}

	enrollIDs, total, err := loadMyEnrollIDsPage(db, userID, page, pageSize)
	if err != nil {
		return MyEnrollJoinListResult{}, err
	}
	enrollMap, err := loadEnrollsByStringID(db, enrollIDs)
	if err != nil {
		return MyEnrollJoinListResult{}, err
	}
	list := make([]model.Enroll, 0, len(enrollIDs))
	for _, id := range enrollIDs {
		if enroll, ok := enrollMap[id]; ok {
			list = append(list, enroll)
		}
	}
	list = populateEnrollFields(list)
	for i := range list {
		idStr := strconv.Itoa(int(list[i].ID))
		list[i].IsJoin = true
		_ = idStr
	}
	return MyEnrollJoinListResult{Enrolls: list, Total: total}, nil
}

const myEnrollIDsBaseSQL = `
SELECT merged.enroll_id, MAX(merged.last_time) AS last_time
FROM (
  SELECT enroll_join_enroll_id AS enroll_id, MAX(enroll_join_add_time) AS last_time
  FROM enroll_joins
  WHERE enroll_join_user_id = ?
  GROUP BY enroll_join_enroll_id
  UNION ALL
  SELECT enroll_user_enroll_id AS enroll_id, MAX(enroll_user_add_time) AS last_time
  FROM enroll_users
  WHERE enroll_user_mini_openid = ?
  GROUP BY enroll_user_enroll_id
) AS merged
JOIN enrolls ON enrolls.id = merged.enroll_id AND enrolls.enroll_status = 1
GROUP BY merged.enroll_id`

func loadMyEnrollIDsPage(db *gorm.DB, userID string, page, pageSize int) ([]string, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var countRow struct {
		Total int64 `gorm:"column:total"`
	}
	if err := db.Raw("SELECT COUNT(*) AS total FROM ("+myEnrollIDsBaseSQL+") AS counted", userID, userID).
		Scan(&countRow).Error; err != nil {
		return nil, 0, err
	}

	var rows []struct {
		EnrollID string `gorm:"column:enroll_id"`
	}
	offset := (page - 1) * pageSize
	if err := db.Raw("SELECT page_ids.enroll_id FROM ("+myEnrollIDsBaseSQL+") AS page_ids ORDER BY page_ids.last_time DESC LIMIT ? OFFSET ?", userID, userID, pageSize, offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	enrollIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		enrollIDs = append(enrollIDs, row.EnrollID)
	}
	return enrollIDs, countRow.Total, nil
}

func collectJoinEnrollIDs(joins []model.EnrollJoin) []string {
	enrollIDs := make([]string, 0, len(joins))
	for _, join := range joins {
		enrollIDs = append(enrollIDs, join.EnrollID)
	}
	return enrollIDs
}

func loadUsersByOpenID(db *gorm.DB, openIDs []string) (map[string]model.User, error) {
	result := map[string]model.User{}
	ids := uniqueNonEmptyStrings(openIDs)
	if len(ids) == 0 {
		return result, nil
	}
	var users []model.User
	if err := db.Where("`user_mini_openid` IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	for _, user := range users {
		result[user.MiniOpenID] = user
	}
	return result, nil
}

func loadEnrollsByStringID(db *gorm.DB, enrollIDs []string) (map[string]model.Enroll, error) {
	result := map[string]model.Enroll{}
	ids := uniqueNonEmptyStrings(enrollIDs)
	if len(ids) == 0 {
		return result, nil
	}
	var enrolls []model.Enroll
	if err := db.Where("`id` IN ?", ids).Find(&enrolls).Error; err != nil {
		return nil, err
	}
	for _, enroll := range enrolls {
		result[strconv.Itoa(int(enroll.ID))] = enroll
	}
	return result, nil
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
