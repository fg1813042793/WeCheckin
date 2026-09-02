package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"wecheckin/backend/internal/modules/scheduledtask/application"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

const (
	workflowVersionLatest   = "latest_published"
	workflowVersionFixed    = "fixed_version"
	maxWorkflowStarterCount = 100
)

type WorkflowStarter interface {
	StartInstance(context.Context, workflowapp.StartInstanceRequest) (*workflowdomain.State, error)
}

type WorkflowHandler struct {
	starter WorkflowStarter
}

type workflowHandlerConfig struct {
	DefinitionID  uint                   `json:"definitionId"`
	VersionPolicy string                 `json:"versionPolicy"`
	FixedVersion  int                    `json:"fixedVersion"`
	StarterIDs    []uint                 `json:"starterIds"`
	StarterID     string                 `json:"starterId,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	FormData      map[string]interface{} `json:"formData"`
	legacyStarter bool
}

func NewWorkflowHandler(starter WorkflowStarter) *WorkflowHandler {
	return &WorkflowHandler{starter: starter}
}

func (handler *WorkflowHandler) Type() string { return "workflow" }

func (handler *WorkflowHandler) Metadata() application.HandlerMetadata {
	return application.HandlerMetadata{
		Type: "workflow", Name: "Start workflow", Description: "Starts one published generic workflow instance for each configured initiator",
		RiskLevel: "medium", ConfigSchema: json.RawMessage(`{
			"type":"object",
			"required":["definitionId","starterIds"],
			"properties":{
				"definitionId":{"type":"integer","minimum":1},
				"versionPolicy":{"type":"string","enum":["latest_published","fixed_version"]},
				"fixedVersion":{"type":"integer","minimum":1},
				"starterIds":{"type":"array","minItems":1,"maxItems":100,"uniqueItems":true,"items":{"type":"integer","minimum":1}},
				"variables":{"type":"object"},
				"formData":{"type":"object"}
			}
		}`),
	}
}

func (handler *WorkflowHandler) ValidateConfig(_ context.Context, raw json.RawMessage) error {
	_, err := decodeWorkflowHandlerConfig(raw)
	return err
}

func (handler *WorkflowHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	if handler == nil || handler.starter == nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "handler_unavailable", Summary: "workflow service is not initialized"}
	}
	config, err := decodeWorkflowHandlerConfig(json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	version := 0
	if config.VersionPolicy == workflowVersionFixed {
		version = config.FixedVersion
	}
	instanceIDs := make([]string, 0, len(config.StarterIDs))
	for _, starterValue := range config.StarterIDs {
		starterID := strconv.FormatUint(uint64(starterValue), 10)
		businessKey := fmt.Sprintf("%s:user:%s", run.RunID, starterID)
		if config.legacyStarter {
			businessKey = run.RunID
		}
		state, err := handler.starter.StartInstance(ctx, workflowapp.StartInstanceRequest{
			DefinitionID: config.DefinitionID, DefinitionVersion: version,
			BusinessType: "scheduled_task", BusinessKey: businessKey,
			StarterID: starterID, OperatorID: starterID,
			Variables: config.Variables, FormData: config.FormData, Idempotent: true,
		})
		if err != nil {
			return application.HandlerResult{}, &application.HandlerError{
				Code: "workflow_start_failed", Summary: fmt.Sprintf("start workflow for user %s: %v", starterID, err),
			}
		}
		if state == nil {
			return application.HandlerResult{}, errors.New("workflow service returned an empty state")
		}
		instanceIDs = append(instanceIDs, state.Instance.ID)
	}
	summary := instanceIDs[0]
	if len(instanceIDs) > 1 {
		summary = fmt.Sprintf("started %d workflow instances", len(instanceIDs))
	}
	return application.HandlerResult{
		Summary: summary,
		Data:    map[string]interface{}{"instanceId": instanceIDs[0], "instanceIds": instanceIDs},
	}, nil
}

func decodeWorkflowHandlerConfig(raw json.RawMessage) (workflowHandlerConfig, error) {
	var config workflowHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, fmt.Errorf("decode workflow task config: %w", err)
	}
	config.StarterID = strings.TrimSpace(config.StarterID)
	config.VersionPolicy = strings.TrimSpace(config.VersionPolicy)
	if config.VersionPolicy == "" {
		config.VersionPolicy = workflowVersionLatest
	}
	if config.DefinitionID == 0 {
		return config, errors.New("workflow definitionId is required")
	}
	if len(config.StarterIDs) == 0 && config.StarterID != "" {
		starterID, parseErr := strconv.ParseUint(config.StarterID, 10, 64)
		if parseErr != nil || starterID == 0 {
			return config, errors.New("workflow starterId is invalid")
		}
		config.StarterIDs = []uint{uint(starterID)}
		config.legacyStarter = true
	}
	seenStarterIDs := make(map[uint]struct{}, len(config.StarterIDs))
	normalizedStarterIDs := make([]uint, 0, len(config.StarterIDs))
	for _, starterID := range config.StarterIDs {
		if starterID == 0 {
			return config, errors.New("workflow starterIds contains an invalid user ID")
		}
		if _, exists := seenStarterIDs[starterID]; exists {
			continue
		}
		seenStarterIDs[starterID] = struct{}{}
		normalizedStarterIDs = append(normalizedStarterIDs, starterID)
	}
	if len(normalizedStarterIDs) == 0 {
		return config, errors.New("workflow starterIds is required")
	}
	if len(normalizedStarterIDs) > maxWorkflowStarterCount {
		return config, fmt.Errorf("workflow starterIds cannot exceed %d users", maxWorkflowStarterCount)
	}
	config.StarterIDs = normalizedStarterIDs
	if config.VersionPolicy != workflowVersionLatest && config.VersionPolicy != workflowVersionFixed {
		return config, errors.New("workflow versionPolicy is invalid")
	}
	if config.VersionPolicy == workflowVersionFixed && config.FixedVersion < 1 {
		return config, errors.New("workflow fixedVersion is required")
	}
	if config.Variables == nil {
		config.Variables = map[string]interface{}{}
	}
	if config.FormData == nil {
		config.FormData = map[string]interface{}{}
	}
	return config, nil
}

var _ application.Handler = (*WorkflowHandler)(nil)
