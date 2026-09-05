package application

import (
	"context"
	"errors"
	"strings"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) StartInstance(ctx context.Context, request StartInstanceRequest) (*workflowdomain.State, error) {
	request.BusinessType = strings.TrimSpace(request.BusinessType)
	request.BusinessKey = strings.TrimSpace(request.BusinessKey)
	request.StarterID = strings.TrimSpace(request.StarterID)
	request.OperatorID = strings.TrimSpace(request.OperatorID)
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if request.StarterID == "" {
		return nil, ErrStarterRequired
	}
	if request.OperatorID == "" {
		return nil, ErrOperatorRequired
	}
	if request.BusinessType == "" || request.BusinessKey == "" {
		return nil, ErrBusinessReferenceRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var state *workflowdomain.State
	var outboxIDs []string
	created := false
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		if request.Idempotent {
			reader, ok := store.(BusinessStateReader)
			if !ok {
				return errors.New("工作流存储不支持业务幂等查询")
			}
			existing, found, err := reader.FindStateByBusiness(ctx, request.BusinessType, request.BusinessKey)
			if err != nil {
				return err
			}
			if found {
				state = existing
				return nil
			}
		}
		definition, publishedVersion, err := store.LoadPublishedDefinition(ctx, request.DefinitionID, request.DefinitionVersion)
		if err != nil {
			return err
		}
		initiator := definitionInitiatorConfig(definition)
		departmentIDs, err := loadInitiatorDepartmentIDs(ctx, store, initiator, request.StarterID)
		if err != nil {
			return err
		}
		if !publishedInitiatorAllows(initiator, request.StarterID, departmentIDs) {
			return ErrStarterNotAllowed
		}
		startedAt := service.currentTime()
		availability := definitionStartAvailabilityConfig(definition)
		if err := startAvailabilityError(workflowcore.EvaluateStartAvailability(availability, startedAt)); err != nil {
			return err
		}
		active, err := store.IsActiveUser(ctx, request.StarterID)
		if err != nil {
			return err
		}
		if !active {
			return ErrStarterInvalid
		}
		if !request.AdminInitiated && request.OperatorID != request.StarterID {
			return ErrStarterAccessDenied
		}
		if request.AdminInitiated {
			allowed, err := store.CanOperatorStartFor(ctx, request.OperatorID, request.StarterID)
			if err != nil {
				return err
			}
			if !allowed {
				return ErrStarterAccessDenied
			}
		}
		if err := workflowcore.ValidateStartFormData(definition, request.FormData); err != nil {
			return err
		}
		startLimit := definitionStartLimitConfig(definition)
		if startLimit.Mode == workflowcore.StartLimitModeLimited {
			window, ok := workflowcore.ResolveStartLimitWindow(&startLimit, availability, startedAt)
			if !ok {
				return errors.New("流程发起次数限制配置无效")
			}
			_, allowed, err := store.ConsumeStartQuota(ctx, request.DefinitionID, request.StarterID, window, startLimit.MaxCount)
			if err != nil {
				return err
			}
			if !allowed {
				return ErrStartLimitExceeded
			}
		}
		state, err = service.engine.Start(ctx, definition, workflowdomain.StartRequest{
			DefinitionID:      request.DefinitionID,
			DefinitionVersion: publishedVersion,
			BusinessType:      request.BusinessType,
			BusinessKey:       request.BusinessKey,
			StarterID:         request.StarterID,
			OperatorID:        request.OperatorID,
			StartTime:         startedAt.UnixMilli(),
			Variables:         request.Variables,
			FormData:          request.FormData,
		})
		if err != nil {
			return err
		}
		if err := store.CreateState(ctx, state); err != nil {
			return err
		}
		if request.ClearStartDraft {
			if err := store.DeleteStartDraft(ctx, request.DefinitionID, request.StarterID); err != nil {
				return err
			}
		}
		created = true
		outboxIDs, err = store.PersistEffects(ctx, state)
		return err
	})
	if err != nil && request.Idempotent {
		if reader, ok := service.store.(BusinessStateReader); ok {
			if existing, found, lookupErr := reader.FindStateByBusiness(ctx, request.BusinessType, request.BusinessKey); lookupErr == nil && found {
				return existing, nil
			}
		}
	}
	if err == nil && created {
		service.dispatchNotifications(ctx, outboxIDs)
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceStarted, InstanceID: state.Instance.ID, ActorID: request.OperatorID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, err
}

func (service *Service) GetStartDraft(ctx context.Context, definitionID uint, starterID string) (*StartDraft, error) {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if starterID == "" {
		return nil, ErrStarterRequired
	}
	if _, err := service.GetPublishedDefinitionForStarter(ctx, definitionID, starterID); err != nil {
		return nil, err
	}
	store, err := service.startDraftStore()
	if err != nil {
		return nil, err
	}
	return store.GetStartDraft(ctx, definitionID, starterID)
}

func (service *Service) SaveStartDraft(ctx context.Context, request SaveStartDraftRequest) (*StartDraft, error) {
	request.StarterID = strings.TrimSpace(request.StarterID)
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if request.StarterID == "" {
		return nil, ErrStarterRequired
	}
	definition, err := service.GetPublishedDefinitionForStarter(ctx, request.DefinitionID, request.StarterID)
	if err != nil {
		return nil, err
	}
	if request.DefinitionVersion > 0 && request.DefinitionVersion != definition.Version {
		return nil, ErrDefinitionVersionChanged
	}
	if request.FormData == nil {
		request.FormData = make(map[string]interface{})
	}
	if err := workflowcore.ValidateFormData(definition.Form, request.FormData, true); err != nil {
		return nil, err
	}
	store, err := service.startDraftStore()
	if err != nil {
		return nil, err
	}
	return store.SaveStartDraft(ctx, StartDraft{
		DefinitionID: request.DefinitionID, DefinitionVersion: definition.Version,
		StarterID: request.StarterID, FormData: request.FormData,
		UpdatedAt: service.currentTime().UnixMilli(),
	})
}

func (service *Service) DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 {
		return ErrDefinitionRequired
	}
	if starterID == "" {
		return ErrStarterRequired
	}
	store, err := service.startDraftStore()
	if err != nil {
		return err
	}
	return store.DeleteStartDraft(ctx, definitionID, starterID)
}

func (service *Service) startDraftStore() (StartDraftStore, error) {
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	store, ok := service.store.(StartDraftStore)
	if !ok {
		return nil, ErrDraftStoreUnavailable
	}
	return store, nil
}

func definitionInitiatorConfig(definition workflowcore.Definition) workflowcore.InitiatorConfig {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart && node.Initiator != nil {
			return *node.Initiator
		}
	}
	return workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}
}
