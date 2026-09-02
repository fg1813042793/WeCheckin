package infrastructure

import (
	"database/sql"
	"fmt"

	"wecheckin/backend/internal/config"
	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func NewHandlerRegistry(cfg config.ScheduledTaskConfig, workflowStarter WorkflowStarter, jobs ...GoJob) (*application.HandlerRegistry, error) {
	registry := application.NewHandlerRegistry()

	goHandler := NewGoHandler()
	for _, job := range jobs {
		if err := goHandler.Register(job); err != nil {
			return nil, err
		}
	}
	if err := registry.Register(goHandler); err != nil {
		return nil, err
	}
	if err := registry.Register(NewWorkflowHandler(workflowStarter)); err != nil {
		return nil, err
	}
	httpHandler, err := NewHTTPHandler(HTTPHandlerPolicy{
		AllowedHosts: cfg.HTTP.AllowedHosts, AllowedCIDRs: cfg.HTTP.AllowedCIDRs,
		AllowPrivateNetworks: cfg.HTTP.AllowPrivateNetworks, MaxRedirects: cfg.HTTP.MaxRedirects,
		MaxRequestBytes: cfg.HTTP.MaxRequestBytes, MaxResponseBytes: cfg.HTTP.MaxResponseBytes,
	}, StaticHTTPCredentials(cfg.HTTP.Credentials))
	if err != nil {
		return nil, err
	}
	if err := registry.Register(httpHandler); err != nil {
		return nil, err
	}
	shellCommands := make(map[string]ShellCommand, len(cfg.ShellCommands))
	for key, item := range cfg.ShellCommands {
		shellCommands[key] = ShellCommand{
			ExecutablePath: item.ExecutablePath, WorkingDir: item.WorkingDir,
			AllowedEnv: item.AllowedEnv, ArgumentPattern: item.ArgumentPattern, MaxArgs: item.MaxArgs,
		}
	}
	shellHandler, err := NewShellHandler(ShellHandlerPolicy{
		Enabled: cfg.EnableShell, MaxOutputBytes: int64(cfg.MaxLogRunBytes), Commands: shellCommands,
	}, nil)
	if err != nil {
		return nil, err
	}
	if err := registry.Register(shellHandler); err != nil {
		return nil, err
	}
	sqlDataSources, err := openSQLDataSources(cfg)
	if err != nil {
		return nil, err
	}
	if err := registry.Register(NewSQLHandler(SQLHandlerPolicy{
		Enabled: cfg.EnableSQL, DataSources: sqlDataSources,
	}, nil)); err != nil {
		return nil, err
	}
	return registry, nil
}

func openSQLDataSources(cfg config.ScheduledTaskConfig) (map[string]SQLDataSource, error) {
	result := make(map[string]SQLDataSource)
	if !cfg.EnableSQL {
		return result, nil
	}
	for key, item := range cfg.SQLDataSources {
		db, err := sql.Open(item.Driver, item.DSN)
		if err != nil {
			return nil, fmt.Errorf("open scheduled task SQL data source %q: %w", key, err)
		}
		if item.MaxOpenConnections > 0 {
			db.SetMaxOpenConns(item.MaxOpenConnections)
		}
		result[key] = SQLDataSource{
			Executor: DBSQLExecutor{DB: db}, ReadOnly: item.ReadOnly, Procedures: item.Procedures,
		}
	}
	return result, nil
}
