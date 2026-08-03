package exam

import (
	"github.com/cloudwego/hertz/pkg/app"

	examservice "wecheckin/backend/internal/app/service/exam"
)

type ClientExamHandler struct {
	svc *examservice.Service
}

func NewClientExamHandler() *ClientExamHandler {
	return &ClientExamHandler{svc: examservice.NewService()}
}

func (h *ClientExamHandler) service() *examservice.Service {
	if h.svc == nil {
		h.svc = examservice.NewService()
	}
	return h.svc
}

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
