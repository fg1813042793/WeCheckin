package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func AdminHome() (map[string]interface{}, error) {
	var userCnt int64
	database.DB.Model(&model.User{}).Count(&userCnt)

	var enrollCnt int64
	database.DB.Model(&model.Enroll{}).Count(&enrollCnt)

	var newsCnt int64
	database.DB.Model(&model.News{}).Count(&newsCnt)

	var joinCnt int64
	database.DB.Model(&model.EnrollJoin{}).Count(&joinCnt)

	result := map[string]interface{}{
		"userCnt":   userCnt,
		"enrollCnt": enrollCnt,
		"newsCnt":   newsCnt,
		"joinCnt":   joinCnt,
	}
	return result, nil
}

func ClearVouchData() error {
	return database.DB.Model(&model.Enroll{}).Where("1 = 1").Update("ENROLL_VOUCH", 0).Error
}

func genRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func InsertLog(logType int, content, adminID, adminName, adminDesc, addIP string) {
	database.DB.Create(&model.Log{
		Type:      logType,
		Content:   content,
		AdminID:   adminID,
		AdminName: adminName,
		AdminDesc: adminDesc,
		AddTime:   database.Now(),
		AddIP:     addIP,
	})
}

func AdminLogin(name, password, addIP string) (map[string]interface{}, error) {
	var admin model.Admin
	err := database.DB.Where("`ADMIN_NAME` = ? AND `ADMIN_PASSWORD` = ?", name, password).First(&admin).Error
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if admin.Status != 1 {
		return nil, fmt.Errorf("账号已禁用")
	}
	token := genRandomString(32)
	admin.Token = token
	admin.TokenTime = database.Now()
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	database.DB.Model(&admin).Updates(map[string]interface{}{
		"ADMIN_TOKEN":      token,
		"ADMIN_TOKEN_TIME": admin.TokenTime,
		"ADMIN_LOGIN_CNT":  admin.LoginCnt,
		"ADMIN_LOGIN_TIME": admin.LoginTime,
	})
	InsertLog(1, "管理员登录", strconv.Itoa(int(admin.ID)), admin.Name, admin.Desc, addIP)
	result := map[string]interface{}{
		"token": token,
		"name":  admin.Name,
		"id":    admin.ID,
		"type":  admin.Type,
	}
	return result, nil
}

func GetMgrList() ([]model.Admin, error) {
	var list []model.Admin
	err := database.DB.Order("`ADMIN_ADD_TIME` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func InsertMgr(name, password, desc, phone, addIP string, typ int) error {
	var cnt int64
	database.DB.Model(&model.Admin{}).Where("`ADMIN_NAME` = ?", name).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("管理员已存在")
	}
	h := md5.Sum([]byte(password))
	admin := model.Admin{
		Name:     name,
		Password: hex.EncodeToString(h[:]),
		Desc:     desc,
		Phone:    phone,
		Status:   1,
		Type:     typ,
		AddTime:  database.Now(),
		AddIP:    addIP,
	}
	return database.DB.Create(&admin).Error
}

func DelMgr(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.Admin{}).Error
}

func GetMgrDetail(id string) (*model.Admin, error) {
	var admin model.Admin
	err := database.DB.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func EditMgr(id, name, desc, phone, addIP string, typ int) error {
	return database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"ADMIN_NAME":  name,
		"ADMIN_DESC":  desc,
		"ADMIN_PHONE": phone,
		"ADMIN_TYPE":  typ,
		"ADMIN_EDIT_TIME":   database.Now(),
		"ADMIN_EDIT_IP":     addIP,
	}).Error
}

func StatusMgr(id string, status int) error {
	return database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("ADMIN_STATUS", status).Error
}

func PwdMgr(id, password string) error {
	h := md5.Sum([]byte(password))
	return database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("ADMIN_PASSWORD", hex.EncodeToString(h[:])).Error
}

func GetLogList() ([]model.Log, error) {
	var list []model.Log
	err := database.DB.Order("`LOG_ADD_TIME` DESC").Limit(100).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func ClearLog() error {
	return database.DB.Where("1 = 1").Delete(&model.Log{}).Error
}

func SetSetup(key, value, typ, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`SETUP_KEY` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			Type:    typ,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Updates(map[string]interface{}{
		"SETUP_VALUE":  value,
		"SETUP_TYPE":   typ,
		"SETUP_EDIT_TIME": database.Now(),
	}).Error
}

func SetContentSetup(key, value, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`SETUP_KEY` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Update("SETUP_VALUE", value).Error
}

func GetUserByOpenID(openID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("`USER_MINI_OPENID` = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserList(keyword string, page, pageSize int) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	query := database.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("`USER_NAME` LIKE ? OR `USER_MOBILE` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("`USER_ADD_TIME` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func DelUser(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.User{}).Error
}

func StatusUser(id string, status int) error {
	return database.DB.Model(&model.User{}).Where("`id` = ?", id).Update("USER_STATUS", status).Error
}

func GetAdminEnrollList(keyword string, page, pageSize int) ([]model.Enroll, int64, error) {
	var list []model.Enroll
	var total int64
	query := database.DB.Model(&model.Enroll{})
	if keyword != "" {
		query = query.Where("`ENROLL_TITLE` LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("`ENROLL_ADD_TIME` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetEnrollDetail(id string) (*model.Enroll, error) {
	var enroll model.Enroll
	err := database.DB.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	return &enroll, nil
}

func UpdateEnrollForms(id, forms string) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("ENROLL_FORMS", forms).Error
}

func GetAdminNewsList(keyword string, page, pageSize int) ([]model.News, int64, error) {
	var list []model.News
	var total int64
	query := database.DB.Model(&model.News{})
	if keyword != "" {
		query = query.Where("`NEWS_TITLE` LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("`NEWS_ADD_TIME` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
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
	return &news, nil
}

func DelNews(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.News{}).Error
}

func GetEnrollJoinList(enrollID string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	var list []model.EnrollJoin
	var total int64
	query := database.DB.Model(&model.EnrollJoin{})
	if enrollID != "" {
		query = query.Where("`ENROLL_JOIN_ENROLL_ID` = ?", enrollID)
	}
	query.Count(&total)
	err := query.Order("`ENROLL_JOIN_ADD_TIME` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetEnrollJoinDataURL(enrollID string) (string, error) {
	return "", nil
}

func DeleteEnrollJoinDataExcel(enrollID string) error {
	return nil
}

func ExportEnrollJoinDataExcel(enrollID string) (string, error) {
	return "", nil
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
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("ENROLL_ORDER", sort).Error
}

func VouchEnroll(id string, vouch int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("ENROLL_VOUCH", vouch).Error
}

func StatusEnroll(id string, status int) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Update("ENROLL_STATUS", status).Error
}

func ClearEnrollAll(id string) error {
	database.DB.Where("`ENROLL_JOIN_ENROLL_ID` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`ENROLL_USER_ENROLL_ID` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"ENROLL_JOIN_CNT": 0,
		"ENROLL_USER_CNT": 0,
	}).Error
}

func DelEnroll(id string) error {
	database.DB.Where("`ENROLL_JOIN_ENROLL_ID` = ?", id).Delete(&model.EnrollJoin{})
	database.DB.Where("`ENROLL_USER_ENROLL_ID` = ?", id).Delete(&model.EnrollUser{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Enroll{}).Error
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
		database.DB.Model(&enroll).UpdateColumn("ENROLL_JOIN_CNT", enroll.JoinCnt-1)
	}
	return nil
}

func SortNews(id, sortStr string) error {
	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("NEWS_ORDER", sort).Error
}

func StatusNews(id string, status int) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("NEWS_STATUS", status).Error
}

func InsertNews(title, desc, cateID, cateName, content, qr, pic, forms, addIP string, status, order int) error {
	news := model.News{
		Title:   title,
		Desc:    desc,
		Status:  status,
		CateID:  cateID,
		CateName: cateName,
		Order:   order,
		Content: content,
		QR:      qr,
		Pic:     pic,
		Forms:   forms,
		AddTime: database.Now(),
		AddIP:   addIP,
	}
	return database.DB.Create(&news).Error
}

func EditNews(id, title, desc, cateID, cateName, content, qr, addIP string, status, order int) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"NEWS_TITLE":     title,
		"NEWS_DESC":      desc,
		"NEWS_STATUS":    status,
		"NEWS_CATE_ID":   cateID,
		"NEWS_CATE_NAME": cateName,
		"NEWS_ORDER":     order,
		"NEWS_CONTENT":   content,
		"NEWS_QR":        qr,
		"NEWS_EDIT_TIME": database.Now(),
		"NEWS_EDIT_IP":   addIP,
	}).Error
}

func UpdateNewsForms(id, forms string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("NEWS_FORMS", forms).Error
}

func UpdateNewsPic(id, pic string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("NEWS_PIC", pic).Error
}

func UpdateNewsContent(id, content string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("NEWS_CONTENT", content).Error
}

func InsertEnroll(title, cateID, cateName, forms, joinForms, qr, addIP string, status, order, dayCnt int, start, end int64) error {
	enroll := model.Enroll{
		Title:     title,
		Status:    status,
		CateID:    cateID,
		CateName:  cateName,
		Start:     start,
		End:       end,
		DayCnt:    dayCnt,
		Order:     order,
		Forms:     forms,
		JoinForms: joinForms,
		QR:        qr,
		AddTime:   database.Now(),
		AddIP:     addIP,
	}
	return database.DB.Create(&enroll).Error
}

func EditEnroll(id, title, cateID, cateName, forms, joinForms, qr, addIP string, status, order, dayCnt int, start, end int64) error {
	return database.DB.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"ENROLL_TITLE":      title,
		"ENROLL_STATUS":     status,
		"ENROLL_CATE_ID":    cateID,
		"ENROLL_CATE_NAME":  cateName,
		"ENROLL_START":      start,
		"ENROLL_END":        end,
		"ENROLL_DAY_CNT":    dayCnt,
		"ENROLL_ORDER":      order,
		"ENROLL_FORMS":      forms,
		"ENROLL_JOIN_FORMS": joinForms,
		"ENROLL_QR":         qr,
		"ENROLL_EDIT_TIME":  database.Now(),
		"ENROLL_EDIT_IP":    addIP,
	}).Error
}
