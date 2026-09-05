package application

import (
	"context"
	"errors"
	"sort"
	"strings"
	"unicode/utf8"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) changeInstanceStatus(ctx context.Context, instanceID, actorID, reason string, adminCancel bool) (*workflowdomain.State, error) {
	instanceID = strings.TrimSpace(instanceID)
	actorID = strings.TrimSpace(actorID)
	if instanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	var state *workflowdomain.State
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		loaded, err := store.LoadStateByInstanceForUpdate(ctx, instanceID)
		if err != nil {
			return err
		}
		if adminCancel {
			err = service.engine.Cancel(loaded, actorID, strings.TrimSpace(reason))
		} else {
			err = service.engine.Withdraw(loaded, actorID, strings.TrimSpace(reason))
		}
		if err != nil {
			return err
		}
		state = loaded
		return store.SaveState(ctx, state)
	})
	return state, err
}

func (service *Service) ListInstances(ctx context.Context, query InstanceQuery) (*InstanceList, error) {
	definitionName, err := normalizeDefinitionNameSearch(query.DefinitionName)
	if err != nil {
		return nil, err
	}
	starterName, err := normalizeStarterNameSearch(query.StarterName)
	if err != nil {
		return nil, err
	}
	query.DefinitionName = definitionName
	query.StarterName = starterName
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.store.ListInstances(ctx, query)
}

func (service *Service) ListMyInstances(ctx context.Context, actorID string, query InstanceQuery) (*InstanceList, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	query.Scope = strings.TrimSpace(query.Scope)
	if query.Scope == "" {
		query.Scope = InstanceScopeStarted
	}
	switch query.Scope {
	case InstanceScopeStarted, InstanceScopeHandled, InstanceScopeCopied:
	default:
		return nil, ErrInstanceScopeInvalid
	}
	query.StarterID = ""
	query.ScopeUserID = actorID
	return service.ListInstances(ctx, query)
}

func (service *Service) GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return nil, errors.New("流程实例不能为空")
	}
	detail, err := service.store.GetInstance(ctx, instanceID)
	if err != nil || detail == nil {
		return detail, err
	}
	value := *detail
	value.Nodes = append([]PublishedNode(nil), detail.Nodes...)
	value.History = append([]HistorySummary(nil), detail.History...)
	detail = &value
	applyTaskCreatedActors(detail)
	applyHistoricalCommentNodeIDs(detail)
	if err := service.resolveNodeAssigneeDisplays(ctx, detail.Nodes, workflowdomain.ProcessInstance{
		StarterID:  detail.Instance.StarterID,
		OperatorID: detail.Instance.OperatorID,
	}); err != nil {
		return nil, err
	}
	applyTaskAssigneeDisplays(detail.Nodes, detail.Tasks)
	return detail, nil
}

type historyActor struct {
	id   string
	name string
}

func applyTaskCreatedActors(detail *InstanceDetail) {
	if detail == nil || len(detail.History) == 0 {
		return
	}
	current := historyActor{
		id:   strings.TrimSpace(detail.Instance.OperatorID),
		name: strings.TrimSpace(detail.Instance.OperatorName),
	}
	if current.id == "" {
		current.id = strings.TrimSpace(detail.Instance.StarterID)
		current.name = strings.TrimSpace(detail.Instance.StarterName)
	}
	indexes := make([]int, len(detail.History))
	for index := range detail.History {
		indexes[index] = index
	}
	sort.SliceStable(indexes, func(left, right int) bool {
		return detail.History[indexes[left]].EventTime < detail.History[indexes[right]].EventTime
	})
	for start := 0; start < len(indexes); {
		end := start + 1
		eventTime := detail.History[indexes[start]].EventTime
		for end < len(indexes) && detail.History[indexes[end]].EventTime == eventTime {
			end++
		}
		trigger := historyActor{}
		for _, index := range indexes[start:end] {
			event := detail.History[index]
			if isTaskCreationTrigger(event.EventType) && strings.TrimSpace(event.ActorID) != "" {
				trigger = historyActor{id: strings.TrimSpace(event.ActorID), name: historyActorName(detail, event.ActorID, event.ActorName)}
			}
		}
		actor := current
		if trigger.id != "" {
			actor = trigger
		}
		if actor.id != "" {
			for _, index := range indexes[start:end] {
				if detail.History[index].EventType == string(workflowdomain.HistoryTaskCreated) {
					detail.History[index].ActorID = actor.id
					detail.History[index].ActorName = actor.name
				}
			}
		}
		if trigger.id != "" {
			current = trigger
		}
		start = end
	}
}

func isTaskCreationTrigger(eventType string) bool {
	switch workflowdomain.HistoryEventType(eventType) {
	case workflowdomain.HistoryInstanceStarted,
		workflowdomain.HistoryTaskApproved,
		workflowdomain.HistoryTaskReturned,
		workflowdomain.HistoryTaskSubmitted:
		return true
	default:
		return false
	}
}

func historyActorName(detail *InstanceDetail, actorID, actorName string) string {
	if name := strings.TrimSpace(actorName); name != "" {
		return name
	}
	actorID = strings.TrimSpace(actorID)
	if name := strings.TrimSpace(detail.UserNames[actorID]); name != "" {
		return name
	}
	if actorID == strings.TrimSpace(detail.Instance.OperatorID) {
		return strings.TrimSpace(detail.Instance.OperatorName)
	}
	if actorID == strings.TrimSpace(detail.Instance.StarterID) {
		return strings.TrimSpace(detail.Instance.StarterName)
	}
	return ""
}

func applyHistoricalCommentNodeIDs(detail *InstanceDetail) {
	if detail == nil || len(detail.History) == 0 {
		return
	}
	assigneesByTask := make(map[string]string, len(detail.Tasks))
	for _, task := range detail.Tasks {
		if taskID := strings.TrimSpace(task.ID); taskID != "" {
			assigneesByTask[taskID] = strings.TrimSpace(task.AssigneeID)
		}
	}
	indexes := make([]int, len(detail.History))
	for index := range detail.History {
		indexes[index] = index
	}
	sort.SliceStable(indexes, func(left, right int) bool {
		return detail.History[indexes[left]].EventTime < detail.History[indexes[right]].EventTime
	})
	activeTasks := make(map[string]string)
	for _, index := range indexes {
		event := &detail.History[index]
		taskID := strings.TrimSpace(event.TaskID)
		nodeID := strings.TrimSpace(event.NodeID)
		switch workflowdomain.HistoryEventType(event.EventType) {
		case workflowdomain.HistoryTaskCreated, workflowdomain.HistoryTaskActivated:
			if taskID != "" && nodeID != "" {
				activeTasks[taskID] = nodeID
			}
		case workflowdomain.HistoryTaskApproved,
			workflowdomain.HistoryTaskRejected,
			workflowdomain.HistoryTaskReturned,
			workflowdomain.HistoryTaskSubmitted,
			workflowdomain.HistoryTaskCancelled:
			delete(activeTasks, taskID)
		case workflowdomain.HistoryInstanceCompleted,
			workflowdomain.HistoryInstanceRejected,
			workflowdomain.HistoryInstanceWithdrawn,
			workflowdomain.HistoryInstanceCancelled:
			clear(activeTasks)
		case workflowdomain.HistoryInstanceCommented:
			if nodeID == "" {
				event.NodeID = commentNodeFromActiveTasks(activeTasks, assigneesByTask, event.ActorID)
			}
		}
	}
}

func currentCommentNodeID(detail *InstanceDetail, actorID string) string {
	if detail == nil {
		return ""
	}
	activeTasks := make(map[string]string)
	assigneesByTask := make(map[string]string)
	for _, task := range detail.Tasks {
		if task.Status != string(workflowdomain.TaskStatusPending) && task.Status != string(workflowdomain.TaskStatusWaiting) {
			continue
		}
		taskID := strings.TrimSpace(task.ID)
		nodeID := strings.TrimSpace(task.NodeID)
		if taskID == "" || nodeID == "" {
			continue
		}
		activeTasks[taskID] = nodeID
		assigneesByTask[taskID] = strings.TrimSpace(task.AssigneeID)
	}
	if nodeID := commentNodeFromActiveTasks(activeTasks, assigneesByTask, actorID); nodeID != "" {
		return nodeID
	}
	processingNodes := make(map[string]struct{})
	for _, node := range detail.NodeProgress {
		if node.Status == NodeProgressProcessing {
			if nodeID := strings.TrimSpace(node.NodeID); nodeID != "" {
				processingNodes[nodeID] = struct{}{}
			}
		}
	}
	return uniqueNodeID(processingNodes)
}

func commentNodeFromActiveTasks(activeTasks, assigneesByTask map[string]string, actorID string) string {
	actorID = strings.TrimSpace(actorID)
	actorNodes := make(map[string]struct{})
	activeNodes := make(map[string]struct{})
	for taskID, nodeID := range activeTasks {
		nodeID = strings.TrimSpace(nodeID)
		if nodeID == "" {
			continue
		}
		activeNodes[nodeID] = struct{}{}
		if actorID != "" && strings.TrimSpace(assigneesByTask[taskID]) == actorID {
			actorNodes[nodeID] = struct{}{}
		}
	}
	if nodeID := uniqueNodeID(actorNodes); nodeID != "" {
		return nodeID
	}
	return uniqueNodeID(activeNodes)
}

func uniqueNodeID(nodes map[string]struct{}) string {
	if len(nodes) != 1 {
		return ""
	}
	for nodeID := range nodes {
		return nodeID
	}
	return ""
}

func applyTaskAssigneeDisplays(nodes []PublishedNode, tasks []TaskSummary) {
	namesByNode := make(map[string][]string)
	seenByNode := make(map[string]map[string]struct{})
	for _, task := range tasks {
		nodeID := strings.TrimSpace(task.NodeID)
		name := strings.TrimSpace(task.HandledByName)
		if name == "" {
			name = strings.TrimSpace(task.AssigneeName)
		}
		if nodeID == "" || name == "" {
			continue
		}
		if seenByNode[nodeID] == nil {
			seenByNode[nodeID] = make(map[string]struct{})
		}
		if _, exists := seenByNode[nodeID][name]; exists {
			continue
		}
		seenByNode[nodeID][name] = struct{}{}
		namesByNode[nodeID] = append(namesByNode[nodeID], name)
	}
	for index := range nodes {
		node := &nodes[index]
		if node.Assignee != nil && node.Assignee.Type == workflowcore.AssigneeTypeInitiator {
			continue
		}
		if names := namesByNode[node.ID]; len(names) > 0 {
			node.AssigneeDisplay = strings.Join(names, "、")
		}
	}
}

func (service *Service) GetMyInstance(ctx context.Context, actorID, instanceID string) (*InstanceDetail, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	detail, err := service.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	allowed := detail.Instance.StarterID == actorID
	if !allowed {
		for _, task := range detail.Tasks {
			if task.AssigneeID == actorID {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		allowed, err = service.store.HasParticipant(ctx, instanceID, actorID, string(workflowdomain.ParticipantRoleCC))
		if err != nil {
			return nil, err
		}
	}
	if !allowed {
		return nil, ErrInstanceAccessDenied
	}
	decorateInstanceReminders(detail, actorID, service.currentTime())
	decorateInstanceFormRevision(detail, actorID)
	return detail, nil
}

func (service *Service) CommentInstance(ctx context.Context, request CommentInstanceRequest) error {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Comment = strings.TrimSpace(request.Comment)
	if request.InstanceID == "" {
		return ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return ErrActorRequired
	}
	images, err := normalizeWorkflowImages(request.Images)
	if err != nil {
		return err
	}
	request.Images = images
	if request.Comment == "" && len(request.Images) == 0 {
		return ErrInstanceCommentRequired
	}
	if utf8.RuneCountInString(request.Comment) > 500 {
		return ErrInstanceCommentTooLong
	}
	if service == nil || service.store == nil || service.ids == nil {
		return errors.New("工作流应用服务未初始化")
	}
	detail, err := service.GetMyInstance(ctx, request.ActorID, request.InstanceID)
	if err != nil {
		return err
	}
	notification, err := service.normalizeCommentNotification(ctx, detail, request.ActorID, request.Notification)
	if err != nil {
		return err
	}
	historyID := service.ids.NewID("history")
	nodeID := currentCommentNodeID(detail, request.ActorID)
	eventTime := service.currentTime().UnixMilli()
	event := workflowdomain.HistoryEvent{
		ID:      historyID,
		Type:    workflowdomain.HistoryInstanceCommented,
		NodeID:  nodeID,
		ActorID: request.ActorID,
		Message: request.Comment,
		Images:  request.Images,
	}

	var outboxIDs []string
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		if err := store.AppendInstanceHistory(ctx, request.InstanceID, event, eventTime); err != nil {
			return err
		}
		if notification == nil {
			return nil
		}
		effects := commentNotificationEffects(detail, event, *notification, service.ids)
		outboxIDs, err = store.PersistEffects(ctx, effects)
		return err
	})
	if err != nil {
		return err
	}
	service.dispatchNotifications(ctx, outboxIDs)
	return nil
}

func (service *Service) DeleteMyInstance(ctx context.Context, actorID, instanceID string) error {
	actorID = strings.TrimSpace(actorID)
	instanceID = strings.TrimSpace(instanceID)
	if actorID == "" {
		return ErrActorRequired
	}
	if instanceID == "" {
		return ErrInstanceIDRequired
	}
	if service == nil || service.store == nil {
		return errors.New("工作流应用服务未初始化")
	}
	detail, err := service.store.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(detail.Instance.StarterID) != actorID {
		return ErrInstanceAccessDenied
	}
	switch detail.Instance.Status {
	case string(workflowdomain.InstanceStatusRunning):
		return ErrRunningInstanceCannotDelete
	case string(workflowdomain.InstanceStatusCompleted),
		string(workflowdomain.InstanceStatusRejected),
		string(workflowdomain.InstanceStatusCancelled),
		"withdrawn":
		return service.store.HideStartedInstance(ctx, instanceID, actorID, service.currentTime().UnixMilli())
	default:
		return ErrInstanceDeleteNotAllowed
	}
}

func (service *Service) DeleteInstances(ctx context.Context, actorID string, instanceIDs []string) (int, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return 0, ErrActorRequired
	}
	instanceIDs = uniqueTrimmedStrings(instanceIDs)
	if len(instanceIDs) == 0 {
		return 0, ErrInstanceIDRequired
	}
	if len(instanceIDs) > 100 {
		return 0, ErrInstanceDeleteTooMany
	}
	if service == nil || service.store == nil {
		return 0, errors.New("工作流应用服务未初始化")
	}

	deleted := 0
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		instances, err := store.LoadInstancesForDelete(ctx, instanceIDs)
		if err != nil {
			return err
		}
		if len(instances) != len(instanceIDs) {
			return ErrInstanceDeleteTargetNotFound
		}
		for _, instance := range instances {
			switch instance.Status {
			case string(workflowdomain.InstanceStatusRunning):
				return ErrRunningInstanceCannotDelete
			case string(workflowdomain.InstanceStatusCompleted),
				string(workflowdomain.InstanceStatusRejected),
				string(workflowdomain.InstanceStatusCancelled),
				"withdrawn":
			default:
				return ErrInstanceDeleteNotAllowed
			}
		}
		count, err := store.SoftDeleteInstances(ctx, instanceIDs, actorID, service.currentTime().UnixMilli())
		if err != nil {
			return err
		}
		if count != int64(len(instanceIDs)) {
			return ErrInstanceDeleteTargetNotFound
		}
		deleted = int(count)
		return nil
	})
	return deleted, err
}

func uniqueTrimmedStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
