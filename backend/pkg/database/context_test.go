package database

import (
	"context"
	"testing"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

func TestQueryContextAddsDefaultTimeout(t *testing.T) {
	ctx, cancel := QueryContext(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("QueryContext should add a default deadline")
	}
	if time.Until(deadline) <= 0 || time.Until(deadline) > DefaultQueryTimeout {
		t.Fatalf("deadline should be within DefaultQueryTimeout, got %v", time.Until(deadline))
	}
}

func TestQueryContextKeepsParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	ctx, cancel := QueryContext(parent)
	defer cancel()
	cancelParent()
	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("child query context should be canceled with parent")
	}
}

func TestDatabaseLoggerHidesAllParameterValues(t *testing.T) {
	filter, ok := newDatabaseLogger(Options{LogLevel: gormlogger.Info}, nil).(interface {
		ParamsFilter(context.Context, string, ...interface{}) (string, []interface{})
	})
	if !ok {
		t.Fatal("database logger should support GORM parameter filtering")
	}
	query := "UPDATE workflow_definitions SET definition_draft_json = ? WHERE id = ?"
	filteredQuery, params := filter.ParamsFilter(context.Background(), query, "short-secret", 1)
	if filteredQuery != query || len(params) != 0 {
		t.Fatalf("SQL shape should be preserved, query=%q params=%#v", filteredQuery, params)
	}
}
