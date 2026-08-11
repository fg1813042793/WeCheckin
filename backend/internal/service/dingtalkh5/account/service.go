package account

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	onlineservice "wecheckin/backend/internal/service/admin/online"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

const perfUserMetaKey = "dingtalkH5Performance"

const perfUserSelectColumns = "`id`, `user_mini_openid`, `user_name`, `user_password`, `user_pic`, `user_status`, `user_role_id`, `user_position_id`, `user_obj`, `user_add_time`, `user_edit_time`"

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

type UserDTO struct {
	ID                     string   `json:"id"`
	Account                string   `json:"account"`
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

type AccountProfilePayload struct {
	Account         string `json:"account"`
	Avatar          string `json:"avatar"`
	CurrentPassword string `json:"currentPassword"`
}

type AccountProfileResponse struct {
	User UserDTO `json:"user"`
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

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func ChangePasswordContext(ctx context.Context, current *model.DingTalkH5PerfUser, currentPassword, newPassword, confirmPassword string) error {
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	confirmPassword = strings.TrimSpace(confirmPassword)
	if current == nil {
		return fmt.Errorf("未登录")
	}
	if !passwordutil.Verify(current.Password, currentPassword) {
		return fmt.Errorf("当前密码不正确")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码至少 6 位")
	}
	if newPassword != confirmPassword {
		return fmt.Errorf("两次输入的新密码不一致")
	}
	hash, err := passwordutil.Hash(newPassword)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.User{}).Where("`user_mini_openid` = ?", current.Account).Update("user_password", hash).Error
}

func UpdateAccountProfileContext(ctx context.Context, current *model.DingTalkH5PerfUser, token string, payload AccountProfilePayload) (*AccountProfileResponse, error) {
	if current == nil {
		return nil, fmt.Errorf("未登录")
	}
	nextAccount := NormalizeUserID(firstNonEmpty(payload.Account, current.Account))
	if nextAccount == "" {
		return nil, fmt.Errorf("请填写账号")
	}
	avatar, err := sanitizeAvatarURL(payload.Avatar)
	if err != nil {
		return nil, err
	}
	accountChanged := nextAccount != NormalizeUserID(current.Account)
	if accountChanged {
		if !passwordutil.Verify(current.Password, strings.TrimSpace(payload.CurrentPassword)) {
			return nil, fmt.Errorf("修改账号需填写正确的当前密码")
		}
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	var updated model.DingTalkH5PerfUser
	if err := db.Transaction(func(tx *gorm.DB) error {
		existing, err := loadPerfUserByIDDB(tx, current.ID)
		if err != nil {
			return fmt.Errorf("登录账号异常")
		}
		if nextAccount != NormalizeUserID(existing.Account) {
			if err := validateAccountAvailable(tx, nextAccount, existing.Account); err != nil {
				return err
			}
		}
		now := database.Now()
		updates := map[string]interface{}{
			"user_mini_openid": nextAccount,
			"user_pic":         avatar,
			"user_edit_time":   now,
		}
		if err := tx.Model(&model.User{}).Where("`id` = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		if accountChanged {
			if err := syncPerfAccountReferences(tx, NormalizeUserID(existing.Account), nextAccount); err != nil {
				return err
			}
		}
		loaded, err := loadPerfUserByIDDB(tx, existing.ID)
		if err != nil {
			return err
		}
		updated = *loaded
		return nil
	}); err != nil {
		return nil, err
	}

	if strings.TrimSpace(token) != "" {
		if err := onlineservice.UpdateDingTalkH5SessionUserContext(ctx, onlineUserFromDingTalkH5User(&updated), strings.TrimSpace(token)); err != nil {
			return nil, err
		}
	}
	return &AccountProfileResponse{User: userDTO(updated)}, nil
}

func validateAccountAvailable(db *gorm.DB, nextAccount, existingAccount string) error {
	nextAccount = NormalizeUserID(nextAccount)
	existingAccount = NormalizeUserID(existingAccount)
	if nextAccount == "" {
		return fmt.Errorf("请填写账号")
	}
	var duplicate model.User
	err := db.Where("`user_mini_openid` = ? AND `user_mini_openid` <> ?", nextAccount, existingAccount).First(&duplicate).Error
	if err == nil {
		return fmt.Errorf("账号已存在")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
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

func syncPerfAccountReferences(tx *gorm.DB, oldAccount, nextAccount string) error {
	oldAccount = NormalizeUserID(oldAccount)
	nextAccount = NormalizeUserID(nextAccount)
	if oldAccount == "" || nextAccount == "" || oldAccount == nextAccount {
		return nil
	}
	now := database.Now()
	reviewUpdates := []struct {
		column string
	}{
		{"employee_account"},
		{"manager_account"},
		{"hrbp_account"},
		{"hrbp_reviewer_account"},
	}
	for _, item := range reviewUpdates {
		if err := tx.Model(&model.DingTalkH5PerfReview{}).
			Where("`deleted_at` = 0").
			Where(item.column+" = ?", oldAccount).
			Updates(map[string]interface{}{item.column: nextAccount, "edit_time": now}).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&model.DingTalkH5PerfHistory{}).
		Where("`by_account` = ?", oldAccount).
		Update("by_account", nextAccount).Error; err != nil {
		return err
	}
	return syncPerfUserMetadataAccountReferences(tx, oldAccount, nextAccount)
}

func syncPerfUserMetadataAccountReferences(tx *gorm.DB, oldAccount, nextAccount string) error {
	var users []model.DingTalkH5PerfUser
	if err := tx.Where("`user_obj` LIKE ?", "%"+oldAccount+"%").Find(&users).Error; err != nil {
		return err
	}
	for _, user := range users {
		hydratePerfUser(&user)
		changed := false
		if NormalizeUserID(user.ManagerAccount) == oldAccount {
			user.ManagerAccount = nextAccount
			changed = true
		}
		if NormalizeUserID(user.HRBPAccount) == oldAccount {
			user.HRBPAccount = nextAccount
			changed = true
		}
		if !changed {
			continue
		}
		if err := tx.Model(&model.User{}).
			Where("`id` = ?", user.ID).
			Updates(map[string]interface{}{
				"user_obj":       encodePerfUserObj(user.Obj, user),
				"user_edit_time": database.Now(),
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func onlineUserFromDingTalkH5User(user *model.DingTalkH5PerfUser) *model.User {
	if user == nil {
		return nil
	}
	return &model.User{
		ID:         user.ID,
		MiniOpenID: user.Account,
		Status:     user.Status,
		Name:       user.Name,
		Pic:        user.Pic,
		Password:   user.Password,
		RoleID:     user.RoleID,
		RoleIDs:    user.RoleIDs,
	}
}

func loadPerfUserByIDDB(db *gorm.DB, userID uint) (*model.DingTalkH5PerfUser, error) {
	var user model.DingTalkH5PerfUser
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := db.Select(perfUserSelectColumns).Where("`id` = ?", userID).First(&user).Error; err != nil {
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
	paths, err := userDeptPathsByUserIDDB(db, uniqueUintIDs(userIDs))
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
		applyDepartmentPathToPerfUser(&users[index], paths[users[index].ID])
		applyPositionNameToPerfUser(&users[index], positionNames)
		applyRoleIDsToPerfUser(&users[index], roleIDsByUser)
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

func userDeptPathsByUserIDDB(db *gorm.DB, userIDs []uint) (map[uint][]string, error) {
	result := make(map[uint][]string, len(userIDs))
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
	for _, row := range rows {
		if row.UserID == 0 || row.DeptID == 0 {
			continue
		}
		if _, exists := result[row.UserID]; exists {
			continue
		}
		if levels := departmentPathLevels(deptByID, row.DeptID); len(levels) > 0 {
			result[row.UserID] = levels
		}
	}
	return result, nil
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

func userDTO(user model.DingTalkH5PerfUser) UserDTO {
	return UserDTO{
		ID:                     user.Account,
		Account:                user.Account,
		Name:                   user.Name,
		Avatar:                 user.Pic,
		Position:               user.Position,
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

func encodeJSON(value interface{}) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func decodeStringList(raw string) []string {
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	return items
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
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
