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
		return orderedUniqueAssignees(strings.Split(assignee.Value, ",")), nil
	case workflowcore.AssigneeTypeVariable:
		return variableAssignees(request.Variables, assignee.Value), nil
	case workflowcore.AssigneeTypeManager:
		return resolver.resolveManager(context.Background(), request, assignee.Value)
	case workflowcore.AssigneeTypeDepartmentLeader:
		return resolver.resolveDepartmentLeader(context.Background(), request, assignee.Value)
	case workflowcore.AssigneeTypeRole:
		return resolver.resolveRoles(context.Background(), assignee.Value)
	case workflowcore.AssigneeTypeOrgIdentity:
		return resolver.resolveOrgIdentity(context.Background(), request, assignee.Value)
	default:
		return nil, fmt.Errorf("不支持的审批人类型 %s", assignee.Type)
	}
}

func (resolver *AssigneeResolver) resolveManager(ctx context.Context, request workflowdomain.AssigneeRequest, rawKey string) ([]string, error) {
	for _, key := range managerVariableKeys(rawKey) {
		if values := variableAssignees(request.Variables, key); len(values) > 0 {
			return values, nil
		}
	}
	if resolver == nil || resolver.db == nil {
		return nil, nil
	}
	starterID := parsePositiveUint(request.Instance.StarterID)
	if starterID == 0 {
		return nil, nil
	}
	var starter struct {
		ManagerUserID uint `gorm:"column:manager_user_id"`
	}
	err := resolver.db.WithContext(ctx).Table("users").
		Select("manager_user_id").
		Where("id = ? AND user_status = 1", starterID).
		Take(&starter).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if starter.ManagerUserID == 0 {
		return nil, nil
	}
	var count int64
	if err := resolver.db.WithContext(ctx).Table("users").
		Where("id = ? AND user_status = 1", starter.ManagerUserID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	return []string{strconv.FormatUint(uint64(starter.ManagerUserID), 10)}, nil
}

func managerVariableKeys(rawKey string) []string {
	key := strings.TrimSpace(rawKey)
	candidates := make([]string, 0, 2)
	if key != "" {
		candidates = append(candidates, key)
	}
	if key == "" || key == "direct_manager" {
		candidates = append(candidates, "managerId")
	}
	return orderedUniqueAssignees(candidates)
}

func (resolver *AssigneeResolver) resolveDepartmentLeader(ctx context.Context, request workflowdomain.AssigneeRequest, rawKey string) ([]string, error) {
	key := strings.TrimSpace(rawKey)
	if key != "" {
		if values := variableAssignees(request.Variables, key); len(values) > 0 {
			return values, nil
		}
	}
	if key == "" || key == "current_department" {
		if values := variableAssignees(request.Variables, "departmentLeaderId"); len(values) > 0 {
			return values, nil
		}
	}
	if resolver == nil || resolver.db == nil {
		return nil, nil
	}
	return resolver.resolveOrgIdentity(ctx, request, "starter_department:department_leader")
}

func (resolver *AssigneeResolver) resolveOrgIdentity(ctx context.Context, request workflowdomain.AssigneeRequest, raw string) ([]string, error) {
	if resolver == nil || resolver.db == nil {
		return nil, errors.New("组织审批身份解析器未配置数据库")
	}
	scope, departmentID, identityCode := parseOrgIdentityValue(raw)
	if identityCode == "" {
		return nil, nil
	}
	departmentIDs := make([]uint, 0, 1)
	switch scope {
	case "starter_department":
		ids, err := resolver.starterDepartmentIDs(ctx, request.Instance.StarterID)
		if err != nil {
			return nil, err
		}
		departmentIDs = ids
	case "department":
		if departmentID > 0 {
			departmentIDs = append(departmentIDs, departmentID)
		}
	default:
		return nil, fmt.Errorf("不支持的组织审批身份范围 %s", scope)
	}
	if len(departmentIDs) == 0 {
		return nil, nil
	}
	var userIDs []uint
	if err := resolver.db.WithContext(ctx).Table("workflow_org_approver_assignments AS a").
		Joins("JOIN workflow_org_approver_identities AS i ON i.identity_code = a.identity_code AND i.identity_status = 1").
		Joins("JOIN users AS u ON u.id = a.user_id AND u.user_status = 1").
		Where("a.department_id IN ? AND a.identity_code = ? AND a.assignment_status = 1", departmentIDs, identityCode).
		Order("a.department_id ASC, a.assignment_sort ASC, a.id ASC").
		Pluck("a.user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return uintIDsToStringsOrdered(userIDs), nil
}

func parseOrgIdentityValue(raw string) (string, uint, string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", 0, ""
	}
	parts := strings.Split(value, ":")
	if len(parts) == 1 {
		return "starter_department", 0, strings.TrimSpace(parts[0])
	}
	scope := strings.TrimSpace(parts[0])
	switch scope {
	case "starter_department":
		if len(parts) < 2 {
			return scope, 0, ""
		}
		return scope, 0, strings.TrimSpace(parts[1])
	case "department":
		if len(parts) < 3 {
			return scope, 0, ""
		}
		return scope, parsePositiveUint(parts[1]), strings.TrimSpace(parts[2])
	default:
		return scope, 0, ""
	}
}

func (resolver *AssigneeResolver) starterDepartmentIDs(ctx context.Context, starterID string) ([]uint, error) {
	userID := parsePositiveUint(starterID)
	if userID == 0 {
		return nil, nil
	}
	var departmentIDs []uint
	if err := resolver.db.WithContext(ctx).Table("user_depts").
		Where("user_dept_user_id = ?", userID).
		Order("id ASC").
		Pluck("user_dept_dept_id", &departmentIDs).Error; err != nil {
		return nil, err
	}
	return normalizeUintIDs(departmentIDs), nil
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

func parsePositiveUint(value string) uint {
	id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || id == 0 {
		return 0
	}
	return uint(id)
}

func normalizeUintIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uintIDsToStringsOrdered(values []uint) []string {
	result := make([]string, 0, len(values))
	for _, value := range normalizeUintIDs(values) {
		result = append(result, strconv.FormatUint(uint64(value), 10))
	}
	return result
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

func orderedUniqueAssignees(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
