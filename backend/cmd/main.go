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
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	gormlogger "gorm.io/gorm/logger"
	_ "wecheckin/backend/docs/swagger"
	"wecheckin/backend/internal/config"
	"wecheckin/backend/internal/middleware"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	notificationoutboxinfra "wecheckin/backend/internal/modules/notificationoutbox/infrastructure"
	poststatservice "wecheckin/backend/internal/service/client/poststat"
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

	databaseLogLevel := gormlogger.Warn
	databaseLogColorful := false
	if cfg.Server.Mode == "debug" {
		databaseLogLevel = gormlogger.Info
		databaseLogColorful = true
	}
	if err := database.ConnectDatabaseWithOptions(database.Options{
		Host: cfg.Database.Host, Port: cfg.Database.Port,
		User: cfg.Database.User, Password: cfg.Database.Password, DBName: cfg.Database.DBName,
		ConnectTimeout: time.Duration(cfg.Database.ConnectTimeoutSec) * time.Second,
		ReadTimeout:    time.Duration(cfg.Database.ReadTimeoutSec) * time.Second,
		WriteTimeout:   time.Duration(cfg.Database.WriteTimeoutSec) * time.Second,
		MaxIdleConns:   cfg.Database.MaxIdleConns, MaxOpenConns: cfg.Database.MaxOpenConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetimeMin) * time.Minute,
		ConnMaxIdleTime: time.Duration(cfg.Database.ConnMaxIdleTimeMin) * time.Minute,
		LogLevel:        databaseLogLevel, Colorful: databaseLogColorful,
	}); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := logger.Init(cfg.Log.Dir, cfg.Log.Level, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		logger.Logger.Printf("Warning: logger init: %v", err)
	}

	if err := rd.Init(cfg.Redis); err != nil {
		logger.Logger.Printf("Warning: Redis init failed: %v", err)
	} else {
		logger.Logger.Println("Redis connected")
	}

	outboxService := notificationoutboxapp.NewService(notificationoutboxinfra.NewGormStore(database.GetDB()))
	poststatservice.ConfigureNotificationDispatcher(poststatservice.NewOutboxDispatcher(outboxService))

	h := server.Default(
		server.WithHostPorts(net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)),
		server.WithReadTimeout(time.Duration(cfg.Server.ReadTimeoutSec)*time.Second),
		server.WithWriteTimeout(time.Duration(cfg.Server.WriteTimeoutSec)*time.Second),
		server.WithIdleTimeout(time.Duration(cfg.Server.IdleTimeoutSec)*time.Second),
		server.WithMaxRequestBodySize(32*1024*1024),
	)

	h.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           time.Hour,
	}))

	h.Use(middleware.AccessLog())

	registerRoutes(h)
	h.Spin()
}
