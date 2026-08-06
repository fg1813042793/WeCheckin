package client

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
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
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}
		roleIDs := ensureClientRoleIDs(ctx, db, user)
		if effect, hit, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, user.ID, required); err == nil && hit && effect == permissionsupport.EffectDeny {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}
		ready, err := permissionsupport.SubjectAPIPermissionReadyWithRoleIDsContext(ctx, db, user.ID, roleIDs, permissionsupport.PlatformClient)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "权限校验失败"})
			c.Abort()
			return
		}
		if !ready {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "无权限访问"})
			c.Abort()
			return
		}
		allowed, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, user.ID, roleIDs, required)
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

func ensureClientRoleIDs(ctx context.Context, db *gorm.DB, user *model.User) []uint {
	if user == nil {
		return nil
	}
	if user.ID == 0 || len(user.RoleIDs) > 0 || db == nil {
		return user.RoleIDs
	}
	var current model.User
	if err := db.WithContext(ctx).Select("user_role_id").Where("`id` = ?", user.ID).First(&current).Error; err == nil {
		user.RoleID = current.RoleID
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, user.ID, user.RoleID)
	if err == nil {
		user.RoleIDs = roleIDs
	}
	return user.RoleIDs
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

func routePatternMatches(pattern, path string) bool {
	patternParts := routePatternParts(pattern)
	pathParts := routePatternParts(path)
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func routePatternParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
