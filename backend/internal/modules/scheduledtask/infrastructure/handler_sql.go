package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

var (
	ErrSQLHandlerDisabled       = errors.New("scheduled task SQL handler is disabled")
	ErrSQLDataSourceNotFound    = errors.New("registered SQL data source not found")
	ErrSQLProcedureNotFound     = errors.New("registered SQL procedure not found")
	ErrSQLRowLimitExceeded      = errors.New("SQL row limit exceeded")
	ErrSQLAffectedLimitExceeded = errors.New("SQL affected row limit exceeded")
)

var procedureNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+(?:\.[A-Za-z0-9_]+)*$`)

type SQLHandlerPolicy struct {
	Enabled            bool
	DefaultMaxRows     int
	DefaultMaxAffected int64
	DataSources        map[string]SQLDataSource
}

type SQLDataSource struct {
	Executor   SQLExecutor
	ReadOnly   bool
	Procedures map[string]string
}

type SQLExecutor interface {
	Query(context.Context, string, []interface{}, int) (SQLQueryResult, error)
	Exec(context.Context, string, []interface{}, int64) (int64, error)
	Call(context.Context, string, []interface{}, int64) (int64, error)
}

type SQLQueryResult struct {
	Rows    int
	Summary string
}

type SQLHandler struct {
	policy SQLHandlerPolicy
}

type sqlHandlerConfig struct {
	DataSourceKey string        `json:"dataSourceKey"`
	Mode          string        `json:"mode"`
	Statement     string        `json:"statement"`
	ProcedureKey  string        `json:"procedureKey"`
	Parameters    []interface{} `json:"parameters"`
	MaxRows       int           `json:"maxRows"`
	MaxAffected   int64         `json:"maxAffected"`
}

type resolvedSQLConfig struct {
	config     sqlHandlerConfig
	dataSource SQLDataSource
	statement  string
	procedure  bool
}

func NewSQLHandler(policy SQLHandlerPolicy, _ interface{}) *SQLHandler {
	if policy.DefaultMaxRows <= 0 {
		policy.DefaultMaxRows = 1000
	}
	if policy.DefaultMaxAffected <= 0 {
		policy.DefaultMaxAffected = 1000
	}
	return &SQLHandler{policy: policy}
}

func (handler *SQLHandler) Type() string { return "sql" }

func (handler *SQLHandler) Metadata() application.HandlerMetadata {
	return application.HandlerMetadata{
		Type: "sql", Name: "Controlled SQL", Description: "Executes one AST-validated statement or a registered procedure",
		RiskLevel: "critical", ConfigSchema: json.RawMessage(`{
			"type":"object","required":["dataSourceKey","mode"],
			"properties":{
				"dataSourceKey":{"type":"string"},"mode":{"type":"string","enum":["read","write"]},
				"statement":{"type":"string"},"procedureKey":{"type":"string"},
				"parameters":{"type":"array"},"maxRows":{"type":"integer","minimum":1},
				"maxAffected":{"type":"integer","minimum":1}
			}
		}`),
	}
}

func (handler *SQLHandler) ValidateConfig(_ context.Context, raw json.RawMessage) error {
	_, err := handler.resolve(raw)
	return err
}

func (handler *SQLHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	resolved, err := handler.resolve(json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	if resolved.procedure {
		affected, err := resolved.dataSource.Executor.Call(ctx, resolved.statement, resolved.config.Parameters, resolved.config.MaxAffected)
		if err != nil {
			return application.HandlerResult{}, sqlExecutionError(err)
		}
		return affectedResult(affected), nil
	}
	if resolved.config.Mode == "read" {
		result, err := resolved.dataSource.Executor.Query(ctx, resolved.statement, resolved.config.Parameters, resolved.config.MaxRows)
		if err != nil {
			return application.HandlerResult{}, sqlExecutionError(err)
		}
		return application.HandlerResult{Summary: result.Summary, Data: map[string]interface{}{"rows": result.Rows}}, nil
	}
	affected, err := resolved.dataSource.Executor.Exec(ctx, resolved.statement, resolved.config.Parameters, resolved.config.MaxAffected)
	if err != nil {
		return application.HandlerResult{}, sqlExecutionError(err)
	}
	return affectedResult(affected), nil
}

func (handler *SQLHandler) resolve(raw json.RawMessage) (resolvedSQLConfig, error) {
	if handler == nil || !handler.policy.Enabled {
		return resolvedSQLConfig{}, ErrSQLHandlerDisabled
	}
	var config sqlHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return resolvedSQLConfig{}, fmt.Errorf("decode SQL task config: %w", err)
	}
	config.DataSourceKey = strings.TrimSpace(config.DataSourceKey)
	config.Mode = strings.TrimSpace(config.Mode)
	config.Statement = strings.TrimSpace(config.Statement)
	config.ProcedureKey = strings.TrimSpace(config.ProcedureKey)
	dataSource, ok := handler.policy.DataSources[config.DataSourceKey]
	if !ok || dataSource.Executor == nil {
		return resolvedSQLConfig{}, fmt.Errorf("%w: %s", ErrSQLDataSourceNotFound, config.DataSourceKey)
	}
	if config.Mode != "read" && config.Mode != "write" {
		return resolvedSQLConfig{}, errors.New("SQL mode must be read or write")
	}
	if dataSource.ReadOnly && config.Mode != "read" {
		return resolvedSQLConfig{}, errors.New("SQL data source is read-only")
	}
	if config.MaxRows <= 0 {
		config.MaxRows = handler.policy.DefaultMaxRows
	}
	if config.MaxAffected <= 0 {
		config.MaxAffected = handler.policy.DefaultMaxAffected
	}
	resolved := resolvedSQLConfig{config: config, dataSource: dataSource}
	if config.ProcedureKey != "" {
		if config.Mode != "write" || config.Statement != "" {
			return resolvedSQLConfig{}, errors.New("registered SQL procedures require write mode and no statement")
		}
		procedureName, ok := dataSource.Procedures[config.ProcedureKey]
		if !ok {
			return resolvedSQLConfig{}, fmt.Errorf("%w: %s", ErrSQLProcedureNotFound, config.ProcedureKey)
		}
		if !procedureNamePattern.MatchString(procedureName) {
			return resolvedSQLConfig{}, errors.New("registered SQL procedure name is invalid")
		}
		resolved.statement = buildCallStatement(procedureName, len(config.Parameters))
		resolved.procedure = true
		return resolved, nil
	}
	if config.Statement == "" {
		return resolvedSQLConfig{}, errors.New("SQL statement is required")
	}
	statement, err := sqlparser.Parse(config.Statement)
	if err != nil {
		return resolvedSQLConfig{}, fmt.Errorf("parse SQL statement: %w", err)
	}
	switch config.Mode {
	case "read":
		if _, ok := statement.(*sqlparser.Select); !ok {
			return resolvedSQLConfig{}, errors.New("SQL read mode only allows SELECT")
		}
	case "write":
		switch statement.(type) {
		case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		default:
			return resolvedSQLConfig{}, errors.New("SQL write mode only allows INSERT, UPDATE, or DELETE")
		}
	}
	bindings := 0
	if err := sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
		if value, ok := node.(*sqlparser.SQLVal); ok && value.Type == sqlparser.ValArg {
			bindings++
		}
		return true, nil
	}, statement); err != nil {
		return resolvedSQLConfig{}, err
	}
	if bindings != len(config.Parameters) {
		return resolvedSQLConfig{}, fmt.Errorf("SQL parameter count %d does not match %d placeholders", len(config.Parameters), bindings)
	}
	resolved.statement = config.Statement
	return resolved, nil
}

func buildCallStatement(procedureName string, parameterCount int) string {
	parts := strings.Split(procedureName, ".")
	for index, part := range parts {
		parts[index] = "`" + part + "`"
	}
	placeholders := make([]string, parameterCount)
	for index := range placeholders {
		placeholders[index] = "?"
	}
	return "CALL " + strings.Join(parts, ".") + "(" + strings.Join(placeholders, ",") + ")"
}

func affectedResult(affected int64) application.HandlerResult {
	return application.HandlerResult{
		Summary: fmt.Sprintf("%d rows affected", affected),
		Data:    map[string]interface{}{"affectedRows": affected},
	}
}

func sqlExecutionError(err error) error {
	temporary := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
	return &application.HandlerError{Code: "sql_failed", Summary: err.Error(), Temporary: temporary}
}

type DBSQLExecutor struct {
	DB *sql.DB
}

func (executor DBSQLExecutor) Query(ctx context.Context, statement string, parameters []interface{}, maxRows int) (SQLQueryResult, error) {
	rows, err := executor.DB.QueryContext(ctx, statement, parameters...)
	if err != nil {
		return SQLQueryResult{}, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return SQLQueryResult{}, err
	}
	count := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		destinations := make([]interface{}, len(columns))
		for index := range values {
			destinations[index] = &values[index]
		}
		if err := rows.Scan(destinations...); err != nil {
			return SQLQueryResult{}, err
		}
		count++
		if count > maxRows {
			return SQLQueryResult{}, ErrSQLRowLimitExceeded
		}
	}
	if err := rows.Err(); err != nil {
		return SQLQueryResult{}, err
	}
	return SQLQueryResult{Rows: count, Summary: fmt.Sprintf("%d rows", count)}, nil
}

func (executor DBSQLExecutor) Exec(ctx context.Context, statement string, parameters []interface{}, maxAffected int64) (int64, error) {
	return executor.execTransaction(ctx, statement, parameters, maxAffected)
}

func (executor DBSQLExecutor) Call(ctx context.Context, statement string, parameters []interface{}, maxAffected int64) (int64, error) {
	return executor.execTransaction(ctx, statement, parameters, maxAffected)
}

func (executor DBSQLExecutor) execTransaction(ctx context.Context, statement string, parameters []interface{}, maxAffected int64) (int64, error) {
	transaction, err := executor.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	result, err := transaction.ExecContext(ctx, statement, parameters...)
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	if affected > maxAffected {
		_ = transaction.Rollback()
		return 0, ErrSQLAffectedLimitExceeded
	}
	if err := transaction.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}

var _ application.Handler = (*SQLHandler)(nil)
