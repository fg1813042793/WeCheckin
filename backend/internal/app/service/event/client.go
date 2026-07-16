package event

import (
	"context"
	"strconv"

	"wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/app/support/publish"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ==================== Client ====================

type ListResponse struct {
	List  []model.Event `json:"list"`
	Total int64         `json:"total"`
}

func GetEventList(page, pageSize int, userID, keyword, typ string) (*ListResponse, error) {
	return GetEventListContext(context.Background(), page, pageSize, userID, keyword, typ)
}

func GetEventListContext(ctx context.Context, page, pageSize int, userID, keyword, typ string) (*ListResponse, error) {
	var list []model.Event
	var total int64
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Model(&model.Event{}).Where("`event_status` = 1")
	if keyword != "" {
		query = query.Where("`event_title` LIKE ?", "%"+keyword+"%")
	}
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	if userID != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL OR " +
				publish.DeptOverlap("event_publish_dept_ids", deptIDs) + ")")
		} else {
			query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)")
		}
	} else {
		query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)")
	}
	query.Count(&total)
	err := query.Order("`event_order` ASC, `event_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEventFields(list)

	if userID != "" {
		participatedIDs := map[string]bool{}
		var parts []model.EventParticipant
		db.Where("`event_part_mini_openid` = ?", userID).Find(&parts)
		for _, p := range parts {
			participatedIDs[strconv.Itoa(int(p.EventID))] = true
		}
		for i := range list {
			idStr := strconv.Itoa(int(list[i].ID))
			list[i].IsJoin = participatedIDs[idStr]
		}
	}

	return &ListResponse{List: list, Total: total}, nil
}
