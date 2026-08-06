package dingtalkh5session

import (
	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
)

const (
	userKey  = "dingtalk_h5_user"
	tokenKey = "dingtalk_h5_token"
)

func SetAuth(c *app.RequestContext, user *model.DingTalkH5PerfUser, token string) {
	c.Set(userKey, user)
	c.Set(tokenKey, token)
}

func CurrentUser(c *app.RequestContext) (*model.DingTalkH5PerfUser, bool) {
	value, ok := c.Get(userKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*model.DingTalkH5PerfUser)
	return user, ok
}

func CurrentToken(c *app.RequestContext) string {
	value, _ := c.Get(tokenKey)
	token, _ := value.(string)
	return token
}
