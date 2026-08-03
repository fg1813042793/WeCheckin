package enroll

import (
	"context"
	"strconv"

	"wecheckin/backend/internal/app/support/dept"
	"wecheckin/backend/internal/app/support/publish"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ListResponse struct {
	List  []model.Enroll `json:"list"`
	Total int64          `json:"total"`
}

var clientEnrollListColumns = []string{
	"id",
	"enroll_title",
	"enroll_status",
	"create_dept_id",
	"enroll_publish_dept_ids",
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
	"enroll_view_cnt",
	"enroll_join_cnt",
	"enroll_user_cnt",
	"add_time",
	"edit_time",
}

func GetEnrollList(page, pageSize int, userID, keyword string) (*ListResponse, error) {
	return GetEnrollListContext(context.Background(), page, pageSize, userID, keyword)
}

func GetEnrollListContext(ctx context.Context, page, pageSize int, userID, keyword string) (*ListResponse, error) {
	var list []model.Enroll
	var total int64
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Model(&model.Enroll{}).Where("`enroll_status` = 1")
	if keyword != "" {
		query = query.Where("`enroll_title` LIKE ? OR `enroll_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Filter by publish departments
	if userID != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL OR " +
				publish.DeptOverlap("enroll_publish_dept_ids", deptIDs) + ")")
		} else {
			query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)")
		}
	} else {
		query = query.Where("(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	err := query.Select(clientEnrollListColumns).Order("`enroll_order` ASC, `add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEnrollFields(list)

	// Get user's joined enroll IDs (from both EnrollJoin and EnrollUser)
	joinedIDs := map[string]bool{}
	if userID != "" && len(list) > 0 {
		enrollIDs := enrollListIDs(list)
		var joins []model.EnrollJoin
		if err := db.Select("enroll_join_enroll_id").
			Where("`enroll_join_user_id` = ? AND `enroll_join_enroll_id` IN ?", userID, enrollIDs).
			Find(&joins).Error; err != nil {
			return nil, err
		}
		for _, j := range joins {
			joinedIDs[j.EnrollID] = true
		}
		var enrollUsers []model.EnrollUser
		if err := db.Select("enroll_user_enroll_id").
			Where("`enroll_user_mini_openid` = ? AND `enroll_user_enroll_id` IN ?", userID, enrollIDs).
			Find(&enrollUsers).Error; err != nil {
			return nil, err
		}
		for _, eu := range enrollUsers {
			joinedIDs[eu.EnrollID] = true
		}
	}
	for i := range list {
		idStr := strconv.Itoa(int(list[i].ID))
		list[i].IsJoin = joinedIDs[idStr]
	}

	return &ListResponse{List: list, Total: total}, nil
}

func enrollListIDs(list []model.Enroll) []string {
	ids := make([]string, 0, len(list))
	for _, item := range list {
		ids = append(ids, strconv.Itoa(int(item.ID)))
	}
	return ids
}
