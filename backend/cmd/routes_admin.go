package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
	admindept "wecheckin-backend/backend/internal/app/handler/admin/department"
	admindict "wecheckin-backend/backend/internal/app/handler/admin/dict"
	adminenroll "wecheckin-backend/backend/internal/app/handler/admin/enroll"
	adminevent "wecheckin-backend/backend/internal/app/handler/admin/event"
	adminhome "wecheckin-backend/backend/internal/app/handler/admin/home"
	adminmenu "wecheckin-backend/backend/internal/app/handler/admin/menu"
	adminmgr "wecheckin-backend/backend/internal/app/handler/admin/mgr"
	adminnews "wecheckin-backend/backend/internal/app/handler/admin/news"
	adminrole "wecheckin-backend/backend/internal/app/handler/admin/role"
	adminsetup "wecheckin-backend/backend/internal/app/handler/admin/setup"
	adminuser "wecheckin-backend/backend/internal/app/handler/admin/user"
	"wecheckin-backend/backend/internal/middleware"
)

func registerAdminRoutes(h *server.Hertz) {
	aMgr := adminmgr.NewAdminMgrHandler()
	h.POST("/admin/login", aMgr.AdminLogin)

	adminGroup := h.Group("/admin", middleware.AdminAuth(), middleware.AdminPerm())
	registerAdminBaseRoutes(adminGroup, aMgr)
	registerAdminContentRoutes(adminGroup)
	registerAdminSystemRoutes(adminGroup)
	registerAdminSurveyRoutes(adminGroup)
	registerAdminExamRoutes(adminGroup)
}

func registerAdminBaseRoutes(adminGroup *route.RouterGroup, aMgr *adminmgr.AdminMgrHandler) {
	aHome := adminhome.NewAdminHomeHandler()
	aSetup := adminsetup.NewAdminSetupHandler()
	aUser := adminuser.NewAdminUserHandler()

	adminGroup.GET("/home", aHome.AdminHome)
	adminGroup.GET("/clear_vouch", aHome.ClearVouchData)
	adminGroup.GET("/mgr_list", aMgr.GetMgrList)
	adminGroup.POST("/mgr_insert", aMgr.InsertMgr)
	adminGroup.POST("/mgr_del", aMgr.DelMgr)
	adminGroup.POST("/mgr_dels", aMgr.DelMgrs)
	adminGroup.GET("/mgr_detail", aMgr.GetMgrDetail)
	adminGroup.POST("/mgr_edit", aMgr.EditMgr)
	adminGroup.POST("/mgr_status", aMgr.StatusMgr)
	adminGroup.POST("/mgr_pwd", aMgr.PwdMgr)
	adminGroup.GET("/log_list", aMgr.GetLogList)
	adminGroup.GET("/log_clear", aMgr.ClearLog)

	adminGroup.POST("/setup_set", aSetup.SetSetup)
	adminGroup.POST("/setup_set_content", aSetup.SetContentSetup)
	adminGroup.GET("/setup_qr", aSetup.GenMiniQr)
	adminGroup.GET("/setup_debug_token", aSetup.DebugTokenConfig)

	adminGroup.GET("/user_list", aUser.GetUserList)
	adminGroup.GET("/user_detail", aUser.GetUserDetail)
	adminGroup.GET("/user_detail_by_id", aUser.GetUserByID)
	adminGroup.POST("/user_add", aUser.AddUser)
	adminGroup.POST("/user_edit", aUser.EditUser)
	adminGroup.POST("/user_del", aUser.DelUser)
	adminGroup.POST("/user_dels", aUser.DelUsers)
	adminGroup.POST("/user_status", aUser.StatusUser)
	adminGroup.POST("/user_reset_pwd", aUser.ResetPassword)
	adminGroup.GET("/user_form_fields", aUser.GetUserFormFields)
	adminGroup.POST("/user_form_field_save", aUser.SaveUserFormFields)
	adminGroup.GET("/user_data_get", aUser.UserDataGet)
	adminGroup.GET("/user_data_export", aUser.UserDataExport)
	adminGroup.POST("/user_data_del", aUser.UserDataDel)

	adminGroup.GET("/user/online", aUser.GetOnlineUsers)
	adminGroup.POST("/user/force_offline", aUser.ForceOfflineUser)
	adminGroup.POST("/user/batch_force_offline", aUser.BatchForceOfflineUser)
	adminGroup.GET("/admin/online", aMgr.GetOnlineAdmins)
	adminGroup.POST("/admin/force_offline", aMgr.ForceOfflineAdmin)
	adminGroup.POST("/admin/batch_force_offline", aMgr.BatchForceOfflineAdmin)
	adminGroup.POST("/admin/logout", aMgr.AdminLogout)
}

func registerAdminContentRoutes(adminGroup *route.RouterGroup) {
	aNews := adminnews.NewAdminNewsHandler()
	aEnroll := adminenroll.NewAdminEnrollHandler()
	aEvent := adminevent.NewAdminEventHandler()

	adminGroup.GET("/news_list", aNews.GetAdminNewsList)
	adminGroup.POST("/news_insert", aNews.InsertNews)
	adminGroup.GET("/news_detail", aNews.GetNewsDetail)
	adminGroup.POST("/news_edit", aNews.EditNews)
	adminGroup.POST("/news_update_forms", aNews.UpdateNewsForms)
	adminGroup.POST("/news_update_pic", aNews.UpdateNewsPic)
	adminGroup.POST("/news_update_content", aNews.UpdateNewsContent)
	adminGroup.POST("/news_del", aNews.DelNews)
	adminGroup.POST("/news_dels", aNews.DelNewsList)
	adminGroup.POST("/news_sort", aNews.SortNews)
	adminGroup.POST("/news_status", aNews.StatusNews)

	adminGroup.GET("/enroll_list", aEnroll.GetAdminEnrollList)
	adminGroup.POST("/enroll_insert", aEnroll.InsertEnroll)
	adminGroup.GET("/enroll_detail", aEnroll.GetEnrollDetail)
	adminGroup.POST("/enroll_edit", aEnroll.EditEnroll)
	adminGroup.POST("/enroll_update_forms", aEnroll.UpdateEnrollForms)
	adminGroup.POST("/enroll_clear", aEnroll.ClearEnrollAll)
	adminGroup.POST("/enroll_del", aEnroll.DelEnroll)
	adminGroup.POST("/enroll_dels", aEnroll.DelEnrolls)
	adminGroup.POST("/enroll_sort", aEnroll.SortEnroll)
	adminGroup.POST("/enroll_vouch", aEnroll.VouchEnroll)
	adminGroup.POST("/enroll_status", aEnroll.StatusEnroll)
	adminGroup.GET("/enroll_join_list", aEnroll.GetEnrollJoinList)
	adminGroup.GET("/enroll_user_list", aEnroll.GetEnrollUserList)
	adminGroup.GET("/enroll_stats", aEnroll.GetEnrollStats)
	adminGroup.POST("/enroll_remove_user", aEnroll.RemoveEnrollUser)
	adminGroup.POST("/enroll_remove_users", aEnroll.RemoveEnrollUsers)
	adminGroup.POST("/enroll_user_forms_edit", aEnroll.EditEnrollUserForms)
	adminGroup.POST("/enroll_join_del", aEnroll.DelEnrollJoin)
	adminGroup.POST("/enroll_join_dels", aEnroll.DelEnrollJoins)
	adminGroup.GET("/enroll_join_data_get", aEnroll.EnrollJoinDataGet)
	adminGroup.GET("/enroll_join_data_export", aEnroll.EnrollJoinDataExport)
	adminGroup.POST("/enroll_join_data_del", aEnroll.EnrollJoinDataDel)

	adminGroup.GET("/event_list", aEvent.GetAdminEventList)
	adminGroup.GET("/event_detail", aEvent.GetAdminEventDetail)
	adminGroup.POST("/event_insert", aEvent.InsertEvent)
	adminGroup.POST("/event_edit", aEvent.EditEvent)
	adminGroup.POST("/event_del", aEvent.DelEvent)
	adminGroup.POST("/event_dels", aEvent.DelEvents)
	adminGroup.POST("/event_status", aEvent.StatusEvent)
	adminGroup.GET("/event_participant_list", aEvent.GetEventParticipantList)
	adminGroup.POST("/event_participant_del", aEvent.DelEventParticipant)
	adminGroup.POST("/event_participant_dels", aEvent.DelEventParticipants)
	adminGroup.POST("/event_participant_edit", aEvent.EditEventParticipant)
	adminGroup.POST("/event_dynamic_add", aEvent.PostEventDynamic)
	adminGroup.GET("/event_dynamics", aEvent.GetEventDynamics)
	adminGroup.POST("/event_dynamic_edit", aEvent.EditEventDynamic)
	adminGroup.POST("/event_dynamic_del", aEvent.DelEventDynamic)
	adminGroup.POST("/event_dynamic_dels", aEvent.DelEventDynamics)
	adminGroup.GET("/event_scores", aEvent.GetEventScores)
	adminGroup.POST("/event_score_edit", aEvent.EditEventScore)
	adminGroup.GET("/dept_users", aEvent.GetDeptUsers)
	adminGroup.POST("/event_vouch", aEvent.VouchEvent)
	adminGroup.POST("/event_top", aEvent.TopEvent)
}

func registerAdminSystemRoutes(adminGroup *route.RouterGroup) {
	aDict := admindict.NewAdminDictHandler()
	aDept := admindept.NewAdminDeptHandler()
	aRole := adminrole.NewAdminRoleHandler()
	aMenu := adminmenu.NewAdminMenuHandler()

	adminGroup.GET("/dict/types", aDict.GetDictTypes)
	adminGroup.GET("/dict/items", aDict.GetDictByType)
	adminGroup.POST("/dict/add", aDict.AddDictItem)
	adminGroup.POST("/dict/edit", aDict.EditDictItem)
	adminGroup.POST("/dict/del", aDict.DelDictItem)
	adminGroup.POST("/dict/clear", aDict.DelDictByType)
	adminGroup.POST("/dict/edit_type_name", aDict.EditDictTypeName)

	adminGroup.GET("/dept/tree", aDept.GetDeptTree)
	adminGroup.POST("/dept/add", aDept.AddDept)
	adminGroup.POST("/dept/edit", aDept.EditDept)
	adminGroup.POST("/dept/del", aDept.DelDept)

	adminGroup.GET("/role/list", aRole.GetRoleList)
	adminGroup.POST("/role/add", aRole.AddRole)
	adminGroup.POST("/role/edit", aRole.EditRole)
	adminGroup.POST("/role/del", aRole.DelRole)
	adminGroup.POST("/role/dels", aRole.DelRoles)

	adminGroup.GET("/menu/tree", aMenu.GetMenuTree)
	adminGroup.GET("/menu/list", aMenu.GetMenuList)
	adminGroup.POST("/menu/add", aMenu.AddMenu)
	adminGroup.POST("/menu/edit", aMenu.EditMenu)
	adminGroup.POST("/menu/del", aMenu.DelMenu)

	adminGroup.GET("/user/menus", aMenu.GetAdminMenus)
	adminGroup.GET("/user/perms", aMenu.GetAdminPerms)
}
