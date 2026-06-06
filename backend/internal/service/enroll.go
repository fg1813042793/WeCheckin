package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/formkit/schema"
	"wecheckin-backend/backend/internal/model"
)

func GetEnrollList(page, pageSize int, userID, keyword string) (map[string]interface{}, error) {
	var list []model.Enroll
	var total int64
	query := database.DB.Model(&model.Enroll{}).Where("`enroll_status` = 1")
	if keyword != "" {
		query = query.Where("`enroll_title` LIKE ? OR `enroll_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Filter by publish departments
	if userID != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL OR "+
				buildDeptOverlap("enroll_publish_dept_ids", deptIDs)+")")
		} else {
			query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)")
		}
	} else {
		query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)")
	}
	query.Count(&total)
	err := query.Order("`enroll_order` ASC, `enroll_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEnrollFields(list)

	// Get user's joined enroll IDs (from both EnrollJoin and EnrollUser)
	joinedIDs := map[string]bool{}
	if userID != "" {
		var joins []model.EnrollJoin
		database.DB.Where("`enroll_join_user_id` = ?", userID).Find(&joins)
		for _, j := range joins {
			joinedIDs[j.EnrollID] = true
		}
		var enrollUsers []model.EnrollUser
		database.DB.Where("`enroll_user_mini_openid` = ?", userID).Find(&enrollUsers)
		for _, eu := range enrollUsers {
			joinedIDs[eu.EnrollID] = true
		}
	}
	for i := range list {
		idStr := strconv.Itoa(int(list[i].ID))
		list[i].IsJoin = joinedIDs[idStr]
	}

	return map[string]interface{}{
		"list":  list,
		"total": total,
	}, nil
}

func ViewEnroll(id, userID string) (*model.Enroll, error) {
	var enroll model.Enroll
	err := database.DB.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	database.DB.Model(&enroll).UpdateColumn("enroll_view_cnt", enroll.ViewCnt+1)

	// Check if current user has joined
	if userID != "" {
		var euCnt int64
		database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", id, userID).Count(&euCnt)
		if euCnt > 0 {
			enroll.IsJoin = true
		}
		var jCnt int64
		database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?", id, userID, time.Now().Format("2006-01-02")).Count(&jCnt)
		if jCnt > 0 {
			enroll.MyEnrollJoinID = "1"
		}
	}

	// Parse OBJ for img/desc/content
	var objMap map[string]interface{}
	if enroll.Obj != "" {
		json.Unmarshal([]byte(enroll.Obj), &objMap)
	}
	if objMap != nil {
		if covers, ok := objMap["cover"].([]interface{}); ok && len(covers) > 0 {
			enroll.Img = GetFullURL(fmt.Sprintf("%v", covers[0]))
		}
		if desc, ok := objMap["desc"].(string); ok {
			enroll.Desc = desc
		}
		if c, ok := objMap["content"].([]interface{}); ok {
			for _, item := range c {
				if m, ok := item.(map[string]interface{}); ok {
					entry := map[string]string{}
					if t, ok := m["type"].(string); ok {
						entry["type"] = t
					}
					if v, ok := m["val"].(string); ok {
						entry["val"] = v
					}
					enroll.Content = append(enroll.Content, entry)
				}
			}
		}
	}

	// Format start/end
	if enroll.Start > 0 {
		enroll.StartStr = time.UnixMilli(enroll.Start).Format("2006-01-02")
	} else {
		enroll.StartStr = "-"
	}
	if enroll.End > 0 {
		enroll.EndStr = time.UnixMilli(enroll.End).Format("2006-01-02")
	} else {
		enroll.EndStr = "-"
	}

	// Status
	now := time.Now().UnixMilli()
	if enroll.Status == 0 {
		enroll.StatusDesc = "已停用"
	} else if enroll.End > 0 && now > enroll.End {
		enroll.StatusDesc = "已结束"
	} else if enroll.Start > 0 && now < enroll.Start {
		enroll.StatusDesc = "未开始"
	} else {
		enroll.StatusDesc = "进行中"
	}

	// DayList from join records
	var days []string
	database.DB.Model(&model.EnrollJoin{}).
		Where("`enroll_join_enroll_id` = ?", id).
		Select("DISTINCT `enroll_join_day`").
		Order("`enroll_join_day` ASC").
		Pluck("enroll_join_day", &days)
	for _, d := range days {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		enroll.DayList = append(enroll.DayList, map[string]string{
			"day":   d,
			"month": fmt.Sprintf("%d月", t.Month()),
			"date":  fmt.Sprintf("%d", t.Day()),
		})
	}

	// RankList from enroll_users
	var enrollUsers []model.EnrollUser
	database.DB.Where("`enroll_user_enroll_id` = ?", id).
		Order("`enroll_user_join_cnt` DESC, `enroll_user_day_cnt` DESC").
		Find(&enrollUsers)

	userMap := map[string]model.User{}
	var allUsers []model.User
	database.DB.Find(&allUsers)
	for _, u := range allUsers {
		userMap[u.MiniOpenID] = u
	}

	for _, eu := range enrollUsers {
		u, ok := userMap[eu.MiniOpenID]
		name := eu.MiniOpenID
		avatar := ""
		if ok {
			name = u.Name
			avatar = GetFullURL(u.Pic)
		}
		enroll.RankList = append(enroll.RankList, map[string]interface{}{
			"userName":   name,
			"userAvatar": avatar,
			"name":       name,
			"avatar":     avatar,
			"joinCount":  eu.JoinCnt,
			"lastDay":    eu.LastDay,
		})
	}

	return &enroll, nil
}

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
			"id":          j.ID,
			"userId":      j.UserID,
			"userName":    u.Name,
			"userAvatar":  GetFullURL(u.Pic),
			"forms":       j.Forms,
			"day":         j.Day,
			"addTime":     j.AddTime,
		}
		// Parse forms JSON (兼容老/新格式)
		formsArr := []map[string]interface{}{}
		if j.Forms != "" {
			fvs := schema.ExtractFieldValues(j.Forms, enrollModel.Forms)
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

func GetEnrollUserRank(enrollID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`enroll_user_enroll_id` = ?", enrollID).
		Order("`enroll_user_join_cnt` DESC, `enroll_user_day_cnt` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetMyEnrollUserList(userID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`enroll_user_mini_openid` = ?", userID).Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	// Populate title, daily limit, today's check-in status, and recalculate dayCnt
	today := time.Now().Format("2006-01-02")
	for i := range list {
		var enroll model.Enroll
		if err := database.DB.Where("`id` = ?", list[i].EnrollID).First(&enroll).Error; err == nil {
			list[i].EnrollTitle = enroll.Title
			list[i].DailyLimit = enroll.DailyLimit
		}
		// Recalculate dayCnt from actual unique check-in days
		var uniqueDays int64
		database.DB.Model(&model.EnrollJoin{}).
			Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", list[i].EnrollID, userID).
			Select("COUNT(DISTINCT `enroll_join_day`)").Scan(&uniqueDays)
		list[i].DayCnt = int(uniqueDays)
		var todayCnt int64
		database.DB.Model(&model.EnrollJoin{}).
			Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?",
				list[i].EnrollID, userID, today).Count(&todayCnt)
		list[i].CheckedInToday = todayCnt > 0
		list[i].TodayJoinCnt = int(todayCnt)
	}
	return list, nil
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

func checkPublishDeptAccess(publishDeptIds string, userDeptIDs []uint) bool {
	if publishDeptIds == "" {
		return true
	}
	ids := strings.Split(publishDeptIds, ",")
	for _, pid := range ids {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		for _, uid := range userDeptIDs {
			if strconv.FormatUint(uint64(uid), 10) == pid {
				return true
			}
		}
	}
	return false
}

func EnrollJoin(enrollID, userID, day, forms, addIP string, status int) error {
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if !checkPublishDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	if !enroll.AllowRepeat {
		var cnt int64
		database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?", enrollID, userID, day).Count(&cnt)
		if cnt > 0 {
			return fmt.Errorf("已打卡")
		}
	}
	join := model.EnrollJoin{
		EnrollID: enrollID,
		UserID:   userID,
		Day:      day,
		Forms:    forms,
		Status:   status,
		AddTime:  database.Now(),
		AddIP:    addIP,
	}
	if err := database.DB.Create(&join).Error; err != nil {
		return err
	}
	database.DB.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt+1)

	var eu model.EnrollUser
	result := database.DB.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).First(&eu)
	if result.Error != nil {
		eu = model.EnrollUser{
			EnrollID:   enrollID,
			MiniOpenID: userID,
			JoinCnt:    1,
			DayCnt:     1,
			LastDay:    day,
			AddTime:    database.Now(),
		}
		database.DB.Create(&eu)
		database.DB.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1)
	} else {
		updates := map[string]interface{}{
			"enroll_user_join_cnt":  eu.JoinCnt + 1,
			"enroll_user_last_day":  day,
			"enroll_user_edit_time": database.Now(),
		}
		// Check if this is a new day
		if eu.LastDay != day {
			updates["enroll_user_day_cnt"] = eu.DayCnt + 1
		}
		database.DB.Model(&eu).Updates(updates)
	}
	return nil
}

func EnrollUserSubmit(enrollID, userID, forms, addIP string) error {
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if !checkPublishDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	var cnt int64
	database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("已参与")
	}
	eu := model.EnrollUser{
		EnrollID:   enrollID,
		MiniOpenID: userID,
		Forms:      forms,
		AddTime:    database.Now(),
		AddIP:      addIP,
	}
	if err := database.DB.Create(&eu).Error; err != nil {
		return err
	}
	database.DB.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1)
	return nil
}

func getJoinStatusDesc(status int) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "已通过"
	case 2:
		return "未通过"
	default:
		return "未知"
	}
}

func getTimeShow(t int64) string {
	return time.UnixMilli(t).Format("2006-01-02 15:04:05")
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

func getUserDeptIDsByMiniOpenID(miniOpenID string) []uint {
	var user model.User
	if err := database.DB.Where("`user_mini_openid` = ?", miniOpenID).First(&user).Error; err != nil {
		return nil
	}
	ids := getUserDeptIDs(user.ID)
	// Include all ancestor departments so users see items published to any parent department
	seen := map[uint]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	for _, id := range ids {
		for _, aid := range getAncestorDeptIDs(id) {
			if !seen[aid] {
				ids = append(ids, aid)
				seen[aid] = true
			}
		}
	}
	return ids
}

func getAncestorDeptIDs(deptID uint) []uint {
	var result []uint
	visited := map[uint]bool{}
	for deptID > 0 {
		if visited[deptID] {
			break
		}
		visited[deptID] = true
		var dept model.Department
		if err := database.DB.First(&dept, deptID).Error; err != nil {
			break
		}
		if dept.ParentID > 0 {
			result = append(result, dept.ParentID)
		}
		deptID = dept.ParentID
	}
	return result
}

func buildDeptOverlap(column string, deptIDs []uint) string {
	if len(deptIDs) == 0 {
		return "1 = 0"
	}
	parts := make([]string, len(deptIDs))
	for i, id := range deptIDs {
		parts[i] = fmt.Sprintf("FIND_IN_SET('%d', `%s`)", id, column)
	}
	return strings.Join(parts, " OR ")
}
