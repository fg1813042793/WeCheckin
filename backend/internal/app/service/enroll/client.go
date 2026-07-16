package enroll

import (
	"context"
	"strconv"

	"wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/publish"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type ListResponse struct {
	List  []model.Enroll `json:"list"`
	Total int64          `json:"total"`
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
	query.Count(&total)
	err := query.Order("`enroll_order` ASC, `enroll_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEnrollFields(list)

	// Get user's joined enroll IDs (from both EnrollJoin and EnrollUser)
	joinedIDs := map[string]bool{}
	if userID != "" {
		var joins []model.EnrollJoin
		db.Where("`enroll_join_user_id` = ?", userID).Find(&joins)
		for _, j := range joins {
			joinedIDs[j.EnrollID] = true
		}
		var enrollUsers []model.EnrollUser
		db.Where("`enroll_user_mini_openid` = ?", userID).Find(&enrollUsers)
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
