package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type ClientExamHandler struct{}

func NewClientExamHandler() *ClientExamHandler { return &ClientExamHandler{} }

type examViewData struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        string                 `json:"tags"`
	Visibility  int                    `json:"visibility"`
	AllowMulti  int                    `json:"allowMulti"`
	Anonymous   int                    `json:"anonymous"`
	ShowResult  int                    `json:"showResult"`
	StartTime   int64                  `json:"startTime"`
	EndTime     int64                  `json:"endTime"`
	Duration    int                    `json:"duration"`
	MaxAttempts int                    `json:"maxAttempts"`
	ShowScore   int                    `json:"showScore"`
	Status      int                    `json:"status"`
	Settings    map[string]interface{} `json:"settings"`
	Schema      string                 `json:"schema"`
	StartAt     int64                  `json:"startAt"`
	Session     string                 `json:"session"`
}

func (h *ClientExamHandler) View(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	var exam model.Exam
	if err := database.DB.Where("`exam_id` = ? AND `exam_status` = 1", id).First(&exam).Error; err != nil {
		response.Fail(c, "考试不存在或未发布")
		return
	}
	var settings map[string]interface{}
	if exam.Settings != "" {
		json.Unmarshal([]byte(exam.Settings), &settings)
	}
	if settings == nil {
		settings = map[string]interface{}{}
	}
	startAt := time.Now().UnixMilli()
	data := examViewData{
		ID:          exam.ID,
		Title:       exam.Title,
		Description: exam.Description,
		Category:    exam.Category,
		Tags:        exam.Tags,
		Visibility:  exam.Visibility,
		AllowMulti:  exam.AllowMulti,
		Anonymous:   exam.Anonymous,
		ShowResult:  exam.ShowResult,
		StartTime:   exam.StartTime,
		EndTime:     exam.EndTime,
		Duration:    exam.Duration,
		MaxAttempts: exam.MaxAttempts,
		ShowScore:   exam.ShowScore,
		Status:      exam.Status,
		Settings:    settings,
		Schema:      exam.Schema,
		StartAt:     startAt,
		Session:     strconv.FormatInt(startAt, 36),
	}
	response.JSON(c, data)
}

type submitReq struct {
	ExamID  uint                   `json:"examId"`
	Answers map[string]interface{} `json:"answers"`
	Device  string                 `json:"device"`
}

func (h *ClientExamHandler) Submit(_ context.Context, c *app.RequestContext) {
	var req submitReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	var exam model.Exam
	if err := database.DB.Where("`exam_id` = ?", req.ExamID).First(&exam).Error; err != nil {
		response.Fail(c, "考试不存在")
		return
	}
	var answersMap map[string]interface{}
	if len(req.Answers) > 0 {
		answersMap = req.Answers
	} else {
		answersMap = map[string]interface{}{}
	}
	ansBytes, _ := json.Marshal(answersMap)
	now := time.Now().UnixMilli()
	record := model.ExamRecord{
		ExamID:     req.ExamID,
		Answers:    string(ansBytes),
		Status:     1,
		StartTime:  now,
		SubmitTime: now,
		AddIP:      c.ClientIP(),
	}
	database.DB.Create(&record)
	response.JSON(c, record)
}

type validateReq struct {
	ExamID  uint                   `json:"examId"`
	Answers map[string]interface{} `json:"answers"`
}

func (h *ClientExamHandler) Validate(_ context.Context, c *app.RequestContext) {
	var req validateReq
	if err := c.BindAndValidate(&req); err != nil {
		response.JSON(c, map[string]interface{}{"ok": false, "errors": []map[string]string{{"message": "参数错误"}}})
		return
	}
	var exam model.Exam
	if err := database.DB.Where("`exam_id` = ?", req.ExamID).First(&exam).Error; err != nil {
		response.JSON(c, map[string]interface{}{"ok": false, "errors": []map[string]string{{"message": "考试不存在"}}})
		return
	}
	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	json.Unmarshal([]byte(exam.Schema), &schema)
	var errors []map[string]string
	for _, q := range schema.Questions {
		qid, _ := q["id"].(string)
		required, _ := q["required"].(bool)
		if required {
			val := req.Answers[qid]
			if val == nil || val == "" {
				title, _ := q["title"].(string)
				errors = append(errors, map[string]string{"message": "请填写: " + title})
			}
		}
	}
	response.JSON(c, map[string]interface{}{"ok": len(errors) == 0, "errors": errors})
}
