package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	calcPkg "wecheckin/backend/internal/formkit/calc"
	questionPkg "wecheckin/backend/internal/formkit/question"
	schemaPkg "wecheckin/backend/internal/formkit/schema"
	"wecheckin/backend/pkg/response"
)

type ParseSchemaRequest struct {
	Schema string `json:"schema"`
}

type EvalExprRequest struct {
	Expr   string                 `json:"expr"`
	Env    map[string]interface{} `json:"env"`
	AsBool bool                   `json:"asBool"`
}

type EvalExprResponse struct {
	Value interface{} `json:"value"`
}

type ValidateAnswersRequest struct {
	Schema  string                 `json:"schema"`
	Answers map[string]interface{} `json:"answers"`
}

type FieldValidationError struct {
	QuestionID string `json:"questionId"`
	Message    string `json:"message"`
}

type ValidateAnswersResponse struct {
	OK     bool                   `json:"ok"`
	Errors []FieldValidationError `json:"errors"`
}

type ApplyFormRequest struct {
	Schema  string                 `json:"schema"`
	Answers map[string]interface{} `json:"answers"`
}

type ApplyFormResponse struct {
	Answers map[string]interface{} `json:"answers"`
	States  map[string]string      `json:"states"`
}

// ParseSchema POST /admin/survey/schema/parse
// @Tags PC端-表单工具
// @Summary 解析 Schema
// @Param schema formData string true "Schema JSON"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ParseSchema(ctx context.Context, c *app.RequestContext) {
	var req ParseSchemaRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "请求参数错误，请稍后重试", err)
		return
	}
	if req.Schema == "" {
		response.Fail(c, "schema 不能为空")
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "schema 解析失败，请稍后重试", err)
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
func (h *AdminSurveyHandler) EvalExpr(ctx context.Context, c *app.RequestContext) {
	var req EvalExprRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "请求参数错误，请稍后重试", err)
		return
	}
	e := calcPkg.New()
	if req.AsBool {
		v, err := e.EvalBool(req.Expr, req.Env)
		if err != nil {
			response.FailInternal(ctx, c, "admin.survey.formkit_tools", "eval 失败，请稍后重试", err)
			return
		}
		response.JSON(c, EvalExprResponse{Value: v})
		return
	}
	v, err := e.Eval(req.Expr, req.Env)
	if err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "eval 失败，请稍后重试", err)
		return
	}
	response.JSON(c, EvalExprResponse{Value: v})
}

// ValidateAnswers POST /admin/survey/validate
func (h *AdminSurveyHandler) ValidateAnswers(ctx context.Context, c *app.RequestContext) {
	var req ValidateAnswersRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "请求参数错误，请稍后重试", err)
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "schema 解析失败，请稍后重试", err)
		return
	}
	var errs []FieldValidationError
	for _, q := range s.Questions {
		v, ok := req.Answers[q.ID]
		if !ok && q.Required {
			errs = append(errs, FieldValidationError{QuestionID: q.ID, Message: "此项为必填"})
			continue
		}
		inst := questionPkg.Get(q.Type)
		if inst == nil {
			continue
		}
		if err := inst.Validate(v, q); err != nil {
			errs = append(errs, FieldValidationError{QuestionID: q.ID, Message: err.Error()})
		}
	}
	response.JSON(c, ValidateAnswersResponse{OK: len(errs) == 0, Errors: errs})
}

// ApplyForm POST /admin/survey/apply
func (h *AdminSurveyHandler) ApplyForm(ctx context.Context, c *app.RequestContext) {
	var req ApplyFormRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "请求参数错误，请稍后重试", err)
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.FailInternal(ctx, c, "admin.survey.formkit_tools", "schema 解析失败，请稍后重试", err)
		return
	}
	eng := calcPkg.New()
	newAns, _ := eng.ApplyCalcValues(s, req.Answers)
	states, _ := eng.ApplyLogic(s, newAns)
	response.JSON(c, ApplyFormResponse{Answers: newAns, States: states})
}
