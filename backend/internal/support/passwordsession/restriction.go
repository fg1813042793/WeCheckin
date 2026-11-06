package passwordsession

import "strings"

type Audience string

const (
	AudienceAdmin      Audience = "admin"
	AudienceClient     Audience = "client"
	AudienceDingTalkH5 Audience = "dingtalk_h5"

	ErrorCode    = "PASSWORD_CHANGE_REQUIRED"
	ErrorMessage = "请先修改登录密码"
)

type allowedRoute struct {
	method string
	path   string
}

var restrictedAllowlists = map[Audience][]allowedRoute{
	AudienceAdmin: {
		{method: "PATCH", path: "/api/v2/admin/me/password"},
		{method: "POST", path: "/api/v2/admin/auth/logout"},
		{method: "GET", path: "/api/v2/admin/password-policy"},
	},
	AudienceClient: {
		{method: "PATCH", path: "/api/v2/me/password"},
		{method: "POST", path: "/api/v2/me/logout"},
	},
	AudienceDingTalkH5: {
		{method: "PATCH", path: "/api/v2/dingtalk/h5/account/password"},
		{method: "POST", path: "/api/v2/dingtalk/h5/logout"},
	},
}

func Allowed(audience Audience, method, path string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	for _, route := range restrictedAllowlists[audience] {
		if route.method == method && route.path == path {
			return true
		}
	}
	return false
}
