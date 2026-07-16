package adminmgr

import (
	"context"
	"fmt"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func ChangePassword(id, oldPassword, newPassword string) error {
	return ChangePasswordContext(context.Background(), id, oldPassword, newPassword)
}

func ChangePasswordContext(ctx context.Context, id, oldPassword, newPassword string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	err := db.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return fmt.Errorf("管理员不存在")
	}
	if !passwordutil.Verify(admin.Password, oldPassword) {
		return fmt.Errorf("旧密码错误")
	}
	hash, err := passwordutil.Hash(newPassword)
	if err != nil {
		return err
	}
	return db.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_password", hash).Error
}
