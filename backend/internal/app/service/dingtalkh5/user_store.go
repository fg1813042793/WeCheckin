package dingtalkh5

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
)

const perfUserMetaKey = "dingtalkH5Performance"

type perfUserMeta struct {
	Role                   string   `json:"role,omitempty"`
	Position               string   `json:"position,omitempty"`
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
	user.Role = strings.TrimSpace(meta.Role)
	if user.Role == "" {
		user.Role = "employee"
	}
	user.Position = strings.TrimSpace(meta.Position)
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
		Role:                   strings.TrimSpace(user.Role),
		Position:               strings.TrimSpace(user.Position),
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
	hydratePerfUser(&user)
	return &user, nil
}

func listPerfUsersDB(db *gorm.DB) ([]model.DingTalkH5PerfUser, error) {
	var users []model.DingTalkH5PerfUser
	if err := db.Order("`user_name` ASC, `id` ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	for index := range users {
		hydratePerfUser(&users[index])
	}
	return users, nil
}
