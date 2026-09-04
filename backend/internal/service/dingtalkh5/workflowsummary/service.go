package workflowsummary

import (
	"context"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/support/access"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

const MaxPageSize = 50
const MaxExportInstances = 50

var (
	ErrDefinitionRequired   = errors.New("流程定义不能为空")
	ErrInstanceRequired     = errors.New("流程实例不能为空")
	ErrExportInstancesEmpty = errors.New("请选择需要导出的流程实例")
	ErrExportInstancesMany  = errors.New("单次最多导出50个流程实例")
	ErrSummaryAccessDenied  = errors.New("流程实例不存在或无权访问")
)

type Runtime interface {
	ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error)
	GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error)
	ListInstances(context.Context, workflowapp.InstanceQuery) (*workflowapp.InstanceList, error)
	GetInstance(context.Context, string) (*workflowapp.InstanceDetail, error)
}

type AccessResolver interface {
	Resolve(context.Context, *model.DingTalkH5PerfUser) (workflowapp.InstanceVisibility, error)
}

type Service struct {
	runtime Runtime
	access  AccessResolver
}

type InstanceQuery struct {
	DefinitionID      uint
	DefinitionVersion int
	DefinitionName    string
	StarterName       string
	Status            string
	StartTimeFrom     int64
	StartTimeTo       int64
	EndTimeFrom       int64
	EndTimeTo         int64
	Page              int
	PageSize          int
}

type ExportRequest struct {
	DefinitionID uint
	InstanceIDs  []string
	Format       ExportFormat
}

func NewService(db *gorm.DB, runtime Runtime) *Service {
	return &Service{runtime: runtime, access: &permissionAccessResolver{db: db}}
}

func NewServiceWithAccess(runtime Runtime, resolver AccessResolver) *Service {
	return &Service{runtime: runtime, access: resolver}
}

func (service *Service) ListDefinitions(ctx context.Context) ([]workflowapp.PublishedDefinition, error) {
	if service == nil || service.runtime == nil {
		return nil, errors.New("流程汇总服务未初始化")
	}
	return service.runtime.ListPublishedDefinitions(ctx)
}

func (service *Service) GetDefinition(ctx context.Context, definitionID uint) (*workflowapp.PublishedDefinition, error) {
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if service == nil || service.runtime == nil {
		return nil, errors.New("流程汇总服务未初始化")
	}
	return service.runtime.GetPublishedDefinition(ctx, definitionID)
}

func (service *Service) ListInstances(ctx context.Context, user *model.DingTalkH5PerfUser, query InstanceQuery) (*workflowapp.InstanceList, error) {
	visibility, err := service.resolveVisibility(ctx, user)
	if err != nil {
		return nil, err
	}
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize != 50 {
		pageSize = 20
	}
	return service.runtime.ListInstances(ctx, workflowapp.InstanceQuery{
		DefinitionID:      query.DefinitionID,
		DefinitionVersion: query.DefinitionVersion,
		DefinitionName:    query.DefinitionName,
		StarterName:       query.StarterName,
		Status:            query.Status,
		StartTimeFrom:     query.StartTimeFrom,
		StartTimeTo:       query.StartTimeTo,
		EndTimeFrom:       query.EndTimeFrom,
		EndTimeTo:         query.EndTimeTo,
		Visibility:        &visibility,
		Page:              page,
		PageSize:          pageSize,
	})
}

func (service *Service) GetInstance(ctx context.Context, user *model.DingTalkH5PerfUser, instanceID string) (*workflowapp.InstanceDetail, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return nil, ErrInstanceRequired
	}
	instances, err := service.authorizedInstances(ctx, user, 0, []string{instanceID})
	if err != nil {
		return nil, err
	}
	if len(instances) != 1 {
		return nil, ErrSummaryAccessDenied
	}
	return service.runtime.GetInstance(ctx, instanceID)
}

func (service *Service) Export(ctx context.Context, user *model.DingTalkH5PerfUser, request ExportRequest) (*ExportResult, error) {
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	instanceIDs := normalizeStringIDs(request.InstanceIDs)
	if len(instanceIDs) == 0 {
		return nil, ErrExportInstancesEmpty
	}
	if len(instanceIDs) > MaxExportInstances {
		return nil, ErrExportInstancesMany
	}
	if !request.Format.Valid() {
		return nil, ErrExportFormatInvalid
	}
	instances, err := service.authorizedInstances(ctx, user, request.DefinitionID, instanceIDs)
	if err != nil {
		return nil, err
	}
	if len(instances) != len(instanceIDs) {
		return nil, ErrSummaryAccessDenied
	}
	details := make([]*workflowapp.InstanceDetail, 0, len(instances))
	for _, instance := range instances {
		detail, err := service.runtime.GetInstance(ctx, instance.ID)
		if err != nil {
			return nil, err
		}
		details = append(details, detail)
	}
	documents := make([]exportDocument, 0, len(details))
	for _, detail := range details {
		documents = append(documents, buildExportDocument(detail))
	}
	return renderWorkflowExport(documents, request.Format)
}

func (service *Service) authorizedInstances(ctx context.Context, user *model.DingTalkH5PerfUser, definitionID uint, instanceIDs []string) ([]workflowapp.InstanceSummary, error) {
	visibility, err := service.resolveVisibility(ctx, user)
	if err != nil {
		return nil, err
	}
	result, err := service.runtime.ListInstances(ctx, workflowapp.InstanceQuery{
		DefinitionID: definitionID,
		InstanceIDs:  instanceIDs,
		Visibility:   &visibility,
		Page:         1,
		PageSize:     MaxExportInstances,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.List, nil
}

func (service *Service) resolveVisibility(ctx context.Context, user *model.DingTalkH5PerfUser) (workflowapp.InstanceVisibility, error) {
	if service == nil || service.runtime == nil || service.access == nil {
		return workflowapp.InstanceVisibility{}, errors.New("流程汇总服务未初始化")
	}
	if user == nil || user.ID == 0 {
		return workflowapp.InstanceVisibility{}, ErrSummaryAccessDenied
	}
	return service.access.Resolve(ctx, user)
}

type permissionAccessResolver struct {
	db *gorm.DB
}

func (resolver *permissionAccessResolver) Resolve(ctx context.Context, user *model.DingTalkH5PerfUser) (workflowapp.InstanceVisibility, error) {
	if resolver == nil || resolver.db == nil || user == nil || user.ID == 0 {
		return workflowapp.InstanceVisibility{}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	db := resolver.db.WithContext(queryCtx)
	roleIDs := append([]uint(nil), user.RoleIDs...)
	if len(roleIDs) == 0 {
		var err error
		roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(queryCtx, db, user.ID, user.RoleID)
		if err != nil {
			return workflowapp.InstanceVisibility{}, err
		}
	}
	scope, extras, err := permissionsupport.DataScopeBundleWithRoleIDsContext(queryCtx, db, user.ID, roleIDs)
	if err != nil {
		return workflowapp.InstanceVisibility{}, err
	}
	needDepartments := (scope.Ready && (scope.Mode == 2 || scope.Mode == 4)) || (extras.Ready && len(extras.DeptIDs) > 0)
	var departments []*model.Department
	var userDepartmentIDs []uint
	if needDepartments {
		if err := db.Find(&departments).Error; err != nil {
			return workflowapp.InstanceVisibility{}, err
		}
	}
	if scope.Ready && scope.Mode == 2 {
		var relations []model.UserDept
		if err := db.Where("user_dept_user_id = ?", user.ID).Find(&relations).Error; err != nil {
			return workflowapp.InstanceVisibility{}, err
		}
		for _, relation := range relations {
			userDepartmentIDs = append(userDepartmentIDs, relation.DeptID)
		}
	}
	return buildInstanceVisibility(user.ID, scope, extras, userDepartmentIDs, departments), nil
}

func buildInstanceVisibility(
	userID uint,
	scope permissionsupport.DataScope,
	extras permissionsupport.DataScopeExtras,
	userDepartmentIDs []uint,
	departments []*model.Department,
) workflowapp.InstanceVisibility {
	visibility := workflowapp.InstanceVisibility{Ready: scope.Ready || extras.Ready}
	if scope.Ready {
		switch scope.Mode {
		case 1:
			visibility.All = true
			return visibility
		case 2:
			visibility.DepartmentIDs = access.DeptDescendantIDs(departments, userDepartmentIDs)
		case 3:
			visibility.UserIDs = append(visibility.UserIDs, userID)
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

func normalizeStringIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sortedUintIDs(values []uint) []uint {
	if len(values) == 0 {
		return nil
	}
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
