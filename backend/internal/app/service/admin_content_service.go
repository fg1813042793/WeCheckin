package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetAdminEnrollList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.Enroll, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Enroll
	var total int64
	query := database.DB.Model(&model.Enroll{})
	if keyword != "" {
		query = query.Where("`enroll_title` LIKE ?", "%"+keyword+"%")
	}
	// Data scope
	where, args := BuildDataScopeFilter(&admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"title":   "`enroll_title`",
		"sort":    "`enroll_order`",
		"status":  "`enroll_status`",
		"isVouch": "`enroll_vouch`",
		"userCnt": "`enroll_user_cnt`",
		"joinCnt": "`enroll_join_cnt`",
		"addTime": "`enroll_add_time`",
	})
	if orderClause == "" {
		orderClause = "`enroll_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEnrollFields(list)
	// Recalculate user count and join count from actual records
	for i := range list {
		eid := strconv.Itoa(int(list[i].ID))
		var userCnt int64
		database.DB.Raw(
			"SELECT COUNT(DISTINCT uid) FROM (SELECT `enroll_join_user_id` AS uid FROM `enroll_joins` WHERE `enroll_join_enroll_id` = ? UNION SELECT `enroll_user_mini_openid` AS uid FROM `enroll_users` WHERE `enroll_user_enroll_id` = ?) AS u",
			eid, eid,
		).Scan(&userCnt)
		list[i].UserCnt = int(userCnt)
		var joinCnt int64
		database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", eid).Count(&joinCnt)
		list[i].JoinCnt = int(joinCnt)
	}
	return list, total, nil
}

func GetEnrollDetail(id string) (*model.Enroll, error) {
	var enroll model.Enroll
	err := database.DB.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	var obj enrollObj
	if enroll.Obj != "" {
		json.Unmarshal([]byte(enroll.Obj), &obj)
	}
	if len(obj.Cover) > 0 {
		enroll.Img = GetFullURL(obj.Cover[0])
	}
	enroll.Desc = obj.Desc
	return &enroll, nil
}

func UpdateEnrollForms(id, forms string) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_forms", forms).Error
}

func GetAdminNewsList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.News, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.News
	var total int64
	query := database.DB.Model(&model.News{})
	if keyword != "" {
		query = query.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	// Data scope
	where, args := BuildDataScopeFilter(&admin, "`news_dept_id`", "`news_create_by`")
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"title":   "`news_title`",
		"type":    "`news_cate_id`",
		"order":   "`news_order`",
		"status":  "`news_status`",
		"vouch":   "`news_vouch`",
		"addTime": "`news_add_time`",
	})
	if orderClause == "" {
		orderClause = "`news_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetNewsDetail(id string) (*model.News, error) {
	var news model.News
	err := database.DB.Where("`id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	news = populateNewsFields([]model.News{news})[0]
	return &news, nil
}

func DelNews(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.News{}).Error
}

func DelNewsList(ids []string) error {
	return database.DB.Where("`id` IN ?", ids).Delete(&model.News{}).Error
}

func GetEnrollUserList(enrollID, keyword string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	query := database.DB.Where("`enroll_user_enroll_id` = ?", enrollID)
	if keyword != "" {
		query = query.Where("`enroll_user_mini_openid` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)", "%"+keyword+"%")
	}
	err := query.Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].MiniOpenID).First(&u)
		list[i].EnrollTitle = u.Name
		list[i].UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].DeptName = dept.Name
				list[i].TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
	}
	return list, nil
}

func GetEnrollJoinList(enrollID, keyword string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	var list []model.EnrollJoin
	var total int64
	query := database.DB.Model(&model.EnrollJoin{})
	if enrollID != "" {
		query = query.Where("`enroll_join_enroll_id` = ?", enrollID)
	}
	if keyword != "" {
		query = query.Where("`enroll_join_user_id` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?) OR `enroll_join_user_id` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	// Populate user names and dept
	userMap := map[string]model.User{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userMap[u.MiniOpenID] = u
	}
	for i := range list {
		u, ok := userMap[list[i].UserID]
		if ok {
			list[i].EnrollTitle = u.Name
			list[i].UserName = u.Name
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].DeptName = dept.Name
				list[i].TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
	}
	return list, total, nil
}

type EnrollStatItem struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	DeptName    string `json:"deptName"`
	TopDeptName string `json:"topDeptName"`
	JoinCnt     int    `json:"joinCnt"`
	DayCnt      int    `json:"dayCnt"`
}

func GetEnrollStats(enrollID, startDay, endDay string) ([]EnrollStatItem, error) {
	var joins []model.EnrollJoin
	query := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		query = query.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		query = query.Where("`enroll_join_day` <= ?", endDay)
	}
	err := query.Find(&joins).Error
	if err != nil {
		return nil, err
	}
	type tmp struct{ cnt, days int }
	agg := map[string]*tmp{}
	daySet := map[string]map[string]bool{}
	for _, j := range joins {
		if _, ok := agg[j.UserID]; !ok {
			agg[j.UserID] = &tmp{}
			daySet[j.UserID] = map[string]bool{}
		}
		agg[j.UserID].cnt++
		daySet[j.UserID][j.Day] = true
	}
	var result []EnrollStatItem
	for uid, t := range agg {
		t.days = len(daySet[uid])
		item := EnrollStatItem{
			UserID:  uid,
			JoinCnt: t.cnt,
			DayCnt:  t.days,
		}
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", uid).First(&u)
		item.UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				item.DeptName = dept.Name
				item.TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func GetEnrollJoinDataURL(enrollID string) (string, error) {
	return "", nil
}

func DeleteEnrollJoinDataExcel(enrollID string) error {
	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	os.Remove(filepath.Join("./uploads", filename))
	return nil
}

func ExportEnrollJoinDataExcel(enrollID, startDay, endDay string) (string, error) {
	var joins []model.EnrollJoin
	query := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		query = query.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		query = query.Where("`enroll_join_day` <= ?", endDay)
	}
	query.Order("`enroll_join_add_time` DESC").Find(&joins)

	// Get enroll title
	var enroll model.Enroll
	database.DB.Where("`id` = ?", enrollID).First(&enroll)

	// Get user names
	userNames := map[string]string{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userNames[u.MiniOpenID] = u.Name
	}

	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	filepath := filepath.Join("./uploads", filename)

	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// BOM for Excel UTF-8 compatibility
	f.WriteString("\xEF\xBB\xBF")

	writer.Write([]string{"打卡项目", enroll.Title})
	writer.Write([]string{"用户ID", "用户姓名", "打卡日期", "打卡内容", "打卡时间", "IP地址"})
	for _, j := range joins {
		joinTime := time.UnixMilli(j.AddTime).Format("2006-01-02 15:04:05")
		writer.Write([]string{
			j.UserID,
			userNames[j.UserID],
			j.Day,
			j.Forms,
			joinTime,
			j.AddIP,
		})
	}

	return filename, nil
}

func GetUserDataURL() (string, error) {
	return "", nil
}

func DeleteUserDataExcel() error {
	return nil
}

func ExportUserDataExcel() (string, error) {
	return "", nil
}

func SortEnroll(id, sortStr string) error {
	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_order", sort).Error
}

func VouchEvent(id string, vouch int) error {
	return database.DB.Model(&model.Event{}).Where("`id` = ?", id).Update("event_vouch", vouch).Error
}

func TopEvent(id string, top int) error {
	return database.DB.Model(&model.Event{}).Where("`id` = ?", id).Update("event_is_top", top).Error
}

func VouchEnroll(id string, vouch int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_vouch", vouch).Error
}

func StatusEnroll(id string, status int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_status", status).Error
}

func ClearEnrollAll(id string) error {
	database.DB.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"enroll_join_cnt": 0,
		"enroll_user_cnt": 0,
	}).Error
}

func DelEnroll(id string) error {
	database.DB.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Enroll{}).Error
}

func DelEnrolls(ids []string) error {
	for _, id := range ids {
		if err := DelEnroll(id); err != nil {
			return err
		}
	}
	return nil
}

func DelEnrollJoin(id string) error {
	var join model.EnrollJoin
	err := database.DB.Where("`id` = ?", id).First(&join).Error
	if err != nil {
		return err
	}
	database.DB.Delete(&join)
	var enroll model.Enroll
	database.DB.Where("`id` = ?", join.EnrollID).First(&enroll)
	if enroll.JoinCnt > 0 {
		database.DB.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt-1)
	}
	return nil
}

func DelEnrollJoins(ids []string) error {
	for _, id := range ids {
		if err := DelEnrollJoin(id); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEnrollUser(enrollID, userID string) error {
	database.DB.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Delete(&model.EnrollUser{})
	database.DB.Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", enrollID, userID).Delete(&model.EnrollJoin{})
	var cnt int64
	database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ?", enrollID).Count(&cnt)
	database.DB.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_user_cnt", cnt)
	return nil
}

func RemoveEnrollUsers(enrollID string, userIDs []string) error {
	for _, uid := range userIDs {
		if err := RemoveEnrollUser(enrollID, uid); err != nil {
			return err
		}
	}
	return nil
}

func EditEnrollUserForms(enrollID, userID, forms string) error {
	return database.DB.Model(&model.EnrollUser{}).
		Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).
		Update("enroll_user_forms", forms).Error
}

func SortNews(id, sortStr string) error {
	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_order", sort).Error
}

func StatusNews(id string, status int) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_status", status).Error
}

func InsertNews(title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds string, status, order int, deptID, createBy uint) error {
	news := model.News{
		Title:          title,
		Desc:           desc,
		Status:         status,
		CateID:         cateID,
		CateName:       cateName,
		Order:          order,
		Content:        content,
		QR:             qr,
		Pic:            pic,
		Forms:          forms,
		DeptID:         deptID,
		PublishDeptIds: publishDeptIds,
		CreateBy:       createBy,
		AddTime:        database.Now(),
		AddIP:          addIP,
	}
	return database.DB.Create(&news).Error
}

func EditNews(id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID uint) error {
	updates := map[string]interface{}{
		"news_title":            title,
		"news_desc":             desc,
		"news_status":           status,
		"news_cate_id":          cateID,
		"news_cate_name":        cateName,
		"news_order":            order,
		"news_content":          content,
		"news_qr":               qr,
		"news_dept_id":          deptID,
		"news_publish_dept_ids": publishDeptIds,
		"news_edit_time":        database.Now(),
		"news_edit_ip":          addIP,
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Updates(updates).Error
}

func UpdateNewsForms(id, forms string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_forms", forms).Error
}

func UpdateNewsPic(id, pic string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_pic", pic).Error
}

func UpdateNewsContent(id, content string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_content", content).Error
}

func InsertEnroll(title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID, createBy uint) error {
	enroll := model.Enroll{
		Title:          title,
		Status:         status,
		CateID:         cateID,
		CateName:       cateName,
		Start:          start,
		End:            end,
		DayCnt:         dayCnt,
		Order:          order,
		Forms:          forms,
		JoinForms:      joinForms,
		QR:             qr,
		Obj:            obj,
		AllowRepeat:    allowRepeat,
		DailyLimit:     dailyLimit,
		DeptID:         deptID,
		PublishDeptIds: publishDeptIds,
		CreateBy:       createBy,
		AddTime:        database.Now(),
		AddIP:          addIP,
	}
	return database.DB.Create(&enroll).Error
}

func EditEnroll(id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID uint) error {
	updates := map[string]interface{}{
		"enroll_title":            title,
		"enroll_status":           status,
		"enroll_cate_id":          cateID,
		"enroll_cate_name":        cateName,
		"enroll_start":            start,
		"enroll_end":              end,
		"enroll_day_cnt":          dayCnt,
		"enroll_order":            order,
		"enroll_forms":            forms,
		"enroll_join_forms":       joinForms,
		"enroll_repeat":           allowRepeat,
		"enroll_limit":            dailyLimit,
		"enroll_dept_id":          deptID,
		"enroll_publish_dept_ids": publishDeptIds,
		"enroll_qr":               qr,
		"enroll_edit_time":        database.Now(),
		"enroll_edit_ip":          addIP,
	}
	if obj != "" {
		updates["enroll_obj"] = obj
	}
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(updates).Error
}
