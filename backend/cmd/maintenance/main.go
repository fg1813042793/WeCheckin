package main

import (
	"flag"
	"log"
	"time"

	gormlogger "gorm.io/gorm/logger"
	"wecheckin/backend/internal/bootstrap"
	"wecheckin/backend/internal/config"
	"wecheckin/backend/pkg/database"
)

func main() {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	enableExam := flag.Bool("exam", false, "初始化在线考试相关权限")
	migrationsDir := flag.String("migrations", "migrations", "版本化 SQL 迁移目录")
	skipSQL := flag.Bool("skip-sql", false, "跳过 migrations 目录中的版本化 SQL")
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
	if err := bootstrap.RunMaintenance(bootstrap.MaintenanceOptions{
		EnableExam:       *enableExam,
		MigrationsDir:    *migrationsDir,
		RunSQLMigrations: !*skipSQL,
	}); err != nil {
		log.Fatalf("Failed to run maintenance tasks: %v", err)
	}
	log.Println("Maintenance tasks completed")
}
