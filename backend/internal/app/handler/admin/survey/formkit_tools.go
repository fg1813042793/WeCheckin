package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	calcPkg "wecheckin/backend/internal/app/formkit/calc"
	questionPkg "wecheckin/backend/internal/app/formkit/question"
	schemaPkg "wecheckin/backend/internal/app/formkit/schema"
	"wecheckin/backend/pkg/response"
)

// ParseSchema POST /admin/survey/schema/parse
// @Tags PC端-表单工具
// @Summary 解析 Schema
// @Param schema formData string true "Schema JSON"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ParseSchema(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema string `json:"schema"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	if req.Schema == "" {
		response.Fail(c, "schema 不能为空")
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	response.JSON(c, s)
}

// EvalExpr POST /admin/survey/eval
// @Tags PC端-表单工具
// @Summary 表达式试算
// @Param expr formData string true "表达式"
// @Param env formData string false "环境变量JSON"
// @Param asBool formData bool false "是否返回布尔值"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) EvalExpr(_ context.Context, c *app.RequestContext) {
	var req struct {
		Expr   string                 `json:"expr"`
		Env    map[string]interface{} `json:"env"`
		AsBool bool                   `json:"asBool"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	e := calcPkg.New()
	if req.AsBool {
		v, err := e.EvalBool(req.Expr, req.Env)
		if err != nil {
			response.Fail(c, "eval 失败: "+err.Error())
			return
		}
		response.JSON(c, map[string]interface{}{"value": v})
		return
	}
	v, err := e.Eval(req.Expr, req.Env)
	if err != nil {
		response.Fail(c, "eval 失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{"value": v})
}

// ValidateAnswers POST /admin/survey/validate
func (h *AdminSurveyHandler) ValidateAnswers(_ context.Context, c *app.RequestContext) {
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
	response.JSON(c, map[string]interface{}{"ok": len(errs) == 0, "errors": errs})
}

// ApplyForm POST /admin/survey/apply
func (h *AdminSurveyHandler) ApplyForm(_ context.Context, c *app.RequestContext) {
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
	response.JSON(c, map[string]interface{}{"answers": newAns, "states": states})
}
