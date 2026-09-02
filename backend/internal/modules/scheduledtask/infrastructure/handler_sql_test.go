package infrastructure

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestSQLHandlerRequiresGlobalSwitchAndRegisteredDataSource(t *testing.T) {
	disabled := NewSQLHandler(SQLHandlerPolicy{Enabled: false}, nil)
	if err := disabled.ValidateConfig(context.Background(), []byte(`{"dataSourceKey":"main","mode":"read","statement":"SELECT 1"}`)); !errors.Is(err, ErrSQLHandlerDisabled) {
		t.Fatalf("disabled error = %v", err)
	}
	enabled := NewSQLHandler(SQLHandlerPolicy{Enabled: true, DataSources: map[string]SQLDataSource{"main": {Executor: &fakeSQLExecutor{}}}}, nil)
	if err := enabled.ValidateConfig(context.Background(), []byte(`{"dataSourceKey":"missing","mode":"read","statement":"SELECT 1"}`)); !errors.Is(err, ErrSQLDataSourceNotFound) {
		t.Fatalf("missing datasource error = %v", err)
	}
}

func TestSQLHandlerUsesASTToEnforceReadAndWriteStatements(t *testing.T) {
	executor := &fakeSQLExecutor{queryResult: SQLQueryResult{Rows: 2, Summary: "2 rows"}, affected: 1}
	handler := NewSQLHandler(SQLHandlerPolicy{
		Enabled: true, DefaultMaxRows: 100, DefaultMaxAffected: 10,
		DataSources: map[string]SQLDataSource{"main": {Executor: executor}},
	}, nil)

	for _, raw := range []string{
		`{"dataSourceKey":"main","mode":"read","statement":"UPDATE users SET status = ? WHERE id = ?","parameters":[1,7]}`,
		`{"dataSourceKey":"main","mode":"write","statement":"CREATE TABLE unsafe(id int)"}`,
		`{"dataSourceKey":"main","mode":"write","statement":"DELETE FROM users; DELETE FROM roles"}`,
	} {
		if err := handler.ValidateConfig(context.Background(), []byte(raw)); err == nil {
			t.Fatalf("unsafe SQL must fail: %s", raw)
		}
	}
	if err := handler.ValidateConfig(context.Background(), []byte(`{"dataSourceKey":"main","mode":"read","statement":"SELECT * FROM users WHERE id = ?","parameters":[]}`)); err == nil {
		t.Fatal("placeholder count mismatch must fail")
	}

	read, err := handler.Execute(context.Background(), application.RunContext{Task: application.TaskSnapshot{HandlerConfigJSON: `{
		"dataSourceKey":"main","mode":"read","statement":"SELECT id FROM users WHERE status = ?","parameters":[1],"maxRows":5
	}`}})
	if err != nil || read.Summary != "2 rows" || executor.queryMaxRows != 5 {
		t.Fatalf("read result/executor = %#v / %#v, err = %v", read, executor, err)
	}
	write, err := handler.Execute(context.Background(), application.RunContext{Task: application.TaskSnapshot{HandlerConfigJSON: `{
		"dataSourceKey":"main","mode":"write","statement":"UPDATE users SET status = ? WHERE id = ?","parameters":[1,7],"maxAffected":3
	}`}})
	if err != nil || write.Data["affectedRows"] != int64(1) || executor.execMaxAffected != 3 {
		t.Fatalf("write result/executor = %#v / %#v, err = %v", write, executor, err)
	}
}

func TestSQLHandlerCallsOnlyRegisteredProcedures(t *testing.T) {
	executor := &fakeSQLExecutor{affected: 4}
	handler := NewSQLHandler(SQLHandlerPolicy{
		Enabled: true,
		DataSources: map[string]SQLDataSource{"main": {
			Executor: executor, Procedures: map[string]string{"refresh-summary": "reporting.refresh_summary"},
		}},
	}, nil)
	if err := handler.ValidateConfig(context.Background(), []byte(`{"dataSourceKey":"main","mode":"write","procedureKey":"missing"}`)); !errors.Is(err, ErrSQLProcedureNotFound) {
		t.Fatalf("missing procedure error = %v", err)
	}
	result, err := handler.Execute(context.Background(), application.RunContext{Task: application.TaskSnapshot{HandlerConfigJSON: `{
		"dataSourceKey":"main","mode":"write","procedureKey":"refresh-summary","parameters":["2026-09"]
	}`}})
	if err != nil {
		t.Fatal(err)
	}
	if executor.callStatement != "CALL `reporting`.`refresh_summary`(?)" || result.Data["affectedRows"] != int64(4) {
		t.Fatalf("call/result = %q / %#v", executor.callStatement, result)
	}
}

type fakeSQLExecutor struct {
	queryResult     SQLQueryResult
	affected        int64
	queryMaxRows    int
	execMaxAffected int64
	callStatement   string
}

func (executor *fakeSQLExecutor) Query(_ context.Context, _ string, _ []interface{}, maxRows int) (SQLQueryResult, error) {
	executor.queryMaxRows = maxRows
	return executor.queryResult, nil
}
func (executor *fakeSQLExecutor) Exec(_ context.Context, _ string, _ []interface{}, maxAffected int64) (int64, error) {
	executor.execMaxAffected = maxAffected
	return executor.affected, nil
}
func (executor *fakeSQLExecutor) Call(_ context.Context, statement string, _ []interface{}, maxAffected int64) (int64, error) {
	executor.callStatement = statement
	executor.execMaxAffected = maxAffected
	return executor.affected, nil
}
