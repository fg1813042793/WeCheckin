package swagger

import (
	notificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/pkg/response"
)

var _ notificationoutboxapp.WebhookRecordList
var _ notificationapp.DingTalkDeliveryRecordList
var _ notificationapp.DingTalkDeliveryRecord
var _ response.Resp

type ScheduledRunDeleteRequest struct {
	IDs []string `json:"ids"`
}

// @Tags API v2-后台管理-通知
// @Summary 查询 Webhook 投递记录
// @Security AdminToken
// @Param status query string false "投递状态"
// @Param sourceType query string false "来源类型"
// @Param webhookType query string false "Webhook 类型"
// @Param addTimeFrom query int false "开始毫秒时间戳"
// @Param addTimeTo query int false "结束毫秒时间戳"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp{data=notificationoutboxapp.WebhookRecordList}
// @Router /api/v2/admin/webhook-notifications [get]
func swaggerWebhookRecordList() {}

// @Tags API v2-后台管理-通知
// @Summary 查看脱敏后的 Webhook 投递详情
// @Security AdminToken
// @Param id path int true "投递记录 ID"
// @Success 200 {object} response.Resp{data=notificationoutboxapp.WebhookRecord}
// @Router /api/v2/admin/webhook-notifications/{id} [get]
func swaggerWebhookRecordDetail() {}

// @Tags API v2-后台管理-通知
// @Summary 删除 Webhook 投递记录
// @Security AdminToken
// @Param id path int true "投递记录 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/webhook-notifications/{id} [delete]
func swaggerWebhookRecordDelete() {}

// @Tags API v2-后台管理-通知
// @Summary 重试 Webhook 投递
// @Description 请求受理后投递状态仍以运行记录为准。
// @Security AdminToken
// @Param id path int true "投递记录 ID"
// @Success 200 {object} response.Resp{data=notificationoutboxapp.WebhookRecord}
// @Router /api/v2/admin/webhook-notifications/{id}/retry [post]
func swaggerWebhookRecordRetry() {}

// @Tags API v2-后台管理-通知
// @Summary 查询钉钉通知投递记录
// @Description 返回后台手工发送、定时任务等来源的钉钉投递数量与终态，不产生站内信未读记录。
// @Security AdminToken
// @Param title query string false "通知标题"
// @Param sourceType query string false "来源类型"
// @Param status query string false "sent、partial、failed_final或skipped"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkDeliveryRecordList}
// @Router /api/v2/admin/dingtalk-notification-records [get]
func swaggerDingTalkDeliveryRecordList() {}

// @Tags API v2-后台管理-通知
// @Summary 查看钉钉通知投递详情
// @Security AdminToken
// @Param id path int true "钉钉投递记录 ID"
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkDeliveryRecord}
// @Router /api/v2/admin/dingtalk-notification-records/{id} [get]
func swaggerDingTalkDeliveryRecordDetail() {}

// @Tags API v2-后台管理-通知
// @Summary 删除钉钉通知投递记录
// @Description 只删除后台记录，不撤回已发送的钉钉通知。
// @Security AdminToken
// @Param id path int true "钉钉投递记录 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk-notification-records/{id} [delete]
func swaggerDingTalkDeliveryRecordDelete() {}

// @Tags API v2-后台管理-定时任务
// @Summary 批量删除终态运行记录及分段日志
// @Security AdminToken
// @Accept json
// @Param request body ScheduledRunDeleteRequest true "运行 ID 列表"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/scheduled-task-runs [delete]
func swaggerScheduledRunBatchDelete() {}
