package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

const maxLoggedSQLValueBytes = 1024

type filteredDatabaseLogger struct {
	logger.Interface
}

func (l filteredDatabaseLogger) ParamsFilter(ctx context.Context, query string, params ...interface{}) (string, []interface{}) {
	filtered := append([]interface{}(nil), params...)
	for index, param := range filtered {
		switch value := param.(type) {
		case string:
			if len(value) > maxLoggedSQLValueBytes {
				filtered[index] = fmt.Sprintf("<large string: %d bytes>", len(value))
			}
		case []byte:
			if len(value) > maxLoggedSQLValueBytes {
				filtered[index] = fmt.Sprintf("<large bytes: %d bytes>", len(value))
			}
		}
	}
	return query, filtered
}

func InitDatabase(host string, port int, user, password, dbname string) {
	if err := ConnectDatabase(host, port, user, password, dbname); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func ConnectDatabase(host string, port int, user, password, dbname string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newDatabaseLogger(),
	})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("open database connection pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func newDatabaseLogger() logger.Interface {
	base := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
	return filteredDatabaseLogger{Interface: base}
}

func Now() int64 {
	return time.Now().UnixMilli()
}

func GetDB() *gorm.DB {
	return DB
}
