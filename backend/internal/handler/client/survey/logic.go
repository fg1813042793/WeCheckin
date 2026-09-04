package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	surveyservice "wecheckin/backend/internal/service/client/survey"
	"wecheckin/backend/pkg/response"
)

// ApplyLogic POST /survey/apply
// 应用 schema 逻辑（计算值 + 显隐）
func (h *ClientSurveyHandler) ApplyLogic(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Answers  map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "client.survey.logic", "参数错误，请稍后重试", err)
		return
	}
	sv, err := h.survey.Get(req.SurveyID)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	out, _ := h.survey.ApplyLogic(sv, req.Answers)
	response.JSON(c, answersResponse{Answers: out.Answers})
}

// Validate POST /survey/validate
func (h *ClientSurveyHandler) Validate(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Answers  map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "client.survey.logic", "参数错误，请稍后重试", err)
		return
	}
	sv, err := h.survey.Get(req.SurveyID)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	errs := surveyservice.ValidateAnswers(sv, req.Answers)
	response.JSON(c, validationResponse{Errors: errs, Valid: len(errs) == 0})
}
