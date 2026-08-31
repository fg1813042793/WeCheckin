package workflow

import (
	"regexp"
	"strings"
)

var definitionKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,99}$`)
var formFieldKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,99}$`)

func ValidateDefinition(definition Definition) []ValidationError {
	errors := make([]ValidationError, 0)
	if definition.SchemaVersion != CurrentSchemaVersion {
		errors = append(errors, ValidationError{Code: ValidationSchemaVersion, Message: "流程定义版本不受支持"})
	}
	if !definitionKeyPattern.MatchString(strings.TrimSpace(definition.Key)) {
		errors = append(errors, ValidationError{Code: ValidationDefinitionKey, Message: "流程编码需以字母开头，仅允许字母、数字、下划线和中划线"})
	}
	formFields, formErrors := validateFormSchema(definition.Form)
	errors = append(errors, formErrors...)

	nodes := make(map[string]Node, len(definition.Nodes))
	incoming := make(map[string][]Edge, len(definition.Nodes))
	outgoing := make(map[string][]Edge, len(definition.Nodes))
	startIDs := make([]string, 0, 1)
	endIDs := make([]string, 0, 1)
	for _, node := range definition.Nodes {
		node.ID = strings.TrimSpace(node.ID)
		if node.ID == "" {
			errors = append(errors, ValidationError{Code: ValidationDuplicateNode, Message: "节点 ID 不能为空"})
			continue
		}
		if _, exists := nodes[node.ID]; exists {
			errors = append(errors, ValidationError{Code: ValidationDuplicateNode, Message: "节点 ID 重复", NodeID: node.ID})
			continue
		}
		nodes[node.ID] = node
		switch node.Type {
		case NodeTypeStart:
			startIDs = append(startIDs, node.ID)
		case NodeTypeEnd:
			endIDs = append(endIDs, node.ID)
		case NodeTypeApproval:
			errors = append(errors, validateApproval(node)...)
			errors = append(errors, validateFieldPermissions(node, formFields)...)
		case NodeTypeExclusive, NodeTypeParallel:
			if node.GatewayMode != GatewayModeSplit && node.GatewayMode != GatewayModeJoin {
				errors = append(errors, ValidationError{Code: ValidationGatewayMode, Message: "网关必须设置为分支或汇聚", NodeID: node.ID})
			}
		default:
			errors = append(errors, ValidationError{Code: ValidationGatewayMode, Message: "不支持的节点类型", NodeID: node.ID})
		}
	}

	if len(startIDs) == 0 {
		errors = append(errors, ValidationError{Code: ValidationMissingStart, Message: "流程必须包含一个开始节点"})
	} else if len(startIDs) > 1 {
		errors = append(errors, ValidationError{Code: ValidationMultipleStarts, Message: "流程只能包含一个开始节点"})
	}
	if len(endIDs) == 0 {
		errors = append(errors, ValidationError{Code: ValidationMissingEnd, Message: "流程必须至少包含一个结束节点"})
	}

	edgeIDs := make(map[string]struct{}, len(definition.Edges))
	for _, edge := range definition.Edges {
		edge.ID = strings.TrimSpace(edge.ID)
		if edge.ID == "" {
			errors = append(errors, ValidationError{Code: ValidationDuplicateEdge, Message: "连线 ID 不能为空"})
			continue
		}
		if _, exists := edgeIDs[edge.ID]; exists {
			errors = append(errors, ValidationError{Code: ValidationDuplicateEdge, Message: "连线 ID 重复", EdgeID: edge.ID})
			continue
		}
		edgeIDs[edge.ID] = struct{}{}
		if _, ok := nodes[edge.Source]; !ok {
			errors = append(errors, ValidationError{Code: ValidationEdgeEndpoint, Message: "连线起点不存在", EdgeID: edge.ID})
			continue
		}
		if _, ok := nodes[edge.Target]; !ok {
			errors = append(errors, ValidationError{Code: ValidationEdgeEndpoint, Message: "连线终点不存在", EdgeID: edge.ID})
			continue
		}
		outgoing[edge.Source] = append(outgoing[edge.Source], edge)
		incoming[edge.Target] = append(incoming[edge.Target], edge)
	}

	for nodeID, node := range nodes {
		if node.Type != NodeTypeStart && len(incoming[nodeID]) == 0 {
			errors = append(errors, ValidationError{Code: ValidationIncomingRequired, Message: "节点缺少进入连线", NodeID: nodeID})
		}
		if node.Type != NodeTypeEnd && len(outgoing[nodeID]) == 0 {
			errors = append(errors, ValidationError{Code: ValidationOutgoingRequired, Message: "节点缺少离开连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeStart && len(incoming[nodeID]) > 0 {
			errors = append(errors, ValidationError{Code: ValidationIncomingRequired, Message: "开始节点不能有进入连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeEnd && len(outgoing[nodeID]) > 0 {
			errors = append(errors, ValidationError{Code: ValidationOutgoingRequired, Message: "结束节点不能有离开连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeExclusive || node.Type == NodeTypeParallel {
			errors = append(errors, validateGateway(node, incoming[nodeID], outgoing[nodeID])...)
		}
	}

	if len(startIDs) == 1 {
		reachable := visitForward(startIDs[0], outgoing)
		for nodeID := range nodes {
			if !reachable[nodeID] {
				errors = append(errors, ValidationError{Code: ValidationUnreachableNode, Message: "节点无法从开始节点到达", NodeID: nodeID})
			}
		}
	}
	if len(endIDs) > 0 {
		canReachEnd := make(map[string]bool, len(nodes))
		for _, endID := range endIDs {
			for nodeID := range visitBackward(endID, incoming) {
				canReachEnd[nodeID] = true
			}
		}
		for nodeID := range nodes {
			if !canReachEnd[nodeID] {
				errors = append(errors, ValidationError{Code: ValidationNoPathToEnd, Message: "节点不存在通往结束节点的路径", NodeID: nodeID})
			}
		}
	}
	return errors
}

func validateFormSchema(fields []FormField) (map[string]FormField, []ValidationError) {
	fieldByKey := make(map[string]FormField, len(fields))
	errors := make([]ValidationError, 0)
	for _, field := range fields {
		field.Key = strings.TrimSpace(field.Key)
		field.Label = strings.TrimSpace(field.Label)
		if !formFieldKeyPattern.MatchString(field.Key) || field.Label == "" {
			errors = append(errors, ValidationError{Code: ValidationFormFieldKey, Message: "表单字段编码或名称无效"})
			continue
		}
		if _, exists := fieldByKey[field.Key]; exists {
			errors = append(errors, ValidationError{Code: ValidationFormFieldDuplicate, Message: "表单字段编码重复：" + field.Key})
			continue
		}
		fieldByKey[field.Key] = field
		switch field.Type {
		case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypeNumber, FormFieldTypeSelect,
			FormFieldTypeMultiSelect, FormFieldTypeDate, FormFieldTypeDateTime, FormFieldTypeUser,
			FormFieldTypeDepartment, FormFieldTypeAttachment, FormFieldTypeBoolean, FormFieldTypeAmount,
			FormFieldTypePhone, FormFieldTypeEmail, FormFieldTypeRadio, FormFieldTypeCheckbox,
			FormFieldTypeTime, FormFieldTypeDateRange, FormFieldTypeUserMulti, FormFieldTypeDepartmentMulti:
		default:
			errors = append(errors, ValidationError{Code: ValidationFormFieldType, Message: "表单字段类型无效：" + field.Key})
		}
		if (field.Type == FormFieldTypeSelect || field.Type == FormFieldTypeMultiSelect ||
			field.Type == FormFieldTypeRadio || field.Type == FormFieldTypeCheckbox) && !validFormOptions(field.Options) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldOptions, Message: "选择字段必须配置不重复的选项：" + field.Key})
		}
		if field.MaxLength < 0 || (field.Min != nil && field.Max != nil && *field.Min > *field.Max) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldRange, Message: "表单字段约束无效：" + field.Key})
		}
		if !validFormFieldSpan(field.Span) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldSpan, Message: "表单字段宽度无效：" + field.Key})
		}
		if field.Default != nil {
			if err := validateFieldValue(field, field.Default); err != nil {
				errors = append(errors, ValidationError{Code: ValidationFormFieldType, Message: "表单字段默认值无效：" + field.Key})
			}
		}
	}
	return fieldByKey, errors
}

func validFormFieldSpan(span int) bool {
	return span == 0 || span == 6 || span == 8 || span == 12 || span == 24
}

func validFormOptions(options []FormOption) bool {
	if len(options) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		value := strings.TrimSpace(option.Value)
		if strings.TrimSpace(option.Label) == "" || value == "" {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validateFieldPermissions(node Node, fields map[string]FormField) []ValidationError {
	errors := make([]ValidationError, 0)
	seen := make(map[string]struct{}, len(node.FormPermissions))
	for _, permission := range node.FormPermissions {
		field := strings.TrimSpace(permission.Field)
		if _, ok := fields[field]; !ok {
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionField, Message: "节点字段权限引用了不存在的字段", NodeID: node.ID})
			continue
		}
		if _, exists := seen[field]; exists {
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionDuplicate, Message: "节点字段权限重复", NodeID: node.ID})
			continue
		}
		seen[field] = struct{}{}
		switch permission.Access {
		case FieldAccessHidden, FieldAccessRead, FieldAccessWrite:
		default:
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionAccess, Message: "节点字段权限无效", NodeID: node.ID})
		}
	}
	return errors
}

func validateApproval(node Node) []ValidationError {
	errors := make([]ValidationError, 0, 2)
	switch node.ApprovalMode {
	case ApprovalModeSingle, ApprovalModeSequential, ApprovalModeParallel, ApprovalModeCountersign:
	default:
		errors = append(errors, ValidationError{Code: ValidationApprovalMode, Message: "审批方式无效", NodeID: node.ID})
	}
	if node.Assignee == nil || strings.TrimSpace(node.Assignee.Type) == "" || strings.TrimSpace(node.Assignee.Value) == "" {
		errors = append(errors, ValidationError{Code: ValidationAssigneeRequired, Message: "审批节点必须配置审批人", NodeID: node.ID})
	} else {
		switch node.Assignee.Type {
		case AssigneeTypeUser, AssigneeTypeRole, AssigneeTypeDepartmentLeader, AssigneeTypeManager, AssigneeTypeVariable, AssigneeTypeOrgIdentity:
		default:
			errors = append(errors, ValidationError{Code: ValidationAssigneeRequired, Message: "审批人类型无效", NodeID: node.ID})
		}
	}
	if node.ApprovalMode == ApprovalModeCountersign && (node.CompletionRate < 1 || node.CompletionRate > 100) {
		errors = append(errors, ValidationError{Code: ValidationCompletionRate, Message: "会签通过比例必须在 1 到 100 之间", NodeID: node.ID})
	}
	return errors
}

func validateGateway(node Node, incoming, outgoing []Edge) []ValidationError {
	errors := make([]ValidationError, 0)
	if node.GatewayMode == GatewayModeSplit && len(outgoing) < 2 {
		errors = append(errors, ValidationError{Code: ValidationBranchCount, Message: "分支网关至少需要两条分支", NodeID: node.ID})
	}
	if node.GatewayMode == GatewayModeJoin && len(incoming) < 2 {
		errors = append(errors, ValidationError{Code: ValidationBranchCount, Message: "汇聚网关至少需要两条进入连线", NodeID: node.ID})
	}
	if node.Type != NodeTypeExclusive || node.GatewayMode != GatewayModeSplit {
		return errors
	}
	defaultCount := 0
	for _, edge := range outgoing {
		if edge.Default {
			defaultCount++
			continue
		}
		if edge.Condition == nil {
			errors = append(errors, ValidationError{Code: ValidationBranchConditionRequired, Message: "条件分支必须配置条件或设为默认分支", NodeID: node.ID, EdgeID: edge.ID})
			continue
		}
		if !validCondition(*edge.Condition) {
			errors = append(errors, ValidationError{Code: ValidationConditionInvalid, Message: "条件表达式不完整", NodeID: node.ID, EdgeID: edge.ID})
		}
	}
	if defaultCount > 1 {
		errors = append(errors, ValidationError{Code: ValidationMultipleDefaultBranches, Message: "条件网关只能有一条默认分支", NodeID: node.ID})
	}
	return errors
}

func validCondition(condition Condition) bool {
	if strings.TrimSpace(condition.Field) == "" || condition.Value == nil {
		return false
	}
	switch condition.Operator {
	case ConditionEQ, ConditionNE, ConditionGT, ConditionGTE, ConditionLT, ConditionLTE:
		return true
	default:
		return false
	}
}

func visitForward(start string, outgoing map[string][]Edge) map[string]bool {
	visited := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, edge := range outgoing[current] {
			stack = append(stack, edge.Target)
		}
	}
	return visited
}

func visitBackward(end string, incoming map[string][]Edge) map[string]bool {
	visited := map[string]bool{}
	stack := []string{end}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, edge := range incoming[current] {
			stack = append(stack, edge.Source)
		}
	}
	return visited
}
