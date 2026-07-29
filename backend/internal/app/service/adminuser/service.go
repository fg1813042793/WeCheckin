package adminuser

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	setupservice "wecheckin-backend/backend/internal/app/service/setup"
	"wecheckin-backend/backend/internal/app/support/access"
	deptsupport "wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/media"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/app/support/query"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func GetUserByOpenID(openID string) (*model.User, error) {
	return GetUserByOpenIDContext(context.Background(), openID)
}

func GetUserByOpenIDContext(ctx context.Context, openID string) (*model.User, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.User
	err := db.Where("`user_mini_openid` = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByOpenIDForAdminContext(ctx context.Context, openID string, adminID uint) (*model.User, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := userVisibleQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := queryBuilder.Where("`user_mini_openid` = ?", openID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type UserDetail struct {
	ID                  uint     `json:"id"`
	Name                string   `json:"name"`
	Mobile              string   `json:"mobile"`
	Avatar              string   `json:"avatar"`
	Pic                 string   `json:"pic"`
	Status              int      `json:"status"`
	Forms               string   `json:"forms"`
	LoginCnt            int      `json:"loginCnt"`
	AddTime             int64    `json:"addTime"`
	LoginTime           int64    `json:"loginTime"`
	PositionID          uint     `json:"positionId"`
	PositionName        string   `json:"positionName"`
	RoleID              uint     `json:"roleId"`
	RoleName            string   `json:"roleName"`
	DeptIDs             []uint   `json:"deptIds"`
	DeptNames           []string `json:"deptNames"`
	TopDeptNames        []string `json:"topDeptNames"`
	AllowPermissionKeys []string `json:"allowPermissionKeys"`
	DenyPermissionKeys  []string `json:"denyPermissionKeys"`
}

func GetUserByID(id string) (UserDetail, error) {
	return GetUserByIDContext(context.Background(), id)
}

func GetUserByIDContext(ctx context.Context, id string) (UserDetail, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.User
	err := db.Where("`id` = ?", id).First(&user).Error
	if err != nil {
		return UserDetail{}, err
	}
	deptIDs := deptsupport.UserDeptIDsContext(ctx, user.ID)
	var deptNames []string
	var topDeptNames []string
	deptNameByID := map[uint]string{}
	if len(deptIDs) > 0 {
		var depts []model.Department
		if err := db.Where("`id` IN ?", deptIDs).Find(&depts).Error; err != nil {
			return UserDetail{}, err
		}
		for _, d := range depts {
			deptNameByID[d.ID] = d.Name
		}
	}
	for _, did := range deptIDs {
		if name := deptNameByID[did]; name != "" {
			deptNames = append(deptNames, name)
		}
		topDeptNames = append(topDeptNames, deptsupport.TopDeptNameContext(ctx, did))
	}
	positionNames, err := loadPositionNameMapContext(ctx, db, []model.User{user})
	if err != nil {
		return UserDetail{}, err
	}
	roleNames, err := loadRoleNameMapContext(ctx, db, []model.User{user})
	if err != nil {
		return UserDetail{}, err
	}
	allowPermissionKeys, denyPermissionKeys, err := permissionsupport.UserApplicationMenuPermissionKeySetsContext(ctx, db, user.ID)
	if err != nil {
		return UserDetail{}, err
	}
	avatar := media.FullURLWithStaticDomain(user.Pic)
	return UserDetail{
		ID:                  user.ID,
		Name:                user.Name,
		Mobile:              user.Mobile,
		Avatar:              avatar,
		Pic:                 avatar,
		Status:              user.Status,
		Forms:               user.Forms,
		LoginCnt:            user.LoginCnt,
		AddTime:             user.AddTime,
		LoginTime:           user.LoginTime,
		PositionID:          user.PositionID,
		PositionName:        positionNames[user.PositionID],
		RoleID:              user.RoleID,
		RoleName:            roleNames[user.RoleID],
		DeptIDs:             deptIDs,
		DeptNames:           deptNames,
		TopDeptNames:        topDeptNames,
		AllowPermissionKeys: allowPermissionKeys,
		DenyPermissionKeys:  denyPermissionKeys,
	}, nil
}

func GetUserByIDForAdminContext(ctx context.Context, id string, adminID uint) (UserDetail, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := userVisibleQueryContext(ctx, db, adminID)
	if err != nil {
		return UserDetail{}, err
	}
	var user model.User
	if err := queryBuilder.Where("`id` = ?", id).First(&user).Error; err != nil {
		return UserDetail{}, err
	}
	deptIDs := deptsupport.UserDeptIDsContext(ctx, user.ID)
	var deptNames []string
	var topDeptNames []string
	deptNameByID := map[uint]string{}
	if len(deptIDs) > 0 {
		var depts []model.Department
		if err := db.Where("`id` IN ?", deptIDs).Find(&depts).Error; err != nil {
			return UserDetail{}, err
		}
		for _, d := range depts {
			deptNameByID[d.ID] = d.Name
		}
	}
	for _, did := range deptIDs {
		if name := deptNameByID[did]; name != "" {
			deptNames = append(deptNames, name)
		}
		topDeptNames = append(topDeptNames, deptsupport.TopDeptNameContext(ctx, did))
	}
	positionNames, err := loadPositionNameMapContext(ctx, db, []model.User{user})
	if err != nil {
		return UserDetail{}, err
	}
	roleNames, err := loadRoleNameMapContext(ctx, db, []model.User{user})
	if err != nil {
		return UserDetail{}, err
	}
	allowPermissionKeys, denyPermissionKeys, err := permissionsupport.UserApplicationMenuPermissionKeySetsContext(ctx, db, user.ID)
	if err != nil {
		return UserDetail{}, err
	}
	avatar := media.FullURLWithStaticDomain(user.Pic)
	return UserDetail{
		ID:                  user.ID,
		Name:                user.Name,
		Mobile:              user.Mobile,
		Avatar:              avatar,
		Pic:                 avatar,
		Status:              user.Status,
		Forms:               user.Forms,
		LoginCnt:            user.LoginCnt,
		AddTime:             user.AddTime,
		LoginTime:           user.LoginTime,
		PositionID:          user.PositionID,
		PositionName:        positionNames[user.PositionID],
		RoleID:              user.RoleID,
		RoleName:            roleNames[user.RoleID],
		DeptIDs:             deptIDs,
		DeptNames:           deptNames,
		TopDeptNames:        topDeptNames,
		AllowPermissionKeys: allowPermissionKeys,
		DenyPermissionKeys:  denyPermissionKeys,
	}, nil
}

func saveUserDepts(ctx context.Context, userID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return saveUserDeptsTx(db, userID, deptIDs)
}

func saveUserDeptsTx(tx *gorm.DB, userID uint, deptIDs []uint) error {
	if err := tx.Where("`user_dept_user_id` = ?", userID).Delete(&model.UserDept{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if deptID > 0 {
			if err := tx.Create(&model.UserDept{UserID: userID, DeptID: deptID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type UserListItem struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Mobile       string `json:"mobile"`
	Avatar       string `json:"avatar"`
	Pic          string `json:"pic"`
	Status       int    `json:"status"`
	LoginCnt     int    `json:"loginCnt"`
	AddTime      int64  `json:"addTime"`
	LoginTime    int64  `json:"loginTime"`
	DeptIDs      []uint `json:"deptIds"`
	PositionID   uint   `json:"positionId"`
	PositionName string `json:"positionName"`
	RoleID       uint   `json:"roleId"`
	RoleName     string `json:"roleName"`
}

var userListColumns = []string{
	"id",
	"user_name",
	"user_mobile",
	"user_position_id",
	"user_role_id",
	"user_pic",
	"user_status",
	"user_login_cnt",
	"user_add_time",
	"user_login_time",
}

func GetUserList(keyword, sortStr string, page, pageSize int, adminID uint) ([]UserListItem, int64, error) {
	return GetUserListContext(context.Background(), keyword, sortStr, page, pageSize, adminID)
}

func userVisibleQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	queryBuilder := db.Model(&model.User{})
	where, args := access.UserDataScopeFilterContext(ctx, &admin)
	if where != "" {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	return queryBuilder, nil
}

func GetUserListContext(ctx context.Context, keyword, sortStr string, page, pageSize int, adminID uint) ([]UserListItem, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return nil, 0, err
	}
	var list []model.User
	var total int64
	queryBuilder := db.Model(&model.User{})
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		queryBuilder = queryBuilder.Where("`user_name` LIKE ? OR `user_mobile` LIKE ?", likeKeyword, likeKeyword)
	}
	where, args := access.UserDataScopeFilterContext(ctx, &admin)
	hasDataScopeFilter := where != ""
	if hasDataScopeFilter {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	useTotalCountCache := keyword == "" && !hasDataScopeFilter
	if useTotalCountCache {
		now := time.Now()
		if cachedTotal, ok := getUserTotalCountCache(now); ok {
			total = cachedTotal
		} else {
			if err := queryBuilder.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			setUserTotalCountCache(total, now)
		}
	} else {
		if err := queryBuilder.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}
	orderClause := query.ParseSort(sortStr, map[string]string{
		"name":     "user_name",
		"mobile":   "user_mobile",
		"status":   "user_status",
		"loginCnt": "user_login_cnt",
		"addTime":  "user_add_time",
	})
	if orderClause == "" {
		orderClause = "`user_add_time` DESC, `id` DESC"
	}
	err := queryBuilder.Select(userListColumns).Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	deptIDsByUser, err := loadUserDeptIDMapContext(ctx, db, list)
	if err != nil {
		return nil, 0, err
	}
	positionNames, err := loadPositionNameMapContext(ctx, db, list)
	if err != nil {
		return nil, 0, err
	}
	roleNames, err := loadRoleNameMapContext(ctx, db, list)
	if err != nil {
		return nil, 0, err
	}
	result := make([]UserListItem, len(list))
	for i, u := range list {
		avatar := media.FullURLWithStaticDomain(u.Pic)
		result[i] = UserListItem{
			ID:           u.ID,
			Name:         u.Name,
			Mobile:       u.Mobile,
			Avatar:       avatar,
			Pic:          avatar,
			Status:       u.Status,
			LoginCnt:     u.LoginCnt,
			AddTime:      u.AddTime,
			LoginTime:    u.LoginTime,
			DeptIDs:      deptIDsByUser[u.ID],
			PositionID:   u.PositionID,
			PositionName: positionNames[u.PositionID],
			RoleID:       u.RoleID,
			RoleName:     roleNames[u.RoleID],
		}
	}
	return result, total, nil
}

func loadUserDeptIDMapContext(ctx context.Context, db *gorm.DB, list []model.User) (map[uint][]uint, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	userIDs := make([]uint, 0, len(list))
	deptIDsByUser := make(map[uint][]uint, len(list))
	for _, item := range list {
		userIDs = append(userIDs, item.ID)
		deptIDsByUser[item.ID] = []uint{}
	}
	if len(userIDs) == 0 {
		return deptIDsByUser, nil
	}
	var rows []model.UserDept
	if err := db.Where("`user_dept_user_id` IN ?", userIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		deptIDsByUser[row.UserID] = append(deptIDsByUser[row.UserID], row.DeptID)
	}
	return deptIDsByUser, nil
}

func loadPositionNameMapContext(ctx context.Context, db *gorm.DB, list []model.User) (map[uint]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	positionIDSet := make(map[uint]struct{})
	for _, item := range list {
		if item.PositionID > 0 {
			positionIDSet[item.PositionID] = struct{}{}
		}
	}
	positionNames := make(map[uint]string, len(positionIDSet))
	if len(positionIDSet) == 0 {
		return positionNames, nil
	}
	positionIDs := make([]uint, 0, len(positionIDSet))
	for id := range positionIDSet {
		positionIDs = append(positionIDs, id)
	}
	var positions []model.Position
	if err := db.Select("id", "position_name").Where("`id` IN ?", positionIDs).Find(&positions).Error; err != nil {
		return nil, err
	}
	for _, item := range positions {
		positionNames[item.ID] = item.Name
	}
	return positionNames, nil
}

func loadRoleNameMapContext(ctx context.Context, db *gorm.DB, list []model.User) (map[uint]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	roleIDSet := make(map[uint]struct{})
	for _, item := range list {
		if item.RoleID > 0 {
			roleIDSet[item.RoleID] = struct{}{}
		}
	}
	roleNames := make(map[uint]string, len(roleIDSet))
	if len(roleIDSet) == 0 {
		return roleNames, nil
	}
	roleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}
	var roles []model.Role
	if err := db.Select("id", "role_name").Where("`id` IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	for _, item := range roles {
		roleNames[item.ID] = item.Name
	}
	return roleNames, nil
}

type AdminAccessInput struct {
	Password              string
	RoleID                uint
	AllowPermissionKeys   []string
	DenyPermissionKeys    []string
	PermissionKeysTouched bool
}

func AddUser(name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint) error {
	return AddUserContext(context.Background(), name, mobile, pic, forms, addIP, positionID, deptIDs)
}

func AddUserContext(ctx context.Context, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint) error {
	return AddUserWithAdminAccessContext(ctx, name, mobile, pic, forms, addIP, positionID, deptIDs, AdminAccessInput{})
}

func AddUserWithAdminAccessContext(ctx context.Context, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminAccess AdminAccessInput) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	now := database.Now()
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", name, now)))
	miniOpenID := hex.EncodeToString(hash[:])
	err := db.Transaction(func(tx *gorm.DB) error {
		user := model.User{
			MiniOpenID: miniOpenID,
			Name:       name,
			Mobile:     mobile,
			PositionID: positionID,
			Pic:        pic,
			Forms:      forms,
			Status:     1,
			AddTime:    now,
			AddIP:      addIP,
			EditTime:   now,
			EditIP:     addIP,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := saveUserDeptsTx(tx, user.ID, deptIDs); err != nil {
			return err
		}
		return saveUserAdminAccessTx(tx, user.ID, adminAccess)
	})
	if err == nil {
		invalidateUserTotalCountCache()
	}
	return err
}

func EditUser(id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint) error {
	return EditUserContext(context.Background(), id, name, mobile, pic, forms, addIP, positionID, deptIDs)
}

func EditUserContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint) error {
	return editUserContext(ctx, id, name, mobile, pic, forms, addIP, positionID, deptIDs, AdminAccessInput{}, false)
}

func EditUserWithAdminAccessContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminAccess AdminAccessInput) error {
	return editUserContext(ctx, id, name, mobile, pic, forms, addIP, positionID, deptIDs, adminAccess, true)
}

func EditUserForAdminContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminID uint) error {
	return editUserForAdminContext(ctx, id, name, mobile, pic, forms, addIP, positionID, deptIDs, AdminAccessInput{}, false, adminID)
}

func EditUserWithAdminAccessForAdminContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminAccess AdminAccessInput, adminID uint) error {
	return editUserForAdminContext(ctx, id, name, mobile, pic, forms, addIP, positionID, deptIDs, adminAccess, true, adminID)
}

func editUserContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminAccess AdminAccessInput, saveAdminAccess bool) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"user_name":        name,
		"user_mobile":      mobile,
		"user_position_id": positionID,
		"user_edit_time":   database.Now(),
		"user_edit_ip":     addIP,
	}
	if pic != "" {
		updates["user_pic"] = pic
	}
	if forms != "" {
		updates["user_forms"] = forms
	}
	uid, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := saveUserDeptsTx(tx, uint(uid), deptIDs); err != nil {
			return err
		}
		if saveAdminAccess {
			return saveUserAdminAccessTx(tx, uint(uid), adminAccess)
		}
		return nil
	})
}

func editUserForAdminContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, positionID uint, deptIDs []uint, adminAccess AdminAccessInput, saveAdminAccess bool, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"user_name":        name,
		"user_mobile":      mobile,
		"user_position_id": positionID,
		"user_edit_time":   database.Now(),
		"user_edit_ip":     addIP,
	}
	if pic != "" {
		updates["user_pic"] = pic
	}
	if forms != "" {
		updates["user_forms"] = forms
	}
	uid, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := userVisibleQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Updates(updates)); err != nil {
			return err
		}
		if err := saveUserDeptsTx(tx, uint(uid), deptIDs); err != nil {
			return err
		}
		if saveAdminAccess {
			return saveUserAdminAccessTx(tx, uint(uid), adminAccess)
		}
		return nil
	})
}

func saveUserAdminAccessTx(tx *gorm.DB, userID uint, adminAccess AdminAccessInput) error {
	var current model.User
	if err := tx.Select("id", "user_password").Where("`id` = ?", userID).First(&current).Error; err != nil {
		return err
	}

	roleID := adminAccess.RoleID
	if roleID > 0 && adminAccess.Password == "" && current.Password == "" {
		return fmt.Errorf("请输入登录密码")
	}

	updates := map[string]interface{}{
		"user_role_id":   roleID,
		"user_edit_time": database.Now(),
	}
	if adminAccess.Password != "" {
		hash, err := passwordutil.Hash(adminAccess.Password)
		if err != nil {
			return err
		}
		updates["user_password"] = hash
	}
	if err := tx.Model(&model.User{}).Where("`id` = ?", userID).Updates(updates).Error; err != nil {
		return err
	}
	if adminAccess.PermissionKeysTouched || len(adminAccess.AllowPermissionKeys) > 0 || len(adminAccess.DenyPermissionKeys) > 0 {
		if err := permissionsupport.SetUserApplicationMenuPermissionOverridesTx(tx, userID, adminAccess.AllowPermissionKeys, adminAccess.DenyPermissionKeys); err != nil {
			return err
		}
	}
	return nil
}

func DelUser(id string) error {
	return DelUserContext(context.Background(), id)
}

func DelUserContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err := db.Where("`id` = ?", id).Delete(&model.User{}).Error
	if err == nil {
		invalidateUserTotalCountCache()
	}
	return err
}

func DelUserForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := userVisibleQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	err = access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Delete(&model.User{}))
	if err == nil {
		invalidateUserTotalCountCache()
	}
	return err
}

func DelUsers(ids []string) error {
	return DelUsersContext(context.Background(), ids)
}

func DelUsersContext(ctx context.Context, ids []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err := db.Where("`id` IN ?", ids).Delete(&model.User{}).Error
	if err == nil {
		invalidateUserTotalCountCache()
	}
	return err
}

func DelUsersForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelUserForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func StatusUser(id string, status int, reason string) error {
	return StatusUserContext(context.Background(), id, status, reason)
}

func StatusUserContext(ctx context.Context, id string, status int, reason string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"user_status": status,
	}
	if status == 1 {
		updates["user_check_reason"] = ""
	} else if reason != "" {
		updates["user_check_reason"] = reason
	}
	return db.Model(&model.User{}).Where("`id` = ?", id).Updates(updates).Error
}

func StatusUserForAdminContext(ctx context.Context, id string, status int, reason string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"user_status": status,
	}
	if status == 1 {
		updates["user_check_reason"] = ""
	} else if reason != "" {
		updates["user_check_reason"] = reason
	}
	queryBuilder, err := userVisibleQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Updates(updates))
}

func ResetUserPassword(id string) error {
	return ResetUserPasswordContext(context.Background(), id)
}

func ResetUserPasswordContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	hash, err := passwordutil.Hash("123456")
	if err != nil {
		return err
	}
	return db.Model(&model.User{}).Where("`id` = ?", id).Update("user_password", hash).Error
}

func ResetUserPasswordForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	hash, err := passwordutil.Hash("123456")
	if err != nil {
		return err
	}
	queryBuilder, err := userVisibleQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("user_password", hash))
}

func GetUserFormFields() ([]model.UserFormField, error) {
	return GetUserFormFieldsContext(context.Background())
}

func GetUserFormFieldsContext(ctx context.Context) ([]model.UserFormField, error) {
	setup, err := setupservice.GetSetupContext(ctx, "SETUP_USER_FORM_FIELDS")
	if err != nil {
		return []model.UserFormField{}, nil
	}
	var list []model.UserFormField
	if setup.Value != "" {
		json.Unmarshal([]byte(setup.Value), &list)
	}
	return list, nil
}

func SaveUserFormFields(fields []model.UserFormField) error {
	return SaveUserFormFieldsContext(context.Background(), fields)
}

func SaveUserFormFieldsContext(ctx context.Context, fields []model.UserFormField) error {
	jsonData, _ := json.Marshal(fields)
	return setupservice.SetContentSetupContext(ctx, "SETUP_USER_FORM_FIELDS", string(jsonData), "")
}
