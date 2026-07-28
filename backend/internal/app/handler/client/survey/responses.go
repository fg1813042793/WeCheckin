package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/response"
)

// MyResponses GET /survey/my_responses
// @Tags 客户端-问卷
// @Summary 我的答卷列表
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /survey/my_responses [get]
func (h *ClientSurveyHandler) MyResponses(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	list, err := h.responses.MyResponsesContext(ctx, uid, 50)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, myResponsesResponse{List: list})
}

// MyResponseDetail GET /survey/my_response?id=
// @Tags 客户端-问卷
// @Summary 查看答卷详情
// @Param id query int true "答卷ID"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /survey/my_response [get]
func (h *ClientSurveyHandler) MyResponseDetail(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	id, _ := strconv.Atoi(c.Query("id"))
	resp, err := h.responses.GetContext(ctx, uint(id))
	if err != nil {
		response.Fail(c, "答卷不存在")
		return
	}
	uidStr := strconv.FormatUint(uint64(uid), 10)
	if resp.UserID != uidStr {
		response.Fail(c, "无权查看")
		return
	}
	sv, _ := h.survey.GetContext(ctx, resp.SurveyID)
	answers := h.responses.ParseAnswers(resp)
	out := myResponseDetailResponse{Response: resp, Answers: answers.Answers}
	if sv != nil && sv.ShowResult == 1 {
		out.Survey = sv
	}
	response.JSON(c, out)
}
