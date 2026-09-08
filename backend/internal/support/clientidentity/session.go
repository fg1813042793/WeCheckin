package clientidentity

import "github.com/cloudwego/hertz/pkg/app"

const openIDContextKey = "user_openid"

// OpenID returns the authenticated client's MiniOpenID populated by ClientAuth.
func OpenID(c *app.RequestContext) (string, bool) {
	value, ok := c.Get(openIDContextKey)
	if !ok {
		return "", false
	}
	openID, ok := value.(string)
	return openID, ok && openID != ""
}
