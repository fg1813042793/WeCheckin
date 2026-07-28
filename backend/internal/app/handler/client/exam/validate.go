package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/pkg/logger"
	"wecheckin-backend/backend/pkg/response"
)

// Validate POST /exam/validate
// @Tags 客户端-考试
// @Summary 校验答案（必填项等）
// @Router /exam/validate [post]
func (h *ClientExamHandler) Validate(ctx context.Context, c *app.RequestContext) {
	raw, _ := c.Body()
	var req struct {
		ExamID   int                    `json:"examId"`
		Answers  map[string]interface{} `json:"answers"`
		Device   string                 `json:"device"`
		DeviceID string                 `json:"deviceId"`
	}
	if err := json.Unmarshal(raw, &req); err != nil || req.ExamID == 0 {
		response.Fail(c, "参数错误")
		return
	}
	e, err := h.service().PublishedExamContext(ctx, uint(req.ExamID))
	if err != nil {
		logger.Logger.Printf("[ExamValidate] 考试不存在或未发布 examId=%d", req.ExamID)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	uid := getUID(c)
	uidStr := ""
	if uid > 0 {
		uidStr = strconv.FormatUint(uint64(uid), 10)
	}
	clientIP := c.ClientIP()
	msg, err := h.service().CheckLimitContext(ctx, e, uidStr, req.Device, req.DeviceID, clientIP)
	if err != nil {
		response.Fail(c, "校验失败: "+err.Error())
		return
	}
	if msg != "" {
		response.JSON(c, examValidationResponse{OK: false, Errors: []map[string]string{{"questionId": "", "message": msg}}})
		return
	}
	type fieldErr struct {
		QuestionID string `json:"questionId"`
		Message    string `json:"message"`
	}
	var errs []fieldErr
	if e.PaperID > 0 {
		// PaperID 模式跳过校验（ExamQuestion 无 Required 字段）
	} else {
		var schMap map[string]interface{}
		_ = json.Unmarshal([]byte(e.Schema), &schMap)
		questions, _ := schMap["questions"].([]interface{})
		for _, qRaw := range questions {
			q, ok := qRaw.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := q["id"].(string)
			if id == "" {
				continue
			}
			required, _ := q["required"].(bool)
			val, ok := req.Answers[id]
			if !ok || val == nil || fmt.Sprintf("%v", val) == "" {
				if required {
					errs = append(errs, fieldErr{QuestionID: id, Message: "此项为必填"})
				}
				continue
			}
			typ, _ := q["type"].(string)
			if typ == "judge" {
				s := fmt.Sprintf("%v", val)
				if s != "true" && s != "false" {
					errs = append(errs, fieldErr{QuestionID: id, Message: "判断题答案格式错误"})
				}
			}
		}
	}
	response.JSON(c, examValidationResponse{OK: len(errs) == 0, Errors: errs})
}
