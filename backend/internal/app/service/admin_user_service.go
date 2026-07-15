package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func GetUserByOpenID(openID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("`user_mini_openid` = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id string) (map[string]interface{}, error) {
	var user model.User
	err := database.DB.Where("`id` = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(id)
	deptIDs := getUserDeptIDs(uint(uid))
	var deptNames []string
	var topDeptNames []string
	for _, did := range deptIDs {
		var d model.Department
		if database.DB.First(&d, did).Error == nil {
			deptNames = append(deptNames, d.Name)
		}
		topDeptNames = append(topDeptNames, getTopDeptName(did))
	}
	m := map[string]interface{}{
		"id": user.ID, "name": user.Name, "mobile": user.Mobile,
		"avatar": GetFullURL(user.Pic), "pic": GetFullURL(user.Pic), "status": user.Status,
		"forms": user.Forms, "loginCnt": user.LoginCnt,
		"addTime": user.AddTime, "loginTime": user.LoginTime,
	}
	m["deptIds"] = deptIDs
	m["deptNames"] = deptNames
	m["topDeptNames"] = topDeptNames
	return m, nil
}

func getUserDeptIDs(userID uint) []uint {
	var depts []model.UserDept
	database.DB.Where("`user_dept_user_id` = ?", userID).Find(&depts)
	ids := make([]uint, len(depts))
	for i, d := range depts {
		ids[i] = d.DeptID
	}
	return ids
}

func saveUserDepts(userID uint, deptIDs []uint) {
	database.DB.Where("`user_dept_user_id` = ?", userID).Delete(&model.UserDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.UserDept{UserID: userID, DeptID: deptID})
		}
	}
}

func GetUserList(keyword, sortStr string, page, pageSize int, adminID uint) ([]map[string]interface{}, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.User
	var total int64
	query := database.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("`user_name` LIKE ? OR `user_mobile` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Data scope
	deptIDs := getDeptVisibleIDs(&admin)
	if deptIDs != nil {
		query = query.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"name":     "`user_name`",
		"mobile":   "`user_mobile`",
		"status":   "`user_status`",
		"loginCnt": "`user_login_cnt`",
		"addTime":  "`user_add_time`",
	})
	if orderClause == "" {
		orderClause = "`user_add_time` DESC"
	}
	err := query.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]map[string]interface{}, len(list))
	for i, u := range list {
		m := map[string]interface{}{
			"id": u.ID, "name": u.Name, "mobile": u.Mobile,
			"avatar": GetFullURL(u.Pic), "pic": GetFullURL(u.Pic), "status": u.Status,
			"loginCnt": u.LoginCnt, "addTime": u.AddTime, "loginTime": u.LoginTime,
		}
		m["deptIds"] = getUserDeptIDs(u.ID)
		result[i] = m
	}
	return result, total, nil
}

func AddUser(name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	now := database.Now()
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", name, now)))
	miniOpenID := hex.EncodeToString(hash[:])
	user := model.User{
		MiniOpenID: miniOpenID,
		Name:       name,
		Mobile:     mobile,
		Pic:        pic,
		Forms:      forms,
		Status:     1,
		AddTime:    now,
		AddIP:      addIP,
		EditTime:   now,
		EditIP:     addIP,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return err
	}
	saveUserDepts(user.ID, deptIDs)
	return nil
}

func EditUser(id, name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	updates := map[string]interface{}{
		"user_name":      name,
		"user_mobile":    mobile,
		"user_edit_time": database.Now(),
		"user_edit_ip":   addIP,
	}
	if pic != "" {
		updates["user_pic"] = pic
	}
	if forms != "" {
		updates["user_forms"] = forms
	}
	if err := database.DB.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	uid, _ := strconv.Atoi(id)
	saveUserDepts(uint(uid), deptIDs)
	return nil
}

func DelUser(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.User{}).Error
}

func DelUsers(ids []string) error {
	return database.DB.Where("`id` IN ?", ids).Delete(&model.User{}).Error
}

func StatusUser(id string, status int, reason string) error {
	updates := map[string]interface{}{
		"user_status": status,
	}
	if status == 1 {
		updates["user_check_reason"] = ""
	} else if reason != "" {
		updates["user_check_reason"] = reason
	}
	return database.DB.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error
}

func ResetUserPassword(id string) error {
	hash, err := passwordutil.Hash("123456")
	if err != nil {
		return err
	}
	return database.DB.Model(&model.User{}).Where("`id` = ?", id).Update("user_password", hash).Error
}

func GetUserFormFields() ([]model.UserFormField, error) {
	var setup model.Setup
	err := database.DB.Where("`setup_key` = ?", "SETUP_USER_FORM_FIELDS").First(&setup).Error
	if err != nil {
		return []model.UserFormField{}, nil
	}
	var list []model.UserFormField
	if setup.Value != "" {
		json.Unmarshal([]byte(setup.Value), &list)
	}
	return list, nil
}

func SaveUserFormFields(fields []model.UserFormField) error {
	jsonData, _ := json.Marshal(fields)
	return SetContentSetup("SETUP_USER_FORM_FIELDS", string(jsonData), "")
}
