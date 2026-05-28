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
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "wecheckin-backend/backend/docs/swagger"
	"wecheckin-backend/backend/internal/api"
	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/middleware"
	"wecheckin-backend/backend/pkg/response"
)

func main() {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	flag.Parse()

	cfg, err := config.LoadConfig(*env)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database.InitDatabase(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)

	h := server.Default(server.WithHostPorts(":" + cfg.Server.Port))

	h.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: true,
		MaxAge:           time.Hour,
	}))

	url := swagger.URL("http://localhost:" + cfg.Server.Port + "/swagger/doc.json")
	h.GET("/swagger", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/swagger/index.html"))
	})
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))

	hm := api.NewHomeHandler()
	pp := api.NewPassportHandler()
	ns := api.NewNewsHandler()
	el := api.NewEnrollHandler()
	fa := api.NewFavHandler()
	aHome := api.NewAdminHomeHandler()
	aMgr := api.NewAdminMgrHandler()
	aSetup := api.NewAdminSetupHandler()
	aUser := api.NewAdminUserHandler()
	aNews := api.NewAdminNewsHandler()
	aEnroll := api.NewAdminEnrollHandler()

	// Public routes
	h.GET("/test/test", func(ctx context.Context, c *app.RequestContext) {
		response.JSON(c, map[string]string{"msg": "ok"})
	})

	h.GET("/home/setup_get", hm.GetSetup)
	h.GET("/home/list", hm.GetHomeList)

	h.POST("/passport/login", pp.Login)
	h.POST("/passport/phone", pp.GetPhone)
	h.GET("/passport/my_detail", pp.GetMyDetail)
	h.POST("/passport/register", pp.Register)
	h.POST("/passport/edit_base", pp.EditBase)

	h.POST("/fav/update", fa.UpdateFav)
	h.POST("/fav/del", fa.DelFav)
	h.GET("/fav/is_fav", fa.IsFav)
	h.GET("/fav/my_list", fa.GetMyFavList)

	h.GET("/news/list", ns.GetNewsList)
	h.GET("/news/view", ns.ViewNews)

	h.GET("/enroll/list", el.GetEnrollList)
	h.GET("/enroll/view", el.ViewEnroll)
	h.GET("/enroll/join_day", el.GetEnrollJoinByDay)
	h.POST("/enroll/join", el.EnrollJoin)
	h.GET("/enroll/my_join_list", el.GetMyEnrollJoinList)
	h.GET("/enroll/my_user_list", el.GetMyEnrollUserList)

	// Admin login (no auth required)
	h.POST("/admin/login", aMgr.AdminLogin)

	// Admin routes (with auth middleware)
	adminGroup := h.Group("/admin", middleware.AdminAuth())

	adminGroup.GET("/home", aHome.AdminHome)
	adminGroup.GET("/clear_vouch", aHome.ClearVouchData)
	adminGroup.GET("/mgr_list", aMgr.GetMgrList)
	adminGroup.POST("/mgr_insert", aMgr.InsertMgr)
	adminGroup.POST("/mgr_del", aMgr.DelMgr)
	adminGroup.GET("/mgr_detail", aMgr.GetMgrDetail)
	adminGroup.POST("/mgr_edit", aMgr.EditMgr)
	adminGroup.POST("/mgr_status", aMgr.StatusMgr)
	adminGroup.POST("/mgr_pwd", aMgr.PwdMgr)
	adminGroup.GET("/log_list", aMgr.GetLogList)
	adminGroup.GET("/log_clear", aMgr.ClearLog)

	adminGroup.POST("/setup_set", aSetup.SetSetup)
	adminGroup.POST("/setup_set_content", aSetup.SetContentSetup)
	adminGroup.GET("/setup_qr", aSetup.GenMiniQr)

	adminGroup.GET("/user_list", aUser.GetUserList)
	adminGroup.GET("/user_detail", aUser.GetUserDetail)
	adminGroup.POST("/user_del", aUser.DelUser)
	adminGroup.POST("/user_status", aUser.StatusUser)
	adminGroup.GET("/user_data_get", aUser.UserDataGet)
	adminGroup.GET("/user_data_export", aUser.UserDataExport)
	adminGroup.POST("/user_data_del", aUser.UserDataDel)

	adminGroup.GET("/news_list", aNews.GetAdminNewsList)
	adminGroup.POST("/news_insert", aNews.InsertNews)
	adminGroup.GET("/news_detail", aNews.GetNewsDetail)
	adminGroup.POST("/news_edit", aNews.EditNews)
	adminGroup.POST("/news_update_forms", aNews.UpdateNewsForms)
	adminGroup.POST("/news_update_pic", aNews.UpdateNewsPic)
	adminGroup.POST("/news_update_content", aNews.UpdateNewsContent)
	adminGroup.POST("/news_del", aNews.DelNews)
	adminGroup.POST("/news_sort", aNews.SortNews)
	adminGroup.POST("/news_status", aNews.StatusNews)

	adminGroup.GET("/enroll_list", aEnroll.GetAdminEnrollList)
	adminGroup.POST("/enroll_insert", aEnroll.InsertEnroll)
	adminGroup.GET("/enroll_detail", aEnroll.GetEnrollDetail)
	adminGroup.POST("/enroll_edit", aEnroll.EditEnroll)
	adminGroup.POST("/enroll_update_forms", aEnroll.UpdateEnrollForms)
	adminGroup.POST("/enroll_clear", aEnroll.ClearEnrollAll)
	adminGroup.POST("/enroll_del", aEnroll.DelEnroll)
	adminGroup.POST("/enroll_sort", aEnroll.SortEnroll)
	adminGroup.POST("/enroll_vouch", aEnroll.VouchEnroll)
	adminGroup.POST("/enroll_status", aEnroll.StatusEnroll)
	adminGroup.GET("/enroll_join_list", aEnroll.GetEnrollJoinList)
	adminGroup.POST("/enroll_join_del", aEnroll.DelEnrollJoin)
	adminGroup.GET("/enroll_join_data_get", aEnroll.EnrollJoinDataGet)
	adminGroup.GET("/enroll_join_data_export", aEnroll.EnrollJoinDataExport)
	adminGroup.POST("/enroll_join_data_del", aEnroll.EnrollJoinDataDel)

	h.Spin()
}
