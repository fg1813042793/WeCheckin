package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type WorkflowBusinessEventDispatcher interface {
	DispatchDueBusinessEvents(context.Context, int) (int, error)
}

type WorkflowBusinessEventJob struct {
	dispatcher WorkflowBusinessEventDispatcher
}

func NewWorkflowBusinessEventJob(dispatcher WorkflowBusinessEventDispatcher) *WorkflowBusinessEventJob {
	return &WorkflowBusinessEventJob{dispatcher: dispatcher}
}

func (*WorkflowBusinessEventJob) Key() string { return "workflow.business-event.dispatch_due" }

func (*WorkflowBusinessEventJob) Name() string { return "派发工作流业务事件" }

func (*WorkflowBusinessEventJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","minimum":1,"maximum":1000}}}`)
}

func (*WorkflowBusinessEventJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeWorkflowBusinessEventParams(raw)
	return err
}

func (job *WorkflowBusinessEventJob) Execute(
	ctx context.Context,
	_ string,
	raw json.RawMessage,
	_ application.RunLogger,
) (application.HandlerResult, error) {
	if job == nil || job.dispatcher == nil {
		return application.HandlerResult{}, errors.New("workflow business event dispatcher is not initialized")
	}
	params, err := decodeWorkflowBusinessEventParams(raw)
	if err != nil {
		return application.HandlerResult{}, err
	}
	count, err := job.dispatcher.DispatchDueBusinessEvents(ctx, params.Limit)
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{
			Code: "workflow_business_event_dispatch_failed", Summary: err.Error(), Temporary: true,
		}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("dispatched %d workflow business events", count),
		Data:    map[string]interface{}{"dispatched": count},
	}, nil
}

type workflowBusinessEventParams struct {
	Limit int `json:"limit"`
}

func decodeWorkflowBusinessEventParams(raw json.RawMessage) (workflowBusinessEventParams, error) {
	params := workflowBusinessEventParams{Limit: 100}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, fmt.Errorf("decode workflow business event params: %w", err)
		}
	}
	if params.Limit < 1 || params.Limit > 1000 {
		return params, errors.New("workflow business event dispatch limit must be between 1 and 1000")
	}
	return params, nil
}

var _ GoJob = (*WorkflowBusinessEventJob)(nil)
