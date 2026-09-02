package httpadmin

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

type GormRiskAuthorizer struct {
	db *gorm.DB
}

func NewGormRiskAuthorizer(db *gorm.DB) *GormRiskAuthorizer {
	return &GormRiskAuthorizer{db: db}
}

func (authorizer *GormRiskAuthorizer) HasPermission(ctx context.Context, admin *model.Admin, code string) (bool, error) {
	if authorizer == nil || authorizer.db == nil {
		return false, errors.New("scheduled task permission database is not initialized")
	}
	if admin == nil || admin.ID == 0 || strings.TrimSpace(code) == "" {
		return false, nil
	}
	roleIDs := admin.RoleIDs
	if len(roleIDs) == 0 {
		var err error
		roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(ctx, authorizer.db, admin.ID, admin.RoleID)
		if err != nil {
			return false, err
		}
	}
	if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, authorizer.db, roleIDs) {
		return true, nil
	}
	key := permissionsupport.AdminAPIPermissionKey(code)
	if effect, ok, err := permissionsupport.SubjectPermissionEffectContext(ctx, authorizer.db, permissionsupport.SubjectUser, admin.ID, key); err != nil {
		return false, err
	} else if ok && effect == permissionsupport.EffectDeny {
		return false, nil
	}
	return permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, authorizer.db, admin.ID, roleIDs, key)
}

var _ RiskAuthorizer = (*GormRiskAuthorizer)(nil)
