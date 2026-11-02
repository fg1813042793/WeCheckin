package dashboard

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

var dashboardPermissionCodes = map[string]string{
	"users":                 "user:list",
	"enrollments":           "enroll:list",
	"content":               "news:list",
	"events":                "event:list",
	"surveys":               "survey:list",
	"surveyResponses":       "response:list",
	"exams":                 "exam:list",
	"managers":              "mgr:list",
	"workflowTasks":         "workflow:task:list",
	"workflowNotifications": "workflow:notification:list",
	"scheduledRuns":         "scheduled-task:run:list",
	"feedback":              "user-feedback:list",
	"dingtalkBindings":      "dingtalk:bindings:list",
	"notifications":         "notification:list",
}

func dashboardCapabilities(ctx context.Context, db *gorm.DB, admin *model.Admin) (map[string]bool, error) {
	result := make(map[string]bool, len(dashboardPermissionCodes))
	roleIDs := append([]uint(nil), admin.RoleIDs...)
	if len(roleIDs) == 0 {
		var err error
		roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
		if err != nil {
			return nil, err
		}
	}
	if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, db, roleIDs) {
		for key := range dashboardPermissionCodes {
			result[key] = true
		}
		return result, nil
	}
	for key, code := range dashboardPermissionCodes {
		allowed, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(
			ctx,
			db,
			admin.ID,
			roleIDs,
			permissionsupport.AdminAPIPermissionKey(code),
		)
		if err != nil {
			return nil, err
		}
		result[key] = allowed
	}
	return result, nil
}
