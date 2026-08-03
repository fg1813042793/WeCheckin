package dingtalkh5

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/appapiperm"
	"wecheckin/backend/internal/app/support/appmenuperm"
	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

const (
	nextObjectiveEditButtonKey   = "dingtalk_h5:button:review:next_objective_edit"
	nextObjectiveAddButtonKey    = "dingtalk_h5:button:review:next_objective_add"
	nextObjectiveDeleteButtonKey = "dingtalk_h5:button:review:next_objective_delete"
)

func BootstrapContext(ctx context.Context, user *model.DingTalkH5PerfUser) (*BootstrapResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return bootstrapForUserDB(ctx, db, user), nil
}

func bootstrapForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) *BootstrapResponse {
	snapshot, err := dingTalkH5PermissionSnapshotForUserDB(ctx, db, user)
	if err != nil {
		snapshot = dingTalkH5PermissionSnapshot{version: permissionVersionFallback(user)}
	}
	appConfig := dingTalkH5AppConfigContext(ctx)
	return &BootstrapResponse{
		User:                  userDTO(*user),
		Menus:                 dingTalkH5MenusByKeysWithLabelsAndIcons(snapshot.menuKeys, snapshot.labels, snapshot.icons),
		AppConfig:             appConfig,
		AppTitle:              appConfig.AppTitle,
		AppName:               appConfig.AppName,
		LogoText:              appConfig.LogoText,
		LogoURL:               appConfig.LogoURL,
		ButtonPermissionKeys:  snapshot.buttonKeys,
		ButtonPermissionReady: snapshot.buttonReady,
		APIPermissionKeys:     snapshot.apiKeys,
		APIPermissionReady:    snapshot.apiReady,
		PermissionVersion:     snapshot.version,
	}
}

func DingTalkH5MenusForUserContext(ctx context.Context, user *model.DingTalkH5PerfUser) []AppMenuDTO {
	if user == nil {
		return nil
	}
	if user.ID == 0 && user.RoleID == 0 && len(user.RoleIDs) == 0 {
		return nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return dingTalkH5MenusForUserDB(ctx, db, user)
}

func dingTalkH5MenusForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) []AppMenuDTO {
	if user == nil {
		return nil
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return nil
	}
	if db == nil || (user.ID == 0 && len(roleIDs) == 0) {
		return nil
	}
	if keys, ready, err := permissionsupport.DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx, db, user.ID, roleIDs); err == nil && ready {
		labels, icons := dingTalkH5MenuMetadataByKeysContext(ctx, db, keys)
		return dingTalkH5MenusByKeysWithLabelsAndIcons(keys, labels, icons)
	}
	return nil
}

func dingTalkH5APIPermissionKeysForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]string, bool) {
	if user == nil || db == nil {
		return nil, false
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return nil, false
	}
	ready, err := permissionsupport.SubjectAPIPermissionReadyWithRoleIDsContext(ctx, db, user.ID, roleIDs, permissionsupport.PlatformDingTalkH5)
	if err != nil || !ready {
		return nil, false
	}
	keys := make([]string, 0)
	for _, declaration := range appapiperm.DingTalkH5APIDeclarations() {
		allowed, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, user.ID, roleIDs, declaration.Key)
		if err == nil && allowed {
			keys = append(keys, declaration.Key)
		}
	}
	return keys, true
}

func dingTalkH5ButtonPermissionKeysForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]string, bool) {
	if user == nil || db == nil {
		return nil, false
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return nil, false
	}
	keys, ready, err := permissionsupport.DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx, db, user.ID, roleIDs)
	if err != nil {
		return nil, false
	}
	return keys, ready
}

func permissionVersionForUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (int64, error) {
	if user == nil {
		return 0, nil
	}
	version := permissionVersionFallback(user)
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return version, err
	}
	if db == nil || (user.ID == 0 && len(roleIDs) == 0) {
		return version, nil
	}
	var grantVersion int64
	query := db.Model(&model.PermissionGrant{}).
		Select("COALESCE(MAX(`grant_edit_time`), 0)").
		Where("`grant_status` = 1").
		Where(
			"(`grant_subject_type` = ? AND `grant_subject_id` = ?) OR (`grant_subject_type` = ? AND `grant_subject_id` IN ?)",
			permissionsupport.SubjectUser,
			user.ID,
			permissionsupport.SubjectRole,
			roleIDs,
		)
	if err := query.Scan(&grantVersion).Error; err != nil {
		return version, err
	}
	if grantVersion > version {
		version = grantVersion
	}
	menuVersion, err := dingTalkH5MenuPermissionEditVersionContext(ctx, db, user)
	if err != nil {
		return version, err
	}
	if menuVersion > version {
		version = menuVersion
	}
	return version, nil
}

func permissionVersionFallback(user *model.DingTalkH5PerfUser) int64 {
	if user == nil {
		return 0
	}
	version := user.EditTime
	if version == 0 {
		version = user.AddTime
	}
	if int64(user.RoleID) > version {
		version = int64(user.RoleID)
	}
	for _, roleID := range user.RoleIDs {
		if int64(roleID) > version {
			version = int64(roleID)
		}
	}
	if int64(user.ID) > version {
		version = int64(user.ID)
	}
	return version
}

func activeRoleIDsForPerfUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]uint, error) {
	if user == nil {
		return nil, nil
	}
	roleIDs := uniqueUintIDs(append([]uint{user.RoleID}, user.RoleIDs...))
	if len(roleIDs) > 0 || db == nil || user.ID == 0 {
		user.RoleIDs = roleIDs
		return roleIDs, nil
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}
	user.RoleIDs = roleIDs
	return roleIDs, nil
}

func dingTalkH5MenusByKeys(keys []string) []AppMenuDTO {
	return dingTalkH5MenusByKeysWithLabels(keys, nil)
}

func dingTalkH5MenusByKeysWithLabels(keys []string, labels map[string]string) []AppMenuDTO {
	return dingTalkH5MenusByKeysWithLabelsAndIcons(keys, labels, nil)
}

func dingTalkH5MenusByKeysWithLabelsAndIcons(keys []string, labels, icons map[string]string) []AppMenuDTO {
	allowed := dingTalkH5AllowedMenuKeySet(keys)
	declarations := appmenuperm.DingTalkH5MenuDeclarations()
	nodes := make(map[string]*AppMenuDTO, len(declarations))
	for _, declaration := range declarations {
		if !allowed[declaration.Key] {
			continue
		}
		node := AppMenuDTO{
			Key:           declaration.Path,
			Label:         dingTalkH5MenuLabel(declaration.Key, declaration.Name, labels),
			Icon:          dingTalkH5MenuIcon(declaration.Key, firstNonEmptyString(declaration.Icon, declaration.Path), icons),
			PermissionKey: declaration.Key,
		}
		nodes[declaration.Key] = &node
	}

	for _, declaration := range declarations {
		node := nodes[declaration.Key]
		if node == nil || declaration.ParentKey == "" {
			continue
		}
		if parent := nodes[declaration.ParentKey]; parent != nil {
			parent.Children = append(parent.Children, *node)
		}
	}

	menus := make([]AppMenuDTO, 0, len(nodes))
	for _, declaration := range declarations {
		node := nodes[declaration.Key]
		if node == nil {
			continue
		}
		if declaration.ParentKey != "" {
			if _, ok := nodes[declaration.ParentKey]; ok {
				continue
			}
		}
		menus = append(menus, *node)
	}
	return menus
}

func dingTalkH5AllowedMenuKeySet(keys []string) map[string]bool {
	allowed := map[string]bool{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			allowed[key] = true
		}
	}
	expandLegacyDingTalkH5MenuKeys(allowed)
	return allowed
}

func dingTalkH5MenuLabel(key, fallback string, labels map[string]string) string {
	if label := strings.TrimSpace(labels[key]); label != "" {
		return label
	}
	return fallback
}

func dingTalkH5MenuIcon(key, fallback string, icons map[string]string) string {
	if icon := strings.TrimSpace(icons[key]); icon != "" {
		return icon
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}

func dingTalkH5MenuLabelsByKeysContext(ctx context.Context, db *gorm.DB, keys []string) map[string]string {
	labels, _ := dingTalkH5MenuMetadataByKeysContext(ctx, db, keys)
	return labels
}

func dingTalkH5MenuMetadataByKeysContext(ctx context.Context, db *gorm.DB, keys []string) (map[string]string, map[string]string) {
	if db == nil || len(keys) == 0 {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, nil
	}
	allowed := dingTalkH5AllowedMenuKeySet(keys)
	if len(allowed) == 0 {
		return nil, nil
	}
	queryKeys := make([]string, 0, len(allowed))
	for key := range allowed {
		queryKeys = append(queryKeys, key)
	}
	var rows []model.Permission
	if err := db.WithContext(ctx).
		Select("`permission_key`, `permission_name`, `permission_icon`").
		Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_type` IN ? AND `permission_status` = 1", queryKeys, permissionsupport.PlatformDingTalkH5, []string{permissionsupport.TypeDirectory, permissionsupport.TypeMenu}).
		Find(&rows).Error; err != nil {
		return nil, nil
	}
	labels := make(map[string]string, len(rows))
	icons := make(map[string]string, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		name := strings.TrimSpace(row.Name)
		if key != "" && name != "" {
			labels[key] = name
		}
		if icon := strings.TrimSpace(row.Icon); key != "" && icon != "" {
			icons[key] = icon
		}
	}
	return labels, icons
}

func dingTalkH5MenuPermissionEditVersionContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (int64, error) {
	if user == nil || db == nil {
		return 0, nil
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return 0, err
	}
	keys, ready, err := permissionsupport.DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx, db, user.ID, roleIDs)
	if err != nil || !ready {
		return 0, err
	}
	buttonKeys, buttonReady, err := permissionsupport.DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx, db, user.ID, roleIDs)
	if err != nil {
		return 0, err
	}
	allowed := dingTalkH5AllowedMenuKeySet(keys)
	if buttonReady {
		for _, key := range buttonKeys {
			if strings.TrimSpace(key) != "" {
				allowed[key] = true
			}
		}
	}
	if len(allowed) == 0 {
		return 0, nil
	}
	queryKeys := make([]string, 0, len(allowed))
	for key := range allowed {
		queryKeys = append(queryKeys, key)
	}
	var version int64
	if err := db.WithContext(ctx).
		Model(&model.Permission{}).
		Select("COALESCE(MAX(`permission_edit_time`), 0)").
		Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_type` IN ? AND `permission_status` = 1", queryKeys, permissionsupport.PlatformDingTalkH5, []string{permissionsupport.TypeDirectory, permissionsupport.TypeMenu, permissionsupport.TypeButton}).
		Scan(&version).Error; err != nil {
		return 0, err
	}
	return version, nil
}

func expandLegacyDingTalkH5MenuKeys(allowed map[string]bool) {
	legacyMap := map[string]string{
		"dingtalk_h5:menu:mine":     "dingtalk_h5:menu:performance:mine",
		"dingtalk_h5:menu:manager":  "dingtalk_h5:menu:performance:manager",
		"dingtalk_h5:menu:hrbp":     "dingtalk_h5:menu:performance:hrbp",
		"dingtalk_h5:menu:summary":  "dingtalk_h5:menu:performance:summary",
		"dingtalk_h5:menu:org":      "dingtalk_h5:menu:performance:org",
		"dingtalk_h5:menu:template": "dingtalk_h5:menu:performance:template",
	}
	for legacy, current := range legacyMap {
		if allowed[legacy] {
			allowed[current] = true
		}
	}
	for key := range allowed {
		if strings.HasPrefix(key, "dingtalk_h5:menu:performance:") {
			allowed["dingtalk_h5:menu:performance"] = true
			return
		}
	}
}

func TemplateContext(ctx context.Context) (TemplateDTO, error) {
	return LoadTemplateContext(ctx)
}

func ListReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters) (*ReviewListResponse, error) {
	normalizeReviewPagination(&filters)
	return listReviewsContext(ctx, user, filters, true)
}

func listReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters, paginate bool) (*ReviewListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var reviews []model.DingTalkH5PerfReview
	query := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{}))
	if filters.Period != "" {
		query = query.Where("period = ?", filters.Period)
	}
	if filters.NotPeriod != "" {
		query = query.Where("period <> ?", filters.NotPeriod)
	}
	if filters.NextPeriod != "" {
		query = query.Where("next_period = ?", filters.NextPeriod)
	}
	statuses := normalizeReviewStatuses(filters.Status, filters.Statuses)
	if len(statuses) == 1 {
		query = query.Where("status = ?", statuses[0])
	} else if len(statuses) > 1 {
		query = query.Where("status IN ?", statuses)
	}
	if filters.EmployeeName != "" {
		query = applyReviewEmployeeNameQuery(query, filters.EmployeeName)
	}
	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
	}
	if len(filters.DepartmentNames) > 0 {
		query = applyReviewDepartmentNamesQuery(query, filters.DepartmentNames)
	} else if filters.DepartmentName != "" {
		query = query.Where("department LIKE ?", "%"+filters.DepartmentName+"%")
	}
	if filters.ManagerID != "" {
		query = query.Where("manager_account = ?", filters.ManagerID)
	}
	if filters.HRBPID != "" {
		query = query.Where("hrbp_account = ?", filters.HRBPID)
	}
	if filters.Grade != "" {
		query = query.Where("(final_grade = ? OR (final_grade = '' AND hrbp_grade = ?))", filters.Grade, filters.Grade)
	}
	query = applyReviewKeywordQuery(query, filters.Keyword)
	query, err := applyReviewVisibilityScopeContext(ctx, db, query, user, filters.Scope)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	query = query.Order("period DESC, id DESC")
	if paginate {
		query = query.Offset((filters.Page - 1) * filters.PageSize).Limit(filters.PageSize)
	}
	if err := query.Find(&reviews).Error; err != nil {
		return nil, err
	}
	participants, err := usersByAccounts(ctx, collectReviewParticipantAccounts(reviews))
	if err != nil {
		return nil, err
	}
	historiesByID := map[uint][]model.DingTalkH5PerfHistory{}
	if !filters.SkipHistory {
		var err error
		historiesByID, err = historiesByReviewIDs(ctx, collectReviewIDs(reviews))
		if err != nil {
			return nil, err
		}
	}
	valueTemplates := loadReviewValueTemplatesContext(ctx)
	result := make([]ReviewDTO, 0, len(reviews))
	for _, review := range reviews {
		dto := reviewDTOWithUsers(review, historiesByID[review.ID], participants)
		hydrateReviewDTOValues(&dto, valueTemplates)
		result = append(result, dto)
	}
	return &ReviewListResponse{List: result, Total: total, Page: filters.Page, PageSize: filters.PageSize}, nil
}

func normalizeReviewStatuses(status string, statuses []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(statuses)+1)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		result = append(result, value)
	}
	add(status)
	for _, item := range statuses {
		add(item)
	}
	return result
}

func applyReviewDepartmentNamesQuery(query *gorm.DB, names []string) *gorm.DB {
	normalized := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		text := strings.TrimSpace(name)
		if text == "" || seen[text] {
			continue
		}
		seen[text] = true
		normalized = append(normalized, text)
	}
	if len(normalized) == 0 {
		return query
	}
	if len(normalized) == 1 {
		return query.Where("department LIKE ?", "%"+normalized[0]+"%")
	}
	parts := make([]string, 0, len(normalized))
	args := make([]interface{}, 0, len(normalized))
	for _, name := range normalized {
		parts = append(parts, "department LIKE ?")
		args = append(args, "%"+name+"%")
	}
	return query.Where("("+strings.Join(parts, " OR ")+")", args...)
}

func normalizeReviewPagination(filters *ReviewFilters) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
}

func applyReviewKeywordQuery(query *gorm.DB, keyword string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return query
	}
	likeKeyword := "%" + keyword + "%"
	return query.Where(
		"`review_no` LIKE ? OR `employee_account` LIKE ? OR `manager_account` LIKE ? OR `hrbp_account` LIKE ? OR `department` LIKE ? OR `period` LIKE ? OR `next_period` LIKE ? OR `status` LIKE ? OR `manager_grade` LIKE ? OR `hrbp_grade` LIKE ? OR `final_grade` LIKE ? OR `employee_account` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)",
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
		likeKeyword,
	)
}

func applyReviewEmployeeNameQuery(query *gorm.DB, employeeName string) *gorm.DB {
	employeeName = strings.TrimSpace(employeeName)
	if employeeName == "" {
		return query
	}
	likeName := "%" + employeeName + "%"
	return query.Where(
		"`employee_account` LIKE ? OR `employee_account` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)",
		likeName,
		likeName,
	)
}

func GetReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) (*ReviewDTO, error) {
	review, err := findVisibleReview(ctx, user, reviewNo)
	if err != nil {
		return nil, err
	}
	histories, err := historiesForReview(ctx, review.ID)
	if err != nil {
		return nil, err
	}
	dto := reviewDTO(*review, histories)
	hydrateReviewDTOValues(&dto, loadReviewValueTemplatesContext(ctx))
	return &dto, nil
}

func CreateReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload) (*ReviewDTO, error) {
	resp, err := CreateReviewsContext(ctx, user, payload)
	if err != nil {
		return nil, err
	}
	if len(resp.List) == 0 {
		if len(resp.Failed) > 0 {
			return nil, fmt.Errorf("%s", resp.Failed[0].Message)
		}
		return nil, fmt.Errorf("考评单创建失败")
	}
	return &resp.List[0], nil
}

func CreateReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload) (*CreateReviewBatchResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("未登录")
	}
	period := strings.TrimSpace(payload.Period)
	nextPeriod := strings.TrimSpace(payload.NextPeriod)
	if !validMonth(period) || !validMonth(nextPeriod) {
		return nil, fmt.Errorf("月份格式应为 YYYY-MM")
	}
	employeeAccounts := reviewPayloadEmployeeIDs(payload)
	if len(employeeAccounts) == 0 {
		employeeAccounts = []string{NormalizeUserID(user.Account)}
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	tpl, err := LoadTemplateContext(ctx)
	if err != nil {
		return nil, err
	}
	accessScope, err := createReviewAccessScopeContext(ctx, db, user)
	if err != nil {
		return nil, err
	}

	result := &CreateReviewBatchResponse{
		List:   make([]ReviewDTO, 0, len(employeeAccounts)),
		Failed: make([]CreateReviewFailure, 0),
	}
	for _, employeeAccount := range employeeAccounts {
		dto, err := createReviewForEmployeeContext(ctx, db, user, accessScope, employeeAccount, period, nextPeriod, tpl)
		if err != nil {
			result.Failed = append(result.Failed, CreateReviewFailure{EmployeeID: employeeAccount, Message: err.Error()})
			continue
		}
		result.List = append(result.List, *dto)
	}
	result.Total = len(result.List)
	if result.Total == 0 {
		messages := make([]string, 0, len(result.Failed))
		for _, item := range result.Failed {
			messages = append(messages, item.Message)
		}
		if len(messages) == 0 {
			return nil, fmt.Errorf("请选择被考评人")
		}
		return nil, fmt.Errorf("%s", strings.Join(messages, "；"))
	}
	return result, nil
}

type createReviewAccessScope struct {
	allowed map[string]struct{}
	all     bool
}

func createReviewAccessScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (createReviewAccessScope, error) {
	if user == nil {
		return createReviewAccessScope{allowed: map[string]struct{}{}}, nil
	}
	scope, err := permissionsupport.DataScopeContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return createReviewAccessScope{}, err
	}
	allowed, all, err := dataScopeUserAccountsContext(ctx, db, user, scope)
	if err != nil {
		return createReviewAccessScope{}, err
	}
	if allowed == nil {
		allowed = map[string]struct{}{}
	}
	return createReviewAccessScope{allowed: allowed, all: all}, nil
}

func (scope createReviewAccessScope) canAccess(account string) bool {
	if scope.all {
		return true
	}
	_, ok := scope.allowed[NormalizeUserID(account)]
	return ok
}

func reviewPayloadEmployeeIDs(payload ReviewPayload) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(payload.EmployeeIDs)+1)
	add := func(value string) {
		account := NormalizeUserID(value)
		if account == "" || seen[account] {
			return
		}
		seen[account] = true
		result = append(result, account)
	}
	for _, id := range payload.EmployeeIDs {
		add(id)
	}
	add(payload.EmployeeID)
	return result
}

func createReviewForEmployeeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, accessScope createReviewAccessScope, employeeAccount, period, nextPeriod string, tpl TemplateDTO) (*ReviewDTO, error) {
	employee, err := loadPerfUserByAccountDB(db, employeeAccount)
	if err != nil || !canBeReviewed(*employee) {
		return nil, fmt.Errorf("请选择有效被考评人")
	}
	if !accessScope.canAccess(employee.Account) {
		return nil, fmt.Errorf("请选择有效被考评人")
	}
	reviewNo := employee.Account + "-" + period
	var count int64
	if err := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{})).Where("review_no = ?", reviewNo).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("该员工这个月份已经有考评单")
	}
	now := database.Now()
	ownerAudit := dingtalkH5AuditMetaForUserContext(ctx, db, employee, now)
	operatorAudit := dingtalkH5AuditMetaForUserContext(ctx, db, user, now)
	previousReview, err := loadPreviousNextObjectivesForCreate(ctx, db, employee.Account, period)
	if err != nil {
		return nil, err
	}
	objectives, objectiveSource := currentObjectivesForNewReview(reviewNo, tpl.ObjectiveDefaults, previousReview)
	review := model.DingTalkH5PerfReview{
		ReviewNo:                reviewNo,
		EmployeeAccount:         employee.Account,
		ManagerAccount:          employee.ManagerAccount,
		HRBPAccount:             fallback(employee.HRBPAccount, "hrbp"),
		Department:              departmentFromUser(*employee),
		DepartmentLevel1:        employee.DepartmentLevel1,
		DepartmentLevel2:        employee.DepartmentLevel2,
		DepartmentLevel3:        employee.DepartmentLevel3,
		Period:                  period,
		NextPeriod:              nextPeriod,
		Status:                  ReviewStatusDraft,
		ObjectiveSourceReviewNo: objectiveSource.reviewNo,
		ObjectiveSourcePeriod:   objectiveSource.period,
		ObjectivesJSON:          encodeJSON(objectives),
		NextObjectivesJSON:      encodeJSON(defaultNextObjectives(tpl.NextObjectiveDefaults, reviewNo)),
		ValuesJSON:              encodeJSON(defaultValues(tpl.Values)),
		AddTime:                 now,
		EditTime:                now,
	}
	applyDingTalkH5CreateAudit(&review.DingTalkH5AuditFields, ownerAudit)
	applyDingTalkH5UpdateAudit(&review.DingTalkH5AuditFields, operatorAudit)
	if err := db.Create(&review).Error; err != nil {
		return nil, err
	}
	if err := addHistoryWithDB(db, &review, user, "创建考评单"); err != nil {
		return nil, err
	}
	histories, _ := historiesForReview(ctx, review.ID)
	dto := reviewDTO(review, histories)
	hydrateReviewDTOValues(&dto, tpl.Values)
	return &dto, nil
}

func loadPreviousNextObjectivesForCreate(ctx context.Context, db *gorm.DB, employeeAccount, period string) (*model.DingTalkH5PerfReview, error) {
	if db == nil {
		return nil, nil
	}
	var review model.DingTalkH5PerfReview
	err := db.WithContext(ctx).
		Scopes(func(tx *gorm.DB) *gorm.DB { return notDeletedReviewQuery(tx) }).
		Where("`employee_account` = ? AND `next_period` = ?", strings.TrimSpace(employeeAccount), strings.TrimSpace(period)).
		Order("`period` DESC, `id` DESC").
		First(&review).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(sanitizeNextObjectives(decodeNextObjectives(review.NextObjectivesJSON))) == 0 {
		return nil, nil
	}
	return &review, nil
}

func SaveSelfContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能修改员工自评")
		}
		if err := ensureNextObjectiveMutationPermissionsContext(ctx, db, user, review, payload); err != nil {
			return err
		}
		copySelfFields(review, payload)
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "保存员工自评")
	})
}

func SubmitSelfContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	var submittedReview model.DingTalkH5PerfReview
	result, err := mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusDraft {
			return fmt.Errorf("当前阶段不能提交员工自评")
		}
		if err := validateSelfSubmitPayload(payload); err != nil {
			return err
		}
		if err := ensureNextObjectiveMutationPermissionsContext(ctx, db, user, review, payload); err != nil {
			return err
		}
		copySelfFields(review, payload)
		review.Status = nextStatusAfterSelfSubmit(DingTalkH5Review{EmployeeID: review.EmployeeAccount, ManagerID: review.ManagerAccount, HRBPID: review.HRBPAccount, Status: review.Status})
		if review.Status == ReviewStatusHRFinal {
			review.HRBPReviewerAccount = review.HRBPAccount
		}
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		action := "提交员工自评"
		if review.Status == ReviewStatusHRFinal {
			action = "提交员工自评，进入 HRBP 归档"
		}
		submittedReview = *review
		return addHistoryWithDB(db, review, user, action)
	})
	if err != nil {
		return nil, err
	}
	notifyReviewTransitionAsync(ctx, submittedReview, user, dingtalkH5NotifyEventSelfSubmitted)
	return result, nil
}

type nextObjectiveMutationSet struct {
	add    bool
	edit   bool
	delete bool
}

func ensureNextObjectiveMutationPermissionsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review *model.DingTalkH5PerfReview, payload ReviewPayload) error {
	changes := nextObjectiveMutations(
		sanitizeNextObjectives(decodeNextObjectives(review.NextObjectivesJSON)),
		sanitizeNextObjectives(payload.NextObjectives),
	)
	if !changes.add && !changes.edit && !changes.delete {
		return nil
	}
	if changes.edit {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveEditButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限编辑下月目标")
		}
	}
	if changes.add {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveAddButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限新增下月目标")
		}
	}
	if changes.delete {
		allowed, err := subjectHasDingTalkH5ButtonPermissionContext(ctx, db, user, nextObjectiveDeleteButtonKey)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("无权限删除下月目标")
		}
	}
	return nil
}

func nextObjectiveMutations(existing []NextObjective, incoming []NextObjective) nextObjectiveMutationSet {
	changes := nextObjectiveMutationSet{}
	existingByID := make(map[string]NextObjective, len(existing))
	for _, item := range existing {
		if key := strings.TrimSpace(item.ID); key != "" {
			existingByID[key] = item
		}
	}
	incomingByID := make(map[string]NextObjective, len(incoming))
	for _, item := range incoming {
		key := strings.TrimSpace(item.ID)
		if key == "" {
			changes.add = true
			continue
		}
		incomingByID[key] = item
		old, ok := existingByID[key]
		if !ok {
			changes.add = true
			continue
		}
		if strings.TrimSpace(old.Target) != strings.TrimSpace(item.Target) || old.Weight != item.Weight {
			changes.edit = true
		}
	}
	for _, item := range existing {
		key := strings.TrimSpace(item.ID)
		if key == "" {
			continue
		}
		if _, ok := incomingByID[key]; !ok {
			changes.delete = true
		}
	}
	return changes
}

func subjectHasDingTalkH5ButtonPermissionContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, key string) (bool, error) {
	if user == nil || db == nil {
		return false, nil
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return false, err
	}
	return permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, user.ID, roleIDs, key)
}

func SubmitManagerContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.ManagerAccount != user.Account || review.Status != ReviewStatusManagerReview {
			return fmt.Errorf("当前阶段不能提交上级评价")
		}
		copyManagerFields(review, payload)
		if review.ManagerGrade == "" {
			return fmt.Errorf("请先选择绩效分档")
		}
		if strings.TrimSpace(review.ManagerComment) == "" {
			return fmt.Errorf("请先填写上级评价")
		}
		if !allStageScoresFilled(decodeValues(review.ValuesJSON), "manager") {
			return fmt.Errorf("请先填写上级价值观评分")
		}
		if shouldSkipHrbpStage(ctx, *review) {
			review.Status = ReviewStatusEmployeeConfirm
			review.HRBPReviewerAccount = fallback(review.HRBPAccount, user.Account)
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmComment = ""
			review.EmployeeConfirmedAt = 0
			if err := db.Save(review).Error; err != nil {
				return err
			}
			return addHistoryWithDB(db, review, user, "提交上级评价，进入员工确认")
		}
		review.Status = ReviewStatusHRBPReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "提交上级评价")
	})
}

func SubmitHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能提交 HRBP 评价")
		}
		copyHrbpFields(review, payload)
		if review.HRBPGrade == "" {
			return fmt.Errorf("请先选择 HRBP绩效分档")
		}
		if review.ManagerGrade == "" {
			return fmt.Errorf("上级绩效分档为空，不能提交 HRBP 评价")
		}
		if review.HRBPGrade != review.ManagerGrade {
			return fmt.Errorf("HRBP绩效分档与上级绩效分档不一致，不能提交给员工确认")
		}
		if strings.TrimSpace(review.HRBPComment) == "" {
			return fmt.Errorf("请先填写 HRBP 评价")
		}
		if !allStageScoresFilled(decodeValues(review.ValuesJSON), "hrbp") {
			return fmt.Errorf("请先填写 HRBP 价值观评分")
		}
		review.Status = ReviewStatusEmployeeConfirm
		review.HRBPReviewerAccount = user.Account
		review.EmployeeConfirmResult = ""
		review.EmployeeConfirmComment = ""
		review.EmployeeConfirmedAt = 0
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "提交 HRBP 评价，进入员工确认")
	})
}

func ConfirmResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusEmployeeConfirm {
			return fmt.Errorf("当前阶段不能确认绩效结果")
		}
		review.EmployeeConfirmComment = strings.TrimSpace(payload.EmployeeConfirmComment)
		review.EmployeeConfirmResult = "confirmed"
		review.EmployeeConfirmedAt = database.Now()
		review.Status = ReviewStatusHRFinal
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		action := "员工确认结果"
		if review.EmployeeConfirmComment != "" {
			action = "员工确认结果：" + review.EmployeeConfirmComment
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func DisputeResultContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.EmployeeAccount != user.Account || review.Status != ReviewStatusEmployeeConfirm {
			return fmt.Errorf("当前阶段不能提出异议")
		}
		review.EmployeeConfirmComment = strings.TrimSpace(payload.EmployeeConfirmComment)
		if review.EmployeeConfirmComment == "" {
			return fmt.Errorf("请填写异议原因")
		}
		review.EmployeeConfirmResult = "disputed"
		review.EmployeeConfirmedAt = database.Now()
		review.Status = ReviewStatusHRBPReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "员工提出异议："+review.EmployeeConfirmComment)
	})
}

func FinalizeContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !canHandleHrbpFinal(user, *review) || (review.Status != ReviewStatusHRFinal && review.Status != ReviewStatusCompleted) {
			return fmt.Errorf("当前阶段不能归档")
		}
		review.FinalGrade = strings.TrimSpace(fallback(payload.FinalGrade, review.HRBPGrade))
		review.FinalNote = strings.TrimSpace(payload.FinalNote)
		if review.FinalGrade == "" {
			return fmt.Errorf("请先选择最终分档")
		}
		review.Status = ReviewStatusCompleted
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "HRBP 归档")
	})
}

func WithdrawContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		action := ""
		switch {
		case review.Status == ReviewStatusManagerReview && review.EmployeeAccount == user.Account:
			if managerReviewStarted(*review) {
				return fmt.Errorf("上级已评价，不能撤回")
			}
			review.Status = ReviewStatusDraft
			action = "撤销员工自评提交"
		case review.Status == ReviewStatusHRBPReview && review.ManagerAccount == user.Account:
			review.Status = ReviewStatusManagerReview
			review.HRBPReviewerAccount = ""
			action = "撤销上级评价提交"
		case review.Status == ReviewStatusEmployeeConfirm && isHrbpReviewer(user, *review):
			review.Status = ReviewStatusHRBPReview
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmComment = ""
			review.EmployeeConfirmedAt = 0
			action = "撤销 HRBP 评价提交"
		case review.Status == ReviewStatusHRFinal && review.EmployeeAccount == user.Account && review.FinalGrade == "":
			review.Status = ReviewStatusEmployeeConfirm
			review.EmployeeConfirmResult = ""
			review.EmployeeConfirmedAt = 0
			action = "撤销员工确认"
		case review.Status == ReviewStatusHRFinal && canHandleHrbpFinal(user, *review) && review.FinalGrade == "":
			review.Status = ReviewStatusHRBPReview
			action = "撤销 HRBP 评价提交"
		default:
			return fmt.Errorf("当前阶段不能撤销提交")
		}
		reason := normalizeReviewReason(payload.ReturnReason)
		if reason == "" {
			return fmt.Errorf("请填写撤回原因")
		}
		action += "：" + reason
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func ReturnEmployeeContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return returnReview(ctx, user, reviewNo, ReviewStatusManagerReview, ReviewStatusDraft, "退回员工修改："+returnReason(payload))
}

func ReturnManagerContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !isHrbpReviewer(user, *review) || review.Status != ReviewStatusHRBPReview {
			return fmt.Errorf("当前阶段不能退回上级修改")
		}
		copyHrbpFields(review, payload)
		review.Status = ReviewStatusManagerReview
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "退回上级修改："+returnReason(payload))
	})
}

func ReturnHRBPContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, payload ReviewPayload) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if !canHandleHrbpFinal(user, *review) || review.Status != ReviewStatusHRFinal {
			return fmt.Errorf("当前阶段不能退回 HRBP 修改")
		}
		review.FinalGrade = strings.TrimSpace(payload.FinalGrade)
		review.FinalNote = strings.TrimSpace(payload.FinalNote)
		review.Status = ReviewStatusHRBPReview
		review.EmployeeConfirmResult = ""
		review.EmployeeConfirmComment = ""
		review.EmployeeConfirmedAt = 0
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, "退回 HRBP 修改："+returnReason(payload))
	})
}

func DeleteReviewContext(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) error {
	if user == nil {
		return fmt.Errorf("未登录")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var review model.DingTalkH5PerfReview
	if err := notDeletedReviewQuery(db).Where("review_no = ?", reviewNo).First(&review).Error; err != nil {
		return fmt.Errorf("没有找到这张考评单")
	}
	visible, err := reviewInDataScopeContext(ctx, db, user, review)
	if err != nil {
		return err
	}
	if !visible {
		return fmt.Errorf("没有找到这张考评单")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		now := database.Now()
		audit := dingtalkH5AuditMetaForUserContext(ctx, tx, user, now)
		applyDingTalkH5DeleteAudit(&review.DingTalkH5AuditFields, audit)
		review.EditTime = now
		updates := dingtalkH5DeleteAuditUpdateValues(audit)
		updates["edit_time"] = now
		if err := tx.Model(&model.DingTalkH5PerfReview{}).
			Where("`id` = ? AND `deleted_at` = 0", review.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		return addHistoryWithDB(tx, &review, user, "删除考评单")
	})
}
