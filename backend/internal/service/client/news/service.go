package news

import (
	"context"
	"encoding/json"

	"wecheckin/backend/internal/model"
	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/publish"
	"wecheckin/backend/pkg/database"
)

type ListResponse struct {
	List  []model.News `json:"list"`
	Total int64        `json:"total"`
}

type CategoryItem struct {
	ID   interface{} `json:"id"`
	Name interface{} `json:"name"`
}

var ListColumns = []string{
	"id",
	"news_title",
	"news_desc",
	"news_status",
	"create_dept_id",
	"news_publish_dept_ids",
	"create_by",
	"update_by",
	"update_dept_id",
	"news_cate_id",
	"news_cate_name",
	"news_order",
	"news_vouch",
	"news_view_cnt",
	"news_pic",
	"add_time",
	"edit_time",
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
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
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
	err := query.Select(ListColumns).Order("`news_order` ASC, `add_time` DESC").
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
	setup, err := setupservice.GetSetupContext(ctx, "news_cate")
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
