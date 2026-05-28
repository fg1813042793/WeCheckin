package service

import (
	"encoding/json"
	"fmt"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func GetPhone(cloudID string) (string, error) {
	return "", nil
}

func RegisterUser(userID, mobile, name, pic string, forms interface{}, status int) (map[string]interface{}, error) {
	var cnt int64
	database.DB.Model(&model.User{}).Where("`USER_MINI_OPENID` = ?", userID).Count(&cnt)
	if cnt > 0 {
		return LoginUser(userID)
	}
	database.DB.Model(&model.User{}).Where("`USER_MOBILE` = ?", mobile).Count(&cnt)
	if cnt > 0 {
		return nil, fmt.Errorf("该手机已注册")
	}
	formsStr := ""
	if forms != nil {
		b, _ := json.Marshal(forms)
		formsStr = string(b)
	}
	user := model.User{
		MiniOpenID: userID,
		Mobile:     mobile,
		Name:       name,
		Pic:        pic,
		Forms:      formsStr,
		Status:     status,
		AddTime:    database.Now(),
		EditTime:   database.Now(),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return LoginUser(userID)
}

func GetMyDetail(userID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("`USER_MINI_OPENID` = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func EditBase(userID, mobile, name, pic string, forms interface{}) error {
	var cnt int64
	database.DB.Model(&model.User{}).Where("`USER_MOBILE` = ? AND `USER_MINI_OPENID` <> ?", mobile, userID).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("该手机已注册")
	}
	formsStr := ""
	if forms != nil {
		b, _ := json.Marshal(forms)
		formsStr = string(b)
	}
	updates := map[string]interface{}{
		"USER_MOBILE":    mobile,
		"USER_NAME":      name,
		"USER_PIC":       pic,
		"USER_FORMS":     formsStr,
		"USER_EDIT_TIME": database.Now(),
	}
	var user model.User
	database.DB.Where("`USER_MINI_OPENID` = ?", userID).First(&user)
	if user.Status == 8 {
		updates["USER_STATUS"] = 0
	}
	return database.DB.Model(&model.User{}).Where("`USER_MINI_OPENID` = ?", userID).Updates(updates).Error
}

func LoginUser(userID string) (map[string]interface{}, error) {
	var user model.User
	err := database.DB.Where("`USER_MINI_OPENID` = ?", userID).First(&user).Error
	if err != nil {
		return map[string]interface{}{"token": nil}, nil
	}
	database.DB.Model(&user).Update("USER_LOGIN_TIME", database.Now())
	database.DB.Model(&user).UpdateColumn("USER_LOGIN_CNT", user.LoginCnt+1)
	token := map[string]interface{}{
		"id":     user.MiniOpenID,
		"key":    user.ID,
		"name":   user.Name,
		"pic":    user.Pic,
		"status": user.Status,
	}
	return map[string]interface{}{"token": token}, nil
}
