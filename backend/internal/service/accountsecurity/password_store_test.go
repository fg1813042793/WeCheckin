package accountsecurity

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/xwb1989/sqlparser"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPasswordHistoryCleanupUsesValidMySQLPagination(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Update().After("gorm:update").Register("password_test:affected", func(tx *gorm.DB) { tx.RowsAffected = 1 }); err != nil {
		t.Fatal(err)
	}
	var cleanupQueries int
	if err := db.Callback().Query().After("gorm:query").Register("password_test:mysql_syntax", func(tx *gorm.DB) {
		cleanupQueries++
		statement := tx.Statement.SQL.String()
		parsed, err := sqlparser.Parse(statement)
		if err != nil {
			tx.AddError(fmt.Errorf("password history cleanup generated invalid MySQL: %s: %w", statement, err))
			return
		}
		query := parsed.(*sqlparser.Select)
		if query.Limit == nil || query.Limit.Offset == nil || query.Limit.Rowcount == nil {
			t.Errorf("history cleanup must have LIMIT and OFFSET: %s", statement)
		}
		if len(tx.Statement.Vars) != 3 || tx.Statement.Vars[0] != uint(7) || tx.Statement.Vars[2] != passwordHistoryRetention {
			t.Errorf("cleanup must retain 20 histories for the target user; args=%v", tx.Statement.Vars)
		}
	}); err != nil {
		t.Fatal(err)
	}
	err = NewGormPasswordStore(db).ApplyPasswordUpdate(context.Background(), PasswordUpdate{
		UserID: 7, NewHash: "test-hash", PreviousHash: "test-previous-hash", ChangedAt: 1000, MustChange: true, ChangeSource: PasswordChangeSourceAdminReset,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cleanupQueries != 1 {
		t.Fatalf("history cleanup queries=%d", cleanupQueries)
	}
}
