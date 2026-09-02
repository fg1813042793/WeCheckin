package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

var (
	ErrShellHandlerDisabled = errors.New("scheduled task Shell handler is disabled")
	ErrShellCommandNotFound = errors.New("registered Shell command not found")
)

type ShellCommand struct {
	ExecutablePath  string
	WorkingDir      string
	AllowedEnv      []string
	ArgumentPattern string
	MaxArgs         int
}

type ShellHandlerPolicy struct {
	Enabled        bool
	MaxOutputBytes int64
	Commands       map[string]ShellCommand
}

type ShellRunner interface {
	Run(context.Context, string, []string, string, map[string]string, int64) (string, string, error)
}

type ShellHandler struct {
	policy   ShellHandlerPolicy
	runner   ShellRunner
	commands map[string]registeredShellCommand
}

type registeredShellCommand struct {
	config     ShellCommand
	argPattern *regexp.Regexp
	allowedEnv map[string]struct{}
}

type shellHandlerConfig struct {
	CommandKey string            `json:"commandKey"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
}

func NewShellHandler(policy ShellHandlerPolicy, runner ShellRunner) (*ShellHandler, error) {
	if policy.MaxOutputBytes <= 0 {
		policy.MaxOutputBytes = 1 << 20
	}
	if runner == nil {
		runner = ExecShellRunner{}
	}
	handler := &ShellHandler{policy: policy, runner: runner, commands: make(map[string]registeredShellCommand)}
	for key, command := range policy.Commands {
		key = strings.TrimSpace(key)
		if key == "" || !filepath.IsAbs(command.ExecutablePath) {
			return nil, fmt.Errorf("Shell command %q must use an absolute executable path", key)
		}
		if command.WorkingDir != "" && !filepath.IsAbs(command.WorkingDir) {
			return nil, fmt.Errorf("Shell command %q must use an absolute working directory", key)
		}
		registered := registeredShellCommand{config: command, allowedEnv: make(map[string]struct{})}
		if command.ArgumentPattern != "" {
			pattern, err := regexp.Compile(command.ArgumentPattern)
			if err != nil {
				return nil, fmt.Errorf("Shell command %q argument pattern: %w", key, err)
			}
			registered.argPattern = pattern
		}
		for _, envKey := range command.AllowedEnv {
			envKey = strings.TrimSpace(envKey)
			if envKey != "" {
				registered.allowedEnv[envKey] = struct{}{}
			}
		}
		handler.commands[key] = registered
	}
	return handler, nil
}

func (handler *ShellHandler) Type() string { return "shell" }

func (handler *ShellHandler) Metadata() application.HandlerMetadata {
	keys := make([]string, 0, len(handler.commands))
	for key := range handler.commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	schema, _ := json.Marshal(map[string]interface{}{
		"type": "object", "required": []string{"commandKey"},
		"properties": map[string]interface{}{
			"commandKey": map[string]interface{}{"type": "string", "enum": keys},
			"args":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"env":        map[string]interface{}{"type": "object"},
		},
	})
	return application.HandlerMetadata{Type: "shell", Name: "Controlled Shell", Description: "Executes a server-registered executable", RiskLevel: "critical", ConfigSchema: schema}
}

func (handler *ShellHandler) ValidateConfig(_ context.Context, raw json.RawMessage) error {
	_, _, err := handler.resolve(raw)
	return err
}

func (handler *ShellHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	config, command, err := handler.resolve(json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	stdout, stderr, err := handler.runner.Run(ctx, command.config.ExecutablePath, config.Args, command.config.WorkingDir, config.Env, handler.policy.MaxOutputBytes)
	summary := formatCommandOutput(stdout, stderr, handler.policy.MaxOutputBytes)
	if err != nil {
		temporary := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil
		return application.HandlerResult{}, &application.HandlerError{Code: "shell_failed", Summary: firstNonEmpty(summary, err.Error()), Temporary: temporary}
	}
	return application.HandlerResult{Summary: summary}, nil
}

func (handler *ShellHandler) resolve(raw json.RawMessage) (shellHandlerConfig, registeredShellCommand, error) {
	if handler == nil || !handler.policy.Enabled {
		return shellHandlerConfig{}, registeredShellCommand{}, ErrShellHandlerDisabled
	}
	var config shellHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, registeredShellCommand{}, fmt.Errorf("decode Shell task config: %w", err)
	}
	config.CommandKey = strings.TrimSpace(config.CommandKey)
	command, ok := handler.commands[config.CommandKey]
	if !ok {
		return config, command, fmt.Errorf("%w: %s", ErrShellCommandNotFound, config.CommandKey)
	}
	if command.config.MaxArgs >= 0 && len(config.Args) > command.config.MaxArgs {
		return config, command, errors.New("Shell argument count exceeds registered limit")
	}
	for _, argument := range config.Args {
		if command.argPattern == nil || !command.argPattern.MatchString(argument) {
			return config, command, fmt.Errorf("Shell argument %q does not match registered rule", argument)
		}
	}
	for key := range config.Env {
		if _, ok := command.allowedEnv[key]; !ok {
			return config, command, fmt.Errorf("Shell environment key %q is not allowed", key)
		}
	}
	return config, command, nil
}

type ExecShellRunner struct{}

func (ExecShellRunner) Run(ctx context.Context, path string, args []string, dir string, env map[string]string, maxOutput int64) (string, string, error) {
	command := exec.CommandContext(ctx, path, args...)
	command.Dir = dir
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	command.Env = make([]string, 0, len(keys))
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+env[key])
	}
	output := newCappedCommandOutput(maxOutput)
	command.Stdout = output.writer(false)
	command.Stderr = output.writer(true)
	err := command.Run()
	stdout, stderr := output.values()
	return stdout, stderr, err
}

type cappedCommandOutput struct {
	mu        sync.Mutex
	remaining int64
	stdout    strings.Builder
	stderr    strings.Builder
	truncated bool
}

type cappedCommandWriter struct {
	output *cappedCommandOutput
	stderr bool
}

func newCappedCommandOutput(limit int64) *cappedCommandOutput {
	return &cappedCommandOutput{remaining: limit}
}

func (output *cappedCommandOutput) writer(stderr bool) cappedCommandWriter {
	return cappedCommandWriter{output: output, stderr: stderr}
}

func (writer cappedCommandWriter) Write(value []byte) (int, error) {
	writer.output.mu.Lock()
	defer writer.output.mu.Unlock()
	originalLength := len(value)
	if int64(len(value)) > writer.output.remaining {
		value = value[:maxInt64(writer.output.remaining, 0)]
		writer.output.truncated = true
	}
	if writer.stderr {
		_, _ = writer.output.stderr.Write(value)
	} else {
		_, _ = writer.output.stdout.Write(value)
	}
	writer.output.remaining -= int64(len(value))
	return originalLength, nil
}

func (output *cappedCommandOutput) values() (string, string) {
	output.mu.Lock()
	defer output.mu.Unlock()
	stdout, stderr := output.stdout.String(), output.stderr.String()
	if output.truncated {
		stderr += "...[truncated]"
	}
	return stdout, stderr
}

func formatCommandOutput(stdout, stderr string, limit int64) string {
	value := strings.TrimSpace(stdout)
	if strings.TrimSpace(stderr) != "" {
		if value != "" {
			value += "\n"
		}
		value += strings.TrimSpace(stderr)
	}
	if int64(len(value)) <= limit {
		return value
	}
	return value[:maxInt64(limit, 0)] + "...[truncated]"
}

func maxInt64(value int64, minimum int64) int {
	if value < minimum {
		value = minimum
	}
	return int(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var _ application.Handler = (*ShellHandler)(nil)
