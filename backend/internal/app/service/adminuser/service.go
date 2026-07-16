package adminuser

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	setupservice "wecheckin-backend/backend/internal/app/service/setup"
	"wecheckin-backend/backend/internal/app/support/access"
	deptsupport "wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/media"
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

type UserDetail struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Mobile       string   `json:"mobile"`
	Avatar       string   `json:"avatar"`
	Pic          string   `json:"pic"`
	Status       int      `json:"status"`
	Forms        string   `json:"forms"`
	LoginCnt     int      `json:"loginCnt"`
	AddTime      int64    `json:"addTime"`
	LoginTime    int64    `json:"loginTime"`
	DeptIDs      []uint   `json:"deptIds"`
	DeptNames    []string `json:"deptNames"`
	TopDeptNames []string `json:"topDeptNames"`
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
	avatar := media.FullURLWithStaticDomain(user.Pic)
	return UserDetail{
		ID:           user.ID,
		Name:         user.Name,
		Mobile:       user.Mobile,
		Avatar:       avatar,
		Pic:          avatar,
		Status:       user.Status,
		Forms:        user.Forms,
		LoginCnt:     user.LoginCnt,
		AddTime:      user.AddTime,
		LoginTime:    user.LoginTime,
		DeptIDs:      deptIDs,
		DeptNames:    deptNames,
		TopDeptNames: topDeptNames,
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
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Mobile    string `json:"mobile"`
	Avatar    string `json:"avatar"`
	Pic       string `json:"pic"`
	Status    int    `json:"status"`
	LoginCnt  int    `json:"loginCnt"`
	AddTime   int64  `json:"addTime"`
	LoginTime int64  `json:"loginTime"`
	DeptIDs   []uint `json:"deptIds"`
}

func GetUserList(keyword, sortStr string, page, pageSize int, adminID uint) ([]UserListItem, int64, error) {
	return GetUserListContext(context.Background(), keyword, sortStr, page, pageSize, adminID)
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
		queryBuilder = queryBuilder.Where("`user_name` LIKE ? OR `user_mobile` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	deptIDs := access.VisibleDeptIDsContext(ctx, &admin)
	if deptIDs != nil {
		queryBuilder = queryBuilder.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
	}
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderClause := query.ParseSort(sortStr, map[string]string{
		"name":     "user_name",
		"mobile":   "user_mobile",
		"status":   "user_status",
		"loginCnt": "user_login_cnt",
		"addTime":  "user_add_time",
	})
	if orderClause == "" {
		orderClause = "`user_add_time` DESC"
	}
	err := queryBuilder.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]UserListItem, len(list))
	for i, u := range list {
		avatar := media.FullURLWithStaticDomain(u.Pic)
		result[i] = UserListItem{
			ID:        u.ID,
			Name:      u.Name,
			Mobile:    u.Mobile,
			Avatar:    avatar,
			Pic:       avatar,
			Status:    u.Status,
			LoginCnt:  u.LoginCnt,
			AddTime:   u.AddTime,
			LoginTime: u.LoginTime,
			DeptIDs:   deptsupport.UserDeptIDsContext(ctx, u.ID),
		}
	}
	return result, total, nil
}

func AddUser(name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	return AddUserContext(context.Background(), name, mobile, pic, forms, addIP, deptIDs)
}

func AddUserContext(ctx context.Context, name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	now := database.Now()
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", name, now)))
	miniOpenID := hex.EncodeToString(hash[:])
	return db.Transaction(func(tx *gorm.DB) error {
		user := model.User{
			MiniOpenID: miniOpenID,
			Name:       name,
			Mobile:     mobile,
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
		return saveUserDeptsTx(tx, user.ID, deptIDs)
	})
}

func EditUser(id, name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	return EditUserContext(context.Background(), id, name, mobile, pic, forms, addIP, deptIDs)
}

func EditUserContext(ctx context.Context, id, name, mobile, pic, forms, addIP string, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"user_name":      name,
		"user_mobile":    mobile,
		"user_edit_time": database.Now(),
		"user_edit_ip":   addIP,
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
		return saveUserDeptsTx(tx, uint(uid), deptIDs)
	})
}

func DelUser(id string) error {
	return DelUserContext(context.Background(), id)
}

func DelUserContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` = ?", id).Delete(&model.User{}).Error
}

func DelUsers(ids []string) error {
	return DelUsersContext(context.Background(), ids)
}

func DelUsersContext(ctx context.Context, ids []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` IN ?", ids).Delete(&model.User{}).Error
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

func GetUserFormFields() ([]model.UserFormField, error) {
	return GetUserFormFieldsContext(context.Background())
}

func GetUserFormFieldsContext(ctx context.Context) ([]model.UserFormField, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var setup model.Setup
	err := db.Where("`setup_key` = ?", "SETUP_USER_FORM_FIELDS").First(&setup).Error
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
