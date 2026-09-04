package httperror

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowinfra "wecheckin/backend/internal/modules/workflow/infrastructure"
	workflowsummary "wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/response"
)

const defaultPublicMessage = "流程操作失败，请稍后重试"

type publicError struct {
	target  error
	message string
}

var permissionErrors = []error{
	workflowapp.ErrStarterAccessDenied,
	workflowapp.ErrInstanceAccessDenied,
	workflowapp.ErrCommentNotificationRecipientDenied,
	workflowapp.ErrReminderStarterOnly,
	workflowdomain.ErrTaskActorMismatch,
	workflowdomain.ErrInstanceStarterMismatch,
	workflowsummary.ErrSummaryAccessDenied,
}

var publicErrors = []publicError{
	{context.Canceled, "请求已取消"},
	{context.DeadlineExceeded, "请求处理超时，请稍后重试"},
	{workflowapp.ErrDefinitionRequired, "流程定义不能为空"},
	{workflowapp.ErrStarterRequired, "流程发起人不能为空"},
	{workflowapp.ErrOperatorRequired, "流程发起操作人不能为空"},
	{workflowapp.ErrStarterInvalid, "流程发起人不存在或已停用"},
	{workflowapp.ErrStarterNotAllowed, "该用户不在流程允许的发起人范围内"},
	{workflowapp.ErrStartNotYetAvailable, "流程尚未到允许发起时间"},
	{workflowapp.ErrStartAvailabilityExpired, "流程已超过允许发起时间"},
	{workflowapp.ErrStartOutsideAvailability, "当前不在流程允许发起时间内"},
	{workflowapp.ErrStartLimitExceeded, "当前周期的流程发起次数已用完"},
	{workflowapp.ErrBusinessReferenceRequired, "业务类型和业务标识不能为空"},
	{workflowapp.ErrTaskIDRequired, "流程任务不能为空"},
	{workflowapp.ErrTaskDeleteNotAllowed, "待激活或待处理的流程任务不能删除"},
	{workflowapp.ErrTaskDeleteTargetNotFound, "流程任务不存在或已删除"},
	{workflowapp.ErrActorRequired, "任务处理人不能为空"},
	{workflowapp.ErrInstanceIDRequired, "流程实例不能为空"},
	{workflowapp.ErrInstanceScopeInvalid, "流程实例查询范围无效"},
	{workflowapp.ErrDefinitionNameSearchTooLong, "流程名称关键字不能超过50个字符"},
	{workflowapp.ErrStarterNameSearchTooLong, "发起人用户名关键字不能超过50个字符"},
	{workflowapp.ErrRunningInstanceCannotDelete, "审批中的申请不能删除，请先撤回"},
	{workflowapp.ErrInstanceDeleteNotAllowed, "当前流程状态不能删除"},
	{workflowapp.ErrInstanceDeleteTargetNotFound, "部分流程实例不存在或已删除"},
	{workflowapp.ErrInstanceDeleteTooMany, "单次最多删除100个流程实例"},
	{workflowapp.ErrInstanceCommentRequired, "流程评论不能为空"},
	{workflowapp.ErrInstanceCommentTooLong, "流程评论不能超过500个字符"},
	{workflowapp.ErrCommentNotificationRecipientsRequired, "请选择评论通知对象"},
	{workflowapp.ErrCommentNotificationRecipientsTooMany, "评论通知对象不能超过100人"},
	{workflowapp.ErrCommentNotificationChannelsRequired, "请选择评论通知方式"},
	{workflowapp.ErrCommentNotificationChannelInvalid, "评论通知方式无效"},
	{workflowapp.ErrTaskRejectCommentRequired, "驳回原因不能为空"},
	{workflowapp.ErrTaskReturnCommentRequired, "退回原因不能为空"},
	{workflowapp.ErrTaskCommentTooLong, "处理意见不能超过500个字符"},
	{workflowapp.ErrReminderNodeRequired, "催办节点不能为空"},
	{workflowapp.ErrReminderNodeUnavailable, "当前节点没有可提醒的待处理任务"},
	{workflowapp.ErrReminderCooldown, "该节点刚刚提醒过，请稍后再试"},
	{workflowapp.ErrReminderDailyLimit, "该节点今天的提醒次数已用完"},
	{workflowapp.ErrDefinitionVersionChanged, "流程定义已更新，请重新打开后保存"},
	{workflowapp.ErrWorkflowImageInvalid, "流程图片数据无效"},
	{workflowapp.ErrWorkflowImageTooMany, "流程图片最多上传9张"},
	{workflowdomain.ErrTaskNotFound, "工作流任务不存在"},
	{workflowdomain.ErrTaskAlreadyHandled, "工作流任务已处理"},
	{workflowdomain.ErrInvalidTaskAction, "工作流任务操作无效"},
	{workflowdomain.ErrInstanceNotRunning, "工作流实例已结束"},
	{workflowdomain.ErrInstanceAlreadyHandled, "流程任务已被处理，不能撤回"},
	{workflowdomain.ErrReturnTargetUnavailable, "当前任务没有可退回的上一节点"},
	{workflowdomain.ErrReturnTargetInvalid, "退回目标必须是已执行过的上游人工节点"},
	{workflowdomain.ErrReturnParallelUnsupported, "并行流程暂不支持退回"},
	{workflowinfra.ErrDefinitionNotPublished, "流程定义尚未发布"},
	{workflowinfra.ErrInstanceNotFound, "流程实例不存在"},
	{workflowinfra.ErrTaskNotFound, "流程任务不存在"},
	{workflowinfra.ErrNotificationNotFound, "通知投递记录不存在"},
	{workflowsummary.ErrDefinitionRequired, "流程定义不能为空"},
	{workflowsummary.ErrInstanceRequired, "流程实例不能为空"},
	{workflowsummary.ErrExportInstancesEmpty, "请选择需要导出的流程实例"},
	{workflowsummary.ErrExportInstancesMany, "单次最多导出50个流程实例"},
	{workflowsummary.ErrExportFormatInvalid, "导出格式仅支持 pdf、xlsx、docx"},
	{workflowsummary.ErrExportBodyTooLarge, "导出文件过大，请减少本次导出数量"},
}

func PublicMessage(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	for _, target := range permissionErrors {
		if errors.Is(err, target) {
			return "无权执行该流程操作", true
		}
	}
	for _, item := range publicErrors {
		if errors.Is(err, item.target) {
			return item.message, true
		}
	}
	var validationErrors workflowcore.ValidationErrors
	if errors.As(err, &validationErrors) {
		return "流程设计数据无效", true
	}
	return "", false
}

func Respond(ctx context.Context, c *app.RequestContext, operation string, err error) {
	if message, ok := PublicMessage(err); ok {
		response.Fail(c, message)
		return
	}
	response.FailInternal(ctx, c, operation, defaultPublicMessage, err)
}
