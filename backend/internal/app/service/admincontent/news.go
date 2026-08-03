package admincontent

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	newsservice "wecheckin/backend/internal/app/service/news"
	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/app/support/query"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func GetAdminNewsList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.News, int64, error) {
	return GetAdminNewsListContext(context.Background(), keyword, sortStr, page, pageSize, adminID)
}

func GetAdminNewsListContext(ctx context.Context, keyword, sortStr string, page, pageSize int, adminID uint) ([]model.News, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	db.First(&admin, adminID)
	var list []model.News
	var total int64
	queryBuilder := db.Model(&model.News{})
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	where, args := access.DataScopeFilterForResourceWithDBContext(ctx, db, &admin, access.NewsAuditFields)
	if where != "" {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	queryBuilder.Count(&total)
	orderClause := query.ParseSort(sortStr, map[string]string{
		"title":   "news_title",
		"type":    "news_cate_id",
		"order":   "news_order",
		"status":  "news_status",
		"vouch":   "news_vouch",
		"addTime": "add_time",
	})
	if orderClause == "" {
		orderClause = "`add_time` DESC"
	}
	err := queryBuilder.Select(newsservice.ListColumns).Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func scopedNewsQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.News{}, access.NewsAuditFields)
}

func GetNewsDetail(id string) (*model.News, error) {
	return GetNewsDetailContext(context.Background(), id)
}

func GetNewsDetailContext(ctx context.Context, id string) (*model.News, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var news model.News
	err := db.Where("`id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	news = newsservice.PopulateFields([]model.News{news})[0]
	return &news, nil
}

func GetNewsDetailForAdminContext(ctx context.Context, id string, adminID uint) (*model.News, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var news model.News
	if err := queryBuilder.Where("`id` = ?", id).First(&news).Error; err != nil {
		return nil, err
	}
	news = newsservice.PopulateFields([]model.News{news})[0]
	return &news, nil
}

func DelNews(id string) error {
	return DelNewsContext(context.Background(), id)
}

func DelNewsContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` = ?", id).Delete(&model.News{}).Error
}

func DelNewsForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Delete(&model.News{}))
}

func DelNewsList(ids []string) error {
	return DelNewsListContext(context.Background(), ids)
}

func DelNewsListContext(ctx context.Context, ids []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` IN ?", ids).Delete(&model.News{}).Error
}

func DelNewsListForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelNewsForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func SortNews(id, sortStr string) error {
	return SortNewsContext(context.Background(), id, sortStr)
}

func SortNewsContext(ctx context.Context, id, sortStr string) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_order", sortValue).Error
}

func SortNewsForAdminContext(ctx context.Context, id, sortStr string, adminID uint) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_order", sortValue))
}

func StatusNews(id string, status int) error {
	return StatusNewsContext(context.Background(), id, status)
}

func StatusNewsContext(ctx context.Context, id string, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_status", status).Error
}

func StatusNewsForAdminContext(ctx context.Context, id string, status int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_status", status))
}

func VouchNews(id string, vouch int) error {
	return VouchNewsContext(context.Background(), id, vouch)
}

func VouchNewsContext(ctx context.Context, id string, vouch int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_vouch", vouch).Error
}

func VouchNewsForAdminContext(ctx context.Context, id string, vouch int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_vouch", vouch))
}

func InsertNews(title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds string, status, order int, deptID, createBy uint) error {
	return InsertNewsContext(context.Background(), title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds, status, order, deptID, createBy)
}

func InsertNewsContext(ctx context.Context, title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds string, status, order int, deptID, createBy uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	news := model.News{
		Title:          title,
		Desc:           desc,
		Status:         status,
		CateID:         cateID,
		CateName:       cateName,
		Order:          order,
		Content:        content,
		QR:             qr,
		Pic:            pic,
		Forms:          forms,
		DeptID:         deptID,
		PublishDeptIds: publishDeptIds,
		CreateBy:       createBy,
		AddTime:        database.Now(),
		EditTime:       database.Now(),
		UpdateBy:       createBy,
		UpdateDeptID:   deptID,
		AddIP:          addIP,
	}
	return db.Create(&news).Error
}

func EditNews(id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID uint) error {
	return EditNewsContext(context.Background(), id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds, status, order, deptID)
}

func EditNewsContext(ctx context.Context, id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"news_title":            title,
		"news_desc":             desc,
		"news_status":           status,
		"news_cate_id":          cateID,
		"news_cate_name":        cateName,
		"news_order":            order,
		"news_content":          content,
		"news_qr":               qr,
		"create_dept_id":        deptID,
		"update_dept_id":        deptID,
		"news_publish_dept_ids": publishDeptIds,
		"edit_time":             database.Now(),
		"news_edit_ip":          addIP,
	}
	return db.Model(&model.News{}).Where("`id` = ?", id).Updates(updates).Error
}

func EditNewsForAdminContext(ctx context.Context, id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"news_title":            title,
		"news_desc":             desc,
		"news_status":           status,
		"news_cate_id":          cateID,
		"news_cate_name":        cateName,
		"news_order":            order,
		"news_content":          content,
		"news_qr":               qr,
		"create_dept_id":        deptID,
		"update_by":             adminID,
		"update_dept_id":        deptID,
		"news_publish_dept_ids": publishDeptIds,
		"edit_time":             database.Now(),
		"news_edit_ip":          addIP,
	}
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Updates(updates))
}

func UpdateNewsForms(id, forms string) error {
	return UpdateNewsFormsContext(context.Background(), id, forms)
}

func UpdateNewsFormsContext(ctx context.Context, id, forms string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_forms", forms).Error
}

func UpdateNewsFormsForAdminContext(ctx context.Context, id, forms string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_forms", forms))
}

func UpdateNewsPic(id, pic string) error {
	return UpdateNewsPicContext(context.Background(), id, pic)
}

func UpdateNewsPicContext(ctx context.Context, id, pic string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_pic", pic).Error
}

func UpdateNewsPicForAdminContext(ctx context.Context, id, pic string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_pic", pic))
}

func UpdateNewsContent(id, content string) error {
	return UpdateNewsContentContext(context.Background(), id, content)
}

func UpdateNewsContentContext(ctx context.Context, id, content string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.News{}).Where("`id` = ?", id).Update("news_content", content).Error
}

func UpdateNewsContentForAdminContext(ctx context.Context, id, content string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedNewsQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("news_content", content))
}
