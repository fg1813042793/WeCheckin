package swagger

import (
	adminmodel "wecheckin/backend/internal/model/admin"
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/pkg/response"
)

var _ adminmodel.Log
var _ adminlogservice.Page
var _ response.Resp

// @Tags API v2-后台管理-日志
// @Summary 查看操作审计详情
// @Description 登录日志需 log:security，导出还需对应 log:export 或 log:security-export。删除需额外授予 log:delete 或 log:security-delete。
// @Description snapshotStatus/snapshotReason独立表示快照采集状态与原因；历史记录默认legacy_unknown。明确白名单的普通配置可返回真实标量，敏感值保持脱敏。
// @Security AdminToken
// @Param id path int true "日志 ID"
// @Success 200 {object} response.Resp{data=adminmodel.Log}
// @Router /api/v2/admin/logs/{id} [get]
func swaggerOperationLogDetail() {}

// @Tags API v2-后台管理-日志
// @Summary 导出操作审计日志（最多1000条）
// @Description 登录日志需 log:security，导出还需对应 log:export 或 log:security-export。拒绝越权查询；不提供删除能力。
// @Security AdminToken
// @Param search query string false "账号、内容或请求 ID"
// @Param source query string false "来源" Enums(admin,client,h5app,system)
// @Param module query string false "模块"
// @Param action query string false "动作"
// @Param outcome query string false "结果" Enums(success,failure,blocked,partial,accepted,unknown)
// @Param actorId query int false "操作者 ID"
// @Param ip query string false "客户端 IP"
// @Param startTime query string false "开始时间：毫秒时间戳或 RFC3339"
// @Param endTime query string false "结束时间：毫秒时间戳或 RFC3339"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量：查询最多100，导出最多1000"
// @Success 200 {object} response.Resp{data=adminlogservice.Page}
// @Router /api/v2/admin/logs/export [get]
func swaggerOperationLogExport() {}

// @Tags API v2-后台管理-日志
// @Summary 查询登录安全日志（独立授权）
// @Description 登录日志需 log:security，导出还需对应 log:export 或 log:security-export。拒绝越权查询；不提供删除能力。
// @Security AdminToken
// @Param search query string false "账号、内容或请求 ID"
// @Param source query string false "来源" Enums(admin,client,h5app,system)
// @Param module query string false "模块"
// @Param action query string false "动作"
// @Param outcome query string false "结果" Enums(success,failure,blocked,partial,accepted,unknown)
// @Param actorId query int false "操作者 ID"
// @Param ip query string false "客户端 IP"
// @Param startTime query string false "开始时间：毫秒时间戳或 RFC3339"
// @Param endTime query string false "结束时间：毫秒时间戳或 RFC3339"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量：查询最多100，导出最多1000"
// @Success 200 {object} response.Resp{data=adminlogservice.Page}
// @Router /api/v2/admin/login-logs [get]
func swaggerLoginLogList() {}

// @Tags API v2-后台管理-日志
// @Summary 查看登录安全日志详情
// @Description 登录日志需 log:security，导出还需对应 log:export 或 log:security-export。拒绝越权查询；不提供删除能力。
// @Security AdminToken
// @Param id path int true "日志 ID"
// @Success 200 {object} response.Resp{data=adminmodel.Log}
// @Router /api/v2/admin/login-logs/{id} [get]
func swaggerLoginLogDetail() {}

// @Tags API v2-后台管理-日志
// @Summary 导出登录安全日志（独立授权，最多1000条）
// @Description 登录日志需 log:security，导出还需对应 log:export 或 log:security-export。拒绝越权查询；不提供删除能力。
// @Security AdminToken
// @Param search query string false "账号、内容或请求 ID"
// @Param source query string false "来源" Enums(admin,client,h5app,system)
// @Param module query string false "模块"
// @Param action query string false "动作"
// @Param outcome query string false "结果" Enums(success,failure,blocked,partial,accepted,unknown)
// @Param actorId query int false "操作者 ID"
// @Param ip query string false "客户端 IP"
// @Param startTime query string false "开始时间：毫秒时间戳或 RFC3339"
// @Param endTime query string false "结束时间：毫秒时间戳或 RFC3339"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量：查询最多100，导出最多1000"
// @Success 200 {object} response.Resp{data=adminlogservice.Page}
// @Router /api/v2/admin/login-logs/export [get]
func swaggerLoginLogExport() {}
