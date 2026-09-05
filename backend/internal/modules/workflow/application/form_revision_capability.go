package application

import (
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func decorateInstanceFormRevision(detail *InstanceDetail, actorID string) {
	if detail == nil {
		return
	}
	detail.FormRevision = FormRevisionCapability{Revision: detail.Instance.FormRevision}
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
		detail.FormRevision.FieldPermissions = permissions
	}
}
