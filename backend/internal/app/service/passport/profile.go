package passport

import (
	"context"
	"encoding/json"
	"fmt"

	"wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type DetailResponse struct {
	User   model.User `json:"user"`
	Domain string     `json:"domain"`
}

func GetMyDetail(userID string) (*DetailResponse, error) {
	return GetMyDetailContext(context.Background(), userID)
}

func GetMyDetailContext(ctx context.Context, userID string) (*DetailResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	err := db.Where("`user_mini_openid` = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	setUserRole(&user)
	var ud model.UserDept
	db.Where("`user_dept_user_id` = ?", user.ID).First(&ud)
	if ud.DeptID > 0 {
		var d model.Department
		if err := db.First(&d, ud.DeptID).Error; err == nil {
			user.DeptName = d.Name
			user.TopDeptName = dept.TopDeptName(ud.DeptID)
		}
	}
	return &DetailResponse{User: user, Domain: media.StaticDomain()}, nil
}

func setUserRole(u *model.User) {
	if u.Status == 9 || u.Status == 0 {
		u.Role = "admin"
	} else {
		u.Role = "user"
	}
}

func EditBase(userID, mobile, name, pic string, forms interface{}) error {
	return EditBaseContext(context.Background(), userID, mobile, name, pic, forms)
}

func EditBaseContext(ctx context.Context, userID, mobile, name, pic string, forms interface{}) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var cnt int64
	db.Model(&model.User{}).Where("`user_mobile` = ? AND `user_mini_openid` <> ?", mobile, userID).Count(&cnt)
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
	db.Where("`user_mini_openid` = ?", userID).First(&user)
	if user.Status == 8 {
		updates["user_status"] = 0
	}
	return db.Model(&model.User{}).Where("`user_mini_openid` = ?", userID).Updates(updates).Error
}
