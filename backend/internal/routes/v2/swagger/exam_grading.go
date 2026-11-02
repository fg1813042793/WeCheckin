package swagger

import (
	examservice "wecheckin/backend/internal/service/client/exam"
	"wecheckin/backend/pkg/response"
)

var _ examservice.RecordGradingDetail
var _ examservice.GradeRecordRequest
var _ response.Resp

// @Tags API v2-后台-考试
// @Summary 获取考试答卷批改详情
// @Description 需要 exam-record:grade 权限及考试数据范围授权。返回题目、作答、参考答案、评分和版本；已完成的答卷只能查看。
// @Security AdminToken
// @Produce json
// @Param id path int true "考试ID"
// @Param recordId path int true "答卷ID"
// @Success 200 {object} response.Resp{data=examservice.RecordGradingDetail}
// @Router /api/v2/admin/exams/{id}/records/{recordId}/grading [get]
func AdminExamRecordGrading() {}

// @Tags API v2-后台-考试
// @Summary 提交考试答卷人工批改
// @Description 需要 exam-record:grade 权限及考试数据范围授权。仅待批改答卷可提交，提交不可重复；版本或状态变化时须刷新。必须为所有人工题提供整数分数，不能修改自动题分数，评语最多1000字。
// @Security AdminToken
// @Accept json
// @Produce json
// @Param id path int true "考试ID"
// @Param recordId path int true "答卷ID"
// @Param body body examservice.GradeRecordRequest true "批改请求，提交不可重复；expectedVersion 取自批改详情，scores 包含全部人工题评分"
// @Success 200 {object} response.Resp{data=examservice.RecordGradingDetail}
// @Router /api/v2/admin/exams/{id}/records/{recordId}/grading [post]
func AdminExamGradeRecord() {}
