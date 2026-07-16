package passport

import (
	"context"
	"encoding/json"
	"fmt"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func RegisterUser(userID, mobile, name, pic string, forms interface{}, status int, addIP, device string) (*LoginResponse, error) {
	return RegisterUserContext(context.Background(), userID, mobile, name, pic, forms, status, addIP, device)
}

func RegisterUserContext(ctx context.Context, userID, mobile, name, pic string, forms interface{}, status int, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var cnt int64
	db.Model(&model.User{}).Where("`user_mini_openid` = ?", userID).Count(&cnt)
	if cnt > 0 {
		return LoginUserContext(ctx, userID, addIP, device)
	}
	db.Model(&model.User{}).Where("`user_mobile` = ?", mobile).Count(&cnt)
	if cnt > 0 {
		return nil, fmt.Errorf("该手机已注册")
	}
	formsStr := ""
	if forms != nil {
		b, _ := json.Marshal(forms)
		formsStr = string(b)
	}
	defaultPwd, err := passwordutil.Hash(mobile)
	if err != nil {
		return nil, err
	}
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
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	return LoginUserContext(ctx, userID, addIP, device)
}
