package exam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	examservice "wecheckin/backend/internal/service/client/exam"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

// Grading carries scalar scores and a short comment, never answer attachments.
const maxRecordGradingBodyBytes = 1 << 20

func (h *AdminExamHandler) RecordGrading(ctx context.Context, c *app.RequestContext) {
	examID, recordID, admin, ok := recordGradingIdentity(c)
	if !ok {
		return
	}
	detail, err := h.svc.RecordGradingContext(ctx, examID, recordID, admin.ID)
	respondRecordGrading(ctx, c, detail, err)
}

func (h *AdminExamHandler) GradeRecord(ctx context.Context, c *app.RequestContext) {
	examID, recordID, admin, ok := recordGradingIdentity(c)
	if !ok {
		return
	}
	recordGradingAudit(ctx, examID, recordID, 0)
	req, err := decodeRecordGradingRequest(c)
	if err != nil {
		respondRecordGrading(ctx, c, nil, err)
		return
	}
	detail, err := h.svc.GradeRecordContext(ctx, examID, recordID, admin.ID, req)
	if err == nil {
		recordGradingAudit(ctx, examID, recordID, 1)
	}
	respondRecordGrading(ctx, c, detail, err)
}

func recordGradingIdentity(c *app.RequestContext) (uint, uint, *model.Admin, bool) {
	examID, examErr := strconv.ParseUint(c.Param("id"), 10, strconv.IntSize)
	recordID, recordErr := strconv.ParseUint(c.Param("recordId"), 10, strconv.IntSize)
	if examErr != nil || examID == 0 || recordErr != nil || recordID == 0 {
		response.Fail(c, "无效的考试或答卷ID")
		return 0, 0, nil, false
	}
	value, _ := c.Get("admin")
	admin, ok := value.(*model.Admin)
	if !ok || admin == nil || admin.ID == 0 {
		response.Fail(c, "请先登录")
		return 0, 0, nil, false
	}
	return uint(examID), uint(recordID), admin, true
}

func decodeRecordGradingRequest(c *app.RequestContext) (examservice.GradeRecordRequest, error) {
	var req examservice.GradeRecordRequest
	if c.Request.Header.ContentLength() > maxRecordGradingBodyBytes {
		return req, examservice.ErrRecordGradingInvalid
	}
	body := c.Request.Body()
	if len(body) == 0 || len(body) > maxRecordGradingBodyBytes {
		return req, examservice.ErrRecordGradingInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return req, examservice.ErrRecordGradingInvalid
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return req, examservice.ErrRecordGradingInvalid
	}
	if strings.TrimSpace(req.ExpectedVersion) == "" || len(req.Scores) == 0 || utf8.RuneCountInString(req.Comment) > 1000 {
		return req, examservice.ErrRecordGradingInvalid
	}
	for _, score := range req.Scores {
		if score.QuestionID == 0 || score.Score == nil || *score.Score < 0 {
			return req, examservice.ErrRecordGradingInvalid
		}
	}
	return req, nil
}

func respondRecordGrading(ctx context.Context, c *app.RequestContext, detail *examservice.RecordGradingDetail, err error) {
	if err == nil {
		response.JSON(c, detail)
		return
	}
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, "考试或答卷不存在")
	case errors.Is(err, examservice.ErrRecordNotGradable):
		response.Fail(c, "当前答卷状态不可批改")
	case errors.Is(err, examservice.ErrRecordGradingChanged):
		response.Fail(c, "答卷已发生变化，请刷新后重试")
	case errors.Is(err, examservice.ErrRecordGradingInvalid):
		response.Fail(c, "批改参数无效，请检查评分和评语")
	case errors.Is(err, examservice.ErrRecordGradingUnavailable):
		response.Fail(c, "答卷批改暂不可用，请联系管理员")
	default:
		response.FailInternal(ctx, c, "admin.exam.grading", "批改操作失败，请稍后重试", err)
	}
}

func recordGradingAudit(ctx context.Context, examID, recordID uint, affected int) {
	auditlog.Update(ctx, func(details *auditlog.Details) {
		details.TargetType = "exam_record"
		details.TargetIDs = []string{strconv.FormatUint(uint64(recordID), 10)}
		details.RelatedType = "exam"
		details.RelatedID = strconv.FormatUint(uint64(examID), 10)
		details.AffectedCount = affected
	})
}
