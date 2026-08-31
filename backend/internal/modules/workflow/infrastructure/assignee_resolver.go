package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowcore "wecheckin/backend/internal/workflow"
)

type AssigneeResolver struct {
	db *gorm.DB
}

func NewAssigneeResolver(db *gorm.DB) *AssigneeResolver {
	return &AssigneeResolver{db: db}
}

func (resolver *AssigneeResolver) Resolve(request workflowdomain.AssigneeRequest) ([]string, error) {
	if request.Node.Assignee == nil {
		return nil, workflowdomain.ErrAssigneeUnavailable
	}
	assignee := request.Node.Assignee
	switch assignee.Type {
	case workflowcore.AssigneeTypeUser:
		return normalizeAssignees(strings.Split(assignee.Value, ",")), nil
	case workflowcore.AssigneeTypeVariable:
		return variableAssignees(request.Variables, assignee.Value), nil
	case workflowcore.AssigneeTypeManager:
		key := strings.TrimSpace(assignee.Value)
		if key == "" {
			key = "managerId"
		}
		return variableAssignees(request.Variables, key), nil
	case workflowcore.AssigneeTypeDepartmentLeader:
		key := strings.TrimSpace(assignee.Value)
		if key == "" {
			key = "departmentLeaderId"
		}
		return variableAssignees(request.Variables, key), nil
	case workflowcore.AssigneeTypeRole:
		return resolver.resolveRoles(context.Background(), assignee.Value)
	default:
		return nil, fmt.Errorf("不支持的审批人类型 %s", assignee.Type)
	}
}

func (resolver *AssigneeResolver) resolveRoles(ctx context.Context, rawRoleIDs string) ([]string, error) {
	if resolver == nil || resolver.db == nil {
		return nil, errors.New("角色审批人解析器未配置数据库")
	}
	roleIDs := make([]uint, 0)
	for _, value := range strings.Split(rawRoleIDs, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		roleIDs = append(roleIDs, uint(id))
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}

	var activeRoleIDs []uint
	if err := resolver.db.WithContext(ctx).Table("roles").Where("id IN ? AND role_status = 1", roleIDs).Pluck("id", &activeRoleIDs).Error; err != nil {
		return nil, err
	}
	if len(activeRoleIDs) == 0 {
		return nil, nil
	}

	var relationUserIDs []uint
	if err := resolver.db.WithContext(ctx).Table("user_roles").
		Where("user_role_role_id IN ? AND user_role_status = 1", activeRoleIDs).
		Pluck("user_role_user_id", &relationUserIDs).Error; err != nil {
		return nil, err
	}
	var legacyUserIDs []uint
	if err := resolver.db.WithContext(ctx).Table("users").
		Where("user_role_id IN ? AND user_status = 1", activeRoleIDs).
		Pluck("id", &legacyUserIDs).Error; err != nil {
		return nil, err
	}
	userIDs := append(relationUserIDs, legacyUserIDs...)
	if len(userIDs) == 0 {
		return nil, nil
	}
	var activeUserIDs []uint
	if err := resolver.db.WithContext(ctx).Table("users").
		Where("id IN ? AND user_status = 1", userIDs).
		Order("id ASC").Pluck("id", &activeUserIDs).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(activeUserIDs))
	for _, id := range activeUserIDs {
		result = append(result, strconv.FormatUint(uint64(id), 10))
	}
	return normalizeAssignees(result), nil
}

func variableAssignees(variables map[string]interface{}, key string) []string {
	if variables == nil {
		return nil
	}
	value, ok := variables[strings.TrimSpace(key)]
	if !ok || value == nil {
		return nil
	}
	values := make([]string, 0)
	switch typed := value.(type) {
	case string:
		values = append(values, strings.Split(typed, ",")...)
	case []string:
		values = append(values, typed...)
	case []interface{}:
		for _, item := range typed {
			values = append(values, scalarString(item))
		}
	default:
		values = append(values, scalarString(typed))
	}
	return normalizeAssignees(values)
}

func scalarString(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	default:
		return fmt.Sprint(typed)
	}
}

func normalizeAssignees(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
