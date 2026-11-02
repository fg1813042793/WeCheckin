package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

func TestWorkflowStartNotificationJobResolvesCurrentSpecifiedAudience(t *testing.T) {
	notifications := &workflowStartNotificationServiceStub{
		template: inappnotificationapp.DingTalkSendTemplate{
			ID: 7, Version: 3, Title: "填写绩效", Content: "请填写", Target: inappnotificationapp.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17},
		},
		result: inappnotificationapp.SendResult{SourceID: "run-1", PlannedCount: 3, SentCount: 2, SkippedCount: 1},
	}
	definitions := &workflowStartDefinitionReaderStub{definition: &workflowapp.PublishedDefinition{
		ID:        17,
		Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{4, 6}, DepartmentIDs: []uint{9}, ExcludedUserIDs: []uint{6}},
	}}
	recipients := &workflowStartRecipientResolverStub{results: map[inappnotificationapp.RecipientScope]inappnotificationapp.RecipientResolution{
		inappnotificationapp.ScopeUsers:       {UserIDs: []uint{4, 6}, SkippedCount: 1},
		inappnotificationapp.ScopeDepartments: {UserIDs: []uint{5, 6, 7}, SkippedCount: 2},
	}}
	logger := &captureRunLogger{}
	job := NewWorkflowStartNotificationJob(notifications, definitions, recipients)

	result, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"templateId":7}`), logger)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "notification.workflow_start.send" || job.Name() != "发送流程开始通知" {
		t.Fatalf("metadata = %q/%q", job.Key(), job.Name())
	}
	if !reflect.DeepEqual(notifications.input.UserIDs, []uint{4, 5, 7}) || notifications.input.Scope != inappnotificationapp.ScopeUsers {
		t.Fatalf("send input = %#v", notifications.input)
	}
	if notifications.input.TemplateID != 7 || notifications.input.TemplateVersion != 3 || notifications.input.Title != "填写绩效" || notifications.input.Content != "请填写" || notifications.input.SourceID != "run-1" {
		t.Fatalf("template send input = %#v", notifications.input)
	}
	if result.Data["sentCount"] != 2 || !strings.Contains(logger.content, "planned=3") || strings.Contains(logger.content, "请填写") {
		t.Fatalf("result=%#v log=%q", result, logger.content)
	}
}

func TestWorkflowStartNotificationJobResolvesAllUsersAndExclusions(t *testing.T) {
	notifications := &workflowStartNotificationServiceStub{template: inappnotificationapp.DingTalkSendTemplate{ID: 8, Version: 1, Title: "题", Content: "文", Target: inappnotificationapp.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 18}}}
	definitions := &workflowStartDefinitionReaderStub{definition: &workflowapp.PublishedDefinition{ID: 18, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll, ExcludedUserIDs: []uint{2}}}}
	recipients := &workflowStartRecipientResolverStub{results: map[inappnotificationapp.RecipientScope]inappnotificationapp.RecipientResolution{inappnotificationapp.ScopeAll: {UserIDs: []uint{1, 2, 3}}}}
	job := NewWorkflowStartNotificationJob(notifications, definitions, recipients)
	if _, err := job.Execute(context.Background(), "run-2", json.RawMessage(`{"templateId":8}`), nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(notifications.input.UserIDs, []uint{1, 3}) {
		t.Fatalf("recipients = %#v", notifications.input.UserIDs)
	}
}

func TestWorkflowStartNotificationJobValidatesTemplateAndMapsNoRecipients(t *testing.T) {
	job := NewWorkflowStartNotificationJob(&workflowStartNotificationServiceStub{}, &workflowStartDefinitionReaderStub{}, &workflowStartRecipientResolverStub{})
	if err := job.Validate(context.Background(), json.RawMessage(`{"templateId":0}`)); err == nil {
		t.Fatal("zero template id must fail")
	}
	notifications := &workflowStartNotificationServiceStub{template: inappnotificationapp.DingTalkSendTemplate{ID: 7, Version: 1, Title: "题", Content: "文", Target: inappnotificationapp.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17}}}
	job = NewWorkflowStartNotificationJob(notifications, &workflowStartDefinitionReaderStub{definition: &workflowapp.PublishedDefinition{ID: 17, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified}}}, &workflowStartRecipientResolverStub{})
	_, err := job.Execute(context.Background(), "run-3", json.RawMessage(`{"templateId":7}`), nil)
	var handlerErr *scheduledtaskapp.HandlerError
	if !errors.As(err, &handlerErr) || handlerErr.Code != "no_recipients" {
		t.Fatalf("error = %v", err)
	}
}

type workflowStartNotificationServiceStub struct {
	template inappnotificationapp.DingTalkSendTemplate
	input    inappnotificationapp.SendInput
	result   inappnotificationapp.SendResult
	err      error
}

func (stub *workflowStartNotificationServiceStub) ResolveScheduledWorkflowStartTemplate(context.Context, uint) (inappnotificationapp.DingTalkSendTemplate, error) {
	return stub.template, stub.err
}

func (stub *workflowStartNotificationServiceStub) SendDingTalk(_ context.Context, input inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error) {
	stub.input = input
	return stub.result, stub.err
}

type workflowStartDefinitionReaderStub struct {
	definition *workflowapp.PublishedDefinition
	err        error
}

func (stub *workflowStartDefinitionReaderStub) GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error) {
	return stub.definition, stub.err
}

type workflowStartRecipientResolverStub struct {
	results map[inappnotificationapp.RecipientScope]inappnotificationapp.RecipientResolution
	err     error
}

func (stub *workflowStartRecipientResolverStub) ResolveRecipients(_ context.Context, rule inappnotificationapp.RecipientRule) (inappnotificationapp.RecipientResolution, error) {
	return stub.results[rule.Scope], stub.err
}
