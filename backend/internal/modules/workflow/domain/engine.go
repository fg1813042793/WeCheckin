package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"wecheckin/backend/internal/workflowcore"
)

const maxAutomaticTransitions = 1000
const timerDueVariablePrefix = "__workflow_timer_due."

var (
	ErrTaskNotFound            = errors.New("工作流任务不存在")
	ErrTaskAlreadyHandled      = errors.New("工作流任务已处理")
	ErrTaskActorMismatch       = errors.New("当前用户不是任务审批人")
	ErrInvalidTaskAction       = errors.New("工作流任务操作无效")
	ErrInstanceNotRunning      = errors.New("工作流实例已结束")
	ErrInstanceStarterMismatch = errors.New("当前用户不是流程发起人")
	ErrInstanceAlreadyHandled  = errors.New("流程任务已被处理，不能撤回")
	ErrNoMatchingBranch        = errors.New("排他网关没有命中条件且未配置默认分支")
	ErrTransitionLimit         = errors.New("流程自动流转次数超过限制")
	ErrAssigneeUnavailable     = errors.New("流程节点未解析到处理人")
)

type Engine struct {
	resolver AssigneeResolver
	ids      IDGenerator
}

func NewEngine(resolver AssigneeResolver, ids IDGenerator) *Engine {
	return &Engine{resolver: resolver, ids: ids}
}

func (engine *Engine) Start(definition workflowcore.Definition, request StartRequest) (*State, error) {
	if validationErrors := workflowcore.ValidateDefinition(definition); len(validationErrors) > 0 {
		return nil, workflowcore.ValidationErrors(validationErrors)
	}
	if engine.resolver == nil || engine.ids == nil {
		return nil, errors.New("工作流引擎依赖未配置")
	}
	if strings.TrimSpace(request.OperatorID) == "" {
		request.OperatorID = request.StarterID
	}
	startNode, ok := findStartNode(definition)
	if !ok {
		return nil, errors.New("流程缺少开始节点")
	}
	state := &State{
		Instance: ProcessInstance{
			ID: engine.ids.NewID("instance"), DefinitionID: request.DefinitionID,
			DefinitionVersion: request.DefinitionVersion, DefinitionKey: definition.Key,
			BusinessType: request.BusinessType, BusinessKey: request.BusinessKey,
			StarterID: request.StarterID, OperatorID: request.OperatorID, Status: InstanceStatusRunning,
		},
		Variables: cloneVariables(request.Variables),
		FormData:  cloneVariables(request.FormData),
	}
	state.Tokens = append(state.Tokens, Token{ID: engine.ids.NewID("token"), NodeID: startNode.ID, Status: TokenStatusActive})
	engine.addHistory(state, HistoryInstanceStarted, startNode.ID, "", request.OperatorID, "流程实例已启动")
	if err := engine.advanceToken(definition, state, len(state.Tokens)-1, 0); err != nil {
		return nil, err
	}
	return state, nil
}

func (engine *Engine) Complete(definition workflowcore.Definition, state *State, request CompleteRequest) error {
	if state == nil || state.Instance.Status != InstanceStatusRunning {
		return ErrInstanceNotRunning
	}
	taskIndex := findTaskIndex(state.Tasks, request.TaskID)
	if taskIndex < 0 {
		return ErrTaskNotFound
	}
	task := &state.Tasks[taskIndex]
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyHandled
	}
	if strings.TrimSpace(task.AssigneeID) != strings.TrimSpace(request.ActorID) {
		return ErrTaskActorMismatch
	}
	node, ok := findNode(definition, task.NodeID)
	if !ok {
		return fmt.Errorf("流程节点 %s 不存在", task.NodeID)
	}
	if node.Type == workflowcore.NodeTypeHandle {
		if request.Action != TaskActionSubmit {
			return ErrInvalidTaskAction
		}
	} else if node.Type != workflowcore.NodeTypeApproval || (request.Action != TaskActionApprove && request.Action != TaskActionReject) {
		return ErrInvalidTaskAction
	}
	for key, value := range request.Variables {
		state.Variables[key] = value
	}
	for key, value := range request.FormData {
		state.FormData[key] = value
	}
	task.Action = request.Action
	task.Comment = request.Comment
	if node.Type == workflowcore.NodeTypeHandle {
		task.Status = TaskStatusCompleted
		engine.addHistory(state, HistoryTaskSubmitted, task.NodeID, task.ID, request.ActorID, request.Comment)
		tokenIndex := findTokenIndex(state.Tokens, task.TokenID)
		if tokenIndex < 0 {
			return errors.New("办理任务对应的流程令牌不存在")
		}
		state.Tokens[tokenIndex].Status = TokenStatusActive
		return engine.leaveNode(definition, state, tokenIndex, 0)
	}
	if request.Action == TaskActionReject {
		task.Status = TaskStatusRejected
		engine.addHistory(state, HistoryTaskRejected, task.NodeID, task.ID, request.ActorID, request.Comment)
		engine.rejectInstance(state, request.ActorID)
		return nil
	}

	task.Status = TaskStatusApproved
	engine.addHistory(state, HistoryTaskApproved, task.NodeID, task.ID, request.ActorID, request.Comment)
	shouldAdvance := false
	switch task.ApprovalMode {
	case workflowcore.ApprovalModeSingle:
		shouldAdvance = true
	case workflowcore.ApprovalModeSequential:
		if next := nextWaitingTask(state.Tasks, task.GroupKey); next >= 0 {
			state.Tasks[next].Status = TaskStatusPending
			engine.addHistory(state, HistoryTaskActivated, state.Tasks[next].NodeID, state.Tasks[next].ID, state.Tasks[next].AssigneeID, "顺序审批任务已激活")
			engine.addTaskNotificationIntent(state, definition.Name, node, state.Tasks[next])
		} else {
			shouldAdvance = true
		}
	case workflowcore.ApprovalModeParallel:
		shouldAdvance = allGroupTasksApproved(state.Tasks, task.GroupKey)
	case workflowcore.ApprovalModeCountersign:
		shouldAdvance = countersignThresholdReached(state.Tasks, task.GroupKey, task.CompletionRate)
		if shouldAdvance {
			engine.cancelGroupTasks(state, task.GroupKey, request.ActorID)
		}
	default:
		return fmt.Errorf("不支持的审批方式 %s", task.ApprovalMode)
	}
	if !shouldAdvance {
		return nil
	}
	tokenIndex := findTokenIndex(state.Tokens, task.TokenID)
	if tokenIndex < 0 {
		return errors.New("审批任务对应的流程令牌不存在")
	}
	state.Tokens[tokenIndex].Status = TokenStatusActive
	return engine.leaveNode(definition, state, tokenIndex, 0)
}

func (engine *Engine) Withdraw(state *State, actorID, reason string) error {
	if state == nil || state.Instance.Status != InstanceStatusRunning {
		return ErrInstanceNotRunning
	}
	if strings.TrimSpace(state.Instance.StarterID) != strings.TrimSpace(actorID) {
		return ErrInstanceStarterMismatch
	}
	for _, task := range state.Tasks {
		if task.Status == TaskStatusApproved || task.Status == TaskStatusRejected {
			return ErrInstanceAlreadyHandled
		}
	}
	engine.terminateInstance(state, actorID, reason, HistoryInstanceWithdrawn)
	return nil
}

func (engine *Engine) Cancel(state *State, actorID, reason string) error {
	if state == nil || state.Instance.Status != InstanceStatusRunning {
		return ErrInstanceNotRunning
	}
	engine.terminateInstance(state, actorID, reason, HistoryInstanceCancelled)
	return nil
}

func (engine *Engine) terminateInstance(state *State, actorID, reason string, eventType HistoryEventType) {
	state.Instance.Status = InstanceStatusCancelled
	for index := range state.Tasks {
		if state.Tasks[index].Status == TaskStatusPending || state.Tasks[index].Status == TaskStatusWaiting {
			state.Tasks[index].Status = TaskStatusCancelled
			engine.addHistory(state, HistoryTaskCancelled, state.Tasks[index].NodeID, state.Tasks[index].ID, actorID, reason)
		}
	}
	for index := range state.Tokens {
		if state.Tokens[index].Status == TokenStatusActive || state.Tokens[index].Status == TokenStatusWaiting {
			state.Tokens[index].Status = TokenStatusCancelled
		}
	}
	engine.addHistory(state, eventType, "", "", actorID, reason)
}

func (engine *Engine) advanceToken(definition workflowcore.Definition, state *State, tokenIndex, depth int) error {
	if depth > maxAutomaticTransitions {
		return ErrTransitionLimit
	}
	token := &state.Tokens[tokenIndex]
	node, ok := findNode(definition, token.NodeID)
	if !ok {
		return fmt.Errorf("流程节点 %s 不存在", token.NodeID)
	}
	switch node.Type {
	case workflowcore.NodeTypeStart:
		return engine.leaveNode(definition, state, tokenIndex, depth+1)
	case workflowcore.NodeTypeApproval:
		return engine.createApprovalTasks(definition.Name, state, tokenIndex, node)
	case workflowcore.NodeTypeHandle:
		return engine.createHandleTask(definition.Name, state, tokenIndex, node)
	case workflowcore.NodeTypeCC:
		return engine.executeCC(definition, state, tokenIndex, node, depth+1)
	case workflowcore.NodeTypeNotify:
		return engine.executeNotify(definition, state, tokenIndex, node, depth+1)
	case workflowcore.NodeTypeAutomation:
		return engine.executeAutomation(definition, state, tokenIndex, node, depth+1)
	case workflowcore.NodeTypeTimer:
		return engine.waitForTimer(state, tokenIndex, node)
	case workflowcore.NodeTypeExclusive:
		if node.GatewayMode == workflowcore.GatewayModeJoin {
			return engine.leaveNode(definition, state, tokenIndex, depth+1)
		}
		edge, err := selectExclusiveEdge(definition, node.ID, state.Variables, state.FormData)
		if err != nil {
			return err
		}
		return engine.followEdge(definition, state, tokenIndex, edge, depth+1)
	case workflowcore.NodeTypeParallel:
		if node.GatewayMode == workflowcore.GatewayModeJoin {
			return engine.joinParallel(definition, state, tokenIndex, node, depth+1)
		}
		return engine.splitParallel(definition, state, tokenIndex, node, depth+1)
	case workflowcore.NodeTypeEnd:
		token.Status = TokenStatusCompleted
		engine.completeInstanceIfIdle(state)
		return nil
	default:
		return fmt.Errorf("不支持的流程节点类型 %s", node.Type)
	}
}

func (engine *Engine) createHandleTask(workflowName string, state *State, tokenIndex int, node workflowcore.Node) error {
	assignees, err := engine.resolver.Resolve(AssigneeRequest{Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables)})
	if err != nil {
		return err
	}
	assignees = uniqueNonEmpty(assignees)
	if len(assignees) == 0 {
		return ErrAssigneeUnavailable
	}
	task := Task{
		ID: engine.ids.NewID("task"), TokenID: state.Tokens[tokenIndex].ID,
		NodeID: node.ID, NodeName: node.Name, GroupKey: engine.ids.NewID("task-group"),
		AssigneeID: assignees[0], ApprovalMode: workflowcore.ApprovalModeSingle,
		CompletionRate: 100, Sequence: 1, Total: 1, Status: TaskStatusPending,
	}
	state.Tasks = append(state.Tasks, task)
	engine.addHistory(state, HistoryTaskCreated, task.NodeID, task.ID, task.AssigneeID, "办理任务已创建")
	engine.addTaskNotificationIntent(state, workflowName, node, task)
	state.Tokens[tokenIndex].Status = TokenStatusWaiting
	return nil
}

func (engine *Engine) executeCC(definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	assignees, err := engine.resolver.Resolve(AssigneeRequest{Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables)})
	if err != nil {
		return err
	}
	assignees = uniqueNonEmpty(assignees)
	if len(assignees) == 0 {
		return ErrAssigneeUnavailable
	}
	for _, assigneeID := range assignees {
		state.Participants = append(state.Participants, Participant{
			ID: engine.ids.NewID("participant"), UserID: assigneeID, Role: ParticipantRoleCC, NodeID: node.ID,
		})
		engine.addHistory(state, HistoryNodeCC, node.ID, "", assigneeID, "流程节点已抄送")
		engine.addNotificationIntent(state, definition.Name, node, "", assigneeID, NotificationKindNodeCC)
	}
	return engine.leaveNode(definition, state, tokenIndex, depth+1)
}

func (engine *Engine) executeNotify(definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	assignees, err := engine.resolver.Resolve(AssigneeRequest{Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables)})
	if err != nil {
		return err
	}
	assignees = uniqueNonEmpty(assignees)
	if len(assignees) == 0 {
		return ErrAssigneeUnavailable
	}
	for _, assigneeID := range assignees {
		engine.addHistory(state, HistoryNodeNotify, node.ID, "", assigneeID, "流程通知节点已触发")
		engine.addNotificationIntent(state, definition.Name, node, "", assigneeID, NotificationKindNodeNotify)
	}
	return engine.leaveNode(definition, state, tokenIndex, depth+1)
}

func (engine *Engine) executeAutomation(definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	for key, value := range node.Automation.Variables {
		state.Variables[key] = value
	}
	engine.addHistory(state, HistoryNodeAutomated, node.ID, "", "", "自动动作已写入流程变量")
	return engine.leaveNode(definition, state, tokenIndex, depth+1)
}

func (engine *Engine) waitForTimer(state *State, tokenIndex int, node workflowcore.Node) error {
	key := timerDueVariableKey(state.Tokens[tokenIndex].ID)
	state.Variables[key] = time.Now().Unix() + node.Timer.DelaySeconds
	state.Tokens[tokenIndex].Status = TokenStatusWaiting
	engine.addHistory(state, HistoryTimerWaiting, node.ID, "", "", "流程进入定时等待")
	return nil
}

func (engine *Engine) ResumeTimers(definition workflowcore.Definition, state *State, now int64) (int, error) {
	if state == nil {
		return 0, ErrInstanceNotRunning
	}
	if state.Instance.Status == InstanceStatusCompleted {
		return 0, nil
	}
	if state.Instance.Status != InstanceStatusRunning {
		return 0, ErrInstanceNotRunning
	}
	advanced := 0
	initialTokenCount := len(state.Tokens)
	for index := 0; index < initialTokenCount; index++ {
		token := state.Tokens[index]
		if token.Status != TokenStatusWaiting {
			continue
		}
		node, ok := findNode(definition, token.NodeID)
		if !ok || node.Type != workflowcore.NodeTypeTimer {
			continue
		}
		key := timerDueVariableKey(token.ID)
		dueAt, ok := timerDueAt(state.Variables[key])
		if !ok {
			return advanced, fmt.Errorf("定时节点 %s 缺少有效到期时间", node.ID)
		}
		if dueAt > now {
			continue
		}
		state.Tokens[index].Status = TokenStatusActive
		delete(state.Variables, key)
		engine.addHistory(state, HistoryTimerResumed, node.ID, "", "", "定时等待已到期")
		if err := engine.leaveNode(definition, state, index, 0); err != nil {
			return advanced, err
		}
		advanced++
	}
	return advanced, nil
}

func timerDueVariableKey(tokenID string) string {
	return timerDueVariablePrefix + tokenID
}

func timerDueAt(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), float64(int64(typed)) == typed
	case json.Number:
		result, err := typed.Int64()
		return result, err == nil
	case string:
		result, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return result, err == nil
	default:
		return 0, false
	}
}

func (engine *Engine) leaveNode(definition workflowcore.Definition, state *State, tokenIndex, depth int) error {
	edges := outgoingEdges(definition, state.Tokens[tokenIndex].NodeID)
	if len(edges) != 1 {
		return fmt.Errorf("节点 %s 需要且只能有一条离开连线", state.Tokens[tokenIndex].NodeID)
	}
	return engine.followEdge(definition, state, tokenIndex, edges[0], depth+1)
}

func (engine *Engine) followEdge(definition workflowcore.Definition, state *State, tokenIndex int, edge workflowcore.Edge, depth int) error {
	state.Tokens[tokenIndex].NodeID = edge.Target
	state.Tokens[tokenIndex].Status = TokenStatusActive
	return engine.advanceToken(definition, state, tokenIndex, depth+1)
}

func (engine *Engine) createApprovalTasks(workflowName string, state *State, tokenIndex int, node workflowcore.Node) error {
	assignees, err := engine.resolver.Resolve(AssigneeRequest{Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables)})
	if err != nil {
		return err
	}
	assignees = uniqueNonEmpty(assignees)
	if len(assignees) == 0 {
		return ErrAssigneeUnavailable
	}
	if node.ApprovalMode == workflowcore.ApprovalModeSingle {
		assignees = assignees[:1]
	}
	groupKey := engine.ids.NewID("task-group")
	for index, assigneeID := range assignees {
		status := TaskStatusPending
		if node.ApprovalMode == workflowcore.ApprovalModeSequential && index > 0 {
			status = TaskStatusWaiting
		}
		task := Task{
			ID: engine.ids.NewID("task"), TokenID: state.Tokens[tokenIndex].ID,
			NodeID: node.ID, NodeName: node.Name, GroupKey: groupKey,
			AssigneeID: assigneeID, ApprovalMode: node.ApprovalMode,
			CompletionRate: node.CompletionRate, Sequence: index + 1,
			Total: len(assignees), Status: status,
		}
		state.Tasks = append(state.Tasks, task)
		engine.addHistory(state, HistoryTaskCreated, task.NodeID, task.ID, task.AssigneeID, "审批任务已创建")
		if task.Status == TaskStatusPending {
			engine.addTaskNotificationIntent(state, workflowName, node, task)
		}
	}
	state.Tokens[tokenIndex].Status = TokenStatusWaiting
	return nil
}

func (engine *Engine) addTaskNotificationIntent(state *State, workflowName string, node workflowcore.Node, task Task) {
	engine.addNotificationIntent(state, workflowName, node, task.ID, task.AssigneeID, NotificationKindTaskArrived)
}

func (engine *Engine) addNotificationIntent(state *State, workflowName string, node workflowcore.Node, taskID, recipientID string, kind NotificationKind) {
	if node.Notification == nil || !node.Notification.Enabled {
		return
	}
	config := *node.Notification
	config.Channels = append([]string(nil), node.Notification.Channels...)
	state.NotificationIntents = append(state.NotificationIntents, NotificationIntent{
		ID: engine.ids.NewID("notification"), Kind: kind, NodeID: node.ID, NodeName: node.Name,
		TaskID: taskID, RecipientUserID: recipientID, WorkflowName: workflowName, Config: config,
	})
}

func (engine *Engine) splitParallel(definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	edges := outgoingEdges(definition, node.ID)
	if len(edges) < 2 {
		return fmt.Errorf("并行分支 %s 至少需要两条离开连线", node.ID)
	}
	state.Tokens[tokenIndex].Status = TokenStatusCompleted
	branchGroup := engine.ids.NewID("branch")
	childIndexes := make([]int, 0, len(edges))
	for _, edge := range edges {
		state.Tokens = append(state.Tokens, Token{
			ID: engine.ids.NewID("token"), NodeID: edge.Target, Status: TokenStatusActive,
			BranchGroup: branchGroup, BranchTotal: len(edges),
		})
		childIndexes = append(childIndexes, len(state.Tokens)-1)
	}
	for _, childIndex := range childIndexes {
		if err := engine.advanceToken(definition, state, childIndex, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) joinParallel(definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	token := &state.Tokens[tokenIndex]
	if token.BranchGroup == "" || token.BranchTotal < 2 {
		return fmt.Errorf("并行汇聚 %s 收到无分支标识的令牌", node.ID)
	}
	token.Status = TokenStatusWaiting
	waiting := make([]int, 0, token.BranchTotal)
	for index := range state.Tokens {
		candidate := state.Tokens[index]
		if candidate.BranchGroup == token.BranchGroup && candidate.NodeID == node.ID && candidate.Status == TokenStatusWaiting {
			waiting = append(waiting, index)
		}
	}
	if len(waiting) < token.BranchTotal {
		return nil
	}
	for _, index := range waiting {
		state.Tokens[index].Status = TokenStatusCompleted
	}
	state.Tokens = append(state.Tokens, Token{ID: engine.ids.NewID("token"), NodeID: node.ID, Status: TokenStatusActive})
	return engine.leaveNode(definition, state, len(state.Tokens)-1, depth+1)
}

func (engine *Engine) rejectInstance(state *State, actorID string) {
	state.Instance.Status = InstanceStatusRejected
	for index := range state.Tasks {
		if state.Tasks[index].Status == TaskStatusPending || state.Tasks[index].Status == TaskStatusWaiting {
			state.Tasks[index].Status = TaskStatusCancelled
			engine.addHistory(state, HistoryTaskCancelled, state.Tasks[index].NodeID, state.Tasks[index].ID, actorID, "实例已拒绝")
		}
	}
	for index := range state.Tokens {
		if state.Tokens[index].Status == TokenStatusActive || state.Tokens[index].Status == TokenStatusWaiting {
			state.Tokens[index].Status = TokenStatusCancelled
		}
	}
	engine.addHistory(state, HistoryInstanceRejected, "", "", actorID, "流程实例已拒绝")
}

func (engine *Engine) cancelGroupTasks(state *State, groupKey, actorID string) {
	for index := range state.Tasks {
		if state.Tasks[index].GroupKey == groupKey && (state.Tasks[index].Status == TaskStatusPending || state.Tasks[index].Status == TaskStatusWaiting) {
			state.Tasks[index].Status = TaskStatusCancelled
			engine.addHistory(state, HistoryTaskCancelled, state.Tasks[index].NodeID, state.Tasks[index].ID, actorID, "会签比例已达到")
		}
	}
}

func (engine *Engine) completeInstanceIfIdle(state *State) {
	for _, token := range state.Tokens {
		if token.Status == TokenStatusActive || token.Status == TokenStatusWaiting {
			return
		}
	}
	state.Instance.Status = InstanceStatusCompleted
	engine.addHistory(state, HistoryInstanceCompleted, "", "", "", "流程实例已完成")
}

func (engine *Engine) addHistory(state *State, eventType HistoryEventType, nodeID, taskID, actorID, message string) {
	state.History = append(state.History, HistoryEvent{
		ID: engine.ids.NewID("history"), Type: eventType, NodeID: nodeID,
		TaskID: taskID, ActorID: actorID, Message: message,
	})
}

func findStartNode(definition workflowcore.Definition) (workflowcore.Node, bool) {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart {
			return node, true
		}
	}
	return workflowcore.Node{}, false
}

func findNode(definition workflowcore.Definition, nodeID string) (workflowcore.Node, bool) {
	for _, node := range definition.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return workflowcore.Node{}, false
}

func outgoingEdges(definition workflowcore.Definition, nodeID string) []workflowcore.Edge {
	result := make([]workflowcore.Edge, 0)
	for _, edge := range definition.Edges {
		if edge.Source == nodeID {
			result = append(result, edge)
		}
	}
	return result
}

func selectExclusiveEdge(definition workflowcore.Definition, nodeID string, variables, formData map[string]interface{}) (workflowcore.Edge, error) {
	var defaultEdge *workflowcore.Edge
	for _, edge := range outgoingEdges(definition, nodeID) {
		if edge.Default {
			copy := edge
			defaultEdge = &copy
			continue
		}
		if edge.Condition != nil {
			actual, ok := conditionValue(edge.Condition.Field, variables, formData)
			if ok && conditionMatches(*edge.Condition, actual) {
				return edge, nil
			}
		}
	}
	if defaultEdge != nil {
		return *defaultEdge, nil
	}
	return workflowcore.Edge{}, ErrNoMatchingBranch
}

func conditionValue(field string, variables, formData map[string]interface{}) (interface{}, bool) {
	if value, ok := variables[field]; ok && value != nil {
		return value, true
	}
	value, ok := formData[field]
	return value, ok && value != nil
}

func conditionMatches(condition workflowcore.Condition, actual interface{}) bool {
	comparison, comparable := compareValues(actual, condition.Value)
	switch condition.Operator {
	case workflowcore.ConditionEQ:
		return comparison == 0
	case workflowcore.ConditionNE:
		return comparison != 0
	case workflowcore.ConditionGT:
		return comparable && comparison > 0
	case workflowcore.ConditionGTE:
		return comparable && comparison >= 0
	case workflowcore.ConditionLT:
		return comparable && comparison < 0
	case workflowcore.ConditionLTE:
		return comparable && comparison <= 0
	default:
		return false
	}
}

func compareValues(left, right interface{}) (int, bool) {
	if leftNumber, ok := rationalValue(left); ok {
		if rightNumber, ok := rationalValue(right); ok {
			return leftNumber.Cmp(rightNumber), true
		}
	}
	leftText := fmt.Sprint(left)
	rightText := fmt.Sprint(right)
	if leftText < rightText {
		return -1, true
	}
	if leftText > rightText {
		return 1, true
	}
	if reflect.DeepEqual(left, right) || leftText == rightText {
		return 0, true
	}
	return 0, false
}

func rationalValue(value interface{}) (*big.Rat, bool) {
	var text string
	switch typed := value.(type) {
	case json.Number:
		text = typed.String()
	case string:
		text = typed
	case int:
		text = strconv.Itoa(typed)
	case int8, int16, int32, int64:
		text = fmt.Sprintf("%d", typed)
	case uint, uint8, uint16, uint32, uint64:
		text = fmt.Sprintf("%d", typed)
	case float32, float64:
		text = fmt.Sprint(typed)
	default:
		return nil, false
	}
	result := new(big.Rat)
	if _, ok := result.SetString(strings.TrimSpace(text)); !ok {
		return nil, false
	}
	return result, true
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
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

func findTaskIndex(tasks []Task, taskID string) int {
	for index := range tasks {
		if tasks[index].ID == taskID {
			return index
		}
	}
	return -1
}

func findTokenIndex(tokens []Token, tokenID string) int {
	for index := range tokens {
		if tokens[index].ID == tokenID {
			return index
		}
	}
	return -1
}

func nextWaitingTask(tasks []Task, groupKey string) int {
	indexes := make([]int, 0)
	for index := range tasks {
		if tasks[index].GroupKey == groupKey && tasks[index].Status == TaskStatusWaiting {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) == 0 {
		return -1
	}
	sort.Slice(indexes, func(i, j int) bool { return tasks[indexes[i]].Sequence < tasks[indexes[j]].Sequence })
	return indexes[0]
}

func allGroupTasksApproved(tasks []Task, groupKey string) bool {
	found := false
	for _, task := range tasks {
		if task.GroupKey != groupKey {
			continue
		}
		found = true
		if task.Status != TaskStatusApproved {
			return false
		}
	}
	return found
}

func countersignThresholdReached(tasks []Task, groupKey string, completionRate int) bool {
	total, approved := 0, 0
	for _, task := range tasks {
		if task.GroupKey != groupKey {
			continue
		}
		total++
		if task.Status == TaskStatusApproved {
			approved++
		}
	}
	return total > 0 && approved*100 >= completionRate*total
}

func cloneVariables(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
