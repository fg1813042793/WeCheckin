package dingtalkh5

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func ListUsersContext(ctx context.Context, current *model.DingTalkH5PerfUser) ([]UserDTO, error) {
	if err := EnsureSeedContext(ctx); err != nil {
		return nil, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	users, err := listPerfUsersDB(db)
	if err != nil {
		return nil, err
	}
	users = visiblePerfUsers(current, users)
	sort.SliceStable(users, func(i, j int) bool {
		left := []string{users[i].DepartmentLevel1, users[i].DepartmentLevel2, users[i].Role, users[i].Name, users[i].Account}
		right := []string{users[j].DepartmentLevel1, users[j].DepartmentLevel2, users[j].Role, users[j].Name, users[j].Account}
		return strings.Join(left, "\x00") < strings.Join(right, "\x00")
	})
	result := make([]UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, userDTO(user))
	}
	return result, nil
}

func visiblePerfUsers(current *model.DingTalkH5PerfUser, users []model.DingTalkH5PerfUser) []model.DingTalkH5PerfUser {
	if current == nil {
		return nil
	}
	if current.Role == "admin" || current.Role == "hrbp_manager" {
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
	result := make([]model.DingTalkH5PerfUser, 0, len(allowed))
	for _, user := range users {
		if _, ok := allowed[user.Account]; ok {
			result = append(result, user)
		}
	}
	return result
}

func canViewPerfUser(current *model.DingTalkH5PerfUser, target model.DingTalkH5PerfUser) bool {
	if current == nil {
		return false
	}
	if current.Role == "admin" || current.Role == "hrbp_manager" {
		return true
	}
	if target.Account == current.Account || target.Account == current.ManagerAccount || target.Account == current.HRBPAccount {
		return true
	}
	if current.Role == "hrbp" {
		return target.HRBPAccount == current.Account || perfUserResponsibleDepartmentScopeMatches(*current, target)
	}
	if _, ok := peopleLeaderRoles[current.Role]; ok {
		return target.ManagerAccount == current.Account || departmentScopeMatches(*current, &target)
	}
	return false
}

func CreateUserContext(ctx context.Context, current *model.DingTalkH5PerfUser, payload UserPayload) (*UserDTO, []UserDTO, error) {
	if !isAdmin(current) {
		return nil, nil, fmt.Errorf("当前账号不能调整组织架构")
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
	if !isAdmin(current) {
		return nil, nil, fmt.Errorf("当前账号不能调整组织架构")
	}
	account = NormalizeUserID(account)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	existing, err := loadPerfUserByAccountDB(db, account)
	if err != nil {
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
			"user_password":  next.Password,
			"user_status":    next.Status,
			"user_obj":       encodePerfUserObj(existing.Obj, next),
			"user_edit_time": next.EditTime,
		}
		if err := tx.Model(&model.User{}).Where("`user_mini_openid` = ?", next.Account).Updates(updates).Error; err != nil {
			return err
		}
		return syncOpenReviewsForUser(tx, next)
	}); err != nil {
		return nil, nil, err
	}
	users, err := ListUsersContext(ctx, current)
	if err != nil {
		return nil, nil, err
	}
	dto := userDTO(next)
	return &dto, users, nil
}

func DeleteUserContext(ctx context.Context, current *model.DingTalkH5PerfUser, account string) ([]UserDTO, error) {
	if !isAdmin(current) {
		return nil, fmt.Errorf("当前账号不能调整组织架构")
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
	if userHasReferences(db, target.Account) {
		return nil, fmt.Errorf("该账号已有组织关系或绩效记录，不能直接删除")
	}
	if err := db.Where("`user_mini_openid` = ?", target.Account).Delete(&model.User{}).Error; err != nil {
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
	role := strings.TrimSpace(payload.Role)
	if role == "" && existing != nil {
		role = existing.Role
	}
	if role == "" {
		role = "employee"
	}
	position := strings.TrimSpace(payload.Position)
	if position == "" && existing != nil {
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
	return model.DingTalkH5PerfUser{
		Account:                account,
		Name:                   name,
		Password:               passwordHash,
		Role:                   role,
		Position:               position,
		Department:             department,
		DepartmentLevel1:       dept1,
		DepartmentLevel2:       dept2,
		DepartmentLevel3:       dept3,
		ManagerAccount:         NormalizeUserID(payload.ManagerID),
		HRBPAccount:            NormalizeUserID(payload.HRBPID),
		ResponsibleDepartments: encodeJSON(normalizeList(payload.ResponsibleDepartments)),
		Status:                 status,
		AddTime:                addTime,
		EditTime:               now,
	}, nil
}

func validateUserPayload(db *gorm.DB, payload model.DingTalkH5PerfUser, existingAccount string) error {
	if payload.Account == "" || payload.Name == "" {
		return fmt.Errorf("请填写账号和姓名")
	}
	if _, ok := editableRoles[payload.Role]; !ok {
		return fmt.Errorf("请选择有效角色")
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
		manager, err := loadPerfUserByAccountDB(db, payload.ManagerAccount)
		if err != nil {
			return fmt.Errorf("直属上级不存在")
		}
		if _, ok := peopleLeaderRoles[manager.Role]; !ok && manager.Role != "admin" && manager.Role != "hrbp_manager" {
			return fmt.Errorf("直属上级需选择总监、经理、主管或管理员")
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
	return db.Model(&model.DingTalkH5PerfReview{}).
		Where("employee_account = ? AND status <> ?", target.Account, ReviewStatusCompleted).
		Updates(updates).Error
}

func userHasReferences(db *gorm.DB, account string) bool {
	var reviewCount int64
	db.Model(&model.DingTalkH5PerfReview{}).
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
