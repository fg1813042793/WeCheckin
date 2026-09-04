package database

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	gormlogger "gorm.io/gorm/logger"
)

func TestMySQLConfigPreservesCredentialsAndTimeouts(t *testing.T) {
	options := Options{
		Host: "db.internal", Port: 3307, User: "worker",
		Password: "p@ss:/?#word", DBName: "wecheckin",
		ConnectTimeout: 7 * time.Second,
		ReadTimeout:    11 * time.Second,
		WriteTimeout:   13 * time.Second,
	}

	config := mysqlConfig(options)
	parsed, err := mysqlDriver.ParseDSN(config.FormatDSN())
	if err != nil {
		t.Fatalf("parse generated DSN: %v", err)
	}
	if parsed.User != options.User || parsed.Passwd != options.Password || parsed.DBName != options.DBName {
		t.Fatalf("generated DSN lost credentials or database name: %#v", parsed)
	}
	if parsed.Addr != "db.internal:3307" {
		t.Fatalf("database address = %q", parsed.Addr)
	}
	if parsed.Timeout != options.ConnectTimeout || parsed.ReadTimeout != options.ReadTimeout || parsed.WriteTimeout != options.WriteTimeout {
		t.Fatalf("database timeouts = connect:%s read:%s write:%s", parsed.Timeout, parsed.ReadTimeout, parsed.WriteTimeout)
	}
	if !parsed.ParseTime || parsed.Params["charset"] != "utf8mb4" {
		t.Fatalf("database protocol options = %#v", parsed)
	}
}

func TestDatabaseLoggerAlwaysUsesParameterizedQueries(t *testing.T) {
	writer := &captureWriter{}
	sqlLogger := newDatabaseLogger(Options{LogLevel: gormlogger.Warn, Colorful: false}, writer)
	filter, ok := sqlLogger.(interface {
		ParamsFilter(context.Context, string, ...interface{}) (string, []interface{})
	})
	if !ok {
		t.Fatal("database logger must expose GORM ParamsFilter")
	}
	query, params := filter.ParamsFilter(context.Background(), "SELECT * FROM users WHERE token = ?", "top-secret")
	if query == "" || len(params) != 0 {
		t.Fatalf("parameterized query filter returned query=%q params=%#v", query, params)
	}

	sqlLogger.Info(context.Background(), "hidden info")
	if writer.String() != "" {
		t.Fatalf("warn logger emitted info output: %q", writer.String())
	}
	sqlLogger.Warn(context.Background(), "slow query")
	if strings.Contains(writer.String(), "\x1b[") {
		t.Fatalf("non-colorful logger emitted ANSI escape sequences: %q", writer.String())
	}
}

type captureWriter struct {
	output strings.Builder
}

func (writer *captureWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&writer.output, format, args...)
}

func (writer *captureWriter) String() string {
	return writer.output.String()
}
