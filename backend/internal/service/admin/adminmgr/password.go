package adminmgr

import (
	"context"
	"fmt"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/pkg/database"
)

func ChangePassword(id, oldPassword, newPassword string) error {
	return ChangePasswordContext(context.Background(), id, oldPassword, newPassword)
}

func ChangePasswordContext(ctx context.Context, id, oldPassword, newPassword string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	err := db.Where("`id` = ?", id).Scopes(adminLoginRoleFilter).First(&admin).Error
	if err != nil {
		return fmt.Errorf("管理员不存在")
	}
	return accountsecurity.ChangeOwnPasswordContext(ctx, admin.ID, oldPassword, newPassword)
}

func ResetPasswordContext(ctx context.Context, id, newPassword string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	if err := db.Where("`id` = ?", id).Scopes(adminLoginRoleFilter).First(&admin).Error; err != nil {
		return fmt.Errorf("管理员不存在")
	}
	return accountsecurity.ResetTemporaryPasswordContext(ctx, admin.ID, newPassword)
}
