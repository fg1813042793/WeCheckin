package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	onlineservice "wecheckin/backend/internal/service/admin/online"
	bootstrapsvc "wecheckin/backend/internal/service/dingtalkh5/bootstrap"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	"wecheckin/backend/pkg/randutil"
	rd "wecheckin/backend/pkg/redis"
)

const (
	DingTalkH5BindRequiredCode = 10020

	perfUserMetaKey = "dingtalkH5Performance"

	maxSessionDeviceLength = 255

	dingTalkH5SelfBindEnabledKey   = "DINGTALK_H5_SELF_BIND_ENABLED"
	dingTalkH5SelfBindTicketTTLKey = "DINGTALK_H5_SELF_BIND_TICKET_TTL"
	dingTalkH5SelfBindTicketPrefix = "dingtalk_h5_bind_ticket:"
	dingTalkH5SelfBindDefaultTTL   = 10 * time.Minute
)

const perfUserSelectColumns = "`id`, `user_mini_openid`, `user_name`, `user_password`, `user_pic`, `user_status`, `user_role_id`, `user_position_id`, `user_obj`, `user_add_time`, `user_edit_time`"

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

type LoginResponse struct {
	Token                 string                           `json:"token"`
	UserInfo              UserDTO                          `json:"userInfo"`
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

type DingTalkH5BindRequiredResponse struct {
	BindTicket           string                           `json:"bindTicket"`
	CorpID               string                           `json:"corpId"`
	DingTalkUserIDMasked string                           `json:"dingTalkUserIdMasked"`
	UnionIDMasked        string                           `json:"unionIdMasked,omitempty"`
	ExpiresIn            int64                            `json:"expiresIn"`
	AppConfig            configsvc.DingTalkH5AppConfigDTO `json:"appConfig"`
	AppTitle             string                           `json:"appTitle"`
	AppName              string                           `json:"appName"`
	LogoText             string                           `json:"logoText"`
	LogoURL              string                           `json:"logoUrl"`
}

type DingTalkH5BindRequiredError struct {
	Response DingTalkH5BindRequiredResponse
}

type dingTalkH5SelfBindTicket struct {
	CorpID         string `json:"corpId"`
	DingTalkUserID string `json:"dingTalkUserId"`
	UnionID        string `json:"unionId"`
	IssuedAt       int64  `json:"issuedAt"`
	ExpiresAt      int64  `json:"expiresAt"`
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

func (e *DingTalkH5BindRequiredError) Error() string {
	return "钉钉账号未绑定系统用户"
}

func PublicConfigContext(ctx context.Context) (*configsvc.PublicConfigResponse, error) {
	return configsvc.PublicConfigContext(ctx)
}

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	name = strings.TrimSpace(name)
	password = strings.TrimSpace(password)
	if name == "" || password == "" {
		return nil, fmt.Errorf("请填写账号和密码")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.DingTalkH5PerfUser
	if err := db.Where("`user_mini_openid` = ? OR `user_name` = ?", NormalizeUserID(name), name).First(&user).Error; err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if !passwordutil.Verify(user.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(user.Password) {
		if hash, err := passwordutil.Hash(password); err == nil {
			_ = db.Model(&model.User{}).Where("`user_mini_openid` = ?", user.Account).Update("user_password", hash).Error
			user.Password = hash
		}
	}

	return issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)
}

func LoginByAuthCodeContext(ctx context.Context, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	return loginByAuthCodeContext(ctx, configsvc.DefaultDingTalkIdentityClient(), corpID, authCode, addIP, device)
}

func loginByAuthCodeContext(ctx context.Context, client configsvc.DingTalkIdentityClient, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	corpID = strings.TrimSpace(corpID)
	if corpID == "" {
		return nil, fmt.Errorf("钉钉企业 CorpId 不能为空")
	}
	authCode = strings.TrimSpace(authCode)
	if authCode == "" {
		return nil, fmt.Errorf("免登授权码不能为空")
	}
	config, err := configsvc.LoadDingTalkH5CorpConfigContext(ctx, corpID)
	if err != nil {
		return nil, err
	}
	identity, err := client.ExchangeAuthCodeContext(ctx, config, authCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(identity.UserID) == "" {
		return nil, fmt.Errorf("钉钉身份异常")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
	}

	binding, err := configsvc.LoadDingTalkH5UserBindingDB(db, corpID, identity.UserID)
	var loadedUser *model.DingTalkH5PerfUser
	if err != nil {
		if configsvc.IsMissingDingTalkH5UserBindingTableError(err) {
			loadedUser, err = loadLegacyDingTalkH5UserByIdentityDB(ctx, db, config, identity)
			if err != nil {
				return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			existingBinding, existingErr := configsvc.LoadAnyDingTalkH5UserBindingDB(db, corpID, identity.UserID)
			if existingErr == nil && existingBinding.Enabled != 1 {
				return nil, fmt.Errorf("钉钉账号绑定已停用，请联系管理员")
			}
			if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
			}
			return nil, newDingTalkH5BindRequiredErrorContext(ctx, corpID, identity)
		} else {
			return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
		}
	} else {
		loadedUser, err = loadPerfUserByIDDB(db, binding.UserID)
		if err != nil {
			return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
		}
	}
	user := *loadedUser
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)
}

func AuthenticateContext(ctx context.Context, token string) (*model.DingTalkH5PerfUser, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("未登录")
	}
	account, err := onlineservice.DingTalkH5SessionAccountContext(ctx, token)
	if err != nil {
		return nil, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.DingTalkH5PerfUser
	if err := db.Select(perfUserSelectColumns).Where("`user_mini_openid` = ? AND `user_status` = 1", account).First(&user).Error; err != nil {
		return nil, fmt.Errorf("登录账号异常")
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func LogoutContext(ctx context.Context, current *model.DingTalkH5PerfUser, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	token = strings.TrimSpace(token)

	userID := uint(0)
	if current != nil {
		userID = current.ID
	}
	if userID == 0 {
		account, err := onlineservice.DingTalkH5SessionAccountContext(ctx, token)
		if err == nil && strings.TrimSpace(account) != "" {
			db, cancel := database.WithContext(ctx)
			defer cancel()
			if db != nil {
				user, err := loadPerfUserByAccountDB(db, account)
				if err == nil && user != nil {
					userID = user.ID
				}
			}
		}
	}

	return onlineservice.RemoveDingTalkH5SessionContext(ctx, userID, token)
}

func BindSelfContext(ctx context.Context, bindTicket, account, password, addIP, device string) (*LoginResponse, error) {
	bindTicket = strings.TrimSpace(bindTicket)
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if !SelfBindEnabledContext(ctx) {
		return nil, fmt.Errorf("钉钉自助绑定已关闭")
	}
	if bindTicket == "" {
		return nil, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	if account == "" || password == "" {
		return nil, fmt.Errorf("请填写系统账号和密码")
	}

	ticket, err := readDingTalkH5SelfBindTicketContext(ctx, bindTicket)
	if err != nil {
		return nil, err
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	baseUser, err := loadSelfBindUserByAccountDB(db, account)
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if baseUser.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if !passwordutil.Verify(baseUser.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(baseUser.Password) {
		if hash, err := passwordutil.Hash(password); err == nil {
			_ = db.Model(&model.User{}).Where("`id` = ?", baseUser.ID).Update("user_password", hash).Error
		}
	}

	if err := bindDingTalkH5IdentityToUserDB(db, ticket, baseUser.ID); err != nil {
		return nil, err
	}
	_ = deleteDingTalkH5SelfBindTicketContext(ctx, bindTicket)

	user, err := loadPerfUserByIDDB(db, baseUser.ID)
	if err != nil {
		return nil, fmt.Errorf("账号异常")
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	return issueDingTalkH5LoginResponseContext(ctx, db, user, addIP, device)
}

func DingTalkH5BindRequiredData(err error) (DingTalkH5BindRequiredResponse, bool) {
	var bindErr *DingTalkH5BindRequiredError
	if errors.As(err, &bindErr) && bindErr != nil {
		return bindErr.Response, true
	}
	return DingTalkH5BindRequiredResponse{}, false
}

func SelfBindEnabledContext(ctx context.Context) bool {
	value := strings.ToLower(strings.TrimSpace(configsvc.SetupValueContext(ctx, dingTalkH5SelfBindEnabledKey)))
	if value == "" {
		return true
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func issueDingTalkH5LoginResponseContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, addIP, device string) (*LoginResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("登录信息异常")
	}
	token := randutil.HexString(32)
	safeDevice := sessionDevice(device)
	if err := onlineservice.StoreDingTalkH5SessionContext(ctx, onlineUserFromDingTalkH5User(user), token, addIP, safeDevice); err != nil {
		return nil, err
	}
	bootstrap, err := bootstrapsvc.BootstrapContext(ctx, user)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token:                 token,
		UserInfo:              bootstrapUserDTO(bootstrap),
		Menus:                 bootstrapAppMenus(bootstrap.Menus),
		AppConfig:             bootstrap.AppConfig,
		AppTitle:              bootstrap.AppTitle,
		AppName:               bootstrap.AppName,
		LogoText:              bootstrap.LogoText,
		LogoURL:               bootstrap.LogoURL,
		ButtonPermissionKeys:  bootstrap.ButtonPermissionKeys,
		ButtonPermissionReady: bootstrap.ButtonPermissionReady,
		APIPermissionKeys:     bootstrap.APIPermissionKeys,
		APIPermissionReady:    bootstrap.APIPermissionReady,
		PermissionVersion:     bootstrap.PermissionVersion,
	}, nil
}

func bootstrapUserDTO(response *bootstrapsvc.BootstrapResponse) UserDTO {
	if response == nil {
		return UserDTO{}
	}
	user := response.User
	return UserDTO{
		ID:                     user.ID,
		Account:                user.Account,
		WorkflowActorID:        user.WorkflowActorID,
		Name:                   user.Name,
		Avatar:                 user.Avatar,
		Position:               user.Position,
		Department:             user.Department,
		DepartmentLevel1:       user.DepartmentLevel1,
		DepartmentLevel2:       user.DepartmentLevel2,
		DepartmentLevel3:       user.DepartmentLevel3,
		DepartmentLevel4:       user.DepartmentLevel4,
		DepartmentLevels:       append([]string(nil), user.DepartmentLevels...),
		ManagerID:              user.ManagerID,
		HRBPID:                 user.HRBPID,
		ResponsibleDepartments: append([]string(nil), user.ResponsibleDepartments...),
		Status:                 user.Status,
	}
}

func bootstrapAppMenus(items []bootstrapsvc.AppMenuDTO) []AppMenuDTO {
	if len(items) == 0 {
		return nil
	}
	result := make([]AppMenuDTO, 0, len(items))
	for _, item := range items {
		result = append(result, AppMenuDTO{
			Key:           item.Key,
			Label:         item.Label,
			Icon:          item.Icon,
			PermissionKey: item.PermissionKey,
			Children:      bootstrapAppMenus(item.Children),
		})
	}
	return result
}

func sessionDevice(device string) string {
	device = strings.TrimSpace(device)
	runes := []rune(device)
	if len(runes) <= maxSessionDeviceLength {
		return device
	}
	return string(runes[:maxSessionDeviceLength])
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

func loadLegacyDingTalkH5UserByIdentityDB(ctx context.Context, db *gorm.DB, config configsvc.DingTalkH5CorpConfig, identity configsvc.DingTalkUserIdentity) (*model.DingTalkH5PerfUser, error) {
	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	legacy := configsvc.LegacyDingTalkH5CorpConfigContext(ctx)
	if strings.TrimSpace(legacy.CorpID) != strings.TrimSpace(config.CorpID) {
		return nil, gorm.ErrRecordNotFound
	}
	return loadPerfUserByAccountDB(db, identity.UserID)
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

func newDingTalkH5BindRequiredErrorContext(ctx context.Context, corpID string, identity configsvc.DingTalkUserIdentity) error {
	if !SelfBindEnabledContext(ctx) {
		return fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
	}
	ticket, expiresIn, err := createDingTalkH5SelfBindTicketContext(ctx, corpID, identity)
	if err != nil {
		return fmt.Errorf("自助绑定暂不可用，请联系管理员")
	}
	appConfig := configsvc.AppConfigContext(ctx)
	return &DingTalkH5BindRequiredError{
		Response: DingTalkH5BindRequiredResponse{
			BindTicket:           ticket,
			CorpID:               strings.TrimSpace(corpID),
			DingTalkUserIDMasked: maskDingTalkIdentity(identity.UserID),
			UnionIDMasked:        maskDingTalkIdentity(identity.UnionID),
			ExpiresIn:            int64(expiresIn.Seconds()),
			AppConfig:            appConfig,
			AppTitle:             appConfig.AppTitle,
			AppName:              appConfig.AppName,
			LogoText:             appConfig.LogoText,
			LogoURL:              appConfig.LogoURL,
		},
	}
}

func bindDingTalkH5IdentityToUserDB(db *gorm.DB, ticket dingTalkH5SelfBindTicket, userID uint) error {
	if db == nil || userID == 0 {
		return fmt.Errorf("绑定信息异常")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		identityBinding, err := configsvc.LoadAnyDingTalkH5UserBindingDB(tx, ticket.CorpID, ticket.DingTalkUserID)
		if err == nil {
			if identityBinding.Enabled != 1 {
				return fmt.Errorf("钉钉账号绑定已停用，请联系管理员")
			}
			if identityBinding.UserID != userID {
				return fmt.Errorf("该钉钉账号已绑定其他系统用户，请联系管理员")
			}
			return tx.Model(identityBinding).Updates(map[string]interface{}{
				"union_id":  ticket.UnionID,
				"edit_time": database.Now(),
			}).Error
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var disabledUserBindingCount int64
		if err := tx.Model(&model.DingTalkH5UserBinding{}).
			Where("`corp_id` = ? AND `user_id` = ? AND `enabled` <> 1", ticket.CorpID, userID).
			Count(&disabledUserBindingCount).Error; err != nil {
			return err
		}
		if disabledUserBindingCount > 0 {
			return fmt.Errorf("该系统账号绑定已停用，请联系管理员")
		}

		var otherUserBindingCount int64
		if err := tx.Model(&model.DingTalkH5UserBinding{}).
			Where("`corp_id` = ? AND `user_id` = ? AND `enabled` = 1 AND `dingtalk_user_id` <> ?", ticket.CorpID, userID, ticket.DingTalkUserID).
			Count(&otherUserBindingCount).Error; err != nil {
			return err
		}
		if otherUserBindingCount > 0 {
			return fmt.Errorf("该系统账号已绑定其他钉钉用户，请联系管理员")
		}

		if strings.TrimSpace(ticket.UnionID) != "" {
			var unionConflictCount int64
			if err := tx.Model(&model.DingTalkH5UserBinding{}).
				Where("`corp_id` = ? AND `union_id` = ? AND `enabled` = 1 AND `user_id` <> ?", ticket.CorpID, ticket.UnionID, userID).
				Count(&unionConflictCount).Error; err != nil {
				return err
			}
			if unionConflictCount > 0 {
				return fmt.Errorf("该钉钉账号已绑定其他系统用户，请联系管理员")
			}
		}

		now := database.Now()
		return tx.Create(&model.DingTalkH5UserBinding{
			CorpID:         ticket.CorpID,
			DingTalkUserID: ticket.DingTalkUserID,
			UnionID:        ticket.UnionID,
			UserID:         userID,
			Enabled:        1,
			AddTime:        now,
			EditTime:       now,
		}).Error
	})
}

func loadSelfBindUserByAccountDB(db *gorm.DB, account string) (*model.User, error) {
	account = strings.TrimSpace(account)
	normalizedAccount := NormalizeUserID(account)
	var user model.User
	if err := db.Where(
		"`user_mini_openid` = ? OR `user_name` = ? OR `user_account` = ? OR `user_mobile` = ?",
		normalizedAccount,
		account,
		account,
		account,
	).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func createDingTalkH5SelfBindTicketContext(ctx context.Context, corpID string, identity configsvc.DingTalkUserIdentity) (string, time.Duration, error) {
	if rd.RDB == nil {
		return "", 0, fmt.Errorf("redis is not initialized")
	}
	ttl := dingTalkH5SelfBindTicketTTLContext(ctx)
	now := time.Now()
	ticket := randutil.HexString(48)
	payload := dingTalkH5SelfBindTicket{
		CorpID:         strings.TrimSpace(corpID),
		DingTalkUserID: strings.TrimSpace(identity.UserID),
		UnionID:        strings.TrimSpace(identity.UnionID),
		IssuedAt:       now.UnixMilli(),
		ExpiresAt:      now.Add(ttl).UnixMilli(),
	}
	if payload.CorpID == "" || payload.DingTalkUserID == "" {
		return "", 0, fmt.Errorf("invalid dingTalk bind ticket payload")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	if err := rd.RDB.Set(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket), string(raw), ttl).Err(); err != nil {
		return "", 0, err
	}
	return ticket, ttl, nil
}

func readDingTalkH5SelfBindTicketContext(ctx context.Context, ticket string) (dingTalkH5SelfBindTicket, error) {
	if rd.RDB == nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	raw, err := rd.RDB.Get(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket)).Result()
	if errors.Is(err, goredis.Nil) {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	if err != nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话读取失败，请稍后重试")
	}
	var payload dingTalkH5SelfBindTicket
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话异常，请重新打开应用")
	}
	if strings.TrimSpace(payload.CorpID) == "" || strings.TrimSpace(payload.DingTalkUserID) == "" || time.Now().UnixMilli() > payload.ExpiresAt {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	return payload, nil
}

func deleteDingTalkH5SelfBindTicketContext(ctx context.Context, ticket string) error {
	if rd.RDB == nil {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	return rd.RDB.Del(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket)).Err()
}

func dingTalkH5SelfBindTicketRedisKey(ticket string) string {
	return dingTalkH5SelfBindTicketPrefix + strings.TrimSpace(ticket)
}

func dingTalkH5SelfBindTicketTTLContext(ctx context.Context) time.Duration {
	value := strings.TrimSpace(configsvc.SetupValueContext(ctx, dingTalkH5SelfBindTicketTTLKey))
	if value == "" {
		return dingTalkH5SelfBindDefaultTTL
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if duration, err := time.ParseDuration(value); err == nil && duration > 0 {
		return duration
	}
	return dingTalkH5SelfBindDefaultTTL
}

func maskDingTalkIdentity(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= 4 {
		return string(runes[:1]) + "***"
	}
	return string(runes[:2]) + "****" + string(runes[len(runes)-2:])
}
