package infrastructure

import (
	"strconv"
	"strings"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/workflowcore"
)

func collectInstanceUserIDs(
	instance workflowmodel.ProcessInstance,
	tasks []workflowmodel.ProcessTask,
	history []workflowmodel.ProcessHistory,
	fields []workflowcore.FormField,
	formData map[string]interface{},
) []uint {
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})
	appendWorkflowUserID(&ids, seen, instance.StarterID)
	appendWorkflowUserID(&ids, seen, instance.OperatorID)
	for _, task := range tasks {
		appendWorkflowUserID(&ids, seen, task.AssigneeID)
		appendWorkflowUserID(&ids, seen, task.HandledBy)
	}
	for _, event := range history {
		appendWorkflowUserID(&ids, seen, event.ActorID)
	}
	collectFormUserIDs(&ids, seen, fields, formData)
	return ids
}

func collectFormUserIDs(ids *[]uint, seen map[uint]struct{}, fields []workflowcore.FormField, data map[string]interface{}) {
	for _, field := range fields {
		if field.Type == workflowcore.FormFieldTypeGroup {
			collectFormUserIDs(ids, seen, field.Fields, data)
			continue
		}
		value := data[field.Key]
		switch field.Type {
		case workflowcore.FormFieldTypeUser:
			appendWorkflowUserID(ids, seen, stringFormValue(value))
		case workflowcore.FormFieldTypeUserMulti:
			for _, item := range stringSliceFormValue(value) {
				appendWorkflowUserID(ids, seen, item)
			}
		case workflowcore.FormFieldTypeDetailList:
			for _, row := range detailFormRows(value) {
				collectFormUserIDs(ids, seen, field.Columns, row)
			}
		}
	}
}

func appendWorkflowUserID(ids *[]uint, seen map[uint]struct{}, value string) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return
	}
	id := uint(parsed)
	if _, exists := seen[id]; exists {
		return
	}
	seen[id] = struct{}{}
	*ids = append(*ids, id)
}

func stringFormValue(value interface{}) string {
	text, _ := value.(string)
	return text
}

func stringSliceFormValue(value interface{}) []string {
	switch values := value.(type) {
	case []string:
		return values
	case []interface{}:
		result := make([]string, 0, len(values))
		for _, item := range values {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func detailFormRows(value interface{}) []map[string]interface{} {
	switch rows := value.(type) {
	case []map[string]interface{}:
		return rows
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(rows))
		for _, item := range rows {
			if row, ok := item.(map[string]interface{}); ok {
				result = append(result, row)
			}
		}
		return result
	default:
		return nil
	}
}
