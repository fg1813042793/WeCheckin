package bootstrap

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

const (
	reviewStatusDraft           = "draft"
	reviewStatusManagerReview   = "manager_review"
	reviewStatusHRBPReview      = "hrbp_review"
	reviewStatusEmployeeConfirm = "employee_confirm"
	reviewStatusHRFinal         = "hr_final"
	reviewStatusCompleted       = "completed"

	reviewScopeDashboard = "dashboard"
)

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

type UserDTO struct {
	ID                     string   `json:"id"`
	Account                string   `json:"account"`
	WorkflowActorID        string   `json:"workflowActorId"`
	Name                   string   `json:"name"`
	Avatar                 string   `json:"avatar"`
	Position               string   `json:"position"`
	Department             string   `json:"department"`
	DepartmentLevel1       string   `json:"departmentLevel1"`
	DepartmentLevel2       string   `json:"departmentLevel2"`
	DepartmentLevel3       string   `json:"departmentLevel3"`
	DepartmentLevel4       string   `json:"departmentLevel4"`
	DepartmentLevels       []string `json:"departmentLevels"`
	ManagerID              string   `json:"managerId"`
	HRBPID                 string   `json:"hrbpId"`
	ResponsibleDepartments []string `json:"responsibleDepartments"`
	Status                 int      `json:"status"`
}

type AppMenuDTO struct {
	Key           string       `json:"key"`
	Label         string       `json:"label"`
	Icon          string       `json:"icon"`
	PermissionKey string       `json:"permissionKey"`
	Children      []AppMenuDTO `json:"children,omitempty"`
}

type BootstrapResponse struct {
	User                  UserDTO                          `json:"user"`
	Menus                 []AppMenuDTO                     `json:"menus"`
	AppConfig             configsvc.DingTalkH5AppConfigDTO `json:"appConfig"`
	AppTitle              string                           `json:"appTitle"`
	AppName               string                           `json:"appName"`
	LogoText              string                           `json:"logoText"`
	LogoURL               string                           `json:"logoUrl"`
	ButtonPermissionKeys  []string                         `json:"buttonPermissionKeys"`
	ButtonPermissionReady bool                             `json:"buttonPermissionReady"`
	APIPermissionKeys     []string                         `json:"apiPermissionKeys"`
	APIPermissionReady    bool                             `json:"apiPermissionReady"`
	PermissionVersion     int64                            `json:"permissionVersion"`
}

type WorkbenchStatsDTO struct {
	Cards []WorkbenchStatCardDTO `json:"cards"`
}

type WorkbenchStatCardDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

type reviewWhereClause struct {
	sql  string
	args []interface{}
}

type workbenchStatusCountRow struct {
	Status string `gorm:"column:status"`
	Count  int64  `gorm:"column:cnt"`
}

type dingTalkH5PermissionSnapshot struct {
	menuKeys    []string
	buttonKeys  []string
	apiKeys     []string
	labels      map[string]string
	icons       map[string]string
	menuReady   bool
	buttonReady bool
	apiReady    bool
	version     int64
}

type dingTalkH5PermissionGrantRow struct {
	SubjectType   string `gorm:"column:grant_subject_type"`
	SubjectID     uint   `gorm:"column:grant_subject_id"`
	PermissionKey string `gorm:"column:grant_permission_key"`
	Effect        string `gorm:"column:grant_effect"`
	EditTime      int64  `gorm:"column:grant_edit_time"`
}

type dingTalkH5PermissionCatalogRow struct {
	Key      string `gorm:"column:permission_key"`
	Name     string `gorm:"column:permission_name"`
	Type     string `gorm:"column:permission_type"`
	Icon     string `gorm:"column:permission_icon"`
	EditTime int64  `gorm:"column:permission_edit_time"`
}

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
	appConfig := configsvc.AppConfigContext(ctx)
	return &BootstrapResponse{
		User:                  userDTO(user),
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

func WorkbenchStatsContext(ctx context.Context, user *model.DingTalkH5PerfUser) (*WorkbenchStatsDTO, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	statusCounts, total, err := workbenchStatusCountsContext(ctx, db, user)
	if err != nil {
		return nil, err
	}
	queue, err := workbenchQueueCountContext(ctx, db, user)
	if err != nil {
		return nil, err
	}
	stats := workbenchStatsFromCounts(statusCounts, total, queue)
	return &stats, nil
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

func dingTalkH5PermissionSnapshotForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (dingTalkH5PermissionSnapshot, error) {
	ctx = normalizedDingTalkH5PermissionContext(ctx)
	snapshot := dingTalkH5PermissionSnapshot{
		labels:  map[string]string{},
		icons:   map[string]string{},
		version: permissionVersionFallback(user),
	}
	if err := ctx.Err(); err != nil {
		return snapshot, err
	}
	roleIDs, err := activeRoleIDsForPerfUserContext(ctx, db, user)
	if err != nil {
		return snapshot, err
	}
	if user == nil || db == nil || (user.ID == 0 && len(roleIDs) == 0) {
		return snapshot, nil
	}
	if !permissionsupport.TablesReady(db) {
		return snapshot, nil
	}

	rows, err := dingTalkH5PermissionGrantRowsForUserContext(ctx, db, user)
	if err != nil {
		return snapshot, err
	}
	roleSet := make(map[uint]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		roleSet[roleID] = true
	}
	roleAllowed := make(map[string]bool, len(rows))
	userAllowed := make(map[string]bool, len(rows))
	denied := make(map[string]bool, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.PermissionKey)
		if key == "" {
			continue
		}
		if row.EditTime > snapshot.version {
			snapshot.version = row.EditTime
		}
		if strings.HasPrefix(key, "dingtalk_h5:api:") && row.Effect == permissionsupport.EffectAllow {
			snapshot.apiReady = true
		}
		switch {
		case row.SubjectType == permissionsupport.SubjectRole && roleSet[row.SubjectID]:
			if row.Effect == permissionsupport.EffectAllow {
				roleAllowed[key] = true
			}
		case row.SubjectType == permissionsupport.SubjectUser && row.SubjectID == user.ID:
			if row.Effect == permissionsupport.EffectDeny {
				denied[key] = true
				continue
			}
			if row.Effect == permissionsupport.EffectAllow {
				userAllowed[key] = true
			}
		}
	}

	selected := make(map[string]bool, len(roleAllowed)+len(userAllowed))
	for key := range roleAllowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	for key := range userAllowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	snapshot.menuKeys = orderedDingTalkH5MenuKeys(selected)
	snapshot.buttonKeys = orderedDingTalkH5ButtonKeys(selected)
	snapshot.apiKeys = orderedDingTalkH5APIKeys(selected)
	snapshot.menuReady = true
	snapshot.buttonReady = true

	labels, icons, permissionVersion, err := dingTalkH5PermissionCatalogContext(ctx, db, snapshot.menuKeys, snapshot.buttonKeys)
	if err != nil {
		return snapshot, err
	}
	snapshot.labels = labels
	snapshot.icons = icons
	if permissionVersion > snapshot.version {
		snapshot.version = permissionVersion
	}
	return snapshot, nil
}

func dingTalkH5PermissionGrantRowsForUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]dingTalkH5PermissionGrantRow, error) {
	subjectWhere, subjectArgs := dingTalkH5PermissionSubjectClause(user)
	if subjectWhere == "" {
		return nil, nil
	}
	permissionWhere, permissionArgs := dingTalkH5PermissionGrantLikeClause()
	var rows []dingTalkH5PermissionGrantRow
	err := db.WithContext(ctx).
		Model(&model.PermissionGrant{}).
		Select("`grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_effect`, `grant_edit_time`").
		Where("`grant_status` = 1").
		Where("("+subjectWhere+")", subjectArgs...).
		Where("("+permissionWhere+")", permissionArgs...).
		Find(&rows).Error
	return rows, err
}

func normalizedDingTalkH5PermissionContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func dingTalkH5PermissionSubjectClause(user *model.DingTalkH5PerfUser) (string, []interface{}) {
	if user == nil {
		return "", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)
	if user.ID > 0 {
		clauses = append(clauses, "(`grant_subject_type` = ? AND `grant_subject_id` = ?)")
		args = append(args, permissionsupport.SubjectUser, user.ID)
	}
	roleIDs := uniqueUintIDs(append([]uint{user.RoleID}, user.RoleIDs...))
	if len(roleIDs) > 0 {
		clauses = append(clauses, "(`grant_subject_type` = ? AND `grant_subject_id` IN ?)")
		args = append(args, permissionsupport.SubjectRole, roleIDs)
	}
	return strings.Join(clauses, " OR "), args
}

func dingTalkH5PermissionGrantLikeClause() (string, []interface{}) {
	prefixes := []string{
		"dingtalk_h5:menu:%",
		"dingtalk_h5:button:%",
		"dingtalk_h5:api:%",
		"data:%",
	}
	clauses := make([]string, 0, len(prefixes))
	args := make([]interface{}, 0, len(prefixes))
	for _, prefix := range prefixes {
		clauses = append(clauses, "`grant_permission_key` LIKE ?")
		args = append(args, prefix)
	}
	return strings.Join(clauses, " OR "), args
}

func orderedDingTalkH5MenuKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appmenuperm.DingTalkH5MenuDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func orderedDingTalkH5ButtonKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appmenuperm.DingTalkH5ButtonDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func orderedDingTalkH5APIKeys(selected map[string]bool) []string {
	keys := make([]string, 0)
	for _, declaration := range appapiperm.DingTalkH5APIDeclarations() {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func dingTalkH5PermissionCatalogContext(ctx context.Context, db *gorm.DB, menuKeys, buttonKeys []string) (map[string]string, map[string]string, int64, error) {
	queryKeys := dingTalkH5PermissionCatalogKeys(menuKeys, buttonKeys)
	if len(queryKeys) == 0 {
		return map[string]string{}, map[string]string{}, 0, nil
	}
	var rows []dingTalkH5PermissionCatalogRow
	if err := db.WithContext(ctx).
		Model(&model.Permission{}).
		Select("`permission_key`, `permission_name`, `permission_type`, `permission_icon`, `permission_edit_time`").
		Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_status` = 1", queryKeys, permissionsupport.PlatformDingTalkH5).
		Find(&rows).Error; err != nil {
		return nil, nil, 0, err
	}
	labels := make(map[string]string, len(rows))
	icons := make(map[string]string, len(rows))
	var version int64
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		name := strings.TrimSpace(row.Name)
		icon := strings.TrimSpace(row.Icon)
		if row.EditTime > version {
			version = row.EditTime
		}
		if key != "" && name != "" && dingTalkH5PermissionCatalogTypeHasMenuLabel(row.Type) {
			labels[key] = name
		}
		if key != "" && icon != "" && dingTalkH5PermissionCatalogTypeHasMenuLabel(row.Type) {
			icons[key] = icon
		}
	}
	return labels, icons, version, nil
}

func dingTalkH5PermissionCatalogKeys(menuKeys, buttonKeys []string) []string {
	allowed := dingTalkH5AllowedMenuKeySet(menuKeys)
	for _, key := range buttonKeys {
		key = strings.TrimSpace(key)
		if key != "" {
			allowed[key] = true
		}
	}
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		keys = append(keys, key)
	}
	return keys
}

func dingTalkH5PermissionCatalogTypeHasMenuLabel(permissionType string) bool {
	return permissionType == permissionsupport.TypeDirectory || permissionType == permissionsupport.TypeMenu
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

func workbenchStatusCountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (map[string]int, int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
	}
	var rows []workbenchStatusCountRow
	query := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{})).Select("status, COUNT(*) AS cnt")
	query, err := applyReviewVisibilityScopeContext(ctx, db, query, user, reviewScopeDashboard)
	if err != nil {
		return nil, 0, err
	}
	if err := query.Group("status").Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	counts := make(map[string]int, len(rows))
	total := 0
	for _, row := range rows {
		value := int(row.Count)
		counts[row.Status] = value
		total += value
	}
	return counts, total, nil
}

func workbenchQueueCountContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
	}
	where, args := workbenchQueueWhere(user)
	var total int64
	query := notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{}))
	query, err := applyReviewVisibilityScopeContext(ctx, db, query, user, reviewScopeDashboard)
	if err != nil {
		return 0, err
	}
	if err := query.Where(where, args...).Count(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

func workbenchQueueWhere(user *model.DingTalkH5PerfUser) (string, []interface{}) {
	if user == nil || user.Account == "" {
		return "1 = 0", nil
	}
	account := NormalizeUserID(user.Account)
	clauses := []reviewWhereClause{
		{sql: "status = ? AND employee_account = ?", args: []interface{}{reviewStatusDraft, account}},
		{sql: "status = ? AND manager_account = ?", args: []interface{}{reviewStatusManagerReview, account}},
		{sql: "status = ? AND employee_account = ?", args: []interface{}{reviewStatusEmployeeConfirm, account}},
		{sql: "status = ? AND (hrbp_reviewer_account = ? OR (hrbp_reviewer_account = '' AND hrbp_account = ?))", args: []interface{}{reviewStatusHRBPReview, account, account}},
		{sql: "status = ? AND (hrbp_reviewer_account = ? OR hrbp_account = ?)", args: []interface{}{reviewStatusHRFinal, account, account}},
	}
	return orReviewWhere(clauses)
}

func workbenchStatsFromCounts(statusCounts map[string]int, total, queue int) WorkbenchStatsDTO {
	reviewing := statusCounts[reviewStatusManagerReview] + statusCounts[reviewStatusHRBPReview] + statusCounts[reviewStatusEmployeeConfirm] + statusCounts[reviewStatusHRFinal]
	return WorkbenchStatsDTO{Cards: []WorkbenchStatCardDTO{
		{Key: "queue", Label: "我的待办", Value: queue},
		{Key: "all", Label: "全部考评单", Value: total},
		{Key: "draft", Label: "员工填写", Value: statusCounts[reviewStatusDraft]},
		{Key: "reviewing", Label: "流转中", Value: reviewing},
		{Key: "completed", Label: "已完成", Value: statusCounts[reviewStatusCompleted]},
	}}
}

func applyReviewVisibilityScopeContext(ctx context.Context, db *gorm.DB, query *gorm.DB, user *model.DingTalkH5PerfUser, scope string) (*gorm.DB, error) {
	where, args := reviewVisibilityWhere(user, scope)
	if where == "" {
		return query, nil
	}
	return query.Where(where, args...), nil
}

func reviewVisibilityWhere(user *model.DingTalkH5PerfUser, scope string) (string, []interface{}) {
	if user == nil || strings.TrimSpace(user.Account) == "" {
		return "1 = 0", nil
	}
	account := NormalizeUserID(user.Account)
	switch strings.TrimSpace(scope) {
	case reviewScopeDashboard:
		return personalReviewVisibilityWhere(account)
	default:
		return personalReviewVisibilityWhere(account)
	}
}

func personalReviewVisibilityWhere(account string) (string, []interface{}) {
	return "(employee_account = ? OR manager_account = ? OR hrbp_account = ? OR hrbp_reviewer_account = ?)",
		[]interface{}{account, account, account, account}
}

func isHrbpReviewer(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return review.HRBPAccount == user.Account
}

func canHandleHrbpFinal(user *model.DingTalkH5PerfUser, review model.DingTalkH5PerfReview) bool {
	if user == nil {
		return false
	}
	if review.HRBPReviewerAccount != "" {
		return review.HRBPReviewerAccount == user.Account
	}
	return review.HRBPAccount == user.Account
}

func notDeletedReviewQuery(db *gorm.DB) *gorm.DB {
	if db == nil {
		return db
	}
	return db.Where("`deleted_at` = 0")
}

func orReviewWhere(clauses []reviewWhereClause) (string, []interface{}) {
	parts := make([]string, 0, len(clauses))
	args := make([]interface{}, 0)
	for _, clause := range clauses {
		if strings.TrimSpace(clause.sql) == "" {
			continue
		}
		parts = append(parts, "("+clause.sql+")")
		args = append(args, clause.args...)
	}
	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return strings.Join(parts, " OR "), args
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

func userDTO(user *model.DingTalkH5PerfUser) UserDTO {
	if user == nil {
		return UserDTO{}
	}
	return UserDTO{
		ID:                     user.Account,
		Account:                user.Account,
		WorkflowActorID:        strconv.FormatUint(uint64(user.ID), 10),
		Name:                   user.Name,
		Avatar:                 user.Pic,
		Position:               user.Position,
		Department:             user.Department,
		DepartmentLevel1:       user.DepartmentLevel1,
		DepartmentLevel2:       user.DepartmentLevel2,
		DepartmentLevel3:       user.DepartmentLevel3,
		DepartmentLevel4:       user.DepartmentLevel4,
		DepartmentLevels:       departmentLevelsFromUser(*user),
		ManagerID:              NormalizeUserID(user.ManagerAccount),
		HRBPID:                 NormalizeUserID(user.HRBPAccount),
		ResponsibleDepartments: decodeStringList(user.ResponsibleDepartments),
		Status:                 user.Status,
	}
}

func departmentLevelsFromUser(user model.DingTalkH5PerfUser) []string {
	levels := normalizeDepartmentLevels(user.DepartmentLevels)
	if len(levels) > 0 {
		return levels
	}
	levels = normalizeDepartmentLevels([]string{
		user.DepartmentLevel1,
		user.DepartmentLevel2,
		user.DepartmentLevel3,
		user.DepartmentLevel4,
	})
	if len(levels) > 0 {
		return levels
	}
	return splitDepartmentText(user.Department)
}

func normalizeDepartmentLevels(levels []string) []string {
	clean := make([]string, 0, len(levels))
	for _, level := range levels {
		if level = strings.TrimSpace(level); level != "" {
			clean = append(clean, level)
		}
	}
	return clean
}

func splitDepartmentText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return normalizeDepartmentLevels(strings.Split(text, " / "))
}

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func decodeStringList(raw string) []string {
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func uniqueUintIDs(items []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(items))
	for _, item := range items {
		if item == 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
