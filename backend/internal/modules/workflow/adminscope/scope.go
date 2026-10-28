package adminscope

import (
	"context"
	"sort"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type Scope struct {
	Visibility     workflowapp.InstanceVisibility
	AllDefinitions bool
	DefinitionIDs  []uint
}

func (scope Scope) AllowsDefinition(definitionID uint) bool {
	if definitionID == 0 {
		return false
	}
	if scope.AllDefinitions {
		return true
	}
	for _, id := range scope.DefinitionIDs {
		if id == definitionID {
			return true
		}
	}
	return false
}

func (scope Scope) ApplyDefinitionQuery(query *gorm.DB) *gorm.DB {
	if query == nil || scope.AllDefinitions {
		return query
	}
	if len(scope.DefinitionIDs) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("workflow_definitions.id IN ?", scope.DefinitionIDs)
}

type Resolver struct {
	db *gorm.DB
}

func NewResolver(db *gorm.DB) *Resolver {
	return &Resolver{db: db}
}

func (resolver *Resolver) ResolveByAdminID(ctx context.Context, adminID uint) (Scope, error) {
	if resolver == nil || resolver.db == nil || adminID == 0 {
		return Scope{}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	var admin model.Admin
	if err := resolver.db.WithContext(queryCtx).First(&admin, adminID).Error; err != nil {
		return Scope{}, err
	}
	return resolver.resolve(queryCtx, resolver.db.WithContext(queryCtx), &admin)
}

func (resolver *Resolver) Resolve(ctx context.Context, admin *model.Admin) (Scope, error) {
	if resolver == nil || resolver.db == nil || admin == nil || admin.ID == 0 {
		return Scope{}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	return resolver.resolve(queryCtx, resolver.db.WithContext(queryCtx), admin)
}

func (resolver *Resolver) resolve(ctx context.Context, db *gorm.DB, admin *model.Admin) (Scope, error) {
	roleIDs := append([]uint(nil), admin.RoleIDs...)
	if len(roleIDs) == 0 {
		var err error
		roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
		if err != nil {
			return Scope{}, err
		}
	}
	if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, db, roleIDs) {
		return Scope{
			Visibility:     workflowapp.InstanceVisibility{Ready: true, All: true},
			AllDefinitions: true,
		}, nil
	}

	scope, extras, err := permissionsupport.DataScopeBundleWithRoleIDsContext(ctx, db, admin.ID, roleIDs)
	if err != nil {
		return Scope{}, err
	}
	needDepartments := (scope.Ready && (scope.Mode == 2 || scope.Mode == 4)) || (extras.Ready && len(extras.DeptIDs) > 0)
	var departments []*model.Department
	if needDepartments {
		if err := db.Find(&departments).Error; err != nil {
			return Scope{}, err
		}
	}
	var ownDepartmentIDs []uint
	if scope.Ready && scope.Mode == 2 {
		var relations []model.UserDept
		if err := db.Where("user_dept_user_id = ?", admin.ID).Find(&relations).Error; err != nil {
			return Scope{}, err
		}
		for _, relation := range relations {
			ownDepartmentIDs = append(ownDepartmentIDs, relation.DeptID)
		}
	}

	visibility := buildVisibility(admin.ID, scope, extras, ownDepartmentIDs, departments)
	result := Scope{Visibility: visibility, AllDefinitions: visibility.Ready && visibility.All}
	if result.AllDefinitions || !visibility.Ready {
		return result, nil
	}
	definitionIDs, err := visibleDefinitionIDs(db, visibility)
	if err != nil {
		return Scope{}, err
	}
	result.DefinitionIDs = definitionIDs
	return result, nil
}

func buildVisibility(
	adminID uint,
	scope permissionsupport.DataScope,
	extras permissionsupport.DataScopeExtras,
	adminDepartmentIDs []uint,
	departments []*model.Department,
) workflowapp.InstanceVisibility {
	visibility := workflowapp.InstanceVisibility{Ready: scope.Ready || extras.Ready}
	if scope.Ready {
		switch scope.Mode {
		case 1:
			visibility.All = true
			return visibility
		case 2:
			visibility.DepartmentIDs = access.DeptDescendantIDs(departments, adminDepartmentIDs)
		case 3:
			visibility.UserIDs = append(visibility.UserIDs, adminID)
		case 4:
			visibility.DepartmentIDs = access.DeptDescendantIDs(departments, scope.DeptIDs)
		}
	}
	if extras.Ready {
		visibility.UserIDs = append(visibility.UserIDs, extras.UserIDs...)
		visibility.DepartmentIDs = append(visibility.DepartmentIDs, access.DeptDescendantIDs(departments, extras.DeptIDs)...)
	}
	visibility.UserIDs = sortedUintIDs(visibility.UserIDs)
	visibility.DepartmentIDs = sortedUintIDs(visibility.DepartmentIDs)
	return visibility
}

func visibleDefinitionIDs(db *gorm.DB, visibility workflowapp.InstanceVisibility) ([]uint, error) {
	if !visibility.Ready || visibility.All {
		return nil, nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if len(visibility.UserIDs) > 0 {
		clauses = append(clauses, "definition_add_user_id IN ?")
		args = append(args, visibility.UserIDs)
	}
	if len(visibility.DepartmentIDs) > 0 {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM user_depts definition_creator_dept
			WHERE definition_creator_dept.user_dept_user_id = workflow_definitions.definition_add_user_id
			AND definition_creator_dept.user_dept_dept_id IN ?
		)`)
		args = append(args, visibility.DepartmentIDs)
	}
	if len(clauses) == 0 {
		return []uint{}, nil
	}
	var ids []uint
	if err := db.Model(&model.WorkflowDefinition{}).
		Where("definition_deleted_at = ?", int64(0)).
		Where("("+joinOR(clauses)+")", args...).
		Order("id ASC").
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func joinOR(clauses []string) string {
	result := ""
	for index, clause := range clauses {
		if index > 0 {
			result += " OR "
		}
		result += clause
	}
	return result
}

func sortedUintIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
