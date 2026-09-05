package swagger

import (
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/pkg/response"
)

var _ response.Resp
var _ workflowapp.WorkflowOverview
var _ dingtalkh5service.AccountProfilePayload
var _ dingtalkh5service.ReviewPayload
var _ dingtalkh5service.TemplateDTO
var _ dingtalkh5service.UserPayload

type H5AppLoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type H5AppSSOLoginRequest struct {
	CorpID   string `json:"corpId"`
	AuthCode string `json:"authCode"`
}

type H5AppBindSelfRequest struct {
	BindTicket string `json:"bindTicket"`
	Account    string `json:"account"`
	Password   string `json:"password"`
}

type H5AppChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type H5AppWorkflowStartRequest struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	BusinessType      string                 `json:"businessType,omitempty"`
	BusinessKey       string                 `json:"businessKey,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	FormData          map[string]interface{} `json:"formData"`
}

type H5AppWorkflowSaveDraftRequest struct {
	DefinitionVersion int                    `json:"definitionVersion"`
	FormData          map[string]interface{} `json:"formData"`
}

type H5AppWorkflowCompleteTaskRequest struct {
	Action             string                 `json:"action" enums:"approve,reject,return,submit"`
	Comment            string                 `json:"comment,omitempty"`
	ReturnTargetNodeID string                 `json:"returnTargetNodeId,omitempty"`
	Images             []H5AppWorkflowImage   `json:"images,omitempty"`
	Variables          map[string]interface{} `json:"variables,omitempty"`
	FormData           map[string]interface{} `json:"formData,omitempty"`
}

type H5AppWorkflowWithdrawRequest struct {
	Reason string `json:"reason,omitempty"`
}

type H5AppWorkflowCommentRequest struct {
	Comment      string                            `json:"comment"`
	Images       []H5AppWorkflowImage              `json:"images,omitempty"`
	Notification *H5AppWorkflowCommentNotification `json:"notification,omitempty"`
}

type H5AppWorkflowCommentNotification struct {
	UserIDs  []string `json:"userIds" example:"7,84"`
	Channels []string `json:"channels" enums:"in_app,dingtalk_oa" example:"in_app,dingtalk_oa"`
}

type H5AppWorkflowImage struct {
	ID       string `json:"id" example:"uploads/workflow/2026/09/04/comment.png"`
	Name     string `json:"name" example:"comment.png"`
	URL      string `json:"url" example:"/uploads/workflow/2026/09/04/comment.png"`
	MimeType string `json:"mimeType" example:"image/png"`
	Size     int64  `json:"size" example:"2048"`
}

type H5AppWorkflowRemindRequest struct {
	NodeID string `json:"nodeId"`
}

type H5AppWorkflowReviseFormRequest struct {
	ExpectedRevision int64                             `json:"expectedRevision" example:"1"`
	FormData         map[string]interface{}            `json:"formData"`
	Reason           string                            `json:"reason" example:"补充实际验收结果"`
	Notification     *H5AppWorkflowCommentNotification `json:"notification,omitempty"`
}

// @Tags API v2-H5App-认证
// @Summary 查询 H5App 公开配置
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/public-config [get]
func swaggerV2H5AppPublicConfigGet() {}

// @Tags API v2-H5App-认证
// @Summary H5App 账号密码登录
// @Param body body H5AppLoginRequest true "登录信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/login [post]
func swaggerV2H5AppLoginPost() {}

// @Tags API v2-H5App-认证
// @Summary H5App 钉钉免登
// @Param body body H5AppSSOLoginRequest true "免登信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/sso-login [post]
func swaggerV2H5AppSSOLoginPost() {}

// @Tags API v2-H5App-认证
// @Summary H5App 绑定本地账号
// @Param body body H5AppBindSelfRequest true "绑定信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/bind-self [post]
func swaggerV2H5AppBindSelfPost() {}

// @Tags API v2-H5App-认证
// @Summary H5App 退出登录
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/logout [post]
func swaggerV2H5AppLogoutPost() {}

// @Tags API v2-H5App-账户
// @Summary 上传 H5App 用户头像
// @Security H5AppToken
// @Accept multipart/form-data
// @Param file formData file true "头像文件"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/account/avatar [post]
func swaggerV2H5AppAccountAvatarPost() {}

// @Tags API v2-H5App-流程审批
// @Summary 上传 H5App 流程附件
// @Security H5AppToken
// @Accept multipart/form-data
// @Param file formData file true "附件文件，单文件最大20MB"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/attachments [post]
func swaggerV2H5AppWorkflowAttachmentsPost() {}

// @Tags API v2-H5App-账户
// @Summary 更新 H5App 账户资料
// @Security H5AppToken
// @Param body body dingtalkh5service.AccountProfilePayload true "账户资料"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/account/profile [patch]
func swaggerV2H5AppAccountProfilePatch() {}

// @Tags API v2-H5App-账户
// @Summary 修改 H5App 账户密码
// @Security H5AppToken
// @Param body body H5AppChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/account/password [patch]
func swaggerV2H5AppAccountPasswordPatch() {}

// @Tags API v2-H5App-工作台
// @Summary 查询 H5App 启动数据和权限
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/bootstrap [get]
func swaggerV2H5AppBootstrapGet() {}

// @Tags API v2-H5App-工作台
// @Summary 查询 H5App 工作台统计
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workbench [get]
func swaggerV2H5AppWorkbenchGet() {}

// @Tags API v2-H5App-站内信
// @Summary 查询当前用户站内信
// @Security H5AppToken
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，最大 100"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/notifications [get]
func swaggerV2H5AppNotificationsGet() {}

// @Tags API v2-H5App-站内信
// @Summary 查询当前用户未读站内信数量
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/notifications/unread-count [get]
func swaggerV2H5AppNotificationsUnreadCountGet() {}

// @Tags API v2-H5App-站内信
// @Summary 标记当前用户全部站内信为已读
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/notifications/read-all [patch]
func swaggerV2H5AppNotificationsReadAllPatch() {}

// @Tags API v2-H5App-站内信
// @Summary 标记当前用户指定站内信为已读
// @Security H5AppToken
// @Param id path int true "站内信 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/notifications/{id}/read [patch]
func swaggerV2H5AppNotificationsIDReadPatch() {}

// @Tags API v2-H5App-站内信
// @Summary 删除当前用户指定站内信
// @Security H5AppToken
// @Param id path int true "站内信 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/notifications/{id} [delete]
func swaggerV2H5AppNotificationsIDDelete() {}

// @Tags API v2-H5App-绩效考核
// @Summary 查询绩效考核列表
// @Security H5AppToken
// @Param scope query string false "数据范围"
// @Param status query string false "考核状态"
// @Param period query string false "考核周期"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews [get]
func swaggerV2H5AppReviewsGet() {}

// @Tags API v2-H5App-绩效考核
// @Summary 创建绩效考核
// @Security H5AppToken
// @Param body body dingtalkh5service.ReviewPayload true "绩效考核数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews [post]
func swaggerV2H5AppReviewsPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 导出绩效考核
// @Security H5AppToken
// @Param scope query string false "数据范围"
// @Param status query string false "考核状态"
// @Param period query string false "考核周期"
// @Success 200 {file} file
// @Router /api/v2/dingtalk/h5/reviews/export [get]
func swaggerV2H5AppReviewsExportGet() {}

// @Tags API v2-H5App-绩效考核
// @Summary 查询绩效考核详情
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id} [get]
func swaggerV2H5AppReviewsIDGet() {}

// @Tags API v2-H5App-绩效考核
// @Summary 删除绩效考核
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id} [delete]
func swaggerV2H5AppReviewsIDDelete() {}

// @Tags API v2-H5App-绩效考核
// @Summary 保存员工自评
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "自评数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/save-self [post]
func swaggerV2H5AppReviewsIDSaveSelfPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 提交员工自评
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "自评数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/submit-self [post]
func swaggerV2H5AppReviewsIDSubmitSelfPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 提交上级评审
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "评审数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/submit-manager [post]
func swaggerV2H5AppReviewsIDSubmitManagerPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 提交 HRBP 评审
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "评审数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/submit-hrbp [post]
func swaggerV2H5AppReviewsIDSubmitHRBPPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 确认绩效结果
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "确认数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/confirm-result [post]
func swaggerV2H5AppReviewsIDConfirmResultPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 申诉绩效结果
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "申诉数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/dispute-result [post]
func swaggerV2H5AppReviewsIDDisputeResultPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 撤回绩效考核
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload false "撤回信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/withdraw [post]
func swaggerV2H5AppReviewsIDWithdrawPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 退回员工自评
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "退回信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/return-employee [post]
func swaggerV2H5AppReviewsIDReturnEmployeePost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 退回上级评审
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "退回信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/return-manager [post]
func swaggerV2H5AppReviewsIDReturnManagerPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 退回 HRBP 评审
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "退回信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/return-hrbp [post]
func swaggerV2H5AppReviewsIDReturnHRBPPost() {}

// @Tags API v2-H5App-绩效考核
// @Summary 归档绩效考核
// @Security H5AppToken
// @Param id path string true "考核编号"
// @Param body body dingtalkh5service.ReviewPayload true "归档数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/reviews/{id}/finalize [post]
func swaggerV2H5AppReviewsIDFinalizePost() {}

// @Tags API v2-H5App-绩效用户
// @Summary 查询绩效用户
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/users [get]
func swaggerV2H5AppUsersGet() {}

// @Tags API v2-H5App-绩效用户
// @Summary 创建绩效用户
// @Security H5AppToken
// @Param body body dingtalkh5service.UserPayload true "用户数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/users [post]
func swaggerV2H5AppUsersPost() {}

// @Tags API v2-H5App-绩效用户
// @Summary 更新绩效用户
// @Security H5AppToken
// @Param id path string true "用户账号"
// @Param body body dingtalkh5service.UserPayload true "用户数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/users/{id} [put]
func swaggerV2H5AppUsersIDPut() {}

// @Tags API v2-H5App-绩效用户
// @Summary 删除绩效用户
// @Security H5AppToken
// @Param id path string true "用户账号"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/users/{id} [delete]
func swaggerV2H5AppUsersIDDelete() {}

// @Tags API v2-H5App-绩效模板
// @Summary 查询绩效模板
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/template [get]
func swaggerV2H5AppTemplateGet() {}

// @Tags API v2-H5App-绩效模板
// @Summary 保存绩效模板
// @Security H5AppToken
// @Param body body dingtalkh5service.TemplateDTO true "绩效模板"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/template [put]
func swaggerV2H5AppTemplatePut() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 工作流概览统计
// @Description 一次返回我的待办、已处理、我的申请和抄送我的数量，仅执行聚合统计，不读取列表记录
// @Security H5AppToken
// @Success 200 {object} response.Resp{data=workflowapp.WorkflowOverview}
// @Router /api/v2/dingtalk/h5/workflows/overview [get]
func swaggerV2H5AppWorkflowOverviewGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 可发起流程
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/definitions [get]
func swaggerV2H5AppWorkflowDefinitionsGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 流程分类
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/categories [get]
func swaggerV2H5AppWorkflowCategoriesGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 流程定义
// @Security H5AppToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/definitions/{id} [get]
func swaggerV2H5AppWorkflowDefinitionsIDGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 流程发起草稿
// @Security H5AppToken
// @Param definitionId path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/drafts/{definitionId} [get]
func swaggerV2H5AppWorkflowDraftsDefinitionIDGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 保存 H5App 流程发起草稿
// @Security H5AppToken
// @Param definitionId path int true "流程定义 ID"
// @Param body body H5AppWorkflowSaveDraftRequest true "草稿数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/drafts/{definitionId} [put]
func swaggerV2H5AppWorkflowDraftsDefinitionIDPut() {}

// @Tags API v2-H5App-OA流程
// @Summary 删除 H5App 流程发起草稿
// @Security H5AppToken
// @Param definitionId path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/drafts/{definitionId} [delete]
func swaggerV2H5AppWorkflowDraftsDefinitionIDDelete() {}

// @Tags API v2-H5App-OA流程
// @Summary H5App 发起 OA 流程
// @Security H5AppToken
// @Param body body H5AppWorkflowStartRequest true "流程发起数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances [post]
func swaggerV2H5AppWorkflowInstancesPost() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 我的 OA 流程
// @Security H5AppToken
// @Param scope query string false "范围: started, handled, copied"
// @Param definitionId query int false "流程定义 ID"
// @Param definitionName query string false "流程名称关键字（模糊匹配，最多 50 个字符）"
// @Param definitionCategory query string false "流程定义分类"
// @Param starterName query string false "发起人用户名关键字"
// @Param status query string false "流程状态"
// @Param businessType query string false "业务类型"
// @Param businessKey query string false "业务键"
// @Param startTimeFrom query int false "开始时间起始值（毫秒）"
// @Param startTimeTo query int false "开始时间截止值（毫秒）"
// @Param endTimeFrom query int false "结束时间起始值（毫秒）"
// @Param endTimeTo query int false "结束时间截止值（毫秒）"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances [get]
func swaggerV2H5AppWorkflowInstancesGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App OA 流程详情
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id} [get]
func swaggerV2H5AppWorkflowInstancesIDGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 从 H5App 我的申请中删除已结束流程
// @Description 仅发起人可删除已结束流程；流程运行数据和审计记录仍由后台保留
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id} [delete]
func swaggerV2H5AppWorkflowInstancesIDDelete() {}

// @Tags API v2-H5App-OA流程
// @Summary 撤回 H5App OA 流程
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Param body body H5AppWorkflowWithdrawRequest false "撤回信息"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id}/withdraw [post]
func swaggerV2H5AppWorkflowInstancesIDWithdrawPost() {}

// @Tags API v2-H5App-OA流程
// @Summary 评论 H5App OA 流程
// @Description 流程发起人、处理人和抄送人可以添加评论，支持文字或最多 9 张已上传的流程图片；可选通知该流程的其他参与人，站内信与钉钉通知按请求渠道发送，钉钉评论通知使用 ActionCard
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Param body body H5AppWorkflowCommentRequest true "评论内容"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id}/comments [post]
func swaggerV2H5AppWorkflowInstancesIDCommentsPost() {}

// @Tags API v2-H5App-流程审批
// @Summary 提醒当前节点处理人
// @Security H5AppToken
// @Param id path string true "流程实例编号"
// @Param body body H5AppWorkflowRemindRequest true "催办节点"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id}/reminders [post]
func swaggerV2H5AppWorkflowInstancesIDRemindersPost() {}

// @Tags API v2-H5App-OA流程
// @Summary 修改已处理流程实例的表单数据
// @Description 仅流程仍在运行、当前用户实际办理过已启用该能力的节点时可用；只允许修改这些节点配置为 write 且不参与路由条件判断的字段。expectedRevision 用于乐观锁校验；通知可省略，收件人仅限流程参与人，渠道支持站内信和钉钉 OA。
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Param body body H5AppWorkflowReviseFormRequest true "表单修改内容"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/instances/{id}/form-data [patch]
func swaggerV2H5AppWorkflowInstancesIDFormDataPatch() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询 H5App 待办任务
// @Security H5AppToken
// @Param instanceId query string false "流程实例 ID"
// @Param status query string false "任务状态"
// @Param definitionName query string false "流程名称关键字（模糊匹配，最多 50 个字符）"
// @Param definitionCategory query string false "流程定义分类"
// @Param starterName query string false "发起人用户名关键字"
// @Param startTimeFrom query int false "流程提交时间起始值（毫秒）"
// @Param startTimeTo query int false "流程提交时间截止值（毫秒）"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/tasks [get]
func swaggerV2H5AppWorkflowTasksGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 完成 H5App OA 流程任务
// @Description reject 会终止整个流程；return 会退回至已执行过的上游人工节点并继续运行，未传 returnTargetNodeId 时默认上一人工节点。reject 和 return 均需填写处理意见，可附加最多 9 张已上传的流程图片
// @Security H5AppToken
// @Param id path string true "流程任务 ID"
// @Param body body H5AppWorkflowCompleteTaskRequest true "任务处理数据"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/tasks/{id}/complete [post]
func swaggerV2H5AppWorkflowTasksIDCompletePost() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询可汇总的已发布流程定义
// @Security H5AppToken
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/summary/definitions [get]
func swaggerV2H5AppWorkflowSummaryDefinitionsGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询可汇总的流程定义详情
// @Security H5AppToken
// @Param id path int true "流程定义 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/summary/definitions/{id} [get]
func swaggerV2H5AppWorkflowSummaryDefinitionsIDGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询流程汇总实例
// @Description 按调用者的统一数据权限查询流程实例，可跨流程定义汇总并覆盖全部历史发布版本
// @Security H5AppToken
// @Param definitionId query int false "流程定义 ID"
// @Param definitionName query string false "流程名称关键字"
// @Param definitionVersion query int false "实例绑定的流程版本"
// @Param starterName query string false "发起人用户名关键字"
// @Param status query string false "流程状态"
// @Param startTimeFrom query int false "发起时间起始值（毫秒）"
// @Param startTimeTo query int false "发起时间截止值（毫秒）"
// @Param endTimeFrom query int false "完成时间起始值（毫秒）"
// @Param endTimeTo query int false "完成时间截止值（毫秒）"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，仅支持 20 或 50"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/summary/instances [get]
func swaggerV2H5AppWorkflowSummaryInstancesGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 查询流程汇总实例详情
// @Description 使用实例绑定的发布版本还原只读表单和完整节点进度，并校验发起人数据范围
// @Security H5AppToken
// @Param id path string true "流程实例 ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/dingtalk/h5/workflows/summary/instances/{id} [get]
func swaggerV2H5AppWorkflowSummaryInstancesIDGet() {}

// @Tags API v2-H5App-OA流程
// @Summary 导出流程汇总表单
// @Description 单个实例直接返回文件；批量 PDF/Word 返回 ZIP，批量 Excel 返回每实例一个工作表的 XLSX
// @Security H5AppToken
// @Param definitionId query int true "流程定义 ID"
// @Param instanceIds query string true "当前页选中的流程实例 ID，逗号分隔，最多 50 个"
// @Param format query string true "导出格式" Enums(pdf,xlsx,docx)
// @Success 200 {file} file
// @Router /api/v2/dingtalk/h5/workflows/summary/export [get]
func swaggerV2H5AppWorkflowSummaryExportGet() {}
