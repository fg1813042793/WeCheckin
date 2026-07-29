package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/app/support/appapiperm"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func ClientPerm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		user, ok := currentClientUser(c)
		if !ok || user.ID == 0 {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "未登录"})
			c.Abort()
			return
		}
		path := requestPath(c)
		required, ok := clientRoutePermission(string(c.Method()), path)
		if !ok {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}

		db, cancel := database.WithContext(ctx)
		defer cancel()
		if !permissionsupport.TablesReady(db) {
			c.Next(ctx)
			return
		}
		roleID := ensureClientRoleID(ctx, db, user)
		if effect, hit, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, user.ID, required); err == nil && hit && effect == permissionsupport.EffectDeny {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}
		ready, err := permissionsupport.SubjectAPIPermissionReadyContext(ctx, db, user.ID, roleID, permissionsupport.PlatformClient)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "权限校验失败"})
			c.Abort()
			return
		}
		if !ready {
			c.Next(ctx)
			return
		}
		allowed, err := permissionsupport.SubjectHasPermissionContext(ctx, db, user.ID, roleID, required)
		if err != nil || !allowed {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

func currentClientUser(c *app.RequestContext) (*model.User, bool) {
	value, ok := c.Get("user")
	if !ok {
		return nil, false
	}
	user, ok := value.(*model.User)
	return user, ok
}

func ensureClientRoleID(ctx context.Context, db *gorm.DB, user *model.User) uint {
	if user == nil || user.ID == 0 || user.RoleID > 0 || db == nil {
		if user == nil {
			return 0
		}
		return user.RoleID
	}
	var current model.User
	if err := db.WithContext(ctx).Select("user_role_id").Where("`id` = ?", user.ID).First(&current).Error; err == nil {
		user.RoleID = current.RoleID
	}
	return user.RoleID
}

func clientRoutePermission(method, path string) (string, bool) {
	method = strings.ToUpper(method)
	for _, route := range appapiperm.ClientRouteDeclarations() {
		if route.Method == method && routePatternMatches(route.Path, path) {
			return route.PermissionKey, true
		}
	}
	return "", false
}

func requestPath(c *app.RequestContext) string {
	path := string(c.Path())
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}
