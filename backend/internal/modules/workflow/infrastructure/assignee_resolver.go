package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
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
	case workflowcore.AssigneeTypeInitiator:
		return orderedUniqueAssignees([]string{request.Instance.StarterID}), nil
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

func (resolver *AssigneeResolver) ResolveDisplayNames(ctx context.Context, request workflowdomain.AssigneeRequest) ([]string, error) {
	assigneeIDs, err := resolver.Resolve(request)
	if err != nil || len(assigneeIDs) == 0 {
		return nil, err
	}
	if resolver == nil || resolver.db == nil {
		return nil, nil
	}
	userIDs := make([]uint, 0, len(assigneeIDs))
	for _, assigneeID := range assigneeIDs {
		if userID := parsePositiveUint(assigneeID); userID > 0 {
			userIDs = append(userIDs, userID)
		}
	}
	if len(userIDs) == 0 {
		return nil, nil
	}
	var users []model.User
	if err := resolver.db.WithContext(ctx).Select("id", "user_name", "user_account").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	return resolvedAssigneeDisplayNames(assigneeIDs, users), nil
}

func resolvedAssigneeDisplayNames(assigneeIDs []string, users []model.User) []string {
	labels := make(map[uint]string, len(users))
	for _, user := range users {
		labels[user.ID] = firstPublishedLabel(user.Name, user.Account)
	}
	result := make([]string, 0, len(assigneeIDs))
	seen := make(map[uint]struct{}, len(assigneeIDs))
	for _, assigneeID := range assigneeIDs {
		userID := parsePositiveUint(assigneeID)
		if userID == 0 {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		if label := strings.TrimSpace(labels[userID]); label != "" {
			result = append(result, label)
		}
	}
	return result
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
	var relation struct {
		ManagerUserID uint `gorm:"column:manager_user_id"`
	}
	now := database.Now()
	err := resolver.db.WithContext(ctx).Table("user_reporting_relations AS r").
		Select("r.manager_user_id").
		Joins("JOIN users AS employee ON employee.id = r.employee_user_id AND employee.user_status = 1").
		Joins("JOIN users AS manager ON manager.id = r.manager_user_id AND manager.user_status = 1").
		Where("r.employee_user_id = ? AND r.relation_type = ? AND r.relation_status = ?", starterID, model.ReportingRelationTypeDirect, model.ReportingRelationStatusOn).
		Where("(r.effective_from = 0 OR r.effective_from <= ?) AND (r.effective_to = 0 OR r.effective_to > ?)", now, now).
		Order("r.is_primary DESC, r.relation_sort ASC, r.id ASC").
		Take(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []string{strconv.FormatUint(uint64(relation.ManagerUserID), 10)}, nil
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
		starterID := parsePositiveUint(request.Instance.StarterID)
		if starterID > 0 {
			userIDs, err := resolver.resolveOrgIdentityAssignments(ctx, model.OrgApproverSubjectTypeUser, []uint{starterID}, identityCode)
			if err != nil {
				return nil, err
			}
			if len(userIDs) > 0 {
				return userIDs, nil
			}
		}
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
	return resolver.resolveOrgIdentityAssignments(ctx, model.OrgApproverSubjectTypeDepartment, departmentIDs, identityCode)
}

func (resolver *AssigneeResolver) resolveOrgIdentityAssignments(ctx context.Context, subjectType string, subjectIDs []uint, identityCode string) ([]string, error) {
	if len(subjectIDs) == 0 {
		return nil, nil
	}
	var userIDs []uint
	if err := resolver.db.WithContext(ctx).Table("workflow_org_approver_assignments AS a").
		Joins("JOIN workflow_org_approver_identities AS i ON i.identity_code = a.identity_code AND i.identity_status = 1").
		Joins("JOIN users AS u ON u.id = a.user_id AND u.user_status = 1").
		Where("a.subject_type = ? AND a.subject_id IN ? AND a.identity_code = ? AND a.assignment_status = 1", subjectType, subjectIDs, identityCode).
		Order("a.subject_id ASC, a.assignment_sort ASC, a.id ASC").
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
