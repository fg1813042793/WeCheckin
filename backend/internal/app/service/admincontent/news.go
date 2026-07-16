package admincontent

import (
	"strconv"

	newsservice "wecheckin-backend/backend/internal/app/service/news"
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/query"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetAdminNewsList(keyword, sortStr string, page, pageSize int, adminID uint) ([]model.News, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.News
	var total int64
	queryBuilder := database.DB.Model(&model.News{})
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	where, args := access.DataScopeFilter(&admin, "`news_dept_id`", "`news_create_by`")
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
		"addTime": "news_add_time",
	})
	if orderClause == "" {
		orderClause = "`news_add_time` DESC"
	}
	err := queryBuilder.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetNewsDetail(id string) (*model.News, error) {
	var news model.News
	err := database.DB.Where("`id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	news = newsservice.PopulateFields([]model.News{news})[0]
	return &news, nil
}

func DelNews(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.News{}).Error
}

func DelNewsList(ids []string) error {
	return database.DB.Where("`id` IN ?", ids).Delete(&model.News{}).Error
}

func SortNews(id, sortStr string) error {
	sortValue, err := strconv.Atoi(sortStr)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_order", sortValue).Error
}

func StatusNews(id string, status int) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_status", status).Error
}

func InsertNews(title, desc, cateID, cateName, content, qr, pic, forms, addIP, publishDeptIds string, status, order int, deptID, createBy uint) error {
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
		AddIP:          addIP,
	}
	return database.DB.Create(&news).Error
}

func EditNews(id, title, desc, cateID, cateName, content, qr, addIP, publishDeptIds string, status, order int, deptID uint) error {
	updates := map[string]interface{}{
		"news_title":            title,
		"news_desc":             desc,
		"news_status":           status,
		"news_cate_id":          cateID,
		"news_cate_name":        cateName,
		"news_order":            order,
		"news_content":          content,
		"news_qr":               qr,
		"news_dept_id":          deptID,
		"news_publish_dept_ids": publishDeptIds,
		"news_edit_time":        database.Now(),
		"news_edit_ip":          addIP,
	}
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Updates(updates).Error
}

func UpdateNewsForms(id, forms string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_forms", forms).Error
}

func UpdateNewsPic(id, pic string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_pic", pic).Error
}

func UpdateNewsContent(id, content string) error {
	return database.DB.Model(&model.News{}).Where("`id` = ?", id).Update("news_content", content).Error
}
