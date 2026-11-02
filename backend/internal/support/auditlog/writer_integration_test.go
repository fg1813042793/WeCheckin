package auditlog

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/randutil"
)

// Opt-in test against a disposable database, never the application's configured DB.
func TestAuditWriterMySQLFinalizesOneEvent(t *testing.T) {
	dsn := os.Getenv("WECHECKIN_AUDIT_TEST_DSN")
	if dsn == "" {
		t.Skip("set disposable MySQL DSN to verify migration and writer")
	}
	cfg, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBName != "audit_verify" || cfg.Addr != "127.0.0.1:18306" {
		t.Fatal("integration test requires isolated audit_verify on localhost:18306")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	old := database.DB
	database.DB = db
	defer func() { database.DB = old }()
	event := &model.Log{EventID: "integration-" + randutil.HexString(8), Category: "login", Source: "admin", Module: "auth", Action: "password_login", Outcome: "unknown", ActorType: "anonymous", AddTime: database.Now(), Changes: json.RawMessage(`[]`), SnapshotStatus: SnapshotPending}
	defer db.Where("log_event_id = ?", event.EventID).Delete(&model.Log{})
	w := DatabaseWriter{}
	if err := w.Write(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	event.Outcome = "failure"
	event.ReasonCode = "invalid_credentials"
	event.SnapshotStatus, event.SnapshotReason = SnapshotOperationFailed, "business_not_successful"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.Write(ctx, event); err != nil {
		t.Fatal(err)
	}
	var rows []model.Log
	if err := db.Where("log_event_id = ?", event.EventID).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Outcome != "failure" || rows[0].ReasonCode != "invalid_credentials" {
		t.Fatalf("stored event=%#v", rows)
	}
	if string(rows[0].Changes) != "[]" {
		t.Fatalf("changes=%s", rows[0].Changes)
	}
	if rows[0].SnapshotStatus != SnapshotOperationFailed || rows[0].SnapshotReason != "business_not_successful" {
		t.Fatalf("snapshot metadata not finalized: %s / %s", rows[0].SnapshotStatus, rows[0].SnapshotReason)
	}
}
