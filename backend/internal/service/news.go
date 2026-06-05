package service

import (
	"encoding/json"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func populateNewsFields(list []model.News) []model.News {
	for i := range list {
		if list[i].Pic != "" {
			var urls []string
			if err := json.Unmarshal([]byte(list[i].Pic), &urls); err == nil && len(urls) > 0 {
				list[i].Img = GetFullURL(urls[0])
			} else {
				list[i].Img = GetFullURL(list[i].Pic)
			}
		}
	}
	return list
}

func GetNewsList(page, pageSize int, keyword, userID string) (map[string]interface{}, error) {
	var list []model.News
	var total int64
	query := database.DB.Model(&model.News{}).Where("`news_status` = 1")
	if keyword != "" {
		query = query.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	if userID != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`news_publish_dept_ids` = '' OR `news_publish_dept_ids` IS NULL OR "+
				buildDeptOverlap("news_publish_dept_ids", deptIDs)+")")
		} else {
			query = query.Where("(`news_publish_dept_ids` = '' OR `news_publish_dept_ids` IS NULL)")
		}
	} else {
		query = query.Where("(`news_publish_dept_ids` = '' OR `news_publish_dept_ids` IS NULL)")
	}
	query.Count(&total)
	err := query.Order("`news_order` ASC, `news_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateNewsFields(list)
	return map[string]interface{}{
		"list":  list,
		"total": total,
	}, nil
}

func ViewNews(id string) (*model.News, error) {
	var news model.News
	err := database.DB.Where("`news_status` = 1 AND `id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	database.DB.Model(&news).UpdateColumn("news_view_cnt", news.ViewCnt+1)
	news = populateNewsFields([]model.News{news})[0]
	return &news, nil
}

func GetNewsCateList() ([]map[string]interface{}, error) {
	var setup model.Setup
	err := database.DB.Where("`setup_key` = ?", "news_cate").First(&setup).Error
	if err != nil {
		return nil, err
	}
	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(setup.Value), &raw); err != nil {
		return nil, err
	}
	// Map title -> name for frontend compatibility
	var result []map[string]interface{}
	for _, item := range raw {
		m := map[string]interface{}{
			"id": item["id"],
			"name": item["title"],
		}
		result = append(result, m)
	}
	return result, nil
}
