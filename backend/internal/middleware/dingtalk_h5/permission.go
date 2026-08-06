package dingtalkh5

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/dingtalkh5session"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/response"
)

func DingTalkH5Perm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		user, ok := dingtalkh5session.CurrentUser(c)
		if !ok || user.ID == 0 {
			response.Fail(c, "未登录")
			c.Abort()
			return
		}
		required, ok := dingTalkH5RoutePermission(string(c.Method()), h5RequestPath(c))
		if !ok {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}

		db, cancel := database.WithContext(ctx)
		defer cancel()
		if !permissionsupport.TablesReady(db) {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}
		if effect, hit, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, user.ID, required); err == nil && hit && effect == permissionsupport.EffectDeny {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}
		roleIDs, err := ensureDingTalkH5RoleIDs(ctx, db, user)
		if err != nil {
			response.Fail(c, "权限校验失败")
			c.Abort()
			return
		}
		ready, err := permissionsupport.SubjectAPIPermissionReadyWithRoleIDsContext(ctx, db, user.ID, roleIDs, permissionsupport.PlatformDingTalkH5)
		if err != nil {
			response.Fail(c, "权限校验失败")
			c.Abort()
			return
		}
		if !ready {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}
		allowed, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, user.ID, roleIDs, required)
		if err != nil || !allowed {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

func dingTalkH5RoutePermission(method, path string) (string, bool) {
	method = strings.ToUpper(method)
	for _, route := range appapiperm.DingTalkH5RouteDeclarations() {
		if route.Method == method && dingTalkH5RoutePatternMatches(route.Path, path) {
			return route.PermissionKey, true
		}
	}
	return "", false
}

func ensureDingTalkH5RoleIDs(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]uint, error) {
	if user == nil {
		return nil, nil
	}
	if len(user.RoleIDs) > 0 {
		return user.RoleIDs, nil
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}
	user.RoleIDs = roleIDs
	return roleIDs, nil
}

func h5RequestPath(c *app.RequestContext) string {
	path := string(c.Path())
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}

func dingTalkH5RoutePatternMatches(pattern, path string) bool {
	patternParts := dingTalkH5RoutePatternParts(pattern)
	pathParts := dingTalkH5RoutePatternParts(path)
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

func dingTalkH5RoutePatternParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
