package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/support/auditlog"
)

func validTemplateInput() DingTalkTemplateInput {
	return DingTalkTemplateInput{Name: "流程提醒", Enabled: true, Title: "请填写", Content: "正文", MessageType: "oa", Target: DingTalkTemplateTarget{Kind: "app_home"}}
}

func TestTemplateValidationAndDefaults(t *testing.T) {
	input, err := normalizeTemplateInput(validTemplateInput())
	if err != nil || input.HeadColor != "FF1677FF" || input.ButtonTitle != "打开页面" {
		t.Fatalf("defaults = %#v, %v", input, err)
	}
	cases := []struct {
		name   string
		change func(*DingTalkTemplateInput)
	}{
		{"name", func(v *DingTalkTemplateInput) { v.Name = strings.Repeat("名", 81) }},
		{"title", func(v *DingTalkTemplateInput) { v.Title = strings.Repeat("题", 256) }},
		{"content", func(v *DingTalkTemplateInput) { v.Content = strings.Repeat("文", 5001) }},
		{"button", func(v *DingTalkTemplateInput) { v.ButtonTitle = strings.Repeat("字", 21) }},
		{"color", func(v *DingTalkTemplateInput) { v.HeadColor = "1677FF" }},
		{"type", func(v *DingTalkTemplateInput) { v.MessageType = "text" }},
		{"empty workflow", func(v *DingTalkTemplateInput) { v.Target.Kind = "workflow_start" }},
		{"conflicting target", func(v *DingTalkTemplateInput) { v.Target.URL = "https://example.test" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := validTemplateInput()
			tc.change(&v)
			if _, err := normalizeTemplateInput(v); !errors.Is(err, ErrTemplateInvalid) {
				t.Fatalf("error=%v", err)
			}
		})
	}
	for _, address := range []string{"http://example.test", "https://u:p@example.test", "https:///no-host", "https://example.test/\n", "https://example.test/%0d%0a", "javascript:alert(1)", "https://example.test/" + strings.Repeat("x", 2048)} {
		v := validTemplateInput()
		v.Target = DingTalkTemplateTarget{Kind: "url", URL: address}
		if _, err := normalizeTemplateInput(v); !errors.Is(err, ErrTemplateInvalid) {
			t.Errorf("URL %q accepted: %v", address, err)
		}
	}
}

func TestTemplateSaveUsesTrustedWorkflowAndActor(t *testing.T) {
	store := &templateStoreFake{workflow: DingTalkTemplateWorkflowOption{ID: 17, Name: "正式流程"}}
	input := validTemplateInput()
	input.Target = DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17, DefinitionName: "伪造名称"}
	result, err := NewService(store).CreateDingTalkTemplate(context.Background(), 9, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != 1 || result.Target.DefinitionName != "正式流程" || store.saved.CreatedBy != 9 || store.saved.UpdatedBy != 9 || result.UpdatedAt <= 0 {
		t.Fatalf("result=%#v saved=%#v", result, store.saved)
	}
	store.workflowErr = ErrTemplateWorkflowUnavailable
	if _, err = NewService(store).CreateDingTalkTemplate(context.Background(), 9, input); !errors.Is(err, ErrTemplateWorkflowUnavailable) {
		t.Fatalf("error=%v", err)
	}
}

func TestResolveScheduledWorkflowStartTemplateUsesLatestEnabledVersion(t *testing.T) {
	store := &templateStoreFake{
		row:      notificationmodel.DingTalkTemplate{ID: 7, Version: 4, Enabled: true, Name: "月度绩效", Title: "请填写", Content: "正文", MessageType: "oa", TargetKind: "workflow_start", DefinitionID: 17},
		workflow: DingTalkTemplateWorkflowOption{ID: 17, Name: "研发部绩效"},
	}
	result, err := NewService(store).ResolveScheduledWorkflowStartTemplate(context.Background(), 7)
	if err != nil || result.ID != 7 || result.Version != 4 || result.Target.DefinitionID != 17 || result.Target.DefinitionName != "研发部绩效" {
		t.Fatalf("result=%#v error=%v", result, err)
	}
	store.row.TargetKind = "app_home"
	if _, err = NewService(store).ResolveScheduledWorkflowStartTemplate(context.Background(), 7); !errors.Is(err, ErrTemplateInvalid) {
		t.Fatalf("non-workflow target error=%v", err)
	}
}

func TestTemplateSendRejectsMissingDisabledChangedAndUnpublishedBeforeDelivery(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		row                       notificationmodel.DingTalkTemplate
		getErr, workflowErr, want error
	}{
		{name: "deleted", getErr: ErrTemplateNotFound, want: ErrTemplateNotFound},
		{name: "disabled", row: notificationmodel.DingTalkTemplate{ID: 1, Version: 3}, want: ErrTemplateDisabled},
		{name: "changed", row: notificationmodel.DingTalkTemplate{ID: 1, Version: 4, Enabled: true}, want: ErrTemplateVersionConflict},
		{name: "unpublished", row: notificationmodel.DingTalkTemplate{ID: 1, Version: 3, Enabled: true, Name: "流程", Title: "默认标题", Content: "正文", MessageType: "oa", TargetKind: "workflow_start", DefinitionID: 17}, workflowErr: ErrTemplateWorkflowUnavailable, want: ErrTemplateWorkflowUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &templateStoreFake{row: tc.row, getErr: tc.getErr, workflowErr: tc.workflowErr}
			store.resolution = RecipientResolution{UserIDs: []uint{4}}
			delivery := &fakeDingTalkDelivery{}
			_, err := NewServiceWithDingTalk(store, delivery).SendDingTalk(context.Background(), SendInput{Title: "本次标题", Content: "本次内容", Scope: ScopeAll, SourceType: SourceAdminManualDingTalk, SourceID: "test", TemplateID: 1, TemplateVersion: 3})
			if !errors.Is(err, tc.want) || delivery.batch.Title != "" || store.rule.Scope != "" {
				t.Fatalf("error=%v delivery=%#v rule=%#v", err, delivery.batch, store.rule)
			}
		})
	}
}

func TestTemplateSendSnapshotsOnceAndKeepsTemporaryContent(t *testing.T) {
	store := &templateStoreFake{row: notificationmodel.DingTalkTemplate{ID: 1, Version: 3, Enabled: true, Name: "模板", Title: "默认标题", Content: "默认正文", MessageType: "link", TargetKind: "url", URL: "https://example.test/path"}}
	store.resolution = RecipientResolution{UserIDs: []uint{4, 9}}
	delivery := &fakeDingTalkDelivery{}
	_, err := NewServiceWithDingTalk(store, delivery).SendDingTalk(context.Background(), SendInput{Title: "本次标题", Content: "本次内容", Scope: ScopeAll, SourceType: SourceAdminManualDingTalk, SourceID: "test", TemplateID: 1, TemplateVersion: 3})
	if err != nil {
		t.Fatal(err)
	}
	if store.getCalls != 1 || delivery.batch.Template == nil || delivery.batch.Template.Target.URL != "https://example.test/path" || delivery.batch.Title != "本次标题" || delivery.batch.Content != "本次内容" {
		t.Fatalf("batch=%#v reads=%d", delivery.batch, store.getCalls)
	}
}

func TestTemplateQueryAndVersionBounds(t *testing.T) {
	store := &templateStoreFake{}
	service := NewService(store)
	for _, page := range []int{-1, 10001} {
		if _, err := service.ListDingTalkTemplates(context.Background(), DingTalkTemplateQuery{Page: page}); !errors.Is(err, ErrTemplateInvalid) {
			t.Fatalf("page %d error=%v", page, err)
		}
	}
	if _, err := service.ListDingTalkTemplates(context.Background(), DingTalkTemplateQuery{PageSize: 101}); !errors.Is(err, ErrTemplateInvalid) {
		t.Fatalf("size error=%v", err)
	}
	result, err := service.DingTalkTemplateOptions(context.Background(), DingTalkTemplateQuery{})
	if err != nil || result.List == nil || !store.enabledOnly || store.query.Page != 1 || store.query.PageSize != 20 {
		t.Fatalf("query=%#v result=%#v error=%v", store.query, result, err)
	}
	if _, err := service.DingTalkTemplateOptions(context.Background(), DingTalkTemplateQuery{TargetKind: "workflow_start"}); err != nil || store.query.TargetKind != "workflow_start" {
		t.Fatalf("target kind query=%#v error=%v", store.query, err)
	}
	if _, err := service.DingTalkTemplateOptions(context.Background(), DingTalkTemplateQuery{TargetKind: "unknown"}); !errors.Is(err, ErrTemplateInvalid) {
		t.Fatalf("invalid target kind error=%v", err)
	}
	if _, err := service.UpdateDingTalkTemplate(context.Background(), 9, 1, validTemplateInput()); !errors.Is(err, ErrTemplateInvalid) {
		t.Fatalf("version error=%v", err)
	}
}

type templateStoreFake struct {
	fakeStore
	row, saved          notificationmodel.DingTalkTemplate
	workflow            DingTalkTemplateWorkflowOption
	getErr, workflowErr error
	getCalls            int
	query               DingTalkTemplateQuery
	enabledOnly         bool
}

func (f *templateStoreFake) ListTemplates(_ context.Context, q DingTalkTemplateQuery, enabledOnly bool) ([]notificationmodel.DingTalkTemplate, int64, error) {
	f.query = q
	f.enabledOnly = enabledOnly
	return nil, 0, nil
}

func (f *templateStoreFake) GetTemplate(context.Context, uint) (notificationmodel.DingTalkTemplate, error) {
	f.getCalls++
	return f.row, f.getErr
}

func (f *templateStoreFake) CreateTemplate(_ context.Context, row *notificationmodel.DingTalkTemplate) error {
	row.ID = 1
	f.saved = *row
	return nil
}

func (f *templateStoreFake) UpdateTemplate(_ context.Context, row *notificationmodel.DingTalkTemplate, _ int) error {
	f.saved = *row
	return nil
}

func (f *templateStoreFake) DeleteTemplate(context.Context, uint, int) error { return nil }

func (f *templateStoreFake) PublishedTemplateWorkflow(context.Context, uint) (DingTalkTemplateWorkflowOption, error) {
	return f.workflow, f.workflowErr
}

func (f *templateStoreFake) TemplateWorkflowOptions(context.Context, DingTalkTemplateQuery) ([]DingTalkTemplateWorkflowOption, int64, error) {
	return nil, 0, nil
}

func TestTemplateDisableClosedWorkflowPreservesSavedTargetName(t *testing.T) {
	store := &templateStoreFake{row: notificationmodel.DingTalkTemplate{ID: 7, Version: 3, TargetKind: "workflow_start", DefinitionID: 17, DefinitionName: "原流程"}, workflowErr: ErrTemplateWorkflowUnavailable}
	input := validTemplateInput()
	input.Enabled = false
	input.Version = 3
	input.Target = DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17, DefinitionName: "伪造"}
	result, err := NewService(store).UpdateDingTalkTemplate(context.Background(), 9, 7, input)
	if err != nil || result.Version != 4 || result.Enabled || result.Target.DefinitionName != "原流程" {
		t.Fatalf("disable=%#v error=%v", result, err)
	}
	input.Enabled = true
	if _, err := NewService(store).UpdateDingTalkTemplate(context.Background(), 9, 7, input); !errors.Is(err, ErrTemplateWorkflowUnavailable) {
		t.Fatalf("enable=%v", err)
	}
	input.Enabled = false
	input.Target.DefinitionID = 18
	if _, err := NewService(store).UpdateDingTalkTemplate(context.Background(), 9, 7, input); !errors.Is(err, ErrTemplateWorkflowUnavailable) {
		t.Fatalf("new target=%v", err)
	}
}

func TestTemplateSendRequiresPositivePairedReference(t *testing.T) {
	for _, ref := range []struct {
		id      uint
		version int
	}{{7, 0}, {7, -1}, {0, 1}, {7, MaxTemplateVersion + 1}, {uint(MaxTemplateID + 1), 1}} {
		store := &templateStoreFake{}
		delivery := &fakeDingTalkDelivery{}
		_, err := NewServiceWithDingTalk(store, delivery).SendDingTalk(context.Background(), SendInput{Title: "题", Content: "文", Scope: ScopeAll, SourceType: SourceAdminManualDingTalk, SourceID: "test", TemplateID: ref.id, TemplateVersion: ref.version})
		if !errors.Is(err, ErrTemplateInvalid) || store.getCalls != 0 || delivery.batch.Title != "" {
			t.Fatalf("ref=%#v error=%v", ref, err)
		}
	}
}

func TestTemplateAuditMetadataForSaveAndRejectedSend(t *testing.T) {
	store := &templateStoreFake{row: notificationmodel.DingTalkTemplate{ID: 7, Name: "模板名称", Version: 3, Enabled: false}}
	ctx, recorder := auditlog.WithRecorder(context.Background())
	if _, err := NewService(store).CreateDingTalkTemplate(ctx, 9, validTemplateInput()); err != nil {
		t.Fatal(err)
	}
	if details := recorder.Snapshot(); details.TargetType != "dingtalk_notification_template" || details.TargetID != "1" || details.TargetName != "流程提醒" || details.AffectedCount != 1 {
		t.Fatalf("create audit=%#v", details)
	}
	ctx, recorder = auditlog.WithRecorder(context.Background())
	_, err := NewServiceWithDingTalk(store, &fakeDingTalkDelivery{}).SendDingTalk(ctx, SendInput{Title: "题", Content: "文", Scope: ScopeAll, SourceType: SourceAdminManualDingTalk, SourceID: "test", TemplateID: 7, TemplateVersion: 3})
	if !errors.Is(err, ErrTemplateDisabled) {
		t.Fatal(err)
	}
	if details := recorder.Snapshot(); details.RelatedType != "dingtalk_notification_template" || details.RelatedID != "7" || details.TargetName != "模板名称" {
		t.Fatalf("send audit=%#v", details)
	}
}
