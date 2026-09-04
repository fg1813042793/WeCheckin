package database

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

type Options struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	ConnectTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogLevel        gormLogger.LogLevel
	Colorful        bool
}

func InitDatabase(host string, port int, user, password, dbname string) {
	if err := ConnectDatabase(host, port, user, password, dbname); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func ConnectDatabase(host string, port int, user, password, dbname string) error {
	return ConnectDatabaseWithOptions(defaultOptions(host, port, user, password, dbname))
}

func ConnectDatabaseWithOptions(options Options) error {
	options = normalizeOptions(options)
	dsn := mysqlConfig(options).FormatDSN()
	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: newDatabaseLogger(options, nil),
	})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("open database connection pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(options.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(options.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(options.ConnMaxIdleTime)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func defaultOptions(host string, port int, user, password, dbname string) Options {
	return normalizeOptions(Options{
		Host: host, Port: port, User: user, Password: password, DBName: dbname,
		LogLevel: gormLogger.Info,
		Colorful: true,
	})
}

func normalizeOptions(options Options) Options {
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = 10 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 30 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 30 * time.Second
	}
	if options.MaxIdleConns <= 0 {
		options.MaxIdleConns = 10
	}
	if options.MaxOpenConns <= 0 {
		options.MaxOpenConns = 100
	}
	if options.ConnMaxLifetime <= 0 {
		options.ConnMaxLifetime = time.Hour
	}
	if options.ConnMaxIdleTime <= 0 {
		options.ConnMaxIdleTime = 10 * time.Minute
	}
	if options.LogLevel == 0 {
		options.LogLevel = gormLogger.Warn
	}
	return options
}

func mysqlConfig(options Options) *mysqlDriver.Config {
	config := mysqlDriver.NewConfig()
	config.User = options.User
	config.Passwd = options.Password
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(options.Host, fmt.Sprintf("%d", options.Port))
	config.DBName = options.DBName
	config.Params = map[string]string{"charset": "utf8mb4"}
	config.ParseTime = true
	config.Loc = time.Local
	config.Timeout = options.ConnectTimeout
	config.ReadTimeout = options.ReadTimeout
	config.WriteTimeout = options.WriteTimeout
	return config
}

func newDatabaseLogger(options Options, writer gormLogger.Writer) gormLogger.Interface {
	if writer == nil {
		writer = log.New(os.Stdout, "\r\n", log.LstdFlags)
	}
	return gormLogger.New(writer, gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  options.LogLevel,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      true,
		Colorful:                  options.Colorful,
	})
}

func Now() int64 {
	return time.Now().UnixMilli()
}

func GetDB() *gorm.DB {
	return DB
}
