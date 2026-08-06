package survey

import (
	"context"

	calcPkg "wecheckin/backend/internal/formkit/calc"
	questionPkg "wecheckin/backend/internal/formkit/question"
	schemaPkg "wecheckin/backend/internal/formkit/schema"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/pkg/response"
)

// PublicValidate POST /survey/validate
// @Tags 客户端-表单工具
// @Summary 校验答案格式（通用）
// @Router /survey/validate [post]
func (h *ClientSurveyHandler) PublicValidate(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Schema   string                 `json:"schema"`
		Answers  map[string]interface{} `json:"answers"`
		Device   string                 `json:"device"`
		DeviceID string                 `json:"deviceId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	schema := req.Schema
	if req.SurveyID > 0 {
		surveySchema, limitMessage, err := h.survey.ValidatePublicSurveyContext(ctx, req.SurveyID, req.DeviceID, c.ClientIP())
		if err != nil {
			response.Fail(c, "问卷不存在或未发布")
			return
		}
		if limitMessage != "" {
			response.JSON(c, publicValidationResponse{Valid: false, Errors: []map[string]string{{"questionId": "", "message": limitMessage}}})
			return
		}
		schema = surveySchema
	}
	s, err := schemaPkg.Parse(schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	type fieldErr struct {
		QuestionID string `json:"questionId"`
		Message    string `json:"message"`
	}
	var errs []fieldErr
	for _, q := range s.Questions {
		v, ok := req.Answers[q.ID]
		if !ok && q.Required {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: "此项为必填"})
			continue
		}
		inst := questionPkg.Get(q.Type)
		if inst == nil {
			continue
		}
		if err := inst.Validate(v, q); err != nil {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: err.Error()})
		}
	}
	response.JSON(c, publicValidationResponse{Valid: len(errs) == 0, Errors: errs})
}

// PublicApply POST /survey/apply
// @Tags 客户端-表单工具
// @Summary 应用表单逻辑（通用）
// @Router /survey/apply [post]
func (h *ClientSurveyHandler) PublicApply(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema  string                 `json:"schema"`
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	eng := calcPkg.New()
	newAns, _ := eng.ApplyCalcValues(s, req.Answers)
	states, _ := eng.ApplyLogic(s, newAns)
	response.JSON(c, publicApplyResponse{Answers: newAns, States: states})
}
