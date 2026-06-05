package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/jwtutil"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
)

func GetPhone(cloudID string) (string, error) {
	return "", nil
}

func RegisterUser(userID, mobile, name, pic string, forms interface{}, status int) (map[string]interface{}, error) {
	var cnt int64
	database.DB.Model(&model.User{}).Where("`user_mini_openid` = ?", userID).Count(&cnt)
	if cnt > 0 {
		return LoginUser(userID)
	}
	database.DB.Model(&model.User{}).Where("`user_mobile` = ?", mobile).Count(&cnt)
	if cnt > 0 {
		return nil, fmt.Errorf("该手机已注册")
	}
	formsStr := ""
	if forms != nil {
		b, _ := json.Marshal(forms)
		formsStr = string(b)
	}
	// Generate a random default password: md5(mobile)
	h := md5.Sum([]byte(mobile))
	defaultPwd := hex.EncodeToString(h[:])
	user := model.User{
		MiniOpenID: userID,
		Mobile:     mobile,
		Name:       name,
		Pic:        pic,
		Forms:      formsStr,
		Status:     status,
		Password:   defaultPwd,
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
	err := database.DB.Where("`user_mini_openid` = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	setUserRole(&user)
	// populate dept info
	var ud model.UserDept
	database.DB.Where("`user_dept_user_id` = ?", user.ID).First(&ud)
	if ud.DeptID > 0 {
		var dept model.Department
		if err := database.DB.First(&dept, ud.DeptID).Error; err == nil {
			user.DeptName = dept.Name
			user.TopDeptName = getTopDeptName(ud.DeptID)
		}
	}
	return &user, nil
}

func setUserRole(u *model.User) {
	if u.Status == 9 || u.Status == 0 {
		u.Role = "admin"
	} else {
		u.Role = "user"
	}
}

func EditBase(userID, mobile, name, pic string, forms interface{}) error {
	var cnt int64
	database.DB.Model(&model.User{}).Where("`user_mobile` = ? AND `user_mini_openid` <> ?", mobile, userID).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("该手机已注册")
	}
	formsStr := ""
	if forms != nil {
		b, _ := json.Marshal(forms)
		formsStr = string(b)
	}
	updates := map[string]interface{}{
		"user_mobile":    mobile,
		"user_name":      name,
		"user_pic":       pic,
		"user_forms":     formsStr,
		"user_edit_time": database.Now(),
	}
	var user model.User
	database.DB.Where("`user_mini_openid` = ?", userID).First(&user)
	if user.Status == 8 {
		updates["user_status"] = 0
	}
	return database.DB.Model(&model.User{}).Where("`user_mini_openid` = ?", userID).Updates(updates).Error
}

func LoginUser(userID string) (map[string]interface{}, error) {
	var user model.User
	err := database.DB.Where("`user_mini_openid` = ?", userID).First(&user).Error
	if err != nil {
		return map[string]interface{}{"token": nil}, nil
	}
	database.DB.Model(&user).Update("user_login_time", database.Now())
	database.DB.Model(&user).UpdateColumn("user_login_cnt", user.LoginCnt+1)
	setUserRole(&user)
	token, err := jwtutil.GenerateToken(user.MiniOpenID, user.Role)
	if err != nil {
		return nil, err
	}
	storeUserToken(user.ID, token)
	return map[string]interface{}{
		"token": token,
		"userInfo": map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     GetFullURL(user.Pic),
			"desc":       "点击完善个人信息",
			"miniOpenID": user.MiniOpenID,
			"role":       user.Role,
		},
	}, nil
}

func LoginByPwd(name, password string) (map[string]interface{}, error) {
	h := md5.Sum([]byte(password))
	passwordMD5 := hex.EncodeToString(h[:])
	var user model.User
	err := database.DB.Where("(`user_name` = ? OR `user_mobile` = ?) AND `user_password` = ?", name, name, passwordMD5).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	database.DB.Model(&user).Update("user_login_time", database.Now())
	database.DB.Model(&user).UpdateColumn("user_login_cnt", user.LoginCnt+1)
	setUserRole(&user)
	token, err := jwtutil.GenerateToken(user.MiniOpenID, user.Role)
	if err != nil {
		return nil, err
	}
	storeUserToken(user.ID, token)
	return map[string]interface{}{
		"token": token,
		"userInfo": map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     GetFullURL(user.Pic),
			"desc":       "点击完善个人信息",
			"miniOpenID": user.MiniOpenID,
			"role":       user.Role,
		},
	}, nil
}

func storeUserToken(userID uint, token string) {
	expire, prefix := tokenutil.GetTokenConfig("user")
	if prefix == "" {
		prefix = "user_token:"
	}
	if rd.RDB != nil {
		rd.RDB.Set(rd.Ctx, prefix+strconv.Itoa(int(userID)), token, expire)
	}
}
