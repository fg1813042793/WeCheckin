package admincontent

import (
	"encoding/json"
	"strconv"

	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/app/support/query"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

func decodeEnrollObj(raw string) enrollObj {
	var obj enrollObj
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &obj)
	}
	return obj
}

func populateEnrollFields(list []model.Enroll) []model.Enroll {
	for i := range list {
		obj := decodeEnrollObj(list[i].Obj)
		if len(obj.Cover) > 0 {
			list[i].Img = media.FullURLWithStaticDomain(obj.Cover[0])
		}
		list[i].Desc = obj.Desc
	}
	return list
}

func GetAdminEnrollList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.Enroll, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Enroll
	var total int64
	queryBuilder := database.DB.Model(&model.Enroll{})
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_title` LIKE ?", "%"+keyword+"%")
	}
	where, args := access.DataScopeFilter(&admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	queryBuilder.Count(&total)
	orderClause := query.ParseSort(sortStr, map[string]string{
		"title":   "enroll_title",
		"sort":    "enroll_order",
		"status":  "enroll_status",
		"isVouch": "enroll_vouch",
		"userCnt": "enroll_user_cnt",
		"joinCnt": "enroll_join_cnt",
		"addTime": "enroll_add_time",
	})
	if orderClause == "" {
		orderClause = "`enroll_add_time` DESC"
	}
	err := queryBuilder.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEnrollFields(list)
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
	obj := decodeEnrollObj(enroll.Obj)
	if len(obj.Cover) > 0 {
		enroll.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	enroll.Desc = obj.Desc
	return &enroll, nil
}

func UpdateEnrollForms(id, forms string) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_forms", forms).Error
}

func SortEnroll(id, sortStr string) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_order", sortValue).Error
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
