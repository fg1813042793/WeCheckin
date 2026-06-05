// WeCheckin API
//
//	@title			WeCheckin API
//	@version		1.0
//	@description	微信小程序打卡项目后端 API
//	@host			localhost:8080
//	@BasePath		/
//	@schemes		http
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "wecheckin-backend/backend/docs/swagger"
	"wecheckin-backend/backend/internal/api/admin"
	"wecheckin-backend/backend/internal/api/client"
	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/middleware"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/logger"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
)

func main() {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	flag.Parse()

	cfg, err := config.LoadConfig(*env)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database.InitDatabase(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)

	if err := logger.Init(cfg.Log.Dir, cfg.Log.Level, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		logger.Logger.Printf("Warning: logger init: %v", err)
	}

	if err := rd.Init(cfg.Redis); err != nil {
		logger.Logger.Printf("Warning: Redis init failed: %v", err)
	} else {
		logger.Logger.Println("Redis connected")
	}

	h := server.Default(server.WithHostPorts(":"+cfg.Server.Port), server.WithMaxRequestBodySize(32*1024*1024))

	h.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: true,
		MaxAge:           time.Hour,
	}))

	h.Use(middleware.AccessLog())

	url := swagger.URL("http://" + cfg.Server.Host + ":" + cfg.Server.Port + "/swagger/doc.json")
	h.GET("/swagger", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/swagger/index.html"))
	})
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))

	hm := client.NewHomeHandler()
	pp := client.NewPassportHandler()
	ns := client.NewNewsHandler()
	el := client.NewEnrollHandler()
	geo := client.NewGeoHandler()
	fa := client.NewFavHandler()
	ev := client.NewEventHandler()
	aHome := admin.NewAdminHomeHandler()
	aMgr := admin.NewAdminMgrHandler()
	aSetup := admin.NewAdminSetupHandler()
	aUser := admin.NewAdminUserHandler()
	aNews := admin.NewAdminNewsHandler()
	aEnroll := admin.NewAdminEnrollHandler()
	aEvent := admin.NewAdminEventHandler()
	aDict := admin.NewAdminDictHandler()
	aDept := admin.NewAdminDeptHandler()
	aRole := admin.NewAdminRoleHandler()
	aMenu := admin.NewAdminMenuHandler()

	// ==================== Public routes (no auth) ====================
	h.GET("/test/test", func(ctx context.Context, c *app.RequestContext) {
		response.JSON(c, map[string]string{"msg": "ok"})
	})
	h.GET("/test/debug_token", func(ctx context.Context, c *app.RequestContext) {
		token := c.Query("token")
		expire, prefix := tokenutil.GetTokenConfig("admin")
		log.Printf("[DebugEndpoint] token=%q raw_prefix=%q", token, prefix)
		if prefix == "" {
			prefix = "admin_token:"
			log.Printf("[DebugEndpoint] applied fallback prefix=%q", prefix)
		}
		key := prefix + "a:" + token
		val, err := rd.RDB.Get(rd.Ctx, key).Result()
		result := map[string]interface{}{
			"prefix":      prefix,
			"key":         key,
			"redisErr":    fmt.Sprintf("%v", err),
			"redisVal":    val,
			"rdbNil":      rd.RDB == nil,
			"expire":      expire.String(),
			"cfgAdmin":    config.Cfg.Token.Admin,
		}
		keys, _, _ := rd.RDB.Scan(rd.Ctx, 0, prefix+"*", 100).Result()
		result["allKeys"] = keys
		response.JSON(c, result)
	})

	h.GET("/home/setup_get", hm.GetSetup)
	h.GET("/home/list", hm.GetHomeList)
	h.GET("/user_form_fields", aUser.GetUserFormFields)

	h.POST("/passport/login", pp.Login)
	h.POST("/passport/login_pwd", pp.LoginByPwd)
	h.POST("/passport/register", pp.Register)
	h.GET("/geo/reverse", geo.ReverseGeocode)

	h.GET("/dict/types", aDict.GetDictTypes)
	h.GET("/dict/items", aDict.GetDictByType)

	// ==================== Client auth routes ====================
	clientGroup := h.Group("/passport", middleware.ClientAuth())
	clientGroup.POST("/phone", pp.GetPhone)
	clientGroup.GET("/my_detail", pp.GetMyDetail)
	clientGroup.POST("/edit_base", pp.EditBase)
	clientGroup.POST("/logout", pp.Logout)

	clientGroup1 := h.Group("/fav", middleware.ClientAuth())
	clientGroup1.POST("/update", fa.UpdateFav)
	clientGroup1.POST("/del", fa.DelFav)
	clientGroup1.GET("/is_fav", fa.IsFav)
	clientGroup1.GET("/my_list", fa.GetMyFavList)

	clientNews := h.Group("/news", middleware.ClientAuth())
	clientNews.GET("/list", ns.GetNewsList)
	clientNews.GET("/view", ns.ViewNews)
	clientNews.GET("/cate_list", ns.GetNewsCateList)

	clientEnroll := h.Group("/enroll", middleware.ClientAuth())
	clientEnroll.GET("/list", el.GetEnrollList)
	clientEnroll.GET("/view", el.ViewEnroll)
	clientEnroll.GET("/join_day", el.GetEnrollJoinByDay)
	clientEnroll.POST("/join", el.EnrollJoin)
	clientEnroll.POST("/enroll_submit", el.EnrollUserSubmit)
	clientEnroll.GET("/my_join_list", el.GetMyEnrollJoinList)
	clientEnroll.GET("/my_user_list", el.GetMyEnrollUserList)
	clientEnroll.GET("/my_records", el.GetMyJoinRecords)
	clientEnroll.GET("/my_calendar", el.GetMyCalendar)
	clientEnroll.GET("/my_day_records", el.GetMyDayRecords)

	clientEvent := h.Group("/event")
	// Public event routes
	clientEvent.GET("/list", ev.GetEventList)
	clientEvent.GET("/view", ev.ViewEvent)
	// Auth'ed event routes
	clientEventAuth := h.Group("/event", middleware.ClientAuth())
	clientEventAuth.POST("/participate", ev.EventParticipate)
	clientEventAuth.GET("/my_list", ev.GetMyEventList)
	clientEventAuth.GET("/my_roles", ev.GetMyEventRoles)
	clientEventAuth.GET("/my_managed", ev.GetMyManagedList)
	clientEventAuth.GET("/dynamics", ev.GetEventDynamics)
	clientEventAuth.POST("/dynamic_post", ev.PostEventDynamic)
	clientEventAuth.GET("/participant_list", ev.GetEventParticipantList)
	clientEventAuth.GET("/scores", ev.GetEventScores)
	clientEventAuth.POST("/score_save", ev.SaveEventScore)

	// ==================== Admin login (no auth) ====================
	h.POST("/admin/login", aMgr.AdminLogin)

	// ==================== Admin routes (with auth + permission middleware) ====================
	adminGroup := h.Group("/admin", middleware.AdminAuth(), middleware.AdminPerm())

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

	// Online users + force offline (no perm middleware needed for the list itself - perm check added per-handler)
	adminGroup.GET("/user/online", aUser.GetOnlineUsers)
	adminGroup.POST("/user/force_offline", aUser.ForceOfflineUser)
	adminGroup.GET("/admin/online", aMgr.GetOnlineAdmins)
	adminGroup.POST("/admin/force_offline", aMgr.ForceOfflineAdmin)
	adminGroup.POST("/admin/logout", aMgr.AdminLogout)

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

	// Admin dict routes
	adminGroup.GET("/dict/types", aDict.GetDictTypes)
	adminGroup.GET("/dict/items", aDict.GetDictByType)
	adminGroup.POST("/dict/add", aDict.AddDictItem)
	adminGroup.POST("/dict/edit", aDict.EditDictItem)
	adminGroup.POST("/dict/del", aDict.DelDictItem)
	adminGroup.POST("/dict/clear", aDict.DelDictByType)
	adminGroup.POST("/dict/edit_type_name", aDict.EditDictTypeName)

	// Department routes
	adminGroup.GET("/dept/tree", aDept.GetDeptTree)
	adminGroup.POST("/dept/add", aDept.AddDept)
	adminGroup.POST("/dept/edit", aDept.EditDept)
	adminGroup.POST("/dept/del", aDept.DelDept)

	// Role routes
	adminGroup.GET("/role/list", aRole.GetRoleList)
	adminGroup.POST("/role/add", aRole.AddRole)
	adminGroup.POST("/role/edit", aRole.EditRole)
	adminGroup.POST("/role/del", aRole.DelRole)
	adminGroup.POST("/role/dels", aRole.DelRoles)

	// Menu routes
	adminGroup.GET("/menu/tree", aMenu.GetMenuTree)
	adminGroup.GET("/menu/list", aMenu.GetMenuList)
	adminGroup.POST("/menu/add", aMenu.AddMenu)
	adminGroup.POST("/menu/edit", aMenu.EditMenu)
	adminGroup.POST("/menu/del", aMenu.DelMenu)

	// Admin's own menu tree and perms (for frontend sidebar)
	adminGroup.GET("/user/menus", aMenu.GetAdminMenus)
	adminGroup.GET("/user/perms", aMenu.GetAdminPerms)

	// ==================== File upload (public) ====================
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	h.POST("/upload", func(ctx context.Context, c *app.RequestContext) {
		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, "上传失败: "+err.Error())
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			// image - continue
		} else if ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".wmv" || ext == ".flv" || ext == ".mkv" {
			// video - continue
		} else {
			response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp/mp4/mov/avi/wmv/flv/mkv")
			return
		}
		if file.Size > 20*1024*1024 {
			response.Fail(c, "上传文件过大，最大20MB")
			return
		}
		now := time.Now()
		dateDir := now.Format("2006/01/02")
		saveDir := filepath.Join(uploadDir, dateDir)
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			response.Fail(c, "创建目录失败")
			return
		}
		filename := fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(file.Filename))
		dst := filepath.Join(saveDir, filename)
		src, err := file.Open()
		if err != nil {
			response.Fail(c, "上传失败")
			return
		}
		defer src.Close()
		out, err := os.Create(dst)
		if err != nil {
			response.Fail(c, "上传失败")
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, src); err != nil {
			response.Fail(c, "上传失败")
			return
		}
		relPath := dateDir + "/" + filename
		absUpload, _ := filepath.Abs(uploadDir)
		relFile := "/uploads/" + relPath
		thumbRelFile := ""
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			thumbName := filename[:len(filename)-len(ext)] + "_thumb.jpg"
			thumbDst := filepath.Join(saveDir, thumbName)
			if err := exec.Command("ffmpeg", "-y", "-i", dst, "-vframes", "1", "-q:v", "2", thumbDst).Run(); err == nil {
				thumbRelFile = "/uploads/" + dateDir + "/" + thumbName
			}
		}
		response.JSON(c, utils.H{"url": relFile, "thumb": thumbRelFile, "path": filepath.Join(absUpload, relPath), "filename": filename})
	})
	absUpload, _ := filepath.Abs(uploadDir)
	h.GET("/uploads/*filepath", func(ctx context.Context, c *app.RequestContext) {
		c.File(filepath.Join(absUpload, c.Param("filepath")))
	})

	h.Spin()
}
