package workflowcore

import (
	"encoding/json"
	"strings"
)

// RuntimeStartConfig contains workflow settings that can change without publishing a new process version.
type RuntimeStartConfig struct {
	Initiator        InitiatorConfig         `json:"initiator"`
	Availability     StartAvailabilityConfig `json:"availability"`
	StartLimit       StartLimitConfig        `json:"startLimit"`
	InstanceIdentity *InstanceIdentityConfig `json:"instanceIdentity,omitempty"`
}

func RuntimeStartConfigFromDefinition(definition Definition) RuntimeStartConfig {
	config := RuntimeStartConfig{
		Initiator:        InitiatorConfig{Scope: InitiatorScopeAll},
		Availability:     DefaultStartAvailability(),
		StartLimit:       DefaultStartLimit(),
		InstanceIdentity: cloneInstanceIdentityConfig(definition.InstanceIdentity),
	}
	for index := range definition.Nodes {
		node := definition.Nodes[index]
		if node.Type != NodeTypeStart {
			continue
		}
		if node.Initiator != nil {
			config.Initiator = cloneInitiatorConfig(*node.Initiator)
		}
		config.Availability = CloneStartAvailability(node.Availability)
		config.StartLimit = CloneStartLimit(node.StartLimit)
		break
	}
	return config
}

func ApplyRuntimeStartConfig(definition *Definition, config RuntimeStartConfig) {
	if definition == nil {
		return
	}
	definition.InstanceIdentity = cloneInstanceIdentityConfig(config.InstanceIdentity)
	for index := range definition.Nodes {
		if definition.Nodes[index].Type != NodeTypeStart {
			continue
		}
		initiator := cloneInitiatorConfig(config.Initiator)
		availability := CloneStartAvailability(&config.Availability)
		limit := CloneStartLimit(&config.StartLimit)
		definition.Nodes[index].Initiator = &initiator
		definition.Nodes[index].Availability = &availability
		definition.Nodes[index].StartLimit = &limit
		return
	}
}

func EncodeRuntimeStartConfig(definition Definition) (string, error) {
	encoded, err := json.Marshal(RuntimeStartConfigFromDefinition(definition))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func ApplyRuntimeStartConfigJSON(definition *Definition, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var config RuntimeStartConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return err
	}
	ApplyRuntimeStartConfig(definition, config)
	return nil
}

func ValidateRuntimeStartConfig(definition Definition) []ValidationError {
	allowed := map[string]struct{}{
		ValidationInitiator:         {},
		ValidationStartAvailability: {},
		ValidationStartLimit:        {},
		ValidationInstanceIdentity:  {},
	}
	result := make([]ValidationError, 0)
	for _, validationError := range ValidateDefinition(definition) {
		if _, ok := allowed[validationError.Code]; ok {
			result = append(result, validationError)
		}
	}
	return result
}

func cloneInitiatorConfig(config InitiatorConfig) InitiatorConfig {
	return InitiatorConfig{
		Scope:           strings.TrimSpace(config.Scope),
		UserIDs:         append([]uint(nil), config.UserIDs...),
		DepartmentIDs:   append([]uint(nil), config.DepartmentIDs...),
		ExcludedUserIDs: append([]uint(nil), config.ExcludedUserIDs...),
	}
}

func cloneInstanceIdentityConfig(config *InstanceIdentityConfig) *InstanceIdentityConfig {
	if config == nil {
		return nil
	}
	cloned := &InstanceIdentityConfig{TitleTemplate: config.TitleTemplate}
	if config.BusinessPeriod != nil {
		period := *config.BusinessPeriod
		cloned.BusinessPeriod = &period
	}
	return cloned
}
