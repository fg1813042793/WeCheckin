package formkitadmin

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func FirstAdminDeptID(adminID uint) (uint, error) {
	return FirstAdminDeptIDContext(context.Background(), adminID)
}

func FirstAdminDeptIDContext(ctx context.Context, adminID uint) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return firstAdminDeptID(db, adminID)
}

func firstAdminDeptID(db *gorm.DB, adminID uint) (uint, error) {
	if adminID == 0 {
		return 0, nil
	}
	var userDept model.UserDept
	err := db.Where("`user_dept_user_id` = ?", adminID).First(&userDept).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return userDept.DeptID, nil
}
