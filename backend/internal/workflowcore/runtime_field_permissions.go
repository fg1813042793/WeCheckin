package workflowcore

import (
	"encoding/json"
	"strings"
)

type RuntimeFieldPermissionsConfig struct {
	Nodes map[string][]FieldPermission `json:"nodes"`
}

func RuntimeFieldPermissionsFromDefinition(definition Definition) RuntimeFieldPermissionsConfig {
	config := RuntimeFieldPermissionsConfig{Nodes: make(map[string][]FieldPermission)}
	for _, node := range definition.Nodes {
		if node.Type != NodeTypeStart && node.Type != NodeTypeApproval && node.Type != NodeTypeHandle {
			continue
		}
		config.Nodes[node.ID] = cloneRuntimeFieldPermissions(node.FormPermissions)
	}
	return config
}

func EncodeRuntimeFieldPermissions(definition Definition) (string, error) {
	encoded, err := json.Marshal(RuntimeFieldPermissionsFromDefinition(definition))
	return string(encoded), err
}

func ApplyRuntimeFieldPermissionsJSON(definition *Definition, raw string) error {
	if definition == nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var config RuntimeFieldPermissionsConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return err
	}
	ApplyRuntimeFieldPermissions(definition, config)
	return nil
}

func ApplyRuntimeFieldPermissions(definition *Definition, config RuntimeFieldPermissionsConfig) {
	if definition == nil {
		return
	}
	fields := make(map[string]FormField)
	for _, field := range dataFormFields(definition.Form) {
		fields[field.Key] = field
	}
	for index := range definition.Nodes {
		configured, ok := config.Nodes[definition.Nodes[index].ID]
		if !ok {
			continue
		}
		filtered := make([]FieldPermission, 0, len(configured))
		for _, permission := range configured {
			field, exists := fields[permission.Field]
			if !exists {
				continue
			}
			cloned := permission
			cloned.Actions = append([]string(nil), permission.Actions...)
			if field.Type == FormFieldTypeCalculation && cloned.Access == FieldAccessWrite {
				cloned.Access = FieldAccessRead
				cloned.Actions = nil
			}
			filtered = append(filtered, cloned)
		}
		definition.Nodes[index].FormPermissions = filtered
	}
}

func cloneRuntimeFieldPermissions(source []FieldPermission) []FieldPermission {
	result := make([]FieldPermission, 0, len(source))
	for _, permission := range source {
		permission.Actions = append([]string(nil), permission.Actions...)
		result = append(result, permission)
	}
	return result
}
