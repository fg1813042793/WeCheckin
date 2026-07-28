package main

import (
	"flag"
	"log"

	"wecheckin-backend/backend/internal/bootstrap"
	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/pkg/database"
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

	database.InitDatabase(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	if err := bootstrap.RunMaintenance(bootstrap.MaintenanceOptions{
		EnableExam:       *enableExam,
		MigrationsDir:    *migrationsDir,
		RunSQLMigrations: !*skipSQL,
	}); err != nil {
		log.Fatalf("Failed to run maintenance tasks: %v", err)
	}
	log.Println("Maintenance tasks completed")
}
