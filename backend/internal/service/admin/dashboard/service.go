package dashboard

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	userfeedbackmodel "wecheckin/backend/internal/model/userfeedback"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/adminaccess"
	"wecheckin/backend/pkg/database"
)

func AdminHome(adminID uint) (AdminHomeResponse, error) {
	return AdminHomeContext(context.Background(), adminID)
}

func AdminHomeContext(ctx context.Context, adminID uint) (AdminHomeResponse, error) {
	return AdminHomeWithDaysContext(ctx, adminID, 7)
}

func AdminHomeWithDaysContext(ctx context.Context, adminID uint, days int) (AdminHomeResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return AdminHomeResponse{}, err
	}
	return loadAdminHome(ctx, db, &admin, normalizeTrendDays(days), time.Now())
}

func loadAdminHome(ctx context.Context, db *gorm.DB, admin *model.Admin, days int, now time.Time) (AdminHomeResponse, error) {
	days = normalizeTrendDays(days)
	todayStart := startOfDay(now).UnixMilli()
	todayEnd := startOfDay(now).AddDate(0, 0, 1).Add(-time.Millisecond).UnixMilli()
	trendStart := startOfDay(now).AddDate(0, 0, -(days - 1)).UnixMilli()
	adminID := strconv.FormatUint(uint64(admin.ID), 10)

	userWhere, userArgs := access.UserDataScopeFilterWithDBContext(ctx, db, admin)
	userQuery := func() *gorm.DB {
		query := db.Model(&model.User{})
		if userWhere != "" {
			query = query.Where(userWhere, userArgs...)
		}
		return query
	}

	enrollWhere, enrollArgs := access.DataScopeFilterForResourceWithDBContext(ctx, db, admin, access.EnrollAuditFields)
	enrollQuery := func() *gorm.DB {
		query := db.Model(&model.Enroll{})
		if enrollWhere != "" {
			query = query.Where(enrollWhere, enrollArgs...)
		}
		return query
	}

	newsWhere, newsArgs := access.DataScopeFilterForResourceWithDBContext(ctx, db, admin, access.NewsAuditFields)
	newsQuery := func() *gorm.DB {
		query := db.Model(&model.News{})
		if newsWhere != "" {
			query = query.Where(newsWhere, newsArgs...)
		}
		return query
	}

	eventWhere, eventArgs := access.DataScopeFilterForResourceWithDBContext(ctx, db, admin, access.EventAuditFields)
	eventQuery := func() *gorm.DB {
		query := db.Model(&model.Event{})
		if eventWhere != "" {
			query = query.Where(eventWhere, eventArgs...)
		}
		return query
	}

	surveyWhere, surveyArgs := access.DataScopeFilterForResourceWithDBContext(ctx, db, admin, access.SurveyAuditFields)
	surveyQuery := func() *gorm.DB {
		query := db.Model(&model.Survey{})
		if surveyWhere != "" {
			query = query.Where(surveyWhere, surveyArgs...)
		}
		return query
	}

	examWhere, examArgs := access.DataScopeFilterForResourceWithDBContext(ctx, db, admin, access.ExamAuditFields)
	examQuery := func() *gorm.DB {
		query := db.Model(&model.Exam{})
		if examWhere != "" {
			query = query.Where(examWhere, examArgs...)
		}
		return query
	}

	capabilities, err := dashboardCapabilities(ctx, db, admin)
	if err != nil {
		return AdminHomeResponse{}, err
	}
	result := AdminHomeResponse{TrendDays: days, UpdatedAt: now.UnixMilli(), Capabilities: capabilities}
	if capabilities["users"] {
		if err := userQuery().Count(&result.UserCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
		if err := userQuery().Where("user_status = ?", 0).Count(&result.PendingUserCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["users"] && capabilities["dingtalkBindings"] {
		if err := userQuery().Where(`NOT EXISTS (
		SELECT 1 FROM dingtalk_h5_user_bindings dashboard_binding
		WHERE dashboard_binding.user_id = users.id AND dashboard_binding.enabled = 1
	)`).Count(&result.UnboundDingTalkUserCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["enrollments"] {
		if err := enrollQuery().Count(&result.EnrollCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["content"] {
		if err := newsQuery().Count(&result.NewsCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["events"] {
		if err := eventQuery().Count(&result.EventCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["surveys"] {
		if err := surveyQuery().Count(&result.SurveyCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["exams"] {
		if err := examQuery().Count(&result.ExamCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["managers"] {
		if err := adminaccess.ApplyUserAdminAccessRoleFilter(db.Model(&model.Admin{})).Count(&result.MgrCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	var enrollIDs, eventIDs, surveyIDs, examIDs []uint
	if capabilities["enrollments"] {
		if err := enrollQuery().Pluck("id", &enrollIDs).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["events"] {
		if err := eventQuery().Pluck("id", &eventIDs).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["surveys"] || capabilities["surveyResponses"] {
		if err := surveyQuery().Pluck("survey_id", &surveyIDs).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["exams"] {
		if err := examQuery().Pluck("exam_id", &examIDs).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	joinQuery := db.Model(&model.EnrollJoin{})
	if len(enrollIDs) == 0 {
		joinQuery = joinQuery.Where("1 = 0")
	} else {
		joinQuery = joinQuery.Where("enroll_join_enroll_id IN ?", enrollIDs)
	}
	if err := joinQuery.Where("enroll_join_add_time BETWEEN ? AND ?", todayStart, todayEnd).Count(&result.JoinCnt).Error; err != nil {
		return AdminHomeResponse{}, err
	}
	if len(eventIDs) > 0 {
		if err := db.Model(&model.EventParticipant{}).Where("event_part_event_id IN ?", eventIDs).Count(&result.EventUserCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	surveyResponseQuery := db.Model(&model.SurveyResponse{}).Where("survey_resp_status = ?", 1)
	if len(surveyIDs) == 0 {
		surveyResponseQuery = surveyResponseQuery.Where("1 = 0")
	} else {
		surveyResponseQuery = surveyResponseQuery.Where("survey_resp_survey_id IN ?", surveyIDs)
	}
	if capabilities["surveyResponses"] {
		if err := surveyResponseQuery.Where("survey_resp_add_time BETWEEN ? AND ?", todayStart, todayEnd).Count(&result.TodaySurveyResponseCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	examRecordQuery := db.Model(&model.ExamRecord{}).Where("exam_r_status >= ?", 1)
	if len(examIDs) == 0 {
		examRecordQuery = examRecordQuery.Where("1 = 0")
	} else {
		examRecordQuery = examRecordQuery.Where("exam_r_exam_id IN ?", examIDs)
	}
	if capabilities["exams"] {
		if err := examRecordQuery.Where("exam_r_submit_time BETWEEN ? AND ?", todayStart, todayEnd).Count(&result.TodayExamRecordCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	if capabilities["workflowTasks"] {
		if err := db.Model(&workflowmodel.ProcessTask{}).
			Where("task_assignee_id = ? AND task_status = ? AND admin_deleted_at = 0", adminID, workflowmodel.TaskStatusPending).
			Where("EXISTS (SELECT 1 FROM workflow_process_instances dashboard_instance WHERE dashboard_instance.id = workflow_process_tasks.instance_id AND dashboard_instance.instance_status = ? AND dashboard_instance.admin_deleted_at = 0)", workflowmodel.InstanceStatusRunning).
			Count(&result.PendingWorkflowTaskCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["workflowNotifications"] {
		if err := db.Model(&workflowmodel.NotificationOutbox{}).
			Where("recipient_user_id = ? AND notification_status IN ? AND admin_deleted_at = 0", adminID, []string{workflowmodel.NotificationStatusFailed, workflowmodel.NotificationStatusDead}).
			Count(&result.FailedWorkflowNotificationCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["scheduledRuns"] {
		if err := db.Model(&scheduledtaskmodel.Run{}).
			Where("run_status = ? AND add_time >= ?", scheduledtaskmodel.RunStatusFailed, now.Add(-24*time.Hour).UnixMilli()).
			Count(&result.FailedScheduledRunCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["feedback"] {
		if err := db.Model(&userfeedbackmodel.Feedback{}).
			Where("feedback_status IN ?", []string{"pending", "processing"}).
			Count(&result.PendingFeedbackCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["notifications"] {
		if err := db.Model(&model.Notify{}).
			Where("notify_user_id = ? AND notify_deleted_at = 0 AND notify_is_read = 0", adminID).
			Count(&result.UnreadNotificationCnt).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}

	timestamps := dashboardTrendTimestamps{}
	if capabilities["users"] {
		if err := userQuery().Where("user_add_time >= ?", trendStart).Pluck("user_add_time", &timestamps.Users).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["enrollments"] {
		if err := joinQuery.Where("enroll_join_add_time >= ?", trendStart).Pluck("enroll_join_add_time", &timestamps.EnrollJoins).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["surveyResponses"] {
		if err := surveyResponseQuery.Where("survey_resp_add_time >= ?", trendStart).Pluck("survey_resp_add_time", &timestamps.SurveyResponses).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	if capabilities["exams"] {
		if err := examRecordQuery.Where("exam_r_submit_time >= ?", trendStart).Pluck("exam_r_submit_time", &timestamps.ExamRecords).Error; err != nil {
			return AdminHomeResponse{}, err
		}
	}
	result.Trend = buildTrendPoints(now, days, timestamps)
	return result, nil
}

func ClearVouchData() error {
	return ClearVouchDataContext(context.Background())
}

func ClearVouchDataContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Enroll{}).Where("1 = 1").Update("enroll_vouch", 0).Error
}
