package exam

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// SaveAnswer POST /exam/save_answer
// @Tags 客户端-考试
// @Summary 保存答案
// @Param recordId formData int true "记录ID"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /exam/save_answer [post]
func (h *ClientExamHandler) SaveAnswer(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	recordID, _ := strconv.Atoi(c.PostForm("recordId"))
	answersJSON := c.PostForm("answers")
	if recordID == 0 {
		response.Fail(c, "recordId 必填")
		return
	}
	var rec model.ExamRecord
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uid).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	if rec.Status == 2 {
		response.Fail(c, "已提交，不可修改")
		return
	}
	if err := db.Model(&rec).Update("exam_r_answers", answersJSON).Error; err != nil {
		response.Fail(c, "保存失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
