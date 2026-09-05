package admincontent

import (
	"context"
	"encoding/json"
	"strconv"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/query"
	"wecheckin/backend/pkg/database"
)

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

var adminEnrollListColumns = []string{
	"id",
	"enroll_title",
	"enroll_status",
	"create_dept_id",
	"enroll_publish_dept_ids",
	"create_by",
	"update_by",
	"update_dept_id",
	"enroll_cate_id",
	"enroll_cate_name",
	"enroll_start",
	"enroll_end",
	"enroll_day_cnt",
	"enroll_order",
	"enroll_vouch",
	"enroll_repeat",
	"enroll_limit",
	"enroll_obj",
	"enroll_qr",
	"enroll_view_cnt",
	"enroll_join_cnt",
	"enroll_user_cnt",
	"add_time",
	"edit_time",
}

func decodeEnrollObj(raw string) enrollObj {
	var obj enrollObj
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &obj)
	}
	return obj
}

func populateEnrollFields(list []model.Enroll) []model.Enroll {
	for i := range list {
		obj := decodeEnrollObj(list[i].Obj)
		if len(obj.Cover) > 0 {
			list[i].Img = media.FullURLWithStaticDomain(obj.Cover[0])
		}
		list[i].Desc = obj.Desc
	}
	return list
}

func GetAdminEnrollList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.Enroll, int64, error) {
	return GetAdminEnrollListContext(context.Background(), keyword, sortStr, page, pageSize, adminID)
}

func GetAdminEnrollListContext(ctx context.Context, keyword, sortStr string, page, pageSize int, adminID uint) ([]model.Enroll, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	db.First(&admin, adminID)
	var list []model.Enroll
	var total int64
	queryBuilder := db.Model(&model.Enroll{})
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_title` LIKE ?", "%"+keyword+"%")
	}
	where, args := access.DataScopeFilterForResourceWithDBContext(ctx, db, &admin, access.EnrollAuditFields)
	if where != "" {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	queryBuilder.Count(&total)
	orderClause := query.ParseSort(sortStr, map[string]string{
		"title":   "enroll_title",
		"sort":    "enroll_order",
		"status":  "enroll_status",
		"isVouch": "enroll_vouch",
		"userCnt": "enroll_user_cnt",
		"joinCnt": "enroll_join_cnt",
		"addTime": "add_time",
	})
	if orderClause == "" {
		orderClause = "`add_time` DESC"
	}
	err := queryBuilder.Select(adminEnrollListColumns).Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEnrollFields(list)
	userCounts, err := loadEnrollUserCountMapContext(ctx, db, list)
	if err != nil {
		return nil, 0, err
	}
	joinCounts, err := loadEnrollJoinCountMapContext(ctx, db, list)
	if err != nil {
		return nil, 0, err
	}
	for i := range list {
		list[i].UserCnt = userCounts[list[i].ID]
		list[i].JoinCnt = joinCounts[list[i].ID]
	}
	return list, total, nil
}

func scopedEnrollQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Enroll{}, access.EnrollAuditFields)
}

func loadEnrollUserCountMapContext(ctx context.Context, db *gorm.DB, list []model.Enroll) (map[uint]int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	ids := enrollIDStrings(list)
	counts := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	type countRow struct {
		EnrollID string `gorm:"column:enroll_id"`
		Count    int    `gorm:"column:cnt"`
	}
	var rows []countRow
	if err := db.Raw(
		"SELECT enroll_id, COUNT(DISTINCT uid) AS cnt FROM (SELECT `enroll_join_enroll_id` AS enroll_id, `enroll_join_user_id` AS uid FROM `enroll_joins` WHERE `enroll_join_enroll_id` IN ? UNION SELECT `enroll_user_enroll_id` AS enroll_id, `enroll_user_mini_openid` AS uid FROM `enroll_users` WHERE `enroll_user_enroll_id` IN ?) AS u GROUP BY enroll_id",
		ids, ids,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		id, err := strconv.Atoi(row.EnrollID)
		if err == nil && id > 0 {
			counts[uint(id)] = row.Count
		}
	}
	return counts, nil
}

func loadEnrollJoinCountMapContext(ctx context.Context, db *gorm.DB, list []model.Enroll) (map[uint]int, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	ids := enrollIDStrings(list)
	counts := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	type countRow struct {
		EnrollID string `gorm:"column:enroll_join_enroll_id"`
		Count    int    `gorm:"column:cnt"`
	}
	var rows []countRow
	if err := db.Model(&model.EnrollJoin{}).
		Select("`enroll_join_enroll_id`, COUNT(*) AS cnt").
		Where("`enroll_join_enroll_id` IN ?", ids).
		Group("`enroll_join_enroll_id`").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		id, err := strconv.Atoi(row.EnrollID)
		if err == nil && id > 0 {
			counts[uint(id)] = row.Count
		}
	}
	return counts, nil
}

func enrollIDStrings(list []model.Enroll) []string {
	ids := make([]string, 0, len(list))
	for _, item := range list {
		ids = append(ids, strconv.Itoa(int(item.ID)))
	}
	return ids
}

func GetEnrollDetail(id string) (*model.Enroll, error) {
	return GetEnrollDetailContext(context.Background(), id)
}

func GetEnrollDetailContext(ctx context.Context, id string) (*model.Enroll, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var enroll model.Enroll
	err := db.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	obj := decodeEnrollObj(enroll.Obj)
	if len(obj.Cover) > 0 {
		enroll.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	enroll.Desc = obj.Desc
	return &enroll, nil
}

func GetEnrollDetailForAdminContext(ctx context.Context, id string, adminID uint) (*model.Enroll, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var enroll model.Enroll
	if err := queryBuilder.Where("`id` = ?", id).First(&enroll).Error; err != nil {
		return nil, err
	}
	obj := decodeEnrollObj(enroll.Obj)
	if len(obj.Cover) > 0 {
		enroll.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	enroll.Desc = obj.Desc
	return &enroll, nil
}

func UpdateEnrollForms(id, forms string) error {
	return UpdateEnrollFormsContext(context.Background(), id, forms)
}

func UpdateEnrollFormsContext(ctx context.Context, id, forms string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_forms", forms).Error
}

func UpdateEnrollFormsForAdminContext(ctx context.Context, id, forms string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("enroll_forms", forms))
}

func SortEnroll(id, sortStr string) error {
	return SortEnrollContext(context.Background(), id, sortStr)
}

func SortEnrollContext(ctx context.Context, id, sortStr string) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_order", sortValue).Error
}

func SortEnrollForAdminContext(ctx context.Context, id, sortStr string, adminID uint) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("enroll_order", sortValue))
}

func VouchEnroll(id string, vouch int) error {
	return VouchEnrollContext(context.Background(), id, vouch)
}

func VouchEnrollContext(ctx context.Context, id string, vouch int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_vouch", vouch).Error
}

func VouchEnrollForAdminContext(ctx context.Context, id string, vouch int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("enroll_vouch", vouch))
}

func StatusEnroll(id string, status int) error {
	return StatusEnrollContext(context.Background(), id, status)
}

func StatusEnrollContext(ctx context.Context, id string, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("`id` = ?", id).Update("enroll_status", status).Error
}

func StatusEnrollForAdminContext(ctx context.Context, id string, status int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("enroll_status", status))
}

func ClearEnrollAll(id string) error {
	return ClearEnrollAllContext(context.Background(), id)
}

func ClearEnrollAllContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
			"enroll_join_cnt": 0,
			"enroll_user_cnt": 0,
		}).Error
	})
}

func ClearEnrollAllForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedEnrollQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`id` = ?", id).First(&model.Enroll{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(map[string]interface{}{
			"enroll_join_cnt": 0,
			"enroll_user_cnt": 0,
		}).Error
	})
}

func DelEnroll(id string) error {
	return DelEnrollContext(context.Background(), id)
}

func DelEnrollContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Enroll{}).Error
	})
}

func DelEnrollForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedEnrollQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`id` = ?", id).First(&model.Enroll{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_join_enroll_id` = ?", id).Delete(&model.EnrollJoin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`enroll_user_enroll_id` = ?", id).Delete(&model.EnrollUser{}).Error; err != nil {
			return err
		}
		return access.RequireRowsAffected(tx.Where("`id` = ?", id).Delete(&model.Enroll{}))
	})
}

func DelEnrolls(ids []string) error {
	return DelEnrollsContext(context.Background(), ids)
}

func DelEnrollsContext(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := DelEnrollContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func DelEnrollsForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelEnrollForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func InsertEnroll(title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID, createBy uint) error {
	return InsertEnrollContext(context.Background(), title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds, status, order, dayCnt, start, end, obj, allowRepeat, dailyLimit, deptID, createBy)
}

func InsertEnrollContext(ctx context.Context, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID, createBy uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	enroll := model.Enroll{
		Title:          title,
		Status:         status,
		CateID:         cateID,
		CateName:       cateName,
		Start:          start,
		End:            end,
		DayCnt:         dayCnt,
		Order:          order,
		Forms:          forms,
		JoinForms:      joinForms,
		QR:             qr,
		Obj:            obj,
		AllowRepeat:    allowRepeat,
		DailyLimit:     dailyLimit,
		DeptID:         deptID,
		PublishDeptIds: publishDeptIds,
		CreateBy:       createBy,
		AddTime:        database.Now(),
		EditTime:       database.Now(),
		UpdateBy:       createBy,
		UpdateDeptID:   deptID,
		AddIP:          addIP,
	}
	return db.Create(&enroll).Error
}

func EditEnroll(id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID uint) error {
	return EditEnrollContext(context.Background(), id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds, status, order, dayCnt, start, end, obj, allowRepeat, dailyLimit, deptID)
}

func EditEnrollContext(ctx context.Context, id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"enroll_title":            title,
		"enroll_status":           status,
		"enroll_cate_id":          cateID,
		"enroll_cate_name":        cateName,
		"enroll_start":            start,
		"enroll_end":              end,
		"enroll_day_cnt":          dayCnt,
		"enroll_order":            order,
		"enroll_forms":            forms,
		"enroll_join_forms":       joinForms,
		"enroll_repeat":           allowRepeat,
		"enroll_limit":            dailyLimit,
		"create_dept_id":          deptID,
		"update_dept_id":          deptID,
		"enroll_publish_dept_ids": publishDeptIds,
		"enroll_qr":               qr,
		"edit_time":               database.Now(),
		"enroll_edit_ip":          addIP,
	}
	if obj != "" {
		updates["enroll_obj"] = obj
	}
	return db.Model(&model.Enroll{}).Where("`id` = ?", id).Updates(updates).Error
}

func EditEnrollForAdminContext(ctx context.Context, id, title, cateID, cateName, forms, joinForms, qr, addIP, publishDeptIds string, status, order, dayCnt int, start, end int64, obj string, allowRepeat bool, dailyLimit int, deptID, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"enroll_title":            title,
		"enroll_status":           status,
		"enroll_cate_id":          cateID,
		"enroll_cate_name":        cateName,
		"enroll_start":            start,
		"enroll_end":              end,
		"enroll_day_cnt":          dayCnt,
		"enroll_order":            order,
		"enroll_forms":            forms,
		"enroll_join_forms":       joinForms,
		"enroll_repeat":           allowRepeat,
		"enroll_limit":            dailyLimit,
		"create_dept_id":          deptID,
		"update_by":               adminID,
		"update_dept_id":          deptID,
		"enroll_publish_dept_ids": publishDeptIds,
		"enroll_qr":               qr,
		"edit_time":               database.Now(),
		"enroll_edit_ip":          addIP,
	}
	if obj != "" {
		updates["enroll_obj"] = obj
	}
	queryBuilder, err := scopedEnrollQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Updates(updates))
}
