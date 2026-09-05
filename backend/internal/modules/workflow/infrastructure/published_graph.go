package infrastructure

import (
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

type publishedAssigneeLabels struct {
	users       map[uint]string
	roles       map[uint]string
	identities  map[string]string
	departments map[uint]string
}

func loadPublishedAssigneeLabels(db *gorm.DB, nodes []workflowcore.Node) (publishedAssigneeLabels, error) {
	labels := publishedAssigneeLabels{
		users:       map[uint]string{},
		roles:       map[uint]string{},
		identities:  map[string]string{},
		departments: map[uint]string{},
	}
	userIDs := map[uint]struct{}{}
	roleIDs := map[uint]struct{}{}
	identityCodes := map[string]struct{}{}
	departmentIDs := map[uint]struct{}{}
	for _, node := range nodes {
		if node.Assignee == nil {
			continue
		}
		switch node.Assignee.Type {
		case workflowcore.AssigneeTypeUser:
			collectPublishedIDs(userIDs, node.Assignee.Value)
		case workflowcore.AssigneeTypeRole:
			collectPublishedIDs(roleIDs, node.Assignee.Value)
		case workflowcore.AssigneeTypeOrgIdentity:
			_, departmentID, identityCode := parseOrgIdentityValue(node.Assignee.Value)
			if identityCode != "" {
				identityCodes[identityCode] = struct{}{}
			}
			if departmentID > 0 {
				departmentIDs[departmentID] = struct{}{}
			}
		}
	}

	if ids := sortedPublishedIDs(userIDs); len(ids) > 0 {
		var rows []model.User
		if err := db.Select("id", "user_name", "user_account").Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return labels, err
		}
		for _, row := range rows {
			labels.users[row.ID] = firstPublishedLabel(row.Name, row.Account)
		}
	}
	if ids := sortedPublishedIDs(roleIDs); len(ids) > 0 {
		var rows []model.Role
		if err := db.Select("id", "role_name").Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return labels, err
		}
		for _, row := range rows {
			labels.roles[row.ID] = strings.TrimSpace(row.Name)
		}
	}
	if codes := sortedPublishedStrings(identityCodes); len(codes) > 0 {
		var rows []workflowmodel.OrgApproverIdentity
		if err := db.Select("identity_code", "identity_name").Where("identity_code IN ?", codes).Find(&rows).Error; err != nil {
			return labels, err
		}
		for _, row := range rows {
			labels.identities[row.Code] = strings.TrimSpace(row.Name)
		}
	}
	if ids := sortedPublishedIDs(departmentIDs); len(ids) > 0 {
		var rows []model.Department
		if err := db.Select("id", "dept_name").Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return labels, err
		}
		for _, row := range rows {
			labels.departments[row.ID] = strings.TrimSpace(row.Name)
		}
	}
	return labels, nil
}

func buildPublishedWorkflowGraph(definition workflowcore.Definition, labels publishedAssigneeLabels) ([]application.PublishedNode, []application.PublishedEdge) {
	nodes := make([]application.PublishedNode, 0, len(definition.Nodes))
	for _, node := range definition.Nodes {
		var position *workflowcore.Position
		if node.Position != nil {
			value := *node.Position
			position = &value
		}
		nodes = append(nodes, application.PublishedNode{
			ID: node.ID, Type: node.Type, Name: node.Name, Position: position,
			ApprovalMode: node.ApprovalMode, GatewayMode: node.GatewayMode,
			AssigneeDisplay: publishedAssigneeDisplay(node.Assignee, labels),
			Assignee:        clonePublishedAssignee(node.Assignee),
			PostHandleEdit:  clonePostHandleEdit(node.PostHandleEdit),
		})
	}
	edges := make([]application.PublishedEdge, 0, len(definition.Edges))
	for _, edge := range definition.Edges {
		edges = append(edges, application.PublishedEdge{
			ID: edge.ID, Source: edge.Source, Target: edge.Target,
			SourceHandle: edge.SourceHandle, TargetHandle: edge.TargetHandle,
			Name: edge.Name, Default: edge.Default, Condition: clonePublishedCondition(edge.Condition),
		})
	}
	return nodes, edges
}

func clonePostHandleEdit(config *workflowcore.PostHandleEditConfig) *workflowcore.PostHandleEditConfig {
	if config == nil {
		return nil
	}
	value := *config
	return &value
}

func clonePublishedCondition(condition *workflowcore.Condition) *workflowcore.Condition {
	if condition == nil {
		return nil
	}
	value := *condition
	return &value
}

func clonePublishedAssignee(assignee *workflowcore.Assignee) *workflowcore.Assignee {
	if assignee == nil {
		return nil
	}
	value := *assignee
	return &value
}

func publishedAssigneeDisplay(assignee *workflowcore.Assignee, labels publishedAssigneeLabels) string {
	if assignee == nil {
		return ""
	}
	switch assignee.Type {
	case workflowcore.AssigneeTypeInitiator:
		return "发起人"
	case workflowcore.AssigneeTypeUser:
		if names := publishedIDLabels(assignee.Value, labels.users); len(names) > 0 {
			return strings.Join(names, "、")
		}
		return "指定用户"
	case workflowcore.AssigneeTypeRole:
		if names := publishedIDLabels(assignee.Value, labels.roles); len(names) > 0 {
			return "系统角色：" + strings.Join(names, "、")
		}
		return "指定系统角色"
	case workflowcore.AssigneeTypeDepartmentLeader:
		return "发起人部门负责人"
	case workflowcore.AssigneeTypeManager:
		return "发起人的直属上级"
	case workflowcore.AssigneeTypeVariable:
		return "由流程变量动态指定"
	case workflowcore.AssigneeTypeOrgIdentity:
		scope, departmentID, identityCode := parseOrgIdentityValue(assignee.Value)
		identityName := firstPublishedLabel(labels.identities[identityCode], "组织审批身份")
		if scope == "department" {
			return firstPublishedLabel(labels.departments[departmentID], "指定部门") + " · " + identityName
		}
		return "发起人部门 · " + identityName
	default:
		return "按流程配置确定"
	}
}

func collectPublishedIDs(target map[uint]struct{}, raw string) {
	for _, value := range strings.Split(raw, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err == nil && id > 0 {
			target[uint(id)] = struct{}{}
		}
	}
}

func publishedIDLabels(raw string, labels map[uint]string) []string {
	result := make([]string, 0)
	seen := map[string]struct{}{}
	for _, value := range strings.Split(raw, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		label := strings.TrimSpace(labels[uint(id)])
		if label == "" {
			continue
		}
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		result = append(result, label)
	}
	return result
}

func sortedPublishedIDs(values map[uint]struct{}) []uint {
	result := make([]uint, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sortedPublishedStrings(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func firstPublishedLabel(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}
