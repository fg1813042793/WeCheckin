package application

import (
	"sort"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func BuildNodeProgress(
	definition workflowcore.Definition,
	instance InstanceSummary,
	tokens []TokenSummary,
	tasks []TaskSummary,
	history []HistorySummary,
) []NodeProgressSummary {
	completed := make(map[string]bool, len(definition.Nodes))
	processing := make(map[string]bool, len(definition.Nodes))
	returned := make(map[string]bool, len(definition.Nodes))
	terminated := make(map[string]bool, len(definition.Nodes))

	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart {
			completed[node.ID] = true
		}
		if node.Type == workflowcore.NodeTypeEnd && instance.Status == string(workflowdomain.InstanceStatusCompleted) {
			completed[node.ID] = true
		}
	}
	collectTokenProgress(tokens, completed, processing, terminated)
	collectTaskProgress(tasks, completed, processing, returned, terminated)
	collectHistoryProgress(history, completed, processing, terminated)
	completeTraversedGateways(definition, completed, processing, terminated)

	orderedNodes := topologicallySortedNodes(definition)
	result := make([]NodeProgressSummary, 0, len(orderedNodes))
	for _, node := range orderedNodes {
		status := defaultNodeProgressStatus(instance.Status)
		switch {
		case processing[node.ID]:
			status = NodeProgressProcessing
		case returned[node.ID]:
			status = NodeProgressReturned
		case terminated[node.ID]:
			status = NodeProgressTerminated
		case completed[node.ID]:
			status = NodeProgressCompleted
		}
		result = append(result, NodeProgressSummary{
			NodeID:      node.ID,
			NodeName:    node.Name,
			NodeType:    node.Type,
			GatewayMode: node.GatewayMode,
			Status:      status,
		})
	}
	return result
}

func collectTokenProgress(
	tokens []TokenSummary,
	completed map[string]bool,
	processing map[string]bool,
	terminated map[string]bool,
) {
	for _, token := range tokens {
		switch workflowdomain.TokenStatus(token.Status) {
		case workflowdomain.TokenStatusActive, workflowdomain.TokenStatusWaiting:
			processing[token.NodeID] = true
		case workflowdomain.TokenStatusCompleted:
			completed[token.NodeID] = true
		case workflowdomain.TokenStatusCancelled:
			terminated[token.NodeID] = true
		}
	}
}

func collectTaskProgress(
	tasks []TaskSummary,
	completed map[string]bool,
	processing map[string]bool,
	returned map[string]bool,
	terminated map[string]bool,
) {
	type taskEvidence struct {
		processing   bool
		cancelled    bool
		lastDecision workflowdomain.TaskStatus
	}
	byNode := make(map[string]taskEvidence)
	for _, task := range tasks {
		evidence := byNode[task.NodeID]
		switch workflowdomain.TaskStatus(task.Status) {
		case workflowdomain.TaskStatusWaiting, workflowdomain.TaskStatusPending:
			evidence.processing = true
		case workflowdomain.TaskStatusCompleted, workflowdomain.TaskStatusApproved:
			evidence.lastDecision = workflowdomain.TaskStatus(task.Status)
		case workflowdomain.TaskStatusRejected:
			evidence.lastDecision = workflowdomain.TaskStatusRejected
		case workflowdomain.TaskStatusReturned:
			evidence.lastDecision = workflowdomain.TaskStatusReturned
		case workflowdomain.TaskStatusCancelled:
			evidence.cancelled = true
		default:
			if task.Status == "submitted" {
				evidence.lastDecision = workflowdomain.TaskStatusCompleted
			}
		}
		byNode[task.NodeID] = evidence
	}
	for nodeID, evidence := range byNode {
		switch {
		case evidence.processing:
			processing[nodeID] = true
		case evidence.lastDecision == workflowdomain.TaskStatusReturned:
			returned[nodeID] = true
		case evidence.lastDecision == workflowdomain.TaskStatusRejected:
			terminated[nodeID] = true
		case evidence.lastDecision == workflowdomain.TaskStatusCompleted || evidence.lastDecision == workflowdomain.TaskStatusApproved:
			completed[nodeID] = true
		case evidence.cancelled:
			terminated[nodeID] = true
		}
	}
}

func collectHistoryProgress(
	history []HistorySummary,
	completed map[string]bool,
	processing map[string]bool,
	terminated map[string]bool,
) {
	type historyEvidence struct {
		completed bool
		waiting   bool
		rejected  bool
		cancelled bool
	}
	byNode := make(map[string]historyEvidence)
	for _, event := range history {
		if event.NodeID == "" {
			continue
		}
		evidence := byNode[event.NodeID]
		switch workflowdomain.HistoryEventType(event.EventType) {
		case workflowdomain.HistoryTaskApproved,
			workflowdomain.HistoryTaskSubmitted,
			workflowdomain.HistoryNodeCC,
			workflowdomain.HistoryNodeNotify,
			workflowdomain.HistoryNodeAutomated,
			workflowdomain.HistoryTimerResumed:
			evidence.completed = true
		case workflowdomain.HistoryTimerWaiting:
			evidence.waiting = true
		case workflowdomain.HistoryTaskRejected:
			evidence.rejected = true
		case workflowdomain.HistoryTaskCancelled:
			evidence.cancelled = true
		}
		byNode[event.NodeID] = evidence
	}
	for nodeID, evidence := range byNode {
		if evidence.completed {
			completed[nodeID] = true
		}
		if evidence.rejected || (evidence.cancelled && !evidence.completed && !completed[nodeID]) {
			terminated[nodeID] = true
		}
		if evidence.waiting && !evidence.completed && !completed[nodeID] && !evidence.rejected {
			processing[nodeID] = true
		}
	}
}

func completeTraversedGateways(
	definition workflowcore.Definition,
	completed map[string]bool,
	processing map[string]bool,
	terminated map[string]bool,
) {
	startNodeID := ""
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart {
			startNodeID = node.ID
			break
		}
	}
	if startNodeID == "" {
		return
	}

	evidenceNodeIDs := make([]string, 0, len(definition.Nodes))
	for _, node := range definition.Nodes {
		if node.Type != workflowcore.NodeTypeExclusive && node.Type != workflowcore.NodeTypeParallel &&
			(completed[node.ID] || processing[node.ID] || terminated[node.ID]) {
			evidenceNodeIDs = append(evidenceNodeIDs, node.ID)
		}
	}
	for _, node := range definition.Nodes {
		if node.Type != workflowcore.NodeTypeExclusive && node.Type != workflowcore.NodeTypeParallel {
			continue
		}
		for _, evidenceNodeID := range evidenceNodeIDs {
			if gatewayDominatesNode(definition, startNodeID, node.ID, evidenceNodeID) {
				completed[node.ID] = true
				break
			}
		}
	}
}

func gatewayDominatesNode(definition workflowcore.Definition, startNodeID, gatewayNodeID, evidenceNodeID string) bool {
	if evidenceNodeID == startNodeID || evidenceNodeID == gatewayNodeID {
		return false
	}
	return isNodeReachable(definition, startNodeID, evidenceNodeID, "") &&
		!isNodeReachable(definition, startNodeID, evidenceNodeID, gatewayNodeID)
}

func isNodeReachable(definition workflowcore.Definition, startNodeID, targetNodeID, excludedNodeID string) bool {
	if startNodeID == excludedNodeID || targetNodeID == excludedNodeID {
		return false
	}
	validNodes := make(map[string]bool, len(definition.Nodes))
	for _, node := range definition.Nodes {
		validNodes[node.ID] = true
	}
	if !validNodes[startNodeID] || !validNodes[targetNodeID] {
		return false
	}
	adjacency := make(map[string][]string, len(definition.Nodes))
	for _, edge := range definition.Edges {
		if validNodes[edge.Source] && validNodes[edge.Target] && edge.Source != excludedNodeID && edge.Target != excludedNodeID {
			adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
		}
	}
	visited := map[string]bool{startNodeID: true}
	queue := []string{startNodeID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == targetNodeID {
			return true
		}
		for _, next := range adjacency[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

func topologicallySortedNodes(definition workflowcore.Definition) []workflowcore.Node {
	if len(definition.Nodes) < 2 {
		return append([]workflowcore.Node(nil), definition.Nodes...)
	}
	nodeByID := make(map[string]workflowcore.Node, len(definition.Nodes))
	definitionIndex := make(map[string]int, len(definition.Nodes))
	indegree := make(map[string]int, len(definition.Nodes))
	adjacency := make(map[string][]string, len(definition.Nodes))
	for index, node := range definition.Nodes {
		nodeByID[node.ID] = node
		definitionIndex[node.ID] = index
		indegree[node.ID] = 0
	}
	for _, edge := range definition.Edges {
		if _, sourceExists := nodeByID[edge.Source]; !sourceExists {
			continue
		}
		if _, targetExists := nodeByID[edge.Target]; !targetExists {
			continue
		}
		adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
		indegree[edge.Target]++
	}

	ready := make([]string, 0, len(definition.Nodes))
	for _, node := range definition.Nodes {
		if indegree[node.ID] == 0 {
			ready = append(ready, node.ID)
		}
	}
	sortReadyNodes(ready, nodeByID, definitionIndex)
	ordered := make([]workflowcore.Node, 0, len(definition.Nodes))
	seen := make(map[string]bool, len(definition.Nodes))
	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		ordered = append(ordered, nodeByID[current])
		for _, next := range adjacency[current] {
			indegree[next]--
			if indegree[next] == 0 {
				ready = append(ready, next)
			}
		}
		sortReadyNodes(ready, nodeByID, definitionIndex)
	}
	for _, node := range definition.Nodes {
		if !seen[node.ID] {
			ordered = append(ordered, node)
		}
	}
	return ordered
}

func sortReadyNodes(ready []string, nodeByID map[string]workflowcore.Node, definitionIndex map[string]int) {
	sort.SliceStable(ready, func(left, right int) bool {
		leftNode := nodeByID[ready[left]]
		rightNode := nodeByID[ready[right]]
		if leftNode.Position != nil && rightNode.Position != nil {
			if leftNode.Position.Y != rightNode.Position.Y {
				return leftNode.Position.Y < rightNode.Position.Y
			}
			if leftNode.Position.X != rightNode.Position.X {
				return leftNode.Position.X < rightNode.Position.X
			}
		}
		return definitionIndex[ready[left]] < definitionIndex[ready[right]]
	})
}

func defaultNodeProgressStatus(instanceStatus string) string {
	switch workflowdomain.InstanceStatus(instanceStatus) {
	case workflowdomain.InstanceStatusCompleted:
		return NodeProgressSkipped
	case workflowdomain.InstanceStatusRunning:
		return NodeProgressNotStarted
	default:
		return NodeProgressTerminated
	}
}
