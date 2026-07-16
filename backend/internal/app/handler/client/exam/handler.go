package exam

import "github.com/cloudwego/hertz/pkg/app"

type ClientExamHandler struct{}

func NewClientExamHandler() *ClientExamHandler { return &ClientExamHandler{} }

func getUID(c *app.RequestContext) uint {
	uidVal, _ := c.Get("user_id")
	if uidVal == nil {
		return 0
	}
	switch v := uidVal.(type) {
	case uint:
		return v
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	}
	return 0
}
