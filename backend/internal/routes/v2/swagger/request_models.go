package swagger

// PublicSurveyApplyRequest describes the JSON payload for form logic evaluation.
type PublicSurveyApplyRequest struct {
	// Schema 是表单 Schema JSON。
	Schema string `json:"schema" example:"{\"questions\":[]}"`
	// Answers 是以字段 ID 为键的当前答案。
	Answers map[string]interface{} `json:"answers"`
}

// PublicSurveyValidateRequest describes the JSON payload for public form validation.
type PublicSurveyValidateRequest struct {
	// SurveyID 是已发布问卷 ID；传入后优先使用问卷保存的 Schema。
	SurveyID uint `json:"surveyId" example:"1"`
	// Schema 是未指定问卷 ID 时使用的表单 Schema JSON。
	Schema string `json:"schema" example:"{\"questions\":[]}"`
	// Answers 是待校验的答案。
	Answers map[string]interface{} `json:"answers"`
	// Device 是设备类型。
	Device string `json:"device" example:"h5"`
	// DeviceID 是设备唯一标识。
	DeviceID string `json:"deviceId" example:"device-001"`
}

// PublicSurveySubmissionRequest describes a public survey submission payload.
type PublicSurveySubmissionRequest struct {
	// Answers 是以题目 ID 为键的答案。
	Answers map[string]interface{} `json:"answers"`
	// Nickname 是匿名答题人昵称。
	Nickname string `json:"nickname" example:"匿名用户"`
	// Session 是匿名答题会话标识。
	Session string `json:"session" example:"survey-session"`
	// StartTime 是开始答题时间的 Unix 毫秒时间戳。
	StartTime int64 `json:"startTime" example:"1788316800000"`
	// Device 是设备或 User-Agent 信息。
	Device string `json:"device" example:"h5"`
	// DeviceID 是设备唯一标识。
	DeviceID string `json:"deviceId" example:"device-001"`
	// AutoSubmit 表示是否由超时机制自动提交。
	AutoSubmit bool `json:"autoSubmit" example:"false"`
}

// PublicExamSubmissionRequest describes an exam submission payload.
type PublicExamSubmissionRequest struct {
	// RecordID 是已开始考试的答题记录 ID。
	RecordID int `json:"recordId" example:"1"`
	// Answers 是以题目 ID 为键的答案。
	Answers map[string]interface{} `json:"answers"`
	// Session 是匿名考试会话标识。
	Session string `json:"session" example:"exam-session"`
	// Device 是设备类型。
	Device string `json:"device" example:"h5"`
	// DeviceID 是设备唯一标识。
	DeviceID string `json:"deviceId" example:"device-001"`
	// AutoSubmit 表示是否由超时机制自动提交。
	AutoSubmit bool `json:"autoSubmit" example:"false"`
}

// PublicExamValidationRequest describes the payload used before submitting an exam.
type PublicExamValidationRequest struct {
	// Answers 是待校验的考试答案。
	Answers map[string]interface{} `json:"answers"`
	// Device 是设备类型。
	Device string `json:"device" example:"h5"`
	// DeviceID 是设备唯一标识。
	DeviceID string `json:"deviceId" example:"device-001"`
}

// WorkflowStartDraftRequest describes a user's saved workflow start form.
type WorkflowStartDraftRequest struct {
	// DefinitionVersion 是草稿对应的流程发布版本。
	DefinitionVersion int `json:"definitionVersion" example:"1"`
	// FormData 是流程发起表单数据。
	FormData map[string]interface{} `json:"formData"`
}

// WorkflowStartInstanceRequest describes a client-started workflow instance.
type WorkflowStartInstanceRequest struct {
	// DefinitionID 是流程定义 ID。
	DefinitionID uint `json:"definitionId" example:"1"`
	// DefinitionVersion 是发起时使用的流程发布版本。
	DefinitionVersion int `json:"definitionVersion" example:"1"`
	// BusinessType 是关联业务类型。
	BusinessType string `json:"businessType" example:"performance_review"`
	// BusinessKey 是关联业务唯一键。
	BusinessKey string `json:"businessKey" example:"review-202609"`
	// Variables 是流程运行变量。
	Variables map[string]interface{} `json:"variables"`
	// FormData 是流程发起表单数据。
	FormData map[string]interface{} `json:"formData"`
}

// AdminWorkflowStartInstanceRequest describes an administrator-started workflow instance.
type AdminWorkflowStartInstanceRequest struct {
	WorkflowStartInstanceRequest
	// StarterID 是业务发起人用户 ID。
	StarterID string `json:"starterId" example:"1001"`
}

// AdminWorkflowInstanceDeleteRequest describes a batch workflow instance deletion.
type AdminWorkflowInstanceDeleteRequest struct {
	// IDs 是待删除的终态流程实例 ID，单次最多 100 个。
	IDs []string `json:"ids" example:"instance-1,instance-2"`
}

// WorkflowCompleteTaskRequest describes an action for a workflow task.
type WorkflowCompleteTaskRequest struct {
	// Action 是任务处理动作；reject 终止整个流程，return 将流程退回至已执行过的上游人工节点并继续运行。
	Action string `json:"action" enums:"approve,reject,return,submit" example:"approve"`
	// Comment 是处理意见；reject 和 return 动作必填。
	Comment string `json:"comment" example:"同意"`
	// ReturnTargetNodeID 是退回目标节点 ID；return 动作省略时默认退回上一人工节点。
	ReturnTargetNodeID string `json:"returnTargetNodeId,omitempty" example:"manager_review"`
	// Variables 是本次处理更新的流程变量。
	Variables map[string]interface{} `json:"variables"`
	// FormData 是本次节点填写或更新的表单数据。
	FormData map[string]interface{} `json:"formData"`
}

// WorkflowReasonRequest describes an optional withdrawal or cancellation reason.
type WorkflowReasonRequest struct {
	// Reason 是撤回或取消原因。
	Reason string `json:"reason" example:"业务申请撤销"`
}

// WorkflowDispatchDueRequest limits one manual notification dispatch batch.
type WorkflowDispatchDueRequest struct {
	// Limit 是本次最多投递的通知数量。
	Limit int `json:"limit" example:"100"`
}

// ScheduledTaskStatusRequest applies an optimistic-locking status change.
type ScheduledTaskStatusRequest struct {
	// Enabled 表示是否启用定时任务。
	Enabled bool `json:"enabled" example:"true"`
	// Version 是用于乐观锁校验的当前版本号。
	Version int64 `json:"version" example:"1"`
}

// ScheduledTaskCronPreviewRequest describes a Cron preview calculation.
type ScheduledTaskCronPreviewRequest struct {
	// Expression 是 Cron 表达式。
	Expression string `json:"expression" example:"0 */5 * * * *"`
	// Precision 是表达式精度。
	Precision string `json:"precision" enums:"second,minute" example:"second"`
	// Timezone 是 IANA 时区名称。
	Timezone string `json:"timezone" example:"Asia/Shanghai"`
	// Count 是需要返回的未来执行时间数量。
	Count int `json:"count" example:"5"`
	// AfterMillis 是预览起点的 Unix 毫秒时间戳。
	AfterMillis int64 `json:"afterMillis" example:"1788316800000"`
}

// InAppNotificationSendRequest describes an administrator-created in-app or DingTalk notification.
type InAppNotificationSendRequest struct {
	// RequestID 是客户端生成的幂等请求标识；为空时由服务端生成。
	RequestID string `json:"requestId" example:"8f9f4f33-f9b8-4f2f-9b26-f83c33d4f17f"`
	// Title 是通知标题。
	Title string `json:"title" example:"系统维护通知"`
	// Content 是通知正文。
	Content string `json:"content" example:"系统将于今晚 22:00 进行维护。"`
	// Scope 是收件范围。
	Scope string `json:"scope" enums:"all,departments,users" example:"departments"`
	// UserIDs 是指定用户范围下的用户 ID。
	UserIDs []uint `json:"userIds" example:"66,67"`
	// DepartmentIDs 是指定部门范围下的部门 ID，执行时包含下级部门。
	DepartmentIDs []uint `json:"departmentIds" example:"3,5"`
}

// DingTalkNotificationSendRequest describes an administrator-created DingTalk notification.
type DingTalkNotificationSendRequest struct {
	// Title 是钉钉通知标题。
	Title string `json:"title" example:"系统维护通知"`
	// Content 是钉钉通知正文。
	Content string `json:"content" example:"系统将于今晚 22:00 进行维护。"`
	// Scope 是收件范围。
	Scope string `json:"scope" enums:"all,departments,users" example:"departments"`
	// UserIDs 是指定用户范围下的用户 ID。
	UserIDs []uint `json:"userIds" example:"66,67"`
	// DepartmentIDs 是指定部门范围下的部门 ID，执行时包含下级部门。
	DepartmentIDs []uint `json:"departmentIds" example:"3,5"`
}

// SurveyQuestionBankRequest describes a reusable survey question.
type SurveyQuestionBankRequest struct {
	// Title 是题目标题。
	Title string `json:"title" example:"满意度"`
	// Type 是题目组件类型。
	Type string `json:"type" example:"radio"`
	// Schema 是题目配置 JSON。
	Schema string `json:"schema" example:"{}"`
	// Category 是题目分类。
	Category string `json:"category" example:"通用"`
	// Tags 是逗号分隔的题目标签。
	Tags string `json:"tags" example:"满意度,通用"`
}

// SurveyTemplatePresetsRequest describes all presets saved for the current administrator.
type SurveyTemplatePresetsRequest struct {
	// Presets 是当前管理员的模板预设列表。
	Presets []map[string]string `json:"presets"`
}
