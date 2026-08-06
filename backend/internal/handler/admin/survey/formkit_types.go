package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	questionPkg "wecheckin/backend/internal/formkit/question"
	"wecheckin/backend/pkg/response"
)

// TypeMeta 题型元信息
type TypeMeta struct {
	Type         string                 `json:"type"`
	DisplayName  string                 `json:"displayName"`
	Category     string                 `json:"category"`
	DefaultProps map[string]interface{} `json:"defaultProps"`
}

// ListTypes GET /admin/survey/types
// @Tags PC端-表单工具
// @Summary 获取题型列表
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ListTypes(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	all := questionPkg.All()
	out := make([]TypeMeta, 0, len(all))
	for _, q := range all {
		out = append(out, TypeMeta{
			Type:         q.Type(),
			DisplayName:  q.DisplayName(),
			Category:     q.Category(),
			DefaultProps: q.DefaultProps(),
		})
	}
	response.JSON(c, out)
}
