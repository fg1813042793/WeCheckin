package swagger

import (
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/pkg/response"
)

var _ adminlogservice.DeleteRequest
var _ adminlogservice.DeleteResult
var _ response.Resp

// @Tags API v2-后台管理-日志
// @Summary 删除单条操作日志
// @Description 需要操作日志查看和删除权限，目标须在当前管理员的数据范围内。
// @Security AdminToken
// @Param id path int true "日志ID"
// @Success 200 {object} response.Resp{data=adminlogservice.DeleteResult}
// @Router /api/v2/admin/logs/{id} [delete]
func swaggerAuditLogDelete() {}

// @Tags API v2-后台管理-日志
// @Summary 批量删除登录日志
// @Description 需要登录日志查看和删除权限；每次1–100个不重复正整数ID，目标必须全部存在且属于登录日志。
// @Security AdminToken
// @Accept json
// @Param body body adminlogservice.DeleteRequest true "待删除日志 ID"
// @Success 200 {object} response.Resp{data=adminlogservice.DeleteResult}
// @Router /api/v2/admin/login-logs [delete]
func swaggerLoginLogsDelete() {}

// @Tags API v2-后台管理-日志
// @Summary 删除单条登录日志
// @Description 需要登录日志查看和删除权限，目标必须属于登录日志。
// @Security AdminToken
// @Param id path int true "日志ID"
// @Success 200 {object} response.Resp{data=adminlogservice.DeleteResult}
// @Router /api/v2/admin/login-logs/{id} [delete]
func swaggerLoginLogDelete() {}
