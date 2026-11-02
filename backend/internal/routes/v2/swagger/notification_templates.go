package swagger

import (
	notificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/pkg/response"
)

var _ notificationapp.DingTalkSendTemplate
var _ notificationapp.DingTalkTemplateInput
var _ notificationapp.DingTalkTemplateList
var _ notificationapp.DingTalkTemplateWorkflowList
var _ response.Resp

// @Tags API v2-后台管理-通知模板
// @Summary 查询钉钉发送模板
// @Description 需要notification:style:list权限；支持查看停用模板。
// @Security AdminToken
// @Param keyword query string false "模板名称或标题，最多80字"
// @Param id query int false "按模板ID定位"
// @Param enabled query bool false "按启用状态筛选"
// @Param page query int false "页码，1至10000" default(1)
// @Param pageSize query int false "每页数量，1至100" default(20)
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkTemplateList}
// @Router /api/v2/admin/dingtalk-notification-templates [get]
func swaggerDingTalkTemplateList() {}

// @Tags API v2-后台管理-通知模板
// @Summary 查询发送时可选的启用模板
// @Description 只需notification:dingtalk:send权限，不要求编辑模板权限；不产生通知投递。
// @Security AdminToken
// @Param keyword query string false "模板名称或标题，最多80字"
// @Param id query int false "按模板ID定位"
// @Param page query int false "页码，1至10000" default(1)
// @Param pageSize query int false "每页数量，1至100" default(20)
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkTemplateList}
// @Router /api/v2/admin/dingtalk-notification-templates/options [get]
func swaggerDingTalkTemplateOptions() {}

// @Tags API v2-后台管理-通知模板
// @Summary 搜索模板跳转目标的已发布流程
// @Description 需要notification:style:edit权限；仅返回ID和显示名称，不返回流程定义正文。
// @Security AdminToken
// @Param keyword query string false "流程名称，最多80字"
// @Param id query int false "按已选流程ID定位"
// @Param page query int false "页码，1至10000" default(1)
// @Param pageSize query int false "每页数量，1至100" default(20)
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkTemplateWorkflowList}
// @Router /api/v2/admin/dingtalk-notification-templates/workflow-options [get]
func swaggerDingTalkTemplateWorkflowOptions() {}

// @Tags API v2-后台管理-通知模板
// @Summary 创建钉钉发送模板
// @Description 需要notification:style:edit权限。类型支持oa、link、action_card；目标支持app_home、workflow_start、url。自定义地址只允许完整HTTPS URL。
// @Security AdminToken
// @Accept json
// @Param body body notificationapp.DingTalkTemplateInput true "模板内容、启用状态和跳转目标"
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkSendTemplate}
// @Router /api/v2/admin/dingtalk-notification-templates [post]
func swaggerDingTalkTemplateCreate() {}

// @Tags API v2-后台管理-通知模板
// @Summary 修改或启停钉钉发送模板
// @Description 需要notification:style:edit权限；version必须为读取时版本，冲突时拒绝覆盖。
// @Security AdminToken
// @Accept json
// @Param id path int true "模板ID"
// @Param body body notificationapp.DingTalkTemplateInput true "完整模板配置及读取时版本"
// @Success 200 {object} response.Resp{data=notificationapp.DingTalkSendTemplate}
// @Router /api/v2/admin/dingtalk-notification-templates/{id} [put]
func swaggerDingTalkTemplateUpdate() {}

// @Tags API v2-后台管理-通知模板
// @Summary 删除钉钉发送模板
// @Description 需要notification:style:edit权限；不撤回已经发出的消息。
// @Security AdminToken
// @Param id path int true "模板ID"
// @Param version query int true "读取时的模板版本，冲突时拒绝删除"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk-notification-templates/{id} [delete]
func swaggerDingTalkTemplateDelete() {}

// @Tags API v2-后台管理-考试管理
// @Summary 查询考试通知模版
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-notification-templates [get]
func swaggerExamNotificationTemplatesGet() {}

// @Tags API v2-后台管理-考试管理
// @Summary 保存考试通知模版
// @Security AdminToken
// @Accept json
// @Param body body ExamNotificationTemplatesSaveRequest true "考试通知模版"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/exam-notification-templates [put]
func swaggerExamNotificationTemplatesPut() {}
