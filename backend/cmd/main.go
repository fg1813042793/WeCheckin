// WeCheckin API
//
//	@title			WeCheckin API
//	@version		2.0
//	@description	微信小程序打卡项目后端 API。包含用户认证、打卡管理、问卷系统、考试系统、报表导出等功能。
//	@host			localhost:8083
//	@BasePath		/
//	@schemes		http
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
//
//	@securityDefinitions.apikey	H5AppToken
//	@in							header
//	@name						Authorization
//	@description				H5App Token，格式: "Bearer {token}"

//go:generate sh -c "cd .. && swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger"
package main

import (
	"flag"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	_ "wecheckin/backend/docs/swagger"
	"wecheckin/backend/internal/config"
	"wecheckin/backend/internal/middleware"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
	rd "wecheckin/backend/pkg/redis"
)

func main() {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	_ = flag.Bool("exam", false, "兼容旧启动参数；初始化菜单请使用 init.sh -exam")
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

	registerRoutes(h)
	h.Spin()
}
