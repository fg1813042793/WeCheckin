package passport

import (
	"context"
	"encoding/json"
	"fmt"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/pkg/database"

	"gorm.io/gorm"
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
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return accountsecurity.SetInitialPasswordWithDBContext(ctx, tx, user.ID, mobile)
	}); err != nil {
		return nil, err
	}
	return LoginUserContext(ctx, userID, addIP, device)
}
