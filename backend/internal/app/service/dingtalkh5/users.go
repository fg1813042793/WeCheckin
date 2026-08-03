package dingtalkh5

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

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
		left := []string{users[i].DepartmentLevel1, users[i].DepartmentLevel2, users[i].Name, users[i].Account}
		right := []string{users[j].DepartmentLevel1, users[j].DepartmentLevel2, users[j].Name, users[j].Account}
		return strings.Join(left, "\x00") < strings.Join(right, "\x00")
	})
	result := make([]UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, userDTO(user))
	}
	return result, nil
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

func visiblePerfUsers(current *model.DingTalkH5PerfUser, users []model.DingTalkH5PerfUser, scope permissionsupport.DataScope) []model.DingTalkH5PerfUser {
	if current == nil {
		return nil
	}
	if !scope.Ready {
		return nil
	}
	if scope.Mode == 1 {
		return users
	}
	allowed := map[string]struct{}{}
	addAccount := func(account string) {
		account = NormalizeUserID(account)
		if account != "" {
			allowed[account] = struct{}{}
		}
	}
	addAccount(current.Account)
	if scope.Mode == 3 {
		return allowedPerfUsers(users, allowed)
	}
	addAccount(current.ManagerAccount)
	addAccount(current.HRBPAccount)
	for _, user := range users {
		if !canViewPerfUser(current, user) {
			continue
		}
		addAccount(user.Account)
		addAccount(user.ManagerAccount)
		addAccount(user.HRBPAccount)
	}
	return allowedPerfUsers(users, allowed)
}

func allowedPerfUsers(users []model.DingTalkH5PerfUser, allowed map[string]struct{}) []model.DingTalkH5PerfUser {
	result := make([]model.DingTalkH5PerfUser, 0, len(allowed))
	for _, user := range users {
		if _, ok := allowed[NormalizeUserID(user.Account)]; ok {
			result = append(result, user)
		}
	}
	return result
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
	dept1 := strings.TrimSpace(payload.DepartmentLevel1)
	dept2 := strings.TrimSpace(payload.DepartmentLevel2)
	dept3 := strings.TrimSpace(payload.DepartmentLevel3)
	if dept1 == "" && existing != nil {
		dept1 = existing.DepartmentLevel1
		dept2 = existing.DepartmentLevel2
		dept3 = existing.DepartmentLevel3
	}
	department := departmentText(dept1, dept2, dept3)
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
		DepartmentLevel1:       dept1,
		DepartmentLevel2:       dept2,
		DepartmentLevel3:       dept3,
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
		if payload.ManagerAccount == payload.Account {
			return fmt.Errorf("直属上级不能选择自己")
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
		Where("employee_account = ? AND status <> ?", target.Account, ReviewStatusCompleted).
		Updates(updates).Error
}

func userHasReferences(db *gorm.DB, account string) bool {
	var reviewCount int64
	notDeletedReviewQuery(db.Model(&model.DingTalkH5PerfReview{})).
		Where("employee_account = ? OR manager_account = ? OR hrbp_account = ?", account, account, account).
		Count(&reviewCount)
	if reviewCount > 0 {
		return true
	}
	users, err := listPerfUsersDB(db)
	if err != nil {
		return false
	}
	for _, user := range users {
		if user.Account == account {
			continue
		}
		if user.ManagerAccount == account || user.HRBPAccount == account {
			return true
		}
	}
	return false
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
