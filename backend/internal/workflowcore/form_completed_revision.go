package workflowcore

import (
	"fmt"
	"sort"
	"strings"
)

func CompletedRevisionCapabilityForNode(definition Definition, nodeID string) (CompletedRevisionCapability, error) {
	nodeID = strings.TrimSpace(nodeID)
	var node *Node
	for index := range definition.Nodes {
		if strings.TrimSpace(definition.Nodes[index].ID) == nodeID {
			node = &definition.Nodes[index]
			break
		}
	}
	if node == nil || (node.Type != NodeTypeApproval && node.Type != NodeTypeHandle) {
		return CompletedRevisionCapability{}, fmt.Errorf("%w：人工任务节点 %s 不存在", ErrFormDataInvalid, nodeID)
	}
	config := node.PostHandleEdit
	if config == nil || config.CompletedRevision == nil || !config.CompletedRevision.Enabled {
		return CompletedRevisionCapability{}, fmt.Errorf("%w：节点 %s 未启用流程完成后修订", ErrFormDataInvalid, nodeID)
	}

	routingFields := revisionRoutingFields(definition)
	permissionByField := make(map[string]FieldPermission, len(node.FormPermissions))
	for _, permission := range node.FormPermissions {
		permissionByField[strings.TrimSpace(permission.Field)] = permission
	}
	fields := dataFormFields(definition.Form)
	permissions := make([]FieldPermission, 0, len(fields))
	writable := make(map[string]struct{})
	for _, field := range fields {
		permission, configured := permissionByField[field.Key]
		access := FieldAccessRead
		if configured {
			switch permission.Access {
			case FieldAccessHidden, FieldAccessRead, FieldAccessWrite:
				access = permission.Access
			}
		}
		if field.Type == FormFieldTypeCalculation {
			access = FieldAccessRead
		}
		if _, blocked := routingFields[field.Key]; blocked {
			access = FieldAccessRead
		}
		actions := []string(nil)
		if access == FieldAccessWrite {
			writable[field.Key] = struct{}{}
			actions = normalizedRevisionActions(permission.Actions)
		}
		permissions = append(permissions, FieldPermission{Field: field.Key, Access: access, Actions: actions})
	}

	directFields := make([]string, 0, len(config.CompletedRevision.DirectFields))
	seen := make(map[string]struct{}, len(config.CompletedRevision.DirectFields))
	for _, raw := range config.CompletedRevision.DirectFields {
		field := strings.TrimSpace(raw)
		if _, ok := writable[field]; !ok {
			return CompletedRevisionCapability{}, fmt.Errorf("%w：字段 %s 不能直接修订", ErrFormDataInvalid, field)
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		directFields = append(directFields, field)
	}
	sort.Strings(directFields)

	return CompletedRevisionCapability{
		NodeID:           node.ID,
		NodeName:         node.Name,
		FieldPermissions: permissions,
		DirectFields:     directFields,
	}, nil
}

func ValidateCompletedRevisionPatch(definition Definition, nodeID string, current, patch map[string]interface{}) error {
	capability, err := CompletedRevisionCapabilityForNode(definition, nodeID)
	if err != nil {
		return err
	}
	writableFields := make(map[string]struct{})
	permissionByField := make(map[string]FieldPermission)
	for _, permission := range capability.FieldPermissions {
		if permission.Access != FieldAccessWrite {
			continue
		}
		writableFields[permission.Field] = struct{}{}
		permissionByField[permission.Field] = permission
	}
	fieldByKey := make(map[string]FormField)
	for _, field := range dataFormFields(definition.Form) {
		fieldByKey[field.Key] = field
	}
	for field := range patch {
		if _, ok := writableFields[field]; !ok {
			return fmt.Errorf("%w：流程完成后无权修改字段 %s", ErrFormDataInvalid, field)
		}
		formField := fieldByKey[field]
		if formField.Type == FormFieldTypeDetailList {
			if err := validateDetailListPatchActions(formField, current[field], patch[field], permissionByField[field]); err != nil {
				return err
			}
		}
	}
	if err := validateSelectedFormData(definition.Form, patch, writableFields, true); err != nil {
		return err
	}
	return validateSelectedFormData(definition.Form, MergeFormData(current, patch), writableFields, false)
}

func CompletedRevisionMode(capability CompletedRevisionCapability, patch map[string]interface{}) string {
	directFields := make(map[string]struct{}, len(capability.DirectFields))
	for _, raw := range capability.DirectFields {
		if field := strings.TrimSpace(raw); field != "" {
			directFields[field] = struct{}{}
		}
	}
	for field := range patch {
		if _, direct := directFields[strings.TrimSpace(field)]; !direct {
			return CompletedRevisionModeDownstreamReview
		}
	}
	return CompletedRevisionModeDirect
}

func BuildCompletedRevisionPlan(definition Definition, actualHumanNodeIDs []string, sourceNodeID string) ([]RevisionPlanStage, error) {
	sourceNodeID = strings.TrimSpace(sourceNodeID)
	nodes := make(map[string]Node, len(definition.Nodes))
	order := make(map[string]int, len(definition.Nodes))
	outgoing := make(map[string][]string, len(definition.Nodes))
	for index, node := range definition.Nodes {
		nodeID := strings.TrimSpace(node.ID)
		nodes[nodeID] = node
		order[nodeID] = index
	}
	if _, ok := nodes[sourceNodeID]; !ok {
		return nil, fmt.Errorf("修订来源节点 %s 不存在", sourceNodeID)
	}
	for _, edge := range definition.Edges {
		outgoing[strings.TrimSpace(edge.Source)] = append(outgoing[strings.TrimSpace(edge.Source)], strings.TrimSpace(edge.Target))
	}

	actual := make(map[string]struct{}, len(actualHumanNodeIDs))
	for _, raw := range actualHumanNodeIDs {
		nodeID := strings.TrimSpace(raw)
		node, ok := nodes[nodeID]
		if !ok || (node.Type != NodeTypeApproval && node.Type != NodeTypeHandle) {
			continue
		}
		actual[nodeID] = struct{}{}
	}
	if _, ok := actual[sourceNodeID]; !ok {
		return nil, fmt.Errorf("修订来源节点 %s 不在实例实际流转路径中", sourceNodeID)
	}

	reachableCache := make(map[string]map[string]bool)
	var visit func(string) map[string]bool
	visit = func(start string) map[string]bool {
		if cached, ok := reachableCache[start]; ok {
			return cached
		}
		result := make(map[string]bool)
		stack := append([]string(nil), outgoing[start]...)
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if result[current] {
				continue
			}
			result[current] = true
			stack = append(stack, outgoing[current]...)
		}
		reachableCache[start] = result
		return result
	}

	candidates := make([]string, 0, len(actual))
	for nodeID := range actual {
		if nodeID != sourceNodeID && visit(sourceNodeID)[nodeID] {
			candidates = append(candidates, nodeID)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return order[candidates[i]] < order[candidates[j]] })
	if len(candidates) == 0 {
		return []RevisionPlanStage{}, nil
	}

	predecessors := make(map[string][]string, len(candidates))
	for _, target := range candidates {
		for _, possible := range candidates {
			if possible == target || !visit(possible)[target] {
				continue
			}
			direct := true
			for _, between := range candidates {
				if between == possible || between == target {
					continue
				}
				if visit(possible)[between] && visit(between)[target] {
					direct = false
					break
				}
			}
			if direct {
				predecessors[target] = append(predecessors[target], possible)
			}
		}
	}

	stagesByNode := make(map[string]int, len(candidates))
	visiting := make(map[string]bool, len(candidates))
	var stageFor func(string) (int, error)
	stageFor = func(nodeID string) (int, error) {
		if stage := stagesByNode[nodeID]; stage > 0 {
			return stage, nil
		}
		if visiting[nodeID] {
			return 0, fmt.Errorf("修订复核节点存在循环：%s", nodeID)
		}
		visiting[nodeID] = true
		stage := 1
		for _, predecessor := range predecessors[nodeID] {
			previousStage, err := stageFor(predecessor)
			if err != nil {
				return 0, err
			}
			if previousStage+1 > stage {
				stage = previousStage + 1
			}
		}
		visiting[nodeID] = false
		stagesByNode[nodeID] = stage
		return stage, nil
	}

	maxStage := 0
	for _, nodeID := range candidates {
		stage, err := stageFor(nodeID)
		if err != nil {
			return nil, err
		}
		if stage > maxStage {
			maxStage = stage
		}
	}
	result := make([]RevisionPlanStage, 0, maxStage)
	for stage := 1; stage <= maxStage; stage++ {
		nodeIDs := make([]string, 0)
		for _, nodeID := range candidates {
			if stagesByNode[nodeID] == stage {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
		if len(nodeIDs) > 0 {
			result = append(result, RevisionPlanStage{Stage: stage, Nodes: nodeIDs})
		}
	}
	return result, nil
}

func revisionRoutingFields(definition Definition) map[string]struct{} {
	result := make(map[string]struct{})
	for _, edge := range definition.Edges {
		if edge.Condition == nil {
			continue
		}
		if field := strings.TrimSpace(edge.Condition.Field); field != "" {
			result[field] = struct{}{}
		}
	}
	return result
}

func normalizedRevisionActions(actions []string) []string {
	result := make([]string, 0, len(actions))
	seen := make(map[string]struct{}, len(actions))
	for _, raw := range actions {
		action := strings.TrimSpace(raw)
		if action != FieldActionAdd && action != FieldActionDelete {
			continue
		}
		if _, exists := seen[action]; exists {
			continue
		}
		seen[action] = struct{}{}
		result = append(result, action)
	}
	sort.Strings(result)
	return result
}
