package dingtalkh5

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	permissionsupport "wecheckin/backend/internal/app/support/permission"
	"wecheckin/backend/internal/model"
)

const perfUserMetaKey = "dingtalkH5Performance"

type perfUserMeta struct {
	Department             string   `json:"department,omitempty"`
	DepartmentLevel1       string   `json:"departmentLevel1,omitempty"`
	DepartmentLevel2       string   `json:"departmentLevel2,omitempty"`
	DepartmentLevel3       string   `json:"departmentLevel3,omitempty"`
	ManagerID              string   `json:"managerId,omitempty"`
	HRBPID                 string   `json:"hrbpId,omitempty"`
	ResponsibleDepartments []string `json:"responsibleDepartments,omitempty"`
}

func hydratePerfUser(user *model.DingTalkH5PerfUser) {
	if user == nil {
		return
	}
	meta := decodePerfUserMeta(user.Obj)
	user.Role = ""
	user.Position = ""
	user.Department = strings.TrimSpace(meta.Department)
	user.DepartmentLevel1 = strings.TrimSpace(meta.DepartmentLevel1)
	user.DepartmentLevel2 = strings.TrimSpace(meta.DepartmentLevel2)
	user.DepartmentLevel3 = strings.TrimSpace(meta.DepartmentLevel3)
	user.ManagerAccount = NormalizeUserID(meta.ManagerID)
	user.HRBPAccount = NormalizeUserID(meta.HRBPID)
	user.ResponsibleDepartments = encodeJSON(uniqueStrings(meta.ResponsibleDepartments))
}

func encodePerfUserObj(raw string, user model.DingTalkH5PerfUser) string {
	obj := decodeObject(raw)
	obj[perfUserMetaKey] = perfUserMeta{
		Department:             strings.TrimSpace(user.Department),
		DepartmentLevel1:       strings.TrimSpace(user.DepartmentLevel1),
		DepartmentLevel2:       strings.TrimSpace(user.DepartmentLevel2),
		DepartmentLevel3:       strings.TrimSpace(user.DepartmentLevel3),
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

func loadPerfUserByAccountDB(db *gorm.DB, account string) (*model.DingTalkH5PerfUser, error) {
	account = NormalizeUserID(account)
	var user model.DingTalkH5PerfUser
	if err := db.Where("`user_mini_openid` = ?", account).First(&user).Error; err != nil {
		return nil, err
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func loadPerfUserByIDDB(db *gorm.DB, userID uint) (*model.DingTalkH5PerfUser, error) {
	var user model.DingTalkH5PerfUser
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := db.Where("`id` = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func listPerfUsersDB(db *gorm.DB) ([]model.DingTalkH5PerfUser, error) {
	var users []model.DingTalkH5PerfUser
	if err := db.Where("`user_status` = 1").Order("`user_name` ASC, `id` ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return hydratePerfUsersWithUserDeptsDB(db, users)
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
	if err := db.Where("`user_mini_openid` IN ? AND `user_status` = 1", accounts).Order("`user_name` ASC, `id` ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return hydratePerfUsersWithUserDeptsDB(db, users)
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
	reversed := make([]string, 0, 3)
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
	if len(cleanLevels) > 1 {
		user.DepartmentLevel2 = cleanLevels[1]
	}
	if len(cleanLevels) > 2 {
		user.DepartmentLevel3 = strings.Join(cleanLevels[2:], " / ")
	}
}
