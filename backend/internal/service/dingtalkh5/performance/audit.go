package performance

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/pkg/database"
)

type dingtalkH5AuditMeta struct {
	UserID       uint
	DeptID       uint
	CurrentMilli int64
}

func dingtalkH5AuditMetaForUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, now int64) dingtalkH5AuditMeta {
	if now == 0 {
		now = database.Now()
	}
	meta := dingtalkH5AuditMeta{CurrentMilli: now}
	if user == nil {
		return meta
	}
	meta.UserID = user.ID
	if db == nil || user.ID == 0 {
		return meta
	}
	deptID := dingtalkH5FirstUserDeptIDContext(ctx, db, user.ID)
	if deptID == 0 {
		return meta
	}
	meta.DeptID = deptID
	return meta
}

func dingtalkH5DBContext(db *gorm.DB) context.Context {
	if db != nil && db.Statement != nil && db.Statement.Context != nil {
		return db.Statement.Context
	}
	return context.Background()
}

func dingtalkH5FirstUserDeptIDContext(ctx context.Context, db *gorm.DB, userID uint) uint {
	if db == nil || userID == 0 {
		return 0
	}
	var row model.UserDept
	err := db.WithContext(ctx).
		Select("`user_dept_dept_id`").
		Where("`user_dept_user_id` = ?", userID).
		Order("`id` ASC").
		First(&row).Error
	if err != nil {
		return 0
	}
	return row.DeptID
}

func applyDingTalkH5CreateAudit(fields *model.DingTalkH5AuditFields, meta dingtalkH5AuditMeta) {
	if fields == nil {
		return
	}
	fields.CreateBy = meta.UserID
	fields.UpdateBy = meta.UserID
	fields.CreateDeptID = meta.DeptID
	fields.UpdateDeptID = meta.DeptID
}

func applyDingTalkH5UpdateAudit(fields *model.DingTalkH5AuditFields, meta dingtalkH5AuditMeta) {
	if fields == nil {
		return
	}
	fields.UpdateBy = meta.UserID
	fields.UpdateDeptID = meta.DeptID
	if fields.CreateBy == 0 {
		fields.CreateBy = meta.UserID
		fields.CreateDeptID = meta.DeptID
	}
}

func applyDingTalkH5DeleteAudit(fields *model.DingTalkH5AuditFields, meta dingtalkH5AuditMeta) {
	if fields == nil {
		return
	}
	applyDingTalkH5UpdateAudit(fields, meta)
	fields.DeleteBy = meta.UserID
	fields.DeleteDeptID = meta.DeptID
	fields.DeletedAt = meta.CurrentMilli
}

func dingtalkH5AuditUpdateValues(meta dingtalkH5AuditMeta) map[string]interface{} {
	fields := access.DingTalkH5AuditFields
	return map[string]interface{}{
		fields.UpdateByColumn:   meta.UserID,
		fields.UpdateDeptColumn: meta.DeptID,
		fields.UpdateTimeColumn: meta.CurrentMilli,
	}
}

func dingtalkH5CreateAuditValues(meta dingtalkH5AuditMeta) map[string]interface{} {
	fields := access.DingTalkH5AuditFields
	values := dingtalkH5AuditUpdateValues(meta)
	values[fields.CreateByColumn] = meta.UserID
	values[fields.CreateDeptColumn] = meta.DeptID
	return values
}

func dingtalkH5DeleteAuditUpdateValues(meta dingtalkH5AuditMeta) map[string]interface{} {
	values := dingtalkH5AuditUpdateValues(meta)
	values["delete_by"] = meta.UserID
	values["delete_dept_id"] = meta.DeptID
	values["deleted_at"] = meta.CurrentMilli
	return values
}
