package domain

import (
	"context"
	"errors"
	"fmt"

	"wecheckin/backend/internal/workflowcore"
)

func (engine *Engine) enterApproval(ctx context.Context, definition workflowcore.Definition, state *State, tokenIndex int, node workflowcore.Node, depth int) error {
	skipped, err := engine.createApprovalTasks(ctx, workflowNameForState(definition, state), state, tokenIndex, node)
	if err != nil || !skipped {
		return err
	}
	engine.addHistory(state, HistoryNodeSkipped, node.ID, "", "", "未解析到审批人，已按节点配置自动跳过审批")
	return engine.leaveNode(ctx, definition, state, tokenIndex, depth+1)
}

// Only successful, normalized resolver results may reach this decision. Resolver
// errors (including ErrAssigneeUnavailable) and chain missing policies stay errors.
func emptyApprovalResult(node workflowcore.Node) (skipped bool, err error) {
	if node.SkipIfNoAssignee {
		return true, nil
	}
	return false, ErrAssigneeUnavailable
}

func (engine *Engine) createApprovalTasks(ctx context.Context, workflowName string, state *State, tokenIndex int, node workflowcore.Node) (skipped bool, err error) {
	if node.DepartmentApprovalChain != nil && node.DepartmentApprovalChain.Enabled {
		resolver, ok := engine.resolver.(ApprovalLayerResolver)
		if !ok {
			return false, errors.New("流程审批人解析器不支持逐级部门审批")
		}
		layers, err := resolver.ResolveApprovalLayers(ctx, AssigneeRequest{
			Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables),
		})
		if err != nil {
			return false, err
		}
		return engine.createLayeredApprovalTasks(workflowName, state, tokenIndex, node, layers)
	}
	assignees, err := engine.resolver.Resolve(ctx, AssigneeRequest{Instance: state.Instance, Node: node, Variables: cloneVariables(state.Variables)})
	if err != nil {
		return false, err
	}
	assignees = uniqueNonEmpty(assignees)
	if len(assignees) == 0 {
		return emptyApprovalResult(node)
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
		engine.addHistory(state, HistoryTaskCreated, task.NodeID, task.ID, taskCreationActorID(state), "审批任务已创建")
		if task.Status == TaskStatusPending {
			engine.addTaskNotificationIntent(state, workflowName, node, task)
		}
	}
	state.Tokens[tokenIndex].Status = TokenStatusWaiting
	return false, nil
}

func (engine *Engine) createLayeredApprovalTasks(
	workflowName string,
	state *State,
	tokenIndex int,
	node workflowcore.Node,
	layers []ApprovalLayer,
) (skipped bool, err error) {
	normalized := make([]ApprovalLayer, 0, len(layers))
	for _, layer := range layers {
		layer.AssigneeIDs = uniqueNonEmpty(layer.AssigneeIDs)
		if node.ApprovalMode == workflowcore.ApprovalModeSingle && len(layer.AssigneeIDs) > 1 {
			layer.AssigneeIDs = layer.AssigneeIDs[:1]
		}
		if len(layer.AssigneeIDs) > 0 {
			normalized = append(normalized, layer)
		}
	}
	if len(normalized) == 0 {
		return emptyApprovalResult(node)
	}

	chainKey := engine.ids.NewID("approval-chain")
	for layerIndex, layer := range normalized {
		groupKey := engine.ids.NewID("task-group")
		for assigneeIndex, assigneeID := range layer.AssigneeIDs {
			status := TaskStatusWaiting
			if layerIndex == 0 && (node.ApprovalMode != workflowcore.ApprovalModeSequential || assigneeIndex == 0) {
				status = TaskStatusPending
			}
			task := Task{
				ID: engine.ids.NewID("task"), TokenID: state.Tokens[tokenIndex].ID,
				NodeID: node.ID, NodeName: node.Name, GroupKey: groupKey,
				AssigneeID: assigneeID, ApprovalMode: node.ApprovalMode,
				CompletionRate: node.CompletionRate, Sequence: assigneeIndex + 1, Total: len(layer.AssigneeIDs),
				ApprovalChainKey: chainKey, ApprovalLayer: layerIndex + 1, ApprovalLayerTotal: len(normalized),
				DepartmentID: layer.DepartmentID, DepartmentName: layer.DepartmentName,
				Status: status,
			}
			state.Tasks = append(state.Tasks, task)
			message := fmt.Sprintf("审批任务已创建（%s，第 %d/%d 级）", layer.DepartmentName, layerIndex+1, len(normalized))
			engine.addHistory(state, HistoryTaskCreated, task.NodeID, task.ID, taskCreationActorID(state), message)
			if task.Status == TaskStatusPending {
				engine.addTaskNotificationIntent(state, workflowName, node, task)
			}
		}
	}
	state.Tokens[tokenIndex].Status = TokenStatusWaiting
	return false, nil
}
