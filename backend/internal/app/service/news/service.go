package news

import (
	"context"
	"encoding/json"

	"wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/app/support/publish"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type ListResponse struct {
	List  []model.News `json:"list"`
	Total int64        `json:"total"`
}

type CategoryItem struct {
	ID   interface{} `json:"id"`
	Name interface{} `json:"name"`
}

func PopulateFields(list []model.News) []model.News {
	for i := range list {
		if list[i].Pic != "" {
			var urls []string
			if err := json.Unmarshal([]byte(list[i].Pic), &urls); err == nil && len(urls) > 0 {
				list[i].Img = media.FullURLWithStaticDomain(urls[0])
			} else {
				list[i].Img = media.FullURLWithStaticDomain(list[i].Pic)
			}
		}
	}
	return list
}

func GetNewsList(page, pageSize int, keyword, userID string) (*ListResponse, error) {
	return GetNewsListContext(context.Background(), page, pageSize, keyword, userID)
}

func GetNewsListContext(ctx context.Context, page, pageSize int, keyword, userID string) (*ListResponse, error) {
	var list []model.News
	var total int64
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Model(&model.News{}).Where("`news_status` = 1")
	if keyword != "" {
		query = query.Where("`news_title` LIKE ?", "%"+keyword+"%")
	}
	if userID != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenID(userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`news_publish_dept_ids` = '' OR `news_publish_dept_ids` IS NULL OR " +
				publish.DeptOverlap("news_publish_dept_ids", deptIDs) + ")")
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
	list = PopulateFields(list)
	return &ListResponse{List: list, Total: total}, nil
}

func ViewNews(id string) (*model.News, error) {
	return ViewNewsContext(context.Background(), id)
}

func ViewNewsContext(ctx context.Context, id string) (*model.News, error) {
	var news model.News
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err := db.Where("`news_status` = 1 AND `id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	db.Model(&news).UpdateColumn("news_view_cnt", news.ViewCnt+1)
	news = PopulateFields([]model.News{news})[0]
	return &news, nil
}

func GetNewsCateList() ([]CategoryItem, error) {
	return GetNewsCateListContext(context.Background())
}

func GetNewsCateListContext(ctx context.Context) ([]CategoryItem, error) {
	var setup model.Setup
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err := db.Where("`setup_key` = ?", "news_cate").First(&setup).Error
	if err != nil {
		return nil, err
	}
	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(setup.Value), &raw); err != nil {
		return nil, err
	}
	var result []CategoryItem
	for _, item := range raw {
		result = append(result, CategoryItem{ID: item["id"], Name: item["title"]})
	}
	return result, nil
}
