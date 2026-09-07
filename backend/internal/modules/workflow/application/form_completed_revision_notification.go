package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

const formRevisionNotificationSourceType = "workflow_form_revision"

type formRevisionNotificationRecipient struct {
	userID string
	nodeID string
	taskID string
}

func (service *Service) buildFormRevisionNotifications(
	ctx context.Context,
	store TransactionStore,
	definition workflowcore.Definition,
	state *workflowdomain.State,
	revision *workflowdomain.FormRevisionRequest,
	kind workflowdomain.NotificationKind,
	recipients []formRevisionNotificationRecipient,
	eventKey string,
) ([]FormRevisionNotificationIntent, error) {
	if service == nil || service.ids == nil || state == nil || revision == nil {
		return nil, nil
	}
	recipients = uniqueFormRevisionNotificationRecipients(recipients)
	if len(recipients) == 0 {
		return nil, nil
	}
	requesterName, err := store.UserDisplayName(ctx, revision.RequesterID)
	if err != nil {
		return nil, err
	}
	workflowName := definition.EffectiveName()
	if workflowName == "" {
		workflowName = state.Instance.DefinitionKey
	}
	title, content := formRevisionNotificationText(kind, workflowName, requesterName, revision)
	view := "workflow:form-revision-detail:" + revision.ID
	intents := make([]FormRevisionNotificationIntent, 0, len(recipients))
	for _, recipient := range recipients {
		intents = append(intents, FormRevisionNotificationIntent{
			ID: service.ids.NewID("notification"), Kind: string(kind),
			NodeID: recipient.nodeID, TaskID: recipient.taskID,
			RecipientUserID: recipient.userID,
			Channels:        []string{workflowcore.NotificationChannelInApp, workflowcore.NotificationChannelDingTalkOA},
			DedupeKeySuffix: strings.TrimSpace(eventKey),
			Payload: NotificationPayload{
				Title: title, Content: content, MessageType: NotificationMessageTypeActionCard,
				WorkflowName: workflowName, InstanceTitle: state.Instance.Title,
				BusinessPeriod: state.Instance.BusinessPeriodLabel,
				StarterID:      state.Instance.StarterID, InstanceID: state.Instance.ID,
				TaskID: recipient.taskID, RecipientUserID: recipient.userID, Kind: string(kind),
				SourceType: formRevisionNotificationSourceType, SourceID: revision.ID, View: view,
			},
		})
	}
	return intents, nil
}

func formRevisionNotificationText(
	kind workflowdomain.NotificationKind,
	workflowName, requesterName string,
	revision *workflowdomain.FormRevisionRequest,
) (string, string) {
	fields := strings.Join(revision.ChangedFieldLabels, "、")
	switch kind {
	case workflowdomain.NotificationKindInstanceFormRevisionRequested:
		return "表单修订待确认", fmt.Sprintf("%s 申请修改《%s》的字段：%s", requesterName, workflowName, fields)
	case workflowdomain.NotificationKindInstanceFormRevisionApproved:
		return "表单修订已生效", fmt.Sprintf("《%s》的表单修订已确认并生效，当前版本：%d", workflowName, revision.AppliedFormRevision)
	case workflowdomain.NotificationKindInstanceFormRevisionRejected:
		return "表单修订已驳回", fmt.Sprintf("《%s》的表单修订未通过确认", workflowName)
	case workflowdomain.NotificationKindInstanceFormRevisionCancelled:
		return "表单修订已取消", fmt.Sprintf("%s 已取消《%s》的表单修订", requesterName, workflowName)
	case workflowdomain.NotificationKindInstanceFormRevisionConflict:
		return "表单修订未生效", fmt.Sprintf("《%s》的表单版本已变化，请重新发起修订", workflowName)
	default:
		return "流程表单有更新", fmt.Sprintf("《%s》的表单状态已更新", workflowName)
	}
}

func uniqueFormRevisionNotificationRecipients(values []formRevisionNotificationRecipient) []formRevisionNotificationRecipient {
	seen := make(map[string]struct{}, len(values))
	result := make([]formRevisionNotificationRecipient, 0, len(values))
	for _, value := range values {
		value.userID = strings.TrimSpace(value.userID)
		if value.userID == "" {
			continue
		}
		if _, exists := seen[value.userID]; exists {
			continue
		}
		seen[value.userID] = struct{}{}
		result = append(result, value)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].userID < result[j].userID })
	return result
}

func formRevisionTaskRecipients(tasks []workflowdomain.FormRevisionTask, taskIDs map[string]struct{}) []formRevisionNotificationRecipient {
	result := make([]formRevisionNotificationRecipient, 0)
	for _, task := range tasks {
		if taskIDs != nil {
			if _, exists := taskIDs[task.ID]; !exists {
				continue
			}
		} else if task.Status != workflowdomain.RevisionTaskStatusPending {
			continue
		}
		result = append(result, formRevisionNotificationRecipient{userID: task.AssigneeID, nodeID: task.NodeID, taskID: task.ID})
	}
	return result
}

func formRevisionResultRecipients(state *workflowdomain.State, revision *workflowdomain.FormRevisionRequest, includeHandlers bool) []formRevisionNotificationRecipient {
	result := []formRevisionNotificationRecipient{{userID: revision.RequesterID}, {userID: state.Instance.StarterID}}
	if includeHandlers {
		for _, task := range revision.Tasks {
			if task.Status != workflowdomain.RevisionTaskStatusApproved {
				continue
			}
			userID := task.HandledBy
			if strings.TrimSpace(userID) == "" {
				userID = task.AssigneeID
			}
			result = append(result, formRevisionNotificationRecipient{userID: userID, nodeID: task.NodeID, taskID: task.ID})
		}
	}
	return result
}

func revisionBusinessEvent(service *Service, state *workflowdomain.State, revision *workflowdomain.FormRevisionRequest) WorkflowBusinessEvent {
	return WorkflowBusinessEvent{
		ID:        service.ids.NewID("business-event"),
		DedupeKey: string(LifecycleInstanceFormRevised) + ":" + revision.ID,
		Event: NewFormRevisedBusinessEvent(
			state.Instance, revision.ID, revision.AppliedFormRevision, revision.Patch,
		),
	}
}
