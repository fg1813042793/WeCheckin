package user

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

const (
	perfUserMetaKey = "dingtalkH5Performance"

	dingtalkH5DataScopeAll    = 1
	dingtalkH5DataScopeDept   = 2
	dingtalkH5DataScopeSelf   = 3
	dingtalkH5DataScopeCustom = 4

	reviewStatusCompleted = "completed"
)

const perfUserSelectColumns = "`id`, `user_mini_openid`, `user_name`, `user_password`, `user_pic`, `user_status`, `user_role_id`, `user_position_id`, `user_obj`, `user_add_time`, `user_edit_time`"

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

type UserDTO struct {
	ID                     string   `json:"id"`
	Account                string   `json:"account"`
	Name                   string   `json:"name"`
	Avatar                 string   `json:"avatar"`
	Position               string   `json:"position"`
	DepartmentID           uint     `json:"departmentId"`
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

type UserPayload struct {
	ID                     string      `json:"id"`
	Account                string      `json:"account"`
	Name                   string      `json:"name"`
	Avatar                 string      `json:"avatar"`
	Password               string      `json:"password"`
	Position               string      `json:"position"`
	Department             string      `json:"department"`
	DepartmentLevel1       string      `json:"departmentLevel1"`
	DepartmentLevel2       string      `json:"departmentLevel2"`
	DepartmentLevel3       string      `json:"departmentLevel3"`
	DepartmentLevel4       string      `json:"departmentLevel4"`
	DepartmentLevels       []string    `json:"departmentLevels"`
	ManagerID              string      `json:"managerId"`
	HRBPID                 string      `json:"hrbpId"`
	ResponsibleDepartments interface{} `json:"responsibleDepartments"`
}

type perfUserMeta struct {
	Department             string   `json:"department,omitempty"`
	DepartmentLevel1       string   `json:"departmentLevel1,omitempty"`
	DepartmentLevel2       string   `json:"departmentLevel2,omitempty"`
	DepartmentLevel3       string   `json:"departmentLevel3,omitempty"`
	DepartmentLevel4       string   `json:"departmentLevel4,omitempty"`
	DepartmentLevels       []string `json:"departmentLevels,omitempty"`
	ManagerID              string   `json:"managerId,omitempty"`
	HRBPID                 string   `json:"hrbpId,omitempty"`
	ResponsibleDepartments []string `json:"responsibleDepartments,omitempty"`
}

type perfUserDepartmentPath struct {
	DeptID uint
	Levels []string
}

func ListUsersContext(ctx context.Context, current *model.DingTalkH5PerfUser) ([]UserDTO, error) {
	if current == nil {
		return nil, fmt.Errorf("未登录")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	scope, err := permissionsupport.DataScopeContext(ctx, db, current.ID, current.RoleID)
	if err != nil {
		return nil, err
	}
	users, err := listPerfUsersByDataScopeContext(ctx, db, current, scope)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(users, func(i, j int) bool {
		left := []string{users[i].DepartmentLevel1, users[i].DepartmentLevel2, users[i].DepartmentLevel3, users[i].DepartmentLevel4, users[i].Name, users[i].Account}
		right := []string{users[j].DepartmentLevel1, users[j].DepartmentLevel2, users[j].DepartmentLevel3, users[j].DepartmentLevel4, users[j].Name, users[j].Account}
		return strings.Join(left, "\x00") < strings.Join(right, "\x00")
	})
	result := make([]UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, userDTO(user))
	}
	return result, nil
}

func CreateUserContext(ctx context.Context, current *model.DingTalkH5PerfUser, payload UserPayload) (*UserDTO, []UserDTO, error) {
	if current == nil {
		return nil, nil, fmt.Errorf("未登录")
	}
	user, err := sanitizeUserPayload(payload, nil)
	if err != nil {
		return nil, nil, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := validateUserPayload(db, user, ""); err != nil {
		return nil, nil, err
	}
	baseUser := model.User{
		MiniOpenID: user.Account,
		Name:       user.Name,
		Password:   user.Password,
		Pic:        user.Pic,
		Status:     user.Status,
		Obj:        encodePerfUserObj("", user),
		AddTime:    user.AddTime,
		EditTime:   user.EditTime,
	}
	if err := db.Create(&baseUser).Error; err != nil {
		return nil, nil, err
	}
	user.ID = baseUser.ID
	users, err := ListUsersContext(ctx, current)
	if err != nil {
		return nil, nil, err
	}
	dto := userDTO(user)
	return &dto, users, nil
}

func UpdateUserContext(ctx context.Context, current *model.DingTalkH5PerfUser, account string, payload UserPayload) (*UserDTO, []UserDTO, error) {
	if current == nil {
		return nil, nil, fmt.Errorf("未登录")
	}
	account = NormalizeUserID(account)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	existing, err := loadPerfUserByAccountDB(db, account)
	if err != nil {
		return nil, nil, fmt.Errorf("没有找到该账号")
	}
	allowed, err := canAccessPerfUserAccountContext(ctx, db, current, existing.Account)
	if err != nil {
		return nil, nil, err
	}
	if !allowed {
		return nil, nil, fmt.Errorf("没有找到该账号")
	}
	next, err := sanitizeUserPayload(payload, existing)
	if err != nil {
		return nil, nil, err
	}
	next.Account = existing.Account
	next.ID = existing.ID
	if err := validateUserPayload(db, next, existing.Account); err != nil {
		return nil, nil, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"user_name":      next.Name,
			"user_pic":       next.Pic,
			"user_status":    next.Status,
			"user_obj":       encodePerfUserObj(existing.Obj, next),
			"user_edit_time": next.EditTime,
		}
		if strings.TrimSpace(payload.Password) != "" {
			updates["user_password"] = next.Password
		}
		if err := tx.Model(&model.User{}).Where("`user_mini_openid` = ?", next.Account).Updates(updates).Error; err != nil {
			return err
		}
		return syncOpenReviewsForUser(tx, next)
	}); err != nil {
		return nil, nil, err
	}
	dto := userDTO(next)
	return &dto, nil, nil
}

func DeleteUserContext(ctx context.Context, current *model.DingTalkH5PerfUser, account string) ([]UserDTO, error) {
	if current == nil {
		return nil, fmt.Errorf("未登录")
	}
	account = NormalizeUserID(account)
	if account == current.Account {
		return nil, fmt.Errorf("不能删除当前登录账号")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	target, err := loadPerfUserByAccountDB(db, account)
	if err != nil {
		return nil, fmt.Errorf("没有找到该账号")
	}
	if target.Status != 1 {
		return nil, fmt.Errorf("没有找到该账号")
	}
	allowed, err := canAccessPerfUserAccountContext(ctx, db, current, target.Account)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("没有找到该账号")
	}
	now := database.Now()
	updates := map[string]interface{}{
		"user_status":    0,
		"user_edit_time": now,
	}
	if err := db.Model(&model.User{}).Where("`user_mini_openid` = ?", target.Account).Updates(updates).Error; err != nil {
		return nil, err
	}
	return ListUsersContext(ctx, current)
}

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func DataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	return dataScopeUserAccountsContext(ctx, db, user, scope)
}

func DataScopeDeptIDsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) ([]uint, error) {
	return dataScopeDeptIDsContext(ctx, db, user, scope)
}

func DataScopeExtraDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]uint, error) {
	return dataScopeExtraDeptIDsContext(ctx, db, deptIDs)
}

func UserAccountsByDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]string, error) {
	return userAccountsByDeptIDsContext(ctx, db, deptIDs)
}

func AccountsByUserIDsContext(ctx context.Context, db *gorm.DB, userIDs []uint) ([]string, error) {
	return accountsByUserIDsContext(ctx, db, userIDs)
}

func LoadPerfUserByAccountDB(db *gorm.DB, account string) (*model.DingTalkH5PerfUser, error) {
	return loadPerfUserByAccountDB(db, account)
}

func ListPerfUsersDB(db *gorm.DB) ([]model.DingTalkH5PerfUser, error) {
	return listPerfUsersDB(db)
}

func ListPerfUsersByAccountsDB(db *gorm.DB, allowed map[string]struct{}) ([]model.DingTalkH5PerfUser, error) {
	return listPerfUsersByAccountsDB(db, allowed)
}

func CanAccessPerfUserAccountContext(ctx context.Context, db *gorm.DB, current *model.DingTalkH5PerfUser, account string) (bool, error) {
	return canAccessPerfUserAccountContext(ctx, db, current, account)
}

func HydratePerfUser(user *model.DingTalkH5PerfUser) {
	hydratePerfUser(user)
}

func SanitizeUserPayload(payload UserPayload, existing *model.DingTalkH5PerfUser) (model.DingTalkH5PerfUser, error) {
	return sanitizeUserPayload(payload, existing)
}

func UserDTOFromModel(user model.DingTalkH5PerfUser) UserDTO {
	return userDTO(user)
}

func EncodePerfUserObj(raw string, user model.DingTalkH5PerfUser) string {
	return encodePerfUserObj(raw, user)
}

func DepartmentFromUser(user model.DingTalkH5PerfUser) string {
	return departmentFromUser(user)
}

func DecodeStringList(raw string) []string {
	return decodeStringList(raw)
}

func UniqueUintIDs(items []uint) []uint {
	return uniqueUintIDs(items)
}

func listPerfUsersByDataScopeContext(ctx context.Context, db *gorm.DB, current *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) ([]model.DingTalkH5PerfUser, error) {
	allowed, all, err := dataScopeUserAccountsContext(ctx, db, current, scope)
	if err != nil {
		return nil, err
	}
	if !all && len(allowed) == 0 {
		return nil, nil
	}
	if all {
		return listPerfUsersDB(db)
	}
	return listPerfUsersByAccountsDB(db, allowed)
}

func canAccessPerfUserAccountContext(ctx context.Context, db *gorm.DB, current *model.DingTalkH5PerfUser, account string) (bool, error) {
	account = NormalizeUserID(account)
	if current == nil || account == "" {
		return false, nil
	}
	if account == NormalizeUserID(current.Account) {
		return true, nil
	}
	scope, err := permissionsupport.DataScopeContext(ctx, db, current.ID, current.RoleID)
	if err != nil {
		return false, err
	}
	allowed, all, err := dataScopeUserAccountsContext(ctx, db, current, scope)
	if err != nil {
		return false, err
	}
	if all {
		return true, nil
	}
	_, ok := allowed[account]
	return ok, nil
}

func dataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	accounts, all, err := baseDataScopeUserAccountsContext(ctx, db, user, scope)
	if err != nil || all {
		return accounts, all, err
	}
	extraAccounts, err := extraDataScopeUserAccountsContext(ctx, db, user)
	if err != nil {
		return nil, false, err
	}
	for account := range extraAccounts {
		accounts[account] = struct{}{}
	}
	return accounts, false, nil
}

func baseDataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	if user == nil {
		return map[string]struct{}{}, false, nil
	}
	if !scope.Ready {
		return map[string]struct{}{NormalizeUserID(user.Account): {}}, false, nil
	}
	switch scope.Mode {
	case dingtalkH5DataScopeAll:
		return nil, true, nil
	case dingtalkH5DataScopeSelf:
		return map[string]struct{}{NormalizeUserID(user.Account): {}}, false, nil
	case dingtalkH5DataScopeDept, dingtalkH5DataScopeCustom:
		accounts := map[string]struct{}{}
		if scope.Mode == dingtalkH5DataScopeDept {
			accounts[NormalizeUserID(user.Account)] = struct{}{}
		}
		deptIDs, err := dataScopeDeptIDsContext(ctx, db, user, scope)
		if err != nil {
			return nil, false, err
		}
		if len(deptIDs) == 0 {
			return accounts, false, nil
		}
		deptAccounts, err := userAccountsByDeptIDsContext(ctx, db, deptIDs)
		if err != nil {
			return nil, false, err
		}
		for _, account := range deptAccounts {
			accounts[account] = struct{}{}
		}
		return accounts, false, nil
	default:
		return map[string]struct{}{}, false, nil
	}
}

func extraDataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (map[string]struct{}, error) {
	result := map[string]struct{}{}
	if user == nil {
		return result, nil
	}
	extras, err := permissionsupport.DataScopeExtrasContext(ctx, db, user.ID, user.RoleID)
	if err != nil || !extras.Ready {
		return result, err
	}
	if len(extras.UserIDs) > 0 {
		accounts, err := accountsByUserIDsContext(ctx, db, extras.UserIDs)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			result[account] = struct{}{}
		}
	}
	if len(extras.DeptIDs) > 0 {
		deptIDs, err := dataScopeExtraDeptIDsContext(ctx, db, extras.DeptIDs)
		if err != nil {
			return nil, err
		}
		if len(deptIDs) > 0 {
			accounts, err := userAccountsByDeptIDsContext(ctx, db, deptIDs)
			if err != nil {
				return nil, err
			}
			for _, account := range accounts {
				result[account] = struct{}{}
			}
		}
	}
	return result, nil
}

func dataScopeDeptIDsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) ([]uint, error) {
	baseIDs := make([]uint, 0)
	switch scope.Mode {
	case dingtalkH5DataScopeDept:
		var rows []model.UserDept
		if err := db.WithContext(ctx).Where("`user_dept_user_id` = ?", user.ID).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row.DeptID > 0 {
				baseIDs = append(baseIDs, row.DeptID)
			}
		}
	case dingtalkH5DataScopeCustom:
		baseIDs = append(baseIDs, scope.DeptIDs...)
	default:
		return nil, nil
	}
	baseIDs = uniqueUintIDs(baseIDs)
	if len(baseIDs) == 0 {
		return nil, nil
	}
	var departments []model.Department
	if err := db.WithContext(ctx).Find(&departments).Error; err != nil {
		return nil, err
	}
	return departmentDescendantIDs(departments, baseIDs), nil
}

func dataScopeExtraDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]uint, error) {
	deptIDs = uniqueUintIDs(deptIDs)
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var departments []model.Department
	if err := db.WithContext(ctx).Find(&departments).Error; err != nil {
		return nil, err
	}
	return departmentDescendantIDs(departments, deptIDs), nil
}

func userAccountsByDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]string, error) {
	deptIDs = uniqueUintIDs(deptIDs)
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var rows []model.DingTalkH5PerfUser
	if err := db.WithContext(ctx).
		Model(&model.DingTalkH5PerfUser{}).
		Distinct("`users`.`user_mini_openid`").
		Joins("JOIN `user_depts` ON `user_depts`.`user_dept_user_id` = `users`.`id`").
		Where("`users`.`user_status` = 1").
		Where("`user_depts`.`user_dept_dept_id` IN ?", deptIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	accounts := make([]string, 0, len(rows))
	for _, row := range rows {
		if account := NormalizeUserID(row.Account); account != "" {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func accountsByUserIDsContext(ctx context.Context, db *gorm.DB, userIDs []uint) ([]string, error) {
	userIDs = uniqueUintIDs(userIDs)
	if len(userIDs) == 0 {
		return nil, nil
	}
	var rows []model.DingTalkH5PerfUser
	if err := db.WithContext(ctx).
		Select("`user_mini_openid`").
		Where("`id` IN ? AND `user_status` = 1", userIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	accounts := make([]string, 0, len(rows))
	for _, row := range rows {
		account := NormalizeUserID(row.Account)
		if account == "" {
			continue
		}
		if _, ok := seen[account]; ok {
			continue
		}
		seen[account] = struct{}{}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func sanitizeUserPayload(payload UserPayload, existing *model.DingTalkH5PerfUser) (model.DingTalkH5PerfUser, error) {
	account := NormalizeUserID(firstNonEmpty(payload.Account, payload.ID))
	if existing != nil {
		account = existing.Account
	}
	if account == "" {
		return model.DingTalkH5PerfUser{}, fmt.Errorf("请填写账号")
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" && existing != nil {
		name = existing.Name
	}
	if name == "" {
		return model.DingTalkH5PerfUser{}, fmt.Errorf("请填写姓名")
	}
	avatar := strings.TrimSpace(payload.Avatar)
	if avatar == "" && existing != nil {
		avatar = existing.Pic
	}
	avatar, err := sanitizeAvatarURL(avatar)
	if err != nil {
		return model.DingTalkH5PerfUser{}, err
	}
	position := ""
	if existing != nil {
		position = existing.Position
	}
	departmentLevels := normalizeDepartmentLevels(payload.DepartmentLevels)
	if len(departmentLevels) == 0 {
		departmentLevels = normalizeDepartmentLevels([]string{
			payload.DepartmentLevel1,
			payload.DepartmentLevel2,
			payload.DepartmentLevel3,
			payload.DepartmentLevel4,
		})
	}
	if len(departmentLevels) == 0 {
		departmentLevels = splitDepartmentText(payload.Department)
	}
	if len(departmentLevels) == 0 && existing != nil {
		departmentLevels = departmentLevelsFromUser(*existing)
	}
	department := departmentText(departmentLevels...)
	if department == "" {
		department = strings.TrimSpace(payload.Department)
	}
	if department == "" && existing != nil {
		department = existing.Department
	}
	password := strings.TrimSpace(payload.Password)
	passwordHash := ""
	if existing != nil {
		passwordHash = existing.Password
	}
	if password != "" {
		hash, err := passwordutil.Hash(password)
		if err != nil {
			return model.DingTalkH5PerfUser{}, err
		}
		passwordHash = hash
	}
	if passwordHash == "" {
		hash, err := passwordutil.Hash("123456")
		if err != nil {
			return model.DingTalkH5PerfUser{}, err
		}
		passwordHash = hash
	}
	now := database.Now()
	addTime := now
	status := 1
	if existing != nil {
		addTime = existing.AddTime
		status = existing.Status
	}
	responsibleDepartments := existingResponsibleDepartments(payload.ResponsibleDepartments, existing)
	return model.DingTalkH5PerfUser{
		Account:                account,
		Name:                   name,
		Password:               passwordHash,
		Pic:                    avatar,
		Role:                   "",
		Position:               position,
		Department:             department,
		DepartmentLevel1:       departmentLevelAt(departmentLevels, 0),
		DepartmentLevel2:       departmentLevelAt(departmentLevels, 1),
		DepartmentLevel3:       departmentLevelAt(departmentLevels, 2),
		DepartmentLevel4:       departmentLevelAt(departmentLevels, 3),
		DepartmentLevels:       departmentLevels,
		ManagerAccount:         NormalizeUserID(payload.ManagerID),
		HRBPAccount:            NormalizeUserID(payload.HRBPID),
		ResponsibleDepartments: responsibleDepartments,
		Status:                 status,
		AddTime:                addTime,
		EditTime:               now,
	}, nil
}

func existingResponsibleDepartments(value interface{}, existing *model.DingTalkH5PerfUser) string {
	if value == nil && existing != nil {
		return existing.ResponsibleDepartments
	}
	return encodeJSON(normalizeList(value))
}

func validateUserPayload(db *gorm.DB, payload model.DingTalkH5PerfUser, existingAccount string) error {
	if payload.Account == "" || payload.Name == "" {
		return fmt.Errorf("请填写账号和姓名")
	}
	var duplicate model.User
	err := db.Where("(`user_mini_openid` = ? OR LOWER(`user_name`) = ?) AND `user_mini_openid` <> ?", payload.Account, strings.ToLower(payload.Name), existingAccount).First(&duplicate).Error
	if err == nil {
		return fmt.Errorf("账号或姓名已存在")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if payload.ManagerAccount != "" {
		if _, err := loadPerfUserByAccountDB(db, payload.ManagerAccount); err != nil {
			return fmt.Errorf("直属上级不存在")
		}
	}
	if payload.HRBPAccount != "" {
		if _, err := loadPerfUserByAccountDB(db, payload.HRBPAccount); err != nil {
			return fmt.Errorf("HRBP不存在")
		}
	}
	return nil
}

func syncOpenReviewsForUser(db *gorm.DB, target model.DingTalkH5PerfUser) error {
	updates := map[string]interface{}{
		"manager_account":   target.ManagerAccount,
		"hrbp_account":      fallback(target.HRBPAccount, "hrbp"),
		"department":        departmentFromUser(target),
		"department_level1": target.DepartmentLevel1,
		"department_level2": target.DepartmentLevel2,
		"department_level3": target.DepartmentLevel3,
		"edit_time":         database.Now(),
	}
	return notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{})).
		Where("employee_account = ? AND status <> ?", target.Account, reviewStatusCompleted).
		Updates(updates).Error
}

func listPerfUsersDB(db *gorm.DB) ([]model.DingTalkH5PerfUser, error) {
	var users []model.DingTalkH5PerfUser
	if err := db.Select(perfUserSelectColumns).Where("`user_status` = 1").Order("`user_name` ASC, `id` ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return hydratePerfUsersForListDB(db, users)
}

func listPerfUsersByAccountsDB(db *gorm.DB, allowed map[string]struct{}) ([]model.DingTalkH5PerfUser, error) {
	accounts := make([]string, 0, len(allowed))
	for account := range allowed {
		account = NormalizeUserID(account)
		if account != "" {
			accounts = append(accounts, account)
		}
	}
	if len(accounts) == 0 {
		return nil, nil
	}
	var users []model.DingTalkH5PerfUser
	if err := db.Select(perfUserSelectColumns).Where("`user_mini_openid` IN ? AND `user_status` = 1", accounts).Order("`user_name` ASC, `id` ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return hydratePerfUsersForListDB(db, users)
}

func loadPerfUserByAccountDB(db *gorm.DB, account string) (*model.DingTalkH5PerfUser, error) {
	account = NormalizeUserID(account)
	var user model.DingTalkH5PerfUser
	if err := db.Select(perfUserSelectColumns).Where("`user_mini_openid` = ?", account).First(&user).Error; err != nil {
		return nil, err
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func hydratePerfUserWithUserDeptDB(db *gorm.DB, user *model.DingTalkH5PerfUser) error {
	if user == nil {
		return nil
	}
	users, err := hydratePerfUsersWithUserDeptsDB(db, []model.DingTalkH5PerfUser{*user})
	if err != nil {
		return err
	}
	if len(users) > 0 {
		*user = users[0]
	}
	return nil
}

func hydratePerfUsersWithUserDeptsDB(db *gorm.DB, users []model.DingTalkH5PerfUser) ([]model.DingTalkH5PerfUser, error) {
	return hydratePerfUsersDB(db, users, false)
}

func hydratePerfUsersForListDB(db *gorm.DB, users []model.DingTalkH5PerfUser) ([]model.DingTalkH5PerfUser, error) {
	return hydratePerfUsersDB(db, users, true)
}

func hydratePerfUsersDB(db *gorm.DB, users []model.DingTalkH5PerfUser, expandDepartments bool) ([]model.DingTalkH5PerfUser, error) {
	for index := range users {
		hydratePerfUser(&users[index])
	}
	if len(users) == 0 {
		return users, nil
	}
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		if user.ID > 0 {
			userIDs = append(userIDs, user.ID)
		}
	}
	paths, err := userDeptPathRowsByUserIDDB(db, uniqueUintIDs(userIDs))
	if err != nil {
		return nil, err
	}
	positionNames, err := positionNamesByIDDB(db, uniquePositionIDs(users))
	if err != nil {
		return nil, err
	}
	roleIDsByUser, err := userRoleIDsByUserIDDB(db, uniqueUintIDs(userIDs))
	if err != nil {
		return nil, err
	}
	for index := range users {
		if userPaths := paths[users[index].ID]; len(userPaths) > 0 {
			users[index].DepartmentID = userPaths[0].DeptID
			applyDepartmentPathToPerfUser(&users[index], userPaths[0].Levels)
		}
		applyPositionNameToPerfUser(&users[index], positionNames)
		applyRoleIDsToPerfUser(&users[index], roleIDsByUser)
	}
	if expandDepartments {
		return expandPerfUsersByDepartmentPaths(users, paths), nil
	}
	return users, nil
}

func hydratePerfUser(user *model.DingTalkH5PerfUser) {
	if user == nil {
		return
	}
	meta := decodePerfUserMeta(user.Obj)
	departmentLevels := normalizeDepartmentLevels(meta.DepartmentLevels)
	if len(departmentLevels) == 0 {
		departmentLevels = normalizeDepartmentLevels([]string{
			meta.DepartmentLevel1,
			meta.DepartmentLevel2,
			meta.DepartmentLevel3,
			meta.DepartmentLevel4,
		})
	}
	user.Role = ""
	user.Position = ""
	user.Department = strings.TrimSpace(meta.Department)
	if user.Department == "" {
		user.Department = departmentText(departmentLevels...)
	}
	user.DepartmentLevel1 = departmentLevelAt(departmentLevels, 0)
	user.DepartmentLevel2 = departmentLevelAt(departmentLevels, 1)
	user.DepartmentLevel3 = departmentLevelAt(departmentLevels, 2)
	user.DepartmentLevel4 = departmentLevelAt(departmentLevels, 3)
	user.DepartmentLevels = departmentLevels
	user.ManagerAccount = NormalizeUserID(meta.ManagerID)
	user.HRBPAccount = NormalizeUserID(meta.HRBPID)
	user.ResponsibleDepartments = encodeJSON(uniqueStrings(meta.ResponsibleDepartments))
}

func encodePerfUserObj(raw string, user model.DingTalkH5PerfUser) string {
	obj := decodeObject(raw)
	departmentLevels := departmentLevelsFromUser(user)
	obj[perfUserMetaKey] = perfUserMeta{
		Department:             strings.TrimSpace(user.Department),
		DepartmentLevel1:       departmentLevelAt(departmentLevels, 0),
		DepartmentLevel2:       departmentLevelAt(departmentLevels, 1),
		DepartmentLevel3:       departmentLevelAt(departmentLevels, 2),
		DepartmentLevel4:       departmentLevelAt(departmentLevels, 3),
		DepartmentLevels:       departmentLevels,
		ManagerID:              NormalizeUserID(user.ManagerAccount),
		HRBPID:                 NormalizeUserID(user.HRBPAccount),
		ResponsibleDepartments: decodeStringList(user.ResponsibleDepartments),
	}
	data, _ := json.Marshal(obj)
	return string(data)
}

func decodePerfUserMeta(raw string) perfUserMeta {
	obj := decodeObject(raw)
	value, ok := obj[perfUserMetaKey]
	if !ok {
		return perfUserMeta{}
	}
	data, _ := json.Marshal(value)
	var meta perfUserMeta
	_ = json.Unmarshal(data, &meta)
	return meta
}

func decodeObject(raw string) map[string]interface{} {
	obj := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return obj
	}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return map[string]interface{}{}
	}
	return obj
}

func userRoleIDsByUserIDDB(db *gorm.DB, userIDs []uint) (map[uint][]uint, error) {
	result := make(map[uint][]uint, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	if !permissionsupport.UserRolesTableReady(db) {
		return result, nil
	}
	var rows []model.UserRole
	if err := db.Select("user_role_user_id", "user_role_role_id").
		Where("`user_role_user_id` IN ? AND `user_role_status` = 1", userIDs).
		Order("`user_role_is_primary` DESC, `id` ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	seenByUser := make(map[uint]map[uint]struct{}, len(userIDs))
	for _, row := range rows {
		if row.UserID == 0 || row.RoleID == 0 {
			continue
		}
		if seenByUser[row.UserID] == nil {
			seenByUser[row.UserID] = map[uint]struct{}{}
		}
		if _, ok := seenByUser[row.UserID][row.RoleID]; ok {
			continue
		}
		seenByUser[row.UserID][row.RoleID] = struct{}{}
		result[row.UserID] = append(result[row.UserID], row.RoleID)
	}
	return result, nil
}

func uniquePositionIDs(users []model.DingTalkH5PerfUser) []uint {
	items := make([]uint, 0, len(users))
	for _, user := range users {
		if user.PositionID > 0 {
			items = append(items, user.PositionID)
		}
	}
	return uniqueUintIDs(items)
}

func positionNamesByIDDB(db *gorm.DB, positionIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string, len(positionIDs))
	if len(positionIDs) == 0 {
		return result, nil
	}
	var positions []model.Position
	if err := db.Select("id", "position_name").Where("`id` IN ?", positionIDs).Find(&positions).Error; err != nil {
		return nil, err
	}
	for _, position := range positions {
		if position.ID == 0 {
			continue
		}
		if name := strings.TrimSpace(position.Name); name != "" {
			result[position.ID] = name
		}
	}
	return result, nil
}

func applyPositionNameToPerfUser(user *model.DingTalkH5PerfUser, positionNames map[uint]string) {
	if user == nil {
		return
	}
	user.Position = ""
	if user.PositionID == 0 {
		return
	}
	user.Position = strings.TrimSpace(positionNames[user.PositionID])
}

func applyRoleIDsToPerfUser(user *model.DingTalkH5PerfUser, roleIDsByUser map[uint][]uint) {
	if user == nil {
		return
	}
	user.RoleIDs = roleIDsByUser[user.ID]
	if len(user.RoleIDs) == 0 && user.RoleID > 0 {
		user.RoleIDs = []uint{user.RoleID}
	}
}

func userDeptPathRowsByUserIDDB(db *gorm.DB, userIDs []uint) (map[uint][]perfUserDepartmentPath, error) {
	result := make(map[uint][]perfUserDepartmentPath, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	var rows []model.UserDept
	if err := db.Where("`user_dept_user_id` IN ?", userIDs).Order("`id` ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return result, nil
	}
	var departments []model.Department
	if err := db.Find(&departments).Error; err != nil {
		return nil, err
	}
	deptByID := make(map[uint]model.Department, len(departments))
	for _, department := range departments {
		if department.ID > 0 {
			deptByID[department.ID] = department
		}
	}
	seen := make(map[uint]map[uint]struct{}, len(userIDs))
	for _, row := range rows {
		if row.UserID == 0 || row.DeptID == 0 {
			continue
		}
		if _, ok := seen[row.UserID]; !ok {
			seen[row.UserID] = map[uint]struct{}{}
		}
		if _, exists := seen[row.UserID][row.DeptID]; exists {
			continue
		}
		if levels := departmentPathLevels(deptByID, row.DeptID); len(levels) > 0 {
			seen[row.UserID][row.DeptID] = struct{}{}
			result[row.UserID] = append(result[row.UserID], perfUserDepartmentPath{DeptID: row.DeptID, Levels: levels})
		}
	}
	return result, nil
}

func expandPerfUsersByDepartmentPaths(users []model.DingTalkH5PerfUser, paths map[uint][]perfUserDepartmentPath) []model.DingTalkH5PerfUser {
	if len(users) == 0 || len(paths) == 0 {
		return users
	}
	expanded := make([]model.DingTalkH5PerfUser, 0, len(users))
	for _, user := range users {
		userPaths := paths[user.ID]
		if len(userPaths) == 0 {
			expanded = append(expanded, user)
			continue
		}
		for _, path := range userPaths {
			next := user
			next.DepartmentID = path.DeptID
			applyDepartmentPathToPerfUser(&next, path.Levels)
			expanded = append(expanded, next)
		}
	}
	return expanded
}

func departmentDescendantIDs(all []model.Department, parentIDs []uint) []uint {
	seen := map[uint]struct{}{}
	queue := uniqueUintIDs(parentIDs)
	for _, id := range queue {
		seen[id] = struct{}{}
	}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		for _, dept := range all {
			if dept.ParentID != parentID || dept.ID == 0 {
				continue
			}
			if _, ok := seen[dept.ID]; ok {
				continue
			}
			seen[dept.ID] = struct{}{}
			queue = append(queue, dept.ID)
		}
	}
	result := make([]uint, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return uniqueUintIDs(result)
}

func departmentPathLevels(deptByID map[uint]model.Department, deptID uint) []string {
	reversed := make([]string, 0, 4)
	visited := map[uint]struct{}{}
	for deptID > 0 {
		if _, exists := visited[deptID]; exists {
			break
		}
		visited[deptID] = struct{}{}
		department, ok := deptByID[deptID]
		if !ok {
			break
		}
		if name := strings.TrimSpace(department.Name); name != "" {
			reversed = append(reversed, name)
		}
		deptID = department.ParentID
	}
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}

func applyDepartmentPathToPerfUser(user *model.DingTalkH5PerfUser, levels []string) {
	if user == nil {
		return
	}
	cleanLevels := make([]string, 0, len(levels))
	for _, level := range levels {
		if level = strings.TrimSpace(level); level != "" {
			cleanLevels = append(cleanLevels, level)
		}
	}
	if len(cleanLevels) == 0 {
		return
	}
	user.Department = strings.Join(cleanLevels, " / ")
	user.DepartmentLevel1 = cleanLevels[0]
	user.DepartmentLevel2 = ""
	user.DepartmentLevel3 = ""
	user.DepartmentLevel4 = ""
	user.DepartmentLevels = cleanLevels
	if len(cleanLevels) > 1 {
		user.DepartmentLevel2 = cleanLevels[1]
	}
	if len(cleanLevels) > 2 {
		user.DepartmentLevel3 = cleanLevels[2]
	}
	if len(cleanLevels) > 3 {
		user.DepartmentLevel4 = cleanLevels[3]
	}
}

func canViewPerfUser(current *model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	if current == nil {
		return false
	}
	if target.Account == current.Account || target.Account == current.ManagerAccount || target.Account == current.HRBPAccount {
		return true
	}
	return target.ManagerAccount == current.Account ||
		target.HRBPAccount == current.Account ||
		perfUserDepartmentScopeMatches(*current, target)
}

func perfUserDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	if departments := decodeStringList(user.ResponsibleDepartments); len(departments) > 0 {
		return perfUserResponsibleDepartmentScopeMatches(user, target)
	}
	return departmentScopeMatches(user, &target)
}

func departmentScopeMatches(leader model.DingTalkH5PerfUser, employee *model.DingTalkH5PerfUser) bool {
	if employee == nil {
		return false
	}
	leaderLevels := departmentLevelsFromUser(leader)
	employeeLevels := departmentLevelsFromUser(*employee)
	hasScope := false
	for _, item := range leaderLevels {
		if item != "" {
			hasScope = true
			break
		}
	}
	if !hasScope {
		return leader.Department != "" && leader.Department == employee.Department
	}
	for index, item := range leaderLevels {
		if item == "" {
			continue
		}
		if index >= len(employeeLevels) || item != employeeLevels[index] {
			return false
		}
	}
	return true
}

func perfUserResponsibleDepartmentScopeMatches(user model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	for _, department := range decodeStringList(user.ResponsibleDepartments) {
		for _, level := range departmentLevelsFromUser(target) {
			if department == level {
				return true
			}
		}
		if target.Department != "" && strings.Contains(target.Department, department) {
			return true
		}
	}
	return false
}

func userDTO(user model.DingTalkH5PerfUser) UserDTO {
	return UserDTO{
		ID:                     user.Account,
		Account:                user.Account,
		Name:                   user.Name,
		Avatar:                 user.Pic,
		Position:               user.Position,
		DepartmentID:           user.DepartmentID,
		Department:             user.Department,
		DepartmentLevel1:       user.DepartmentLevel1,
		DepartmentLevel2:       user.DepartmentLevel2,
		DepartmentLevel3:       user.DepartmentLevel3,
		DepartmentLevel4:       user.DepartmentLevel4,
		DepartmentLevels:       departmentLevelsFromUser(user),
		ManagerID:              user.ManagerAccount,
		HRBPID:                 user.HRBPAccount,
		ResponsibleDepartments: decodeStringList(user.ResponsibleDepartments),
		Status:                 user.Status,
	}
}

func departmentFromUser(user model.DingTalkH5PerfUser) string {
	if text := departmentText(departmentLevelsFromUser(user)...); text != "" {
		return text
	}
	return user.Department
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

func departmentLevelAt(levels []string, index int) string {
	if index < 0 || index >= len(levels) {
		return ""
	}
	return levels[index]
}

func departmentText(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, " / ")
}

func sanitizeAvatarURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > 500 {
		return "", fmt.Errorf("头像地址不能超过 500 个字符")
	}
	if value == "" {
		return "", nil
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(value, "/") {
		return value, nil
	}
	return "", fmt.Errorf("头像地址需使用 http(s) 或站内相对路径")
}

func notDeletedReviewQuery(db *gorm.DB) *gorm.DB {
	if db == nil {
		return db
	}
	return db.Where("`deleted_at` = 0")
}

func fallback(value, fallbackValue string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallbackValue)
}

func encodeJSON(value interface{}) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func decodeStringList(raw string) []string {
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func normalizeList(value interface{}) []string {
	switch typed := value.(type) {
	case []string:
		return uniqueStrings(typed)
	case []interface{}:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.ReplaceAll(toString(item), "，", ","), "、", ","), ",")))
		}
		return uniqueStrings(items)
	default:
		return uniqueStrings(strings.FieldsFunc(toString(value), func(r rune) bool {
			return r == ',' || r == '，' || r == '、' || r == ';' || r == '；' || r == '\n'
		}))
	}
}

func toString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		data, _ := json.Marshal(typed)
		return strings.Trim(string(data), `"`)
	}
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
