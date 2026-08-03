package admincontent

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func ensureEnrollVisibleContext(ctx context.Context, db *gorm.DB, enrollID string, adminID uint) error {
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return queryBuilder.Where("`id` = ?", enrollID).First(&model.Enroll{}).Error
}

func EnsureEnrollVisibleForAdminContext(ctx context.Context, enrollID string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return ensureEnrollVisibleContext(ctx, db, enrollID, adminID)
}

func GetEnrollUserList(enrollID, keyword string) ([]model.EnrollUser, error) {
	return GetEnrollUserListContext(context.Background(), enrollID, keyword)
}

func GetEnrollUserListContext(ctx context.Context, enrollID, keyword string) ([]model.EnrollUser, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EnrollUser
	queryBuilder := db.Where("`enroll_user_enroll_id` = ?", enrollID)
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_user_mini_openid` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)", "%"+keyword+"%")
	}
	err := queryBuilder.Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return enrichEnrollUsersWithUserInfoContext(ctx, db, list), nil
}

func GetEnrollUserListForAdminContext(ctx context.Context, enrollID, keyword string, adminID uint) ([]model.EnrollUser, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEnrollVisibleContext(ctx, db, enrollID, adminID); err != nil {
		return nil, err
	}
	return GetEnrollUserListContext(ctx, enrollID, keyword)
}

func GetEnrollJoinList(enrollID, keyword string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	return GetEnrollJoinListContext(context.Background(), enrollID, keyword, page, pageSize)
}

func GetEnrollJoinListContext(ctx context.Context, enrollID, keyword string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EnrollJoin
	var total int64
	queryBuilder := db.Model(&model.EnrollJoin{})
	if enrollID != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_enroll_id` = ?", enrollID)
	}
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_user_id` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?) OR `enroll_join_user_id` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	queryBuilder.Count(&total)
	err := queryBuilder.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = enrichEnrollJoinsWithUserInfoContext(ctx, db, list)
	return list, total, nil
}

func GetEnrollJoinListForAdminContext(ctx context.Context, enrollID, keyword string, page, pageSize int, adminID uint) ([]model.EnrollJoin, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEnrollVisibleContext(ctx, db, enrollID, adminID); err != nil {
		return nil, 0, err
	}
	return GetEnrollJoinListContext(ctx, enrollID, keyword, page, pageSize)
}

type EnrollStatItem struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	DeptName    string `json:"deptName"`
	TopDeptName string `json:"topDeptName"`
	JoinCnt     int    `json:"joinCnt"`
	DayCnt      int    `json:"dayCnt"`
}

func GetEnrollStats(enrollID, startDay, endDay string) ([]EnrollStatItem, error) {
	return GetEnrollStatsContext(context.Background(), enrollID, startDay, endDay)
}

func GetEnrollStatsContext(ctx context.Context, enrollID, startDay, endDay string) ([]EnrollStatItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder := db.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` <= ?", endDay)
	}

	type statRow struct {
		UserID  string `gorm:"column:user_id"`
		JoinCnt int    `gorm:"column:join_cnt"`
		DayCnt  int    `gorm:"column:day_cnt"`
	}
	var rows []statRow
	if err := queryBuilder.Model(&model.EnrollJoin{}).
		Select("`enroll_join_user_id` AS user_id, COUNT(*) AS join_cnt, COUNT(DISTINCT `enroll_join_day`) AS day_cnt").
		Group("`enroll_join_user_id`").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []EnrollStatItem{}, nil
	}

	openIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		openIDs = append(openIDs, row.UserID)
	}
	infoByOpenID, _ := loadUserDeptInfoByOpenIDContext(ctx, db, openIDs)

	result := make([]EnrollStatItem, 0, len(rows))
	for _, row := range rows {
		info := infoByOpenID[row.UserID]
		item := EnrollStatItem{
			UserID:      row.UserID,
			UserName:    info.User.Name,
			DeptName:    info.DeptName,
			TopDeptName: info.TopDeptName,
			JoinCnt:     row.JoinCnt,
			DayCnt:      row.DayCnt,
		}
		result = append(result, item)
	}
	return result, nil
}

func GetEnrollStatsForAdminContext(ctx context.Context, enrollID, startDay, endDay string, adminID uint) ([]EnrollStatItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEnrollVisibleContext(ctx, db, enrollID, adminID); err != nil {
		return nil, err
	}
	return GetEnrollStatsContext(ctx, enrollID, startDay, endDay)
}

func DelEnrollJoin(id string) error {
	return DelEnrollJoinContext(context.Background(), id)
}

func DelEnrollJoinContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var join model.EnrollJoin
		if err := tx.Where("`id` = ?", id).First(&join).Error; err != nil {
			return err
		}
		if err := tx.Delete(&join).Error; err != nil {
			return err
		}
		var joinCnt int64
		if err := tx.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", join.EnrollID).Count(&joinCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", join.EnrollID).Update("enroll_join_cnt", joinCnt).Error
	})
}

func DelEnrollJoinForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var join model.EnrollJoin
		if err := tx.Where("`id` = ?", id).First(&join).Error; err != nil {
			return err
		}
		if err := ensureEnrollVisibleContext(ctx, tx, join.EnrollID, adminID); err != nil {
			return err
		}
		if err := tx.Delete(&join).Error; err != nil {
			return err
		}
		var joinCnt int64
		if err := tx.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", join.EnrollID).Count(&joinCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", join.EnrollID).Update("enroll_join_cnt", joinCnt).Error
	})
}

func DelEnrollJoins(ids []string) error {
	return DelEnrollJoinsContext(context.Background(), ids)
}

func DelEnrollJoinsContext(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := DelEnrollJoinContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func DelEnrollJoinsForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelEnrollJoinForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEnrollUser(enrollID, userID string) error {
	return RemoveEnrollUserContext(context.Background(), enrollID, userID)
}

func RemoveEnrollUserContext(ctx context.Context, enrollID, userID string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", enrollID, userID).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		var userCnt int64
		if err := tx.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ?", enrollID).Count(&userCnt).Error; err != nil {
			return err
		}
		var joinCnt int64
		if err := tx.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", enrollID).Count(&joinCnt).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_user_cnt", userCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_join_cnt", joinCnt).Error
	})
}

func RemoveEnrollUserForAdminContext(ctx context.Context, enrollID, userID string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ensureEnrollVisibleContext(ctx, tx, enrollID, adminID); err != nil {
			return err
		}
		if err := tx.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", enrollID, userID).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		var userCnt int64
		if err := tx.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ?", enrollID).Count(&userCnt).Error; err != nil {
			return err
		}
		var joinCnt int64
		if err := tx.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ?", enrollID).Count(&joinCnt).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_user_cnt", userCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_join_cnt", joinCnt).Error
	})
}

func RemoveEnrollUsers(enrollID string, userIDs []string) error {
	return RemoveEnrollUsersContext(context.Background(), enrollID, userIDs)
}

func RemoveEnrollUsersContext(ctx context.Context, enrollID string, userIDs []string) error {
	for _, uid := range userIDs {
		if err := RemoveEnrollUserContext(ctx, enrollID, uid); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEnrollUsersForAdminContext(ctx context.Context, enrollID string, userIDs []string, adminID uint) error {
	for _, uid := range userIDs {
		if err := RemoveEnrollUserForAdminContext(ctx, enrollID, uid, adminID); err != nil {
			return err
		}
	}
	return nil
}

func EditEnrollUserForms(enrollID, userID, forms string) error {
	return EditEnrollUserFormsContext(context.Background(), enrollID, userID, forms)
}

func EditEnrollUserFormsContext(ctx context.Context, enrollID, userID, forms string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.EnrollUser{}).
		Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).
		Update("enroll_user_forms", forms).Error
}

func EditEnrollUserFormsForAdminContext(ctx context.Context, enrollID, userID, forms string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEnrollVisibleContext(ctx, db, enrollID, adminID); err != nil {
		return err
	}
	return access.RequireRowsAffected(db.Model(&model.EnrollUser{}).
		Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).
		Update("enroll_user_forms", forms))
}

type enrollUserDeptInfo struct {
	User        model.User
	DeptName    string
	TopDeptName string
}

func enrichEnrollUsersWithUserInfoContext(ctx context.Context, db *gorm.DB, list []model.EnrollUser) []model.EnrollUser {
	openIDs := make([]string, 0, len(list))
	for _, item := range list {
		openIDs = append(openIDs, item.MiniOpenID)
	}
	infoByOpenID, err := loadUserDeptInfoByOpenIDContext(ctx, db, openIDs)
	if err != nil {
		return list
	}
	for i := range list {
		info := infoByOpenID[list[i].MiniOpenID]
		list[i].EnrollTitle = info.User.Name
		list[i].UserName = info.User.Name
		list[i].DeptName = info.DeptName
		list[i].TopDeptName = info.TopDeptName
	}
	return list
}

func enrichEnrollJoinsWithUserInfoContext(ctx context.Context, db *gorm.DB, list []model.EnrollJoin) []model.EnrollJoin {
	openIDs := make([]string, 0, len(list))
	for _, item := range list {
		openIDs = append(openIDs, item.UserID)
	}
	infoByOpenID, err := loadUserDeptInfoByOpenIDContext(ctx, db, openIDs)
	if err != nil {
		return list
	}
	for i := range list {
		info, ok := infoByOpenID[list[i].UserID]
		if !ok {
			continue
		}
		list[i].EnrollTitle = info.User.Name
		list[i].UserName = info.User.Name
		list[i].DeptName = info.DeptName
		list[i].TopDeptName = info.TopDeptName
	}
	return list
}

func loadUserDeptInfoByOpenIDContext(ctx context.Context, db *gorm.DB, openIDs []string) (map[string]enrollUserDeptInfo, error) {
	result := make(map[string]enrollUserDeptInfo)
	uniqueOpenIDs := uniqueNonEmptyStrings(openIDs)
	if len(uniqueOpenIDs) == 0 {
		return result, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result, err
		}
	}

	var users []model.User
	if err := db.Select("id", "user_mini_openid", "user_name").
		Where("`user_mini_openid` IN ?", uniqueOpenIDs).
		Find(&users).Error; err != nil {
		return result, err
	}
	if len(users) == 0 {
		return result, nil
	}

	userIDToOpenID := make(map[uint]string, len(users))
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		result[user.MiniOpenID] = enrollUserDeptInfo{User: user}
		userIDToOpenID[user.ID] = user.MiniOpenID
		userIDs = append(userIDs, user.ID)
	}

	var userDepts []model.UserDept
	if err := db.Select("id", "user_dept_user_id", "user_dept_dept_id").
		Where("`user_dept_user_id` IN ?", userIDs).
		Order("`id` ASC").
		Find(&userDepts).Error; err != nil {
		return result, err
	}
	if len(userDepts) == 0 {
		return result, nil
	}

	deptIDByUserID := make(map[uint]uint, len(userDepts))
	for _, userDept := range userDepts {
		if deptIDByUserID[userDept.UserID] == 0 {
			deptIDByUserID[userDept.UserID] = userDept.DeptID
		}
	}

	var departments []model.Department
	if err := db.Select("id", "dept_name", "dept_parent_id").Find(&departments).Error; err != nil {
		return result, err
	}
	deptByID := make(map[uint]model.Department, len(departments))
	for _, department := range departments {
		deptByID[department.ID] = department
	}

	for userID, deptID := range deptIDByUserID {
		openID := userIDToOpenID[userID]
		if openID == "" {
			continue
		}
		info := result[openID]
		if department, ok := deptByID[deptID]; ok {
			info.DeptName = department.Name
		}
		info.TopDeptName = topDeptNameFromDepartmentMap(deptID, deptByID)
		result[openID] = info
	}
	return result, nil
}

func topDeptNameFromDepartmentMap(deptID uint, deptByID map[uint]model.Department) string {
	visited := make(map[uint]struct{})
	for deptID > 0 {
		if _, ok := visited[deptID]; ok {
			return ""
		}
		visited[deptID] = struct{}{}

		department, ok := deptByID[deptID]
		if !ok {
			return ""
		}
		if department.ParentID == 0 {
			return department.Name
		}
		deptID = department.ParentID
	}
	return ""
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
