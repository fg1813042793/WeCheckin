package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

const WorkflowStartNotificationJobKey = "notification.workflow_start.send"

type WorkflowStartNotificationService interface {
	ResolveScheduledWorkflowStartTemplate(context.Context, uint) (inappnotificationapp.DingTalkSendTemplate, error)
	SendDingTalk(context.Context, inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error)
}

type WorkflowStartDefinitionReader interface {
	GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error)
}

type WorkflowStartRecipientResolver interface {
	ResolveRecipients(context.Context, inappnotificationapp.RecipientRule) (inappnotificationapp.RecipientResolution, error)
}

type WorkflowStartNotificationJob struct {
	notifications WorkflowStartNotificationService
	definitions   WorkflowStartDefinitionReader
	recipients    WorkflowStartRecipientResolver
}

func NewWorkflowStartNotificationJob(notifications WorkflowStartNotificationService, definitions WorkflowStartDefinitionReader, recipients WorkflowStartRecipientResolver) *WorkflowStartNotificationJob {
	return &WorkflowStartNotificationJob{notifications: notifications, definitions: definitions, recipients: recipients}
}

func (job *WorkflowStartNotificationJob) Key() string  { return WorkflowStartNotificationJobKey }
func (job *WorkflowStartNotificationJob) Name() string { return "发送流程开始通知" }
func (job *WorkflowStartNotificationJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","required":["templateId"],"properties":{"templateId":{"type":"integer","minimum":1}}}`)
}

type workflowStartNotificationParams struct {
	TemplateID uint `json:"templateId"`
}

func (job *WorkflowStartNotificationJob) Validate(ctx context.Context, raw json.RawMessage) error {
	params, err := decodeWorkflowStartNotificationParams(raw)
	if err != nil {
		return err
	}
	if job == nil || job.notifications == nil || job.definitions == nil || job.recipients == nil {
		return errors.New("workflow start notification service is not initialized")
	}
	template, err := job.notifications.ResolveScheduledWorkflowStartTemplate(ctx, params.TemplateID)
	if err != nil {
		return err
	}
	_, err = job.definitions.GetPublishedDefinition(ctx, template.Target.DefinitionID)
	return err
}

func (job *WorkflowStartNotificationJob) Execute(ctx context.Context, runID string, raw json.RawMessage, logger scheduledtaskapp.RunLogger) (scheduledtaskapp.HandlerResult, error) {
	if job == nil || job.notifications == nil || job.definitions == nil || job.recipients == nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "handler_unavailable", Summary: "workflow start notification service is not initialized"}
	}
	params, err := decodeWorkflowStartNotificationParams(raw)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	template, err := job.notifications.ResolveScheduledWorkflowStartTemplate(ctx, params.TemplateID)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "template_unavailable", Summary: err.Error()}
	}
	definition, err := job.definitions.GetPublishedDefinition(ctx, template.Target.DefinitionID)
	if err != nil || definition == nil {
		if err == nil {
			err = inappnotificationapp.ErrTemplateWorkflowUnavailable
		}
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "workflow_unavailable", Summary: err.Error()}
	}
	userIDs, err := resolveWorkflowStartRecipients(ctx, job.recipients, definition.Initiator)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "recipient_resolution_failed", Summary: err.Error(), Temporary: true}
	}
	if len(userIDs) == 0 {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "no_recipients", Summary: inappnotificationapp.ErrNoRecipients.Error()}
	}
	result, err := job.notifications.SendDingTalk(ctx, inappnotificationapp.SendInput{
		TemplateID: template.ID, TemplateVersion: template.Version,
		Title: template.Title, Content: template.Content,
		Scope: inappnotificationapp.ScopeUsers, UserIDs: userIDs,
		SourceType: inappnotificationapp.SourceScheduledTaskRun, SourceID: runID,
	})
	if err != nil {
		if errors.Is(err, inappnotificationapp.ErrNoRecipients) {
			return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "no_recipients", Summary: err.Error()}
		}
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "notification_delivery_failed", Summary: err.Error(), Temporary: true}
	}
	if logger == nil {
		logger = scheduledtaskapp.NopRunLogger{}
	}
	_ = logger.Log(ctx, "info", "delivery", fmt.Sprintf("planned=%d sent=%d skipped=%d failed=%d", result.PlannedCount, result.SentCount, result.SkippedCount, result.FailedCount))
	return scheduledtaskapp.HandlerResult{Summary: fmt.Sprintf("sent %d workflow start notifications", result.SentCount), Data: map[string]interface{}{
		"templateId": template.ID, "definitionId": template.Target.DefinitionID, "sourceId": result.SourceID,
		"plannedCount": result.PlannedCount, "sentCount": result.SentCount, "skippedCount": result.SkippedCount, "failedCount": result.FailedCount,
	}}, nil
}

func decodeWorkflowStartNotificationParams(raw json.RawMessage) (workflowStartNotificationParams, error) {
	var params workflowStartNotificationParams
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return params, fmt.Errorf("decode workflow start notification params: %w", err)
	}
	if params.TemplateID == 0 || uint64(params.TemplateID) > inappnotificationapp.MaxTemplateID {
		return params, errors.New("workflow start notification templateId is required")
	}
	return params, nil
}

func resolveWorkflowStartRecipients(ctx context.Context, resolver WorkflowStartRecipientResolver, initiator workflowcore.InitiatorConfig) ([]uint, error) {
	ids := make([]uint, 0)
	if initiator.Scope != workflowcore.InitiatorScopeSpecified {
		resolved, err := resolver.ResolveRecipients(ctx, inappnotificationapp.RecipientRule{Scope: inappnotificationapp.ScopeAll})
		if err != nil {
			return nil, err
		}
		ids = append(ids, resolved.UserIDs...)
	} else {
		if len(initiator.UserIDs) > 0 {
			resolved, err := resolver.ResolveRecipients(ctx, inappnotificationapp.RecipientRule{Scope: inappnotificationapp.ScopeUsers, UserIDs: initiator.UserIDs})
			if err != nil {
				return nil, err
			}
			ids = append(ids, resolved.UserIDs...)
		}
		if len(initiator.DepartmentIDs) > 0 {
			resolved, err := resolver.ResolveRecipients(ctx, inappnotificationapp.RecipientRule{Scope: inappnotificationapp.ScopeDepartments, DepartmentIDs: initiator.DepartmentIDs})
			if err != nil {
				return nil, err
			}
			ids = append(ids, resolved.UserIDs...)
		}
	}
	excluded := make(map[uint]struct{}, len(initiator.ExcludedUserIDs))
	for _, id := range initiator.ExcludedUserIDs {
		excluded[id] = struct{}{}
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, blocked := excluded[id]; blocked {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

var _ GoJob = (*WorkflowStartNotificationJob)(nil)
