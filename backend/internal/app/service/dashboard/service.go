package dashboard

import (
	"context"
	"time"

	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/adminaccess"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type AdminHomeResponse struct {
	UserCnt      int64 `json:"userCnt"`
	EnrollCnt    int64 `json:"enrollCnt"`
	NewsCnt      int64 `json:"newsCnt"`
	JoinCnt      int64 `json:"joinCnt"`
	EventCnt     int64 `json:"eventCnt"`
	EventUserCnt int64 `json:"eventUserCnt"`
	MgrCnt       int64 `json:"mgrCnt"`
}

func AdminHome(adminID uint) (AdminHomeResponse, error) {
	return AdminHomeContext(context.Background(), adminID)
}

func AdminHomeContext(ctx context.Context, adminID uint) (AdminHomeResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	var userCnt int64
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) || admin.RoleID == 0 {
		if err := db.Model(&model.User{}).Count(&userCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	} else {
		q := db.Model(&model.User{})
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = access.AdminDeptIDsContext(ctx, admin.ID)
				} else {
					deptIDs = access.RoleDeptIDsContext(ctx, admin.RoleID)
				}
				if len(deptIDs) > 0 {
					q = q.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
				}
			} else if role.DataScope == 3 {
				q = q.Where("1 = 0")
			}
		}
		if err := q.Count(&userCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	var enrollCnt int64
	q := db.Model(&model.Enroll{})
	where, args := access.DataScopeFilterContext(ctx, &admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		q = q.Where(where, args...)
	}
	if err := q.Count(&enrollCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	var newsCnt int64
	q2 := db.Model(&model.News{})
	where2, args2 := access.DataScopeFilterContext(ctx, &admin, "`news_dept_id`", "`news_create_by`")
	if where2 != "" {
		q2 = q2.Where(where2, args2...)
	}
	if err := q2.Count(&newsCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999, now.Location()).UnixMilli()
	var joinCnt int64
	q3 := db.Model(&model.EnrollJoin{}).Where("`enroll_join_add_time` BETWEEN ? AND ?", todayStart, todayEnd)
	if where != "" {
		q3 = q3.Where("`enroll_join_enroll_id` IN (SELECT `id` FROM `enrolls` WHERE "+where+")", args...)
	}
	if err := q3.Count(&joinCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	var eventCnt int64
	q4 := db.Model(&model.Event{})
	where4, args4 := access.DataScopeFilterContext(ctx, &admin, "`event_dept_id`", "`event_create_by`")
	if where4 != "" {
		q4 = q4.Where(where4, args4...)
	}
	if err := q4.Count(&eventCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	var eventUserCnt int64
	if err := db.Model(&model.EventParticipant{}).Count(&eventUserCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	var mgrCnt int64
	if err := adminaccess.ApplyUserAdminAccessRoleFilter(db.Model(&model.Admin{})).
		Count(&mgrCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}

	return AdminHomeResponse{
		UserCnt:      userCnt,
		EnrollCnt:    enrollCnt,
		NewsCnt:      newsCnt,
		JoinCnt:      joinCnt,
		EventCnt:     eventCnt,
		EventUserCnt: eventUserCnt,
		MgrCnt:       mgrCnt,
	}, nil
}

func ClearVouchData() error {
	return ClearVouchDataContext(context.Background())
}

func ClearVouchDataContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("1 = 1").Update("enroll_vouch", 0).Error
}
