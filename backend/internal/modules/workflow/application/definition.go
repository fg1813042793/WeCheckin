package application

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error) {
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	definitions, err := service.store.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	return decoratePublishedDefinitionsAvailability(definitions, service.currentTime()), nil
}

func (service *Service) ListPublishedDefinitionCategories(ctx context.Context) ([]string, error) {
	definitions, err := service.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(definitions))
	categories := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		category := strings.TrimSpace(definition.Category)
		if category == "" {
			continue
		}
		if _, exists := seen[category]; exists {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories, nil
}

func (service *Service) ListPublishedDefinitionsForStarter(ctx context.Context, starterID string) ([]PublishedDefinition, error) {
	definitions, err := service.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	var departmentIDs []uint
	for _, definition := range definitions {
		if initiatorNeedsDepartments(definition.Initiator, starterID) {
			departmentIDs, err = service.store.UserDepartmentIDs(ctx, starterID)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	result := make([]PublishedDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if publishedInitiatorAllows(definition.Initiator, starterID, departmentIDs) {
			result = append(result, definition)
		}
	}
	if err := service.decorateStartLimitStatuses(ctx, result, starterID); err != nil {
		return nil, err
	}
	return result, nil
}

func (service *Service) GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error) {
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	definition, err := service.store.GetPublishedDefinition(ctx, definitionID)
	if err != nil || definition == nil {
		return definition, err
	}
	definition.AvailabilityStatus = workflowcore.EvaluateStartAvailability(&definition.Availability, service.currentTime())
	return definition, nil
}

func (service *Service) GetPublishedDefinitionForStarter(ctx context.Context, definitionID uint, starterID string) (*PublishedDefinition, error) {
	definition, err := service.GetPublishedDefinition(ctx, definitionID)
	if err != nil {
		return nil, err
	}
	departmentIDs, err := loadInitiatorDepartmentIDs(ctx, service.store, definition.Initiator, starterID)
	if err != nil {
		return nil, err
	}
	if !publishedInitiatorAllows(definition.Initiator, starterID, departmentIDs) {
		return nil, ErrStarterNotAllowed
	}
	value := *definition
	value.Nodes = append([]PublishedNode(nil), definition.Nodes...)
	definition = &value
	if err := service.resolvePublishedAssigneeDisplays(ctx, definition, starterID); err != nil {
		return nil, err
	}
	definitions := []PublishedDefinition{*definition}
	if err := service.decorateStartLimitStatuses(ctx, definitions, starterID); err != nil {
		return nil, err
	}
	*definition = definitions[0]
	return definition, nil
}

func (service *Service) resolvePublishedAssigneeDisplays(ctx context.Context, definition *PublishedDefinition, starterID string) error {
	if definition == nil {
		return nil
	}
	instance := workflowdomain.ProcessInstance{StarterID: starterID, OperatorID: starterID}
	return service.resolveNodeAssigneeDisplays(ctx, definition.Nodes, instance)
}

func (service *Service) resolveNodeAssigneeDisplays(ctx context.Context, nodes []PublishedNode, instance workflowdomain.ProcessInstance) error {
	resolver, ok := service.resolver.(assigneeDisplayResolver)
	if !ok {
		return nil
	}
	for index := range nodes {
		node := &nodes[index]
		if node.Assignee == nil {
			continue
		}
		names, err := resolver.ResolveDisplayNames(ctx, workflowdomain.AssigneeRequest{
			Instance: instance,
			Node: workflowcore.Node{
				ID: node.ID, Type: node.Type, Name: node.Name, ApprovalMode: node.ApprovalMode,
				GatewayMode: node.GatewayMode, Assignee: node.Assignee,
			},
		})
		if err != nil {
			return err
		}
		resolved := make([]string, 0, len(names))
		for _, name := range names {
			if value := strings.TrimSpace(name); value != "" {
				resolved = append(resolved, value)
			}
		}
		if len(resolved) > 0 {
			node.AssigneeDisplay = strings.Join(resolved, "、")
		}
	}
	return nil
}

func publishedInitiatorAllows(initiator workflowcore.InitiatorConfig, starterID string, departmentIDs []uint) bool {
	if initiatorContainsUser(initiator.ExcludedUserIDs, starterID) {
		return false
	}
	if initiator.Scope != workflowcore.InitiatorScopeSpecified {
		return true
	}
	if initiatorContainsUser(initiator.UserIDs, starterID) {
		return true
	}
	allowedDepartments := make(map[uint]struct{}, len(initiator.DepartmentIDs))
	for _, departmentID := range initiator.DepartmentIDs {
		allowedDepartments[departmentID] = struct{}{}
	}
	for _, departmentID := range departmentIDs {
		if _, allowed := allowedDepartments[departmentID]; allowed {
			return true
		}
	}
	return false
}

func loadInitiatorDepartmentIDs(
	ctx context.Context,
	reader UserDepartmentReader,
	initiator workflowcore.InitiatorConfig,
	starterID string,
) ([]uint, error) {
	if !initiatorNeedsDepartments(initiator, starterID) {
		return nil, nil
	}
	return reader.UserDepartmentIDs(ctx, starterID)
}

func initiatorNeedsDepartments(initiator workflowcore.InitiatorConfig, starterID string) bool {
	if initiator.Scope != workflowcore.InitiatorScopeSpecified || len(initiator.DepartmentIDs) == 0 {
		return false
	}
	if initiatorContainsUser(initiator.ExcludedUserIDs, starterID) {
		return false
	}
	return !initiatorContainsUser(initiator.UserIDs, starterID)
}

func initiatorContainsUser(userIDs []uint, starterID string) bool {
	starterID = strings.TrimSpace(starterID)
	for _, userID := range userIDs {
		if starterID == strconv.FormatUint(uint64(userID), 10) {
			return true
		}
	}
	return false
}

func definitionStartAvailabilityConfig(definition workflowcore.Definition) *workflowcore.StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return definition.Nodes[index].Availability
		}
	}
	return nil
}

func definitionStartLimitConfig(definition workflowcore.Definition) workflowcore.StartLimitConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartLimit(definition.Nodes[index].StartLimit)
		}
	}
	return workflowcore.DefaultStartLimit()
}

func (service *Service) decorateStartLimitStatuses(ctx context.Context, definitions []PublishedDefinition, starterID string) error {
	now := service.currentTime()
	for index := range definitions {
		definition := &definitions[index]
		definition.StartLimitStatus = StartLimitStatus{Allowed: true}
		if definition.StartLimit.Mode != workflowcore.StartLimitModeLimited {
			continue
		}
		window, ok := workflowcore.ResolveStartLimitWindow(&definition.StartLimit, &definition.Availability, now)
		if !ok {
			return errors.New("流程发起次数限制配置无效")
		}
		usedCount, err := service.store.CountStartQuotaUsage(ctx, definition.ID, starterID, window)
		if err != nil {
			return err
		}
		remainingCount := definition.StartLimit.MaxCount - usedCount
		if remainingCount < 0 {
			remainingCount = 0
		}
		definition.StartLimitStatus = StartLimitStatus{
			Allowed: usedCount < definition.StartLimit.MaxCount, UsedCount: usedCount,
			RemainingCount: remainingCount, ResetsAt: window.EndsAt,
		}
	}
	return nil
}

func decoratePublishedDefinitionsAvailability(definitions []PublishedDefinition, now time.Time) []PublishedDefinition {
	result := append([]PublishedDefinition(nil), definitions...)
	for index := range result {
		result[index].AvailabilityStatus = workflowcore.EvaluateStartAvailability(&result[index].Availability, now)
	}
	return result
}

func startAvailabilityError(state string) error {
	switch state {
	case workflowcore.StartAvailabilityStateAvailable:
		return nil
	case workflowcore.StartAvailabilityStateNotStarted:
		return ErrStartNotYetAvailable
	case workflowcore.StartAvailabilityStateExpired:
		return ErrStartAvailabilityExpired
	default:
		return ErrStartOutsideAvailability
	}
}

func (service *Service) currentTime() time.Time {
	if service != nil && service.now != nil {
		return service.now()
	}
	return time.Now()
}
