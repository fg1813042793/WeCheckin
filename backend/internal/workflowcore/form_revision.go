package workflowcore

import (
	"fmt"
	"sort"
	"strings"
)

func PostHandleFormPermissions(definition Definition, handledNodeIDs []string) []FieldPermission {
	handled := make(map[string]struct{}, len(handledNodeIDs))
	for _, nodeID := range handledNodeIDs {
		if nodeID = strings.TrimSpace(nodeID); nodeID != "" {
			handled[nodeID] = struct{}{}
		}
	}
	routingFields := make(map[string]struct{})
	for _, edge := range definition.Edges {
		if edge.Condition == nil {
			continue
		}
		if field := strings.TrimSpace(edge.Condition.Field); field != "" {
			routingFields[field] = struct{}{}
		}
	}
	fields := dataFormFields(definition.Form)
	accessByField := make(map[string]string, len(fields))
	actionsByField := make(map[string]map[string]struct{})
	hasEligibleNode := false
	for _, node := range definition.Nodes {
		if _, ok := handled[strings.TrimSpace(node.ID)]; !ok || node.PostHandleEdit == nil || !node.PostHandleEdit.Enabled {
			continue
		}
		if node.Type != NodeTypeApproval && node.Type != NodeTypeHandle {
			continue
		}
		hasEligibleNode = true
		permissionByField := make(map[string]FieldPermission, len(node.FormPermissions))
		for _, permission := range node.FormPermissions {
			field := strings.TrimSpace(permission.Field)
			if field == "" {
				continue
			}
			permissionByField[field] = permission
		}
		for _, field := range fields {
			access := FieldAccessRead
			permission, configured := permissionByField[field.Key]
			if configured {
				switch permission.Access {
				case FieldAccessHidden, FieldAccessRead, FieldAccessWrite:
					access = permission.Access
				}
			}
			if access == FieldAccessWrite {
				if _, blocked := routingFields[field.Key]; blocked {
					access = FieldAccessRead
				}
			}
			if fieldAccessPriority(access) > fieldAccessPriority(accessByField[field.Key]) {
				accessByField[field.Key] = access
			}
			if access != FieldAccessWrite {
				continue
			}
			if actionsByField[field.Key] == nil {
				actionsByField[field.Key] = make(map[string]struct{})
			}
			for _, action := range permission.Actions {
				if action = strings.TrimSpace(action); action != "" {
					actionsByField[field.Key][action] = struct{}{}
				}
			}
		}
	}

	if !hasEligibleNode {
		return nil
	}
	result := make([]FieldPermission, 0, len(fields))
	for _, field := range fields {
		access := accessByField[field.Key]
		actions := make([]string, 0, len(actionsByField[field.Key]))
		if access == FieldAccessWrite {
			for action := range actionsByField[field.Key] {
				actions = append(actions, action)
			}
			sort.Strings(actions)
		}
		result = append(result, FieldPermission{Field: field.Key, Access: access, Actions: actions})
	}
	return result
}

func PostHandleEditablePermissions(definition Definition, handledNodeIDs []string) []FieldPermission {
	permissions := PostHandleFormPermissions(definition, handledNodeIDs)
	result := make([]FieldPermission, 0, len(permissions))
	for _, permission := range permissions {
		if permission.Access == FieldAccessWrite {
			result = append(result, permission)
		}
	}
	return result
}

func fieldAccessPriority(access string) int {
	switch access {
	case FieldAccessWrite:
		return 3
	case FieldAccessRead:
		return 2
	case FieldAccessHidden:
		return 1
	default:
		return 0
	}
}

func ValidatePostHandleFormPatch(definition Definition, handledNodeIDs []string, current, patch map[string]interface{}) error {
	permissions := PostHandleEditablePermissions(definition, handledNodeIDs)
	writableFields := make(map[string]struct{}, len(permissions))
	permissionByField := make(map[string]FieldPermission, len(permissions))
	for _, permission := range permissions {
		writableFields[permission.Field] = struct{}{}
		permissionByField[permission.Field] = permission
	}
	fieldByKey := make(map[string]FormField)
	for _, field := range dataFormFields(definition.Form) {
		fieldByKey[field.Key] = field
	}
	for field := range patch {
		if _, ok := writableFields[field]; !ok {
			return fmt.Errorf("%w：办理完成后无权修改字段 %s", ErrFormDataInvalid, field)
		}
		formField := fieldByKey[field]
		if formField.Type == FormFieldTypeDetailList {
			if err := validateDetailListPatchActions(formField, current[field], patch[field], permissionByField[field]); err != nil {
				return err
			}
		}
	}
	if err := validateSelectedFormData(definition.Form, patch, writableFields, true); err != nil {
		return err
	}
	return validateSelectedFormData(definition.Form, MergeFormData(current, patch), writableFields, false)
}
