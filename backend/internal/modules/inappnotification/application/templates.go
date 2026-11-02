package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/support/auditlog"
)

func (s *Service) templateStore() (TemplateStore, error) {
	if s == nil {
		return nil, ErrTemplateStoreUnavailable
	}
	store, ok := s.store.(TemplateStore)
	if !ok {
		return nil, ErrTemplateStoreUnavailable
	}
	return store, nil
}

func (s *Service) ListDingTalkTemplates(ctx context.Context, query DingTalkTemplateQuery) (DingTalkTemplateList, error) {
	return s.listDingTalkTemplates(ctx, query, false)
}

func (s *Service) DingTalkTemplateOptions(ctx context.Context, query DingTalkTemplateQuery) (DingTalkTemplateList, error) {
	return s.listDingTalkTemplates(ctx, query, true)
}

func (s *Service) listDingTalkTemplates(ctx context.Context, query DingTalkTemplateQuery, enabledOnly bool) (DingTalkTemplateList, error) {
	query, err := NormalizeDingTalkTemplateQuery(query)
	if err != nil {
		return DingTalkTemplateList{}, err
	}
	store, err := s.templateStore()
	if err != nil {
		return DingTalkTemplateList{}, err
	}
	rows, total, err := store.ListTemplates(ctx, query, enabledOnly)
	if err != nil {
		return DingTalkTemplateList{}, err
	}
	result := DingTalkTemplateList{List: make([]DingTalkSendTemplate, 0, len(rows)), Total: total, Page: query.Page, PageSize: query.PageSize}
	for _, row := range rows {
		result.List = append(result.List, templateDTO(row))
	}
	return result, nil
}

func (s *Service) DingTalkTemplateWorkflowOptions(ctx context.Context, query DingTalkTemplateQuery) (DingTalkTemplateWorkflowList, error) {
	query, err := NormalizeDingTalkTemplateQuery(query)
	if err != nil {
		return DingTalkTemplateWorkflowList{}, err
	}
	store, err := s.templateStore()
	if err != nil {
		return DingTalkTemplateWorkflowList{}, err
	}
	rows, total, err := store.TemplateWorkflowOptions(ctx, query)
	if err != nil {
		return DingTalkTemplateWorkflowList{}, err
	}
	if rows == nil {
		rows = []DingTalkTemplateWorkflowOption{}
	}
	return DingTalkTemplateWorkflowList{List: rows, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *Service) ResolveScheduledWorkflowStartTemplate(ctx context.Context, id uint) (DingTalkSendTemplate, error) {
	if id == 0 || uint64(id) > MaxTemplateID {
		return DingTalkSendTemplate{}, invalidTemplate("请选择有效的流程开始通知模板")
	}
	store, err := s.templateStore()
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	row, err := store.GetTemplate(ctx, id)
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	if !row.Enabled {
		return DingTalkSendTemplate{}, ErrTemplateDisabled
	}
	if row.TargetKind != "workflow_start" || row.DefinitionID == 0 {
		return DingTalkSendTemplate{}, invalidTemplate("定时任务只能选择跳转目标为发起流程的启用模板")
	}
	resolved, err := s.resolveSendTemplate(ctx, id, row.Version)
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	if resolved == nil {
		return DingTalkSendTemplate{}, ErrTemplateNotFound
	}
	return *resolved, nil
}

func (s *Service) CreateDingTalkTemplate(ctx context.Context, actorID uint, input DingTalkTemplateInput) (DingTalkSendTemplate, error) {
	if actorID == 0 {
		return DingTalkSendTemplate{}, ErrUnauthenticated
	}
	store, err := s.templateStore()
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	row, err := prepareTemplate(ctx, store, input)
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	row.Version = 1
	row.CreatedBy = actorID
	row.UpdatedBy = actorID
	row.CreatedAt = time.Now().UnixMilli()
	row.UpdatedAt = row.CreatedAt
	if err = store.CreateTemplate(ctx, &row); err != nil {
		return DingTalkSendTemplate{}, err
	}
	auditTemplate(ctx, row)
	return templateDTO(row), nil
}

func (s *Service) UpdateDingTalkTemplate(ctx context.Context, actorID, id uint, input DingTalkTemplateInput) (DingTalkSendTemplate, error) {
	if actorID == 0 {
		return DingTalkSendTemplate{}, ErrUnauthenticated
	}
	if id == 0 || uint64(id) > MaxTemplateID {
		return DingTalkSendTemplate{}, invalidTemplate("模板ID无效")
	}
	if err := validateTemplateVersion(input.Version); err != nil {
		return DingTalkSendTemplate{}, err
	}
	if input.Version == MaxTemplateVersion {
		return DingTalkSendTemplate{}, ErrTemplateVersionConflict
	}
	store, err := s.templateStore()
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	var existing *notificationmodel.DingTalkTemplate
	if !input.Enabled {
		current, err := store.GetTemplate(ctx, id)
		if err != nil {
			return DingTalkSendTemplate{}, err
		}
		if current.Version != input.Version {
			return DingTalkSendTemplate{}, ErrTemplateVersionConflict
		}
		existing = &current
	}
	row, err := prepareTemplateWithExisting(ctx, store, input, existing)
	if err != nil {
		return DingTalkSendTemplate{}, err
	}
	row.ID = id
	row.Version = input.Version + 1
	row.UpdatedBy = actorID
	row.UpdatedAt = time.Now().UnixMilli()
	if err = store.UpdateTemplate(ctx, &row, input.Version); err != nil {
		return DingTalkSendTemplate{}, err
	}
	auditTemplate(ctx, row)
	return templateDTO(row), nil
}

func (s *Service) DeleteDingTalkTemplate(ctx context.Context, actorID, id uint, version int) error {
	if actorID == 0 {
		return ErrUnauthenticated
	}
	if id == 0 || uint64(id) > MaxTemplateID {
		return invalidTemplate("模板ID无效")
	}
	if err := validateTemplateVersion(version); err != nil {
		return err
	}
	store, err := s.templateStore()
	if err != nil {
		return err
	}
	row, err := store.GetTemplate(ctx, id)
	if err != nil {
		return err
	}
	if err = store.DeleteTemplate(ctx, id, version); err != nil {
		return err
	}
	auditTemplate(ctx, row)
	return nil
}

func prepareTemplate(ctx context.Context, store TemplateStore, input DingTalkTemplateInput) (notificationmodel.DingTalkTemplate, error) {
	return prepareTemplateWithExisting(ctx, store, input, nil)
}

func prepareTemplateWithExisting(ctx context.Context, store TemplateStore, input DingTalkTemplateInput, existing *notificationmodel.DingTalkTemplate) (notificationmodel.DingTalkTemplate, error) {
	input, err := normalizeTemplateInput(input)
	if err != nil {
		return notificationmodel.DingTalkTemplate{}, err
	}
	if input.Target.Kind == "workflow_start" {
		option, err := store.PublishedTemplateWorkflow(ctx, input.Target.DefinitionID)
		if err != nil {
			// Allow disabling the existing target after its workflow closes, but never accepting a new unavailable target.
			if !input.Enabled && existing != nil && existing.TargetKind == "workflow_start" && existing.DefinitionID == input.Target.DefinitionID && errors.Is(err, ErrTemplateWorkflowUnavailable) {
				input.Target.DefinitionName = existing.DefinitionName
			} else {
				return notificationmodel.DingTalkTemplate{}, err
			}
		} else {
			input.Target.DefinitionName = option.Name
		}
	}
	return notificationmodel.DingTalkTemplate{Name: input.Name, Enabled: input.Enabled, Title: input.Title, Content: input.Content, MessageType: input.MessageType, TargetKind: input.Target.Kind, DefinitionID: input.Target.DefinitionID, DefinitionName: input.Target.DefinitionName, URL: input.Target.URL, ButtonTitle: input.ButtonTitle, HeadColor: input.HeadColor}, nil
}

func templateDTO(row notificationmodel.DingTalkTemplate) DingTalkSendTemplate {
	return DingTalkSendTemplate{ID: row.ID, Name: row.Name, Enabled: row.Enabled, Title: row.Title, Content: row.Content, MessageType: row.MessageType, Target: DingTalkTemplateTarget{Kind: row.TargetKind, DefinitionID: row.DefinitionID, DefinitionName: row.DefinitionName, URL: row.URL}, ButtonTitle: row.ButtonTitle, HeadColor: row.HeadColor, Version: row.Version, UpdatedAt: row.UpdatedAt}
}

func (s *Service) resolveSendTemplate(ctx context.Context, id uint, version int) (*DingTalkSendTemplate, error) {
	if id == 0 && version == 0 {
		return nil, nil
	}
	if id == 0 || uint64(id) > MaxTemplateID {
		return nil, invalidTemplate("请选择有效模板并同时提供模板版本号")
	}
	if err := validateTemplateVersion(version); err != nil {
		return nil, err
	}
	store, err := s.templateStore()
	if err != nil {
		return nil, err
	}
	auditlog.Update(ctx, func(d *auditlog.Details) {
		d.RelatedType = "dingtalk_notification_template"
		d.RelatedID = strconv.FormatUint(uint64(id), 10)
	})
	row, err := store.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	auditlog.Update(ctx, func(d *auditlog.Details) { d.TargetName = row.Name })
	if !row.Enabled {
		return nil, ErrTemplateDisabled
	}
	if row.Version != version {
		return nil, ErrTemplateVersionConflict
	}
	// Copy and validate once before recipient resolution; all corporations use this snapshot.
	checked, err := prepareTemplate(ctx, store, DingTalkTemplateInput{Name: row.Name, Enabled: row.Enabled, Title: row.Title, Content: row.Content, MessageType: row.MessageType, Target: DingTalkTemplateTarget{Kind: row.TargetKind, DefinitionID: row.DefinitionID, URL: row.URL}, ButtonTitle: row.ButtonTitle, HeadColor: row.HeadColor})
	if err != nil {
		return nil, err
	}
	checked.ID = row.ID
	checked.Version = row.Version
	checked.UpdatedAt = row.UpdatedAt
	result := templateDTO(checked)
	return &result, nil
}

func auditTemplate(ctx context.Context, row notificationmodel.DingTalkTemplate) {
	auditlog.Update(ctx, func(d *auditlog.Details) {
		d.TargetType = "dingtalk_notification_template"
		d.TargetID = strconv.FormatUint(uint64(row.ID), 10)
		d.TargetName = row.Name
		d.AffectedCount = 1
	})
}
