// WeCheckin API
//
//	@title			WeCheckin API
//	@version		2.0
//	@description	微信小程序打卡项目后端 API。包含用户认证、打卡管理、问卷系统、考试系统、报表导出等功能。
//	@host			localhost:8083
//	@BasePath		/
//	@schemes		http
// @tag.name		PC端-用户管理
// @tag.description	后台管理用户相关接口
// @tag.name		PC端-通知公告
// @tag.description	后台管理通知公告相关接口
// @tag.name		PC端-赛事活动管理
// @tag.description	后台管理赛事活动相关接口
// @tag.name		PC端-打卡管理
// @tag.description	后台管理打卡相关接口
// @tag.name		PC端-菜单管理
// @tag.description	后台管理菜单相关接口
// @tag.name		PC端-角色管理
// @tag.description	后台管理角色相关接口
// @tag.name		PC端-字典管理
// @tag.description	后台管理字典相关接口
// @tag.name		PC端-部门管理
// @tag.description	后台管理部门相关接口
// @tag.name		PC端-管理员管理
// @tag.description	后台管理管理员相关接口
// @tag.name		PC端-系统设置
// @tag.description	后台系统设置相关接口
// @tag.name		PC端-管理后台首页
// @tag.description	后台首页数据接口
// @tag.name		PC端-考试管理
// @tag.description	后台管理考试相关接口
// @tag.name		PC端-问卷管理
// @tag.description	后台管理问卷相关接口
// @tag.name		PC端-表单工具
// @tag.description	后台表单工具相关接口
// @tag.name		PC端-在线用户
// @tag.description	在线用户管理接口
// @tag.name		PC端-在线管理员
// @tag.description	在线管理员管理接口
// @tag.name		客户端-通行证
// @tag.description	客户端用户认证相关接口
// @tag.name		客户端-打卡
// @tag.description	客户端打卡相关接口
// @tag.name		客户端-赛事活动
// @tag.description	客户端赛事活动相关接口
// @tag.name		客户端-考试
// @tag.description	客户端考试相关接口
// @tag.name		客户端-问卷
// @tag.description	客户端问卷相关接口
// @tag.name		客户端-地理编码
// @tag.description	客户端地理编码相关接口
// @tag.name		客户端-通知公告
// @tag.description	客户端通知公告相关接口
// @tag.name		客户端-首页
// @tag.description	客户端首页数据接口
// @tag.name		客户端-收藏
// @tag.description	客户端收藏相关接口
// @tag.name		客户端-表单工具
// @tag.description	客户端表单工具相关接口
//
//	@securityDefinitions.apikey	AdminToken
//	@in							header
//	@name						Authorization
//	@description				管理员 Token，格式: "Bearer {token}"
//
//	@securityDefinitions.apikey	ClientToken
//	@in							header
//	@name						Authorization
//	@description				用户 Token，格式: "Bearer {token}"

//go:generate sh -c "cd .. && swag init -g cmd/main.go --output docs/swagger"
package main

import (
	"flag"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	_ "wecheckin-backend/backend/docs/swagger"
	"wecheckin-backend/backend/internal/bootstrap"
	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/internal/middleware"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	rd "wecheckin-backend/backend/pkg/redis"
)

func main() {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	examFlag := flag.Bool("exam", false, "启用在线考试菜单")
	flag.Parse()

	cfg, err := config.LoadConfig(*env)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database.InitDatabase(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	if err := bootstrap.InitBusiness(*examFlag); err != nil {
		log.Fatalf("Failed to initialize business data: %v", err)
	}

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

	registerRoutes(h)
	h.Spin()
}
