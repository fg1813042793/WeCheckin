package exam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	examservice "wecheckin/backend/internal/service/client/exam"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

func gradingRequest(examID, recordID, body string) *app.RequestContext {
	c := &app.RequestContext{}
	c.Params = param.Params{{Key: "id", Value: examID}, {Key: "recordId", Value: recordID}}
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.SetBodyString(body)
	c.Set("admin", &model.Admin{ID: 7})
	return c
}

func gradingResponse(t *testing.T, c *app.RequestContext) response.Resp {
	t.Helper()
	var result response.Resp
	if err := json.Unmarshal(c.Response.Body(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRecordGradingRejectsInvalidPathBeforeService(t *testing.T) {
	h := &AdminExamHandler{}
	for _, handler := range []struct {
		name string
		call func(context.Context, *app.RequestContext)
	}{{"GET", h.RecordGrading}, {"POST", h.GradeRecord}} {
		for _, invalid := range []string{"", "0", "-1", "+1", "1.5", " 1", "1x", "18446744073709551616"} {
			for _, field := range []string{"exam", "record"} {
				t.Run(handler.name+"/"+field+"/"+invalid, func(t *testing.T) {
					c := gradingRequest("8", "99", `{"expectedVersion":"v1","scores":[{"questionId":1,"score":0}]}`)
					if field == "exam" {
						c.Params[0].Value = invalid
					} else {
						c.Params[1].Value = invalid
					}
					// Query parameters cannot rescue or replace an invalid path ID.
					c.Request.URI().QueryArgs().Set("id", "8")
					c.Request.URI().QueryArgs().Set("recordId", "99")
					handler.call(context.Background(), c)
					if got := gradingResponse(t, c); got.Code == 0 {
						t.Fatalf("invalid path accepted: %+v", got)
					}
				})
			}
		}
	}
}

func TestGradeRecordRejectsInvalidJSONBeforeService(t *testing.T) {
	h := &AdminExamHandler{}
	for name, body := range map[string]string{
		"empty": "", "whitespace": "  ", "null": "null", "object": "{}", "array": "[]",
		"broken": `{`, "trailing": `{"expectedVersion":"v1","scores":[{"questionId":1,"score":0}]} {}`,
		"unknown":         `{"expectedVersion":"v1","scores":[{"questionId":1,"score":0}],"answers":"secret"}`,
		"missing version": `{"scores":[{"questionId":1,"score":0}]}`,
		"missing scores":  `{"expectedVersion":"v1"}`,
		"empty scores":    `{"expectedVersion":"v1","scores":[]}`,
		"missing score":   `{"expectedVersion":"v1","scores":[{"questionId":1}]}`,
		"null score":      `{"expectedVersion":"v1","scores":[{"questionId":1,"score":null}]}`,
		"fraction":        `{"expectedVersion":"v1","scores":[{"questionId":1,"score":1.5}]}`,
		"negative score":  `{"expectedVersion":"v1","scores":[{"questionId":1,"score":-1}]}`,
		"zero question":   `{"expectedVersion":"v1","scores":[{"questionId":0,"score":1}]}`,
		"long comment":    `{"expectedVersion":"v1","scores":[{"questionId":1,"score":0}],"comment":"` + strings.Repeat("评", 1001) + `"}`,
	} {
		t.Run(name, func(t *testing.T) {
			c := gradingRequest("8", "99", body)
			h.GradeRecord(context.Background(), c)
			if got := gradingResponse(t, c); got.Code == 0 {
				t.Fatalf("invalid JSON accepted: %+v", got)
			}
		})
	}
}

func TestGradeRecordRejectsOversizedBodyBeforeService(t *testing.T) {
	c := gradingRequest("8", "99", `{}`)
	c.Request.Header.SetContentLength(64*1024*1024 + 1)
	(&AdminExamHandler{}).GradeRecord(context.Background(), c)
	if got := gradingResponse(t, c); got.Code == 0 {
		t.Fatalf("oversized body accepted: %+v", got)
	}
}

func TestRecordGradingRequiresAuthenticatedAdmin(t *testing.T) {
	h := &AdminExamHandler{}
	for _, call := range []func(context.Context, *app.RequestContext){h.RecordGrading, h.GradeRecord} {
		for _, actor := range []interface{}{nil, (*model.Admin)(nil), &model.Admin{}, "7"} {
			c := gradingRequest("8", "99", `{}`)
			c.Set("admin", actor)
			call(context.Background(), c)
			if got := gradingResponse(t, c); got.Code == 0 || got.Msg != "请先登录" {
				t.Fatalf("invalid actor accepted: %+v", got)
			}
		}
	}
}

func TestRecordGradingUsesPathIDs(t *testing.T) {
	c := gradingRequest("8", "99", `{}`)
	examID, recordID, admin, ok := recordGradingIdentity(c)
	if !ok || examID != 8 || recordID != 99 || admin.ID != 7 {
		t.Fatalf("identity = %d, %d, %+v, %v", examID, recordID, admin, ok)
	}
}

func TestRecordGradingJSONAcceptsZeroAndUnicodeComment(t *testing.T) {
	c := gradingRequest("8", "99", `{"expectedVersion":"v1","scores":[{"questionId":1,"score":0}],"comment":"`+strings.Repeat("评", 1000)+`"}`)
	req, err := decodeRecordGradingRequest(c)
	if err != nil || len(req.Scores) != 1 || req.Scores[0].Score == nil || *req.Scores[0].Score != 0 || req.ExpectedVersion != "v1" {
		t.Fatalf("zero score or valid unicode comment rejected: %+v, %v", req, err)
	}
}

func TestRecordGradingErrorsUseStablePublicMessages(t *testing.T) {
	for _, test := range []struct {
		err     error
		message string
	}{
		{examservice.ErrRecordNotGradable, "当前答卷状态不可批改"},
		{examservice.ErrRecordGradingChanged, "答卷已发生变化，请刷新后重试"},
		{examservice.ErrRecordGradingInvalid, "批改参数无效，请检查评分和评语"},
		{examservice.ErrRecordGradingUnavailable, "答卷批改暂不可用，请联系管理员"},
		{gorm.ErrRecordNotFound, "考试或答卷不存在"},
		{errors.New("internal database failure"), "批改操作失败，请稍后重试"},
	} {
		t.Run(test.message, func(t *testing.T) {
			c := gradingRequest("8", "99", "")
			respondRecordGrading(context.Background(), c, nil, fmt.Errorf("private SQL or credentials: %w", test.err))
			got := gradingResponse(t, c)
			if got.Code == 0 || got.Msg != test.message {
				t.Fatalf("public error = %+v", got)
			}
		})
	}
}

func TestRecordGradingSuccessReturnsDetail(t *testing.T) {
	c := gradingRequest("8", "99", "")
	detail := &examservice.RecordGradingDetail{ExamTitle: "考试", Version: "v1", CanGrade: true}
	respondRecordGrading(context.Background(), c, detail, nil)
	var result struct {
		Code int                             `json:"code"`
		Data examservice.RecordGradingDetail `json:"data"`
	}
	if err := json.Unmarshal(c.Response.Body(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 0 || result.Data.ExamTitle != "考试" || result.Data.Version != "v1" || !result.Data.CanGrade {
		t.Fatalf("grading response = %+v", result)
	}
}

func TestRecordGradingAuditContainsOnlyRecordIdentity(t *testing.T) {
	ctx, recorder := auditlog.WithRecorder(context.Background())
	recordGradingAudit(ctx, 8, 99, 0)
	if details := recorder.Snapshot(); details.AffectedCount != 0 {
		t.Fatalf("failed operation marked successful: %+v", details)
	}
	recordGradingAudit(ctx, 8, 99, 1)
	details := recorder.Snapshot()
	if details.AffectedCount != 1 || details.TargetType != "exam_record" || len(details.TargetIDs) != 1 || details.TargetIDs[0] != "99" || details.RelatedType != "exam" || details.RelatedID != "8" {
		t.Fatalf("grading audit = %+v", details)
	}
	if len(details.Changes) != 0 || details.Reason != "" || details.TargetName != "" {
		t.Fatalf("unexpected grading content in audit: %+v", details)
	}
}
