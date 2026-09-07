package application

import (
	"sort"
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func decorateInstanceFormRevision(detail *InstanceDetail, actorID string) {
	if detail == nil {
		return
	}
	detail.FormRevision = FormRevisionCapability{Revision: detail.Instance.FormRevision}
	if detail.Instance.Status == string(workflowdomain.InstanceStatusCompleted) {
		decorateCompletedInstanceFormRevision(detail, actorID)
		return
	}
	if detail.Instance.Status != string(workflowdomain.InstanceStatusRunning) {
		return
	}
	handledNodeIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, event := range detail.History {
		if strings.TrimSpace(event.ActorID) != strings.TrimSpace(actorID) {
			continue
		}
		switch workflowdomain.HistoryEventType(event.EventType) {
		case workflowdomain.HistoryTaskApproved, workflowdomain.HistoryTaskSubmitted, workflowdomain.HistoryTaskReturned:
		default:
			continue
		}
		nodeID := strings.TrimSpace(event.NodeID)
		if nodeID == "" {
			continue
		}
		if _, exists := seen[nodeID]; exists {
			continue
		}
		seen[nodeID] = struct{}{}
		handledNodeIDs = append(handledNodeIDs, nodeID)
	}
	definition := workflowcore.Definition{Form: detail.Form}
	for _, node := range detail.Nodes {
		definition.Nodes = append(definition.Nodes, workflowcore.Node{
			ID: node.ID, Type: node.Type,
			FormPermissions: detail.FieldPermissions[node.ID],
			PostHandleEdit:  node.PostHandleEdit,
		})
	}
	for _, edge := range detail.Edges {
		definition.Edges = append(definition.Edges, workflowcore.Edge{Condition: edge.Condition})
	}
	permissions := workflowcore.PostHandleFormPermissions(definition, handledNodeIDs)
	for _, permission := range permissions {
		if permission.Access == workflowcore.FieldAccessWrite {
			detail.FormRevision.Allowed = true
			break
		}
	}
	if detail.FormRevision.Allowed {
		detail.FormRevision.RunningRevisionAllowed = true
		detail.FormRevision.FieldPermissions = permissions
	}
}

func decorateCompletedInstanceFormRevision(detail *InstanceDetail, actorID string) {
	definition := workflowcore.Definition{Form: detail.Form}
	for _, node := range detail.Nodes {
		definition.Nodes = append(definition.Nodes, workflowcore.Node{
			ID: node.ID, Type: node.Type, Name: node.Name,
			FormPermissions: detail.FieldPermissions[node.ID], PostHandleEdit: node.PostHandleEdit,
		})
	}
	for _, edge := range detail.Edges {
		definition.Edges = append(definition.Edges, workflowcore.Edge{Source: edge.Source, Target: edge.Target, Condition: edge.Condition})
	}

	latestByNode := make(map[string]int64)
	for _, event := range detail.History {
		if strings.TrimSpace(event.ActorID) != strings.TrimSpace(actorID) {
			continue
		}
		switch workflowdomain.HistoryEventType(event.EventType) {
		case workflowdomain.HistoryTaskApproved, workflowdomain.HistoryTaskSubmitted, workflowdomain.HistoryTaskReturned:
		default:
			continue
		}
		if event.EventTime > latestByNode[event.NodeID] {
			latestByNode[event.NodeID] = event.EventTime
		}
	}
	type orderedCapability struct {
		capability workflowcore.CompletedRevisionCapability
		time       int64
	}
	ordered := make([]orderedCapability, 0, len(latestByNode))
	for nodeID, eventTime := range latestByNode {
		capability, err := workflowcore.CompletedRevisionCapabilityForNode(definition, nodeID)
		if err != nil || !hasWritableCompletedRevisionField(capability) {
			continue
		}
		ordered = append(ordered, orderedCapability{capability: capability, time: eventTime})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].time == ordered[j].time {
			return ordered[i].capability.NodeID < ordered[j].capability.NodeID
		}
		return ordered[i].time > ordered[j].time
	})
	for _, item := range ordered {
		detail.FormRevision.CompletedRevisionNodes = append(detail.FormRevision.CompletedRevisionNodes, item.capability)
	}
}

func hasWritableCompletedRevisionField(capability workflowcore.CompletedRevisionCapability) bool {
	for _, permission := range capability.FieldPermissions {
		if permission.Access == workflowcore.FieldAccessWrite {
			return true
		}
	}
	return false
}
