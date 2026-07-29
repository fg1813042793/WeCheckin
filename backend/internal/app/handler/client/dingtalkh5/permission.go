package dingtalkh5

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/app/support/appapiperm"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

func (h *Handler) ApiPerm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		user, ok := currentUser(c)
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
			c.Next(ctx)
			return
		}
		if effect, hit, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, user.ID, required); err == nil && hit && effect == permissionsupport.EffectDeny {
			response.Fail(c, "无权限访问")
			c.Abort()
			return
		}
		ready, err := permissionsupport.SubjectAPIPermissionReadyContext(ctx, db, user.ID, user.RoleID, permissionsupport.PlatformDingTalkH5)
		if err != nil {
			response.Fail(c, "权限校验失败")
			c.Abort()
			return
		}
		if !ready {
			c.Next(ctx)
			return
		}
		allowed, err := permissionsupport.SubjectHasPermissionContext(ctx, db, user.ID, user.RoleID, required)
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
		if route.Method == method && routePatternMatches(route.Path, path) {
			return route.PermissionKey, true
		}
	}
	return "", false
}

func h5RequestPath(c *app.RequestContext) string {
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
