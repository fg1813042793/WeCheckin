package event

import (
	"context"

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

var clientEventListColumns = []string{
	"id",
	"event_title",
	"event_type",
	"event_status",
	"event_dept_id",
	"event_publish_dept_ids",
	"event_cate_id",
	"event_cate_name",
	"event_reg_start",
	"event_reg_end",
	"event_event_start",
	"event_event_end",
	"event_order",
	"event_vouch",
	"event_is_top",
	"event_obj",
	"event_view_cnt",
	"event_join_cnt",
	"event_user_cnt",
	"event_add_time",
	"event_edit_time",
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
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	err := query.Select(clientEventListColumns).Order("`event_order` ASC, `event_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEventFields(list)

	if userID != "" && len(list) > 0 {
		eventIDs := eventListIDs(list)
		participatedIDs := map[uint]bool{}
		var parts []model.EventParticipant
		if err := db.Select("event_part_event_id").
			Where("`event_part_mini_openid` = ? AND `event_part_event_id` IN ?", userID, eventIDs).
			Find(&parts).Error; err != nil {
			return nil, err
		}
		for _, p := range parts {
			participatedIDs[p.EventID] = true
		}
		for i := range list {
			list[i].IsJoin = participatedIDs[list[i].ID]
		}
	}

	return &ListResponse{List: list, Total: total}, nil
}

func eventListIDs(list []model.Event) []uint {
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.ID)
	}
	return ids
}
