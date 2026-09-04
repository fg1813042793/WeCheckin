package workflowsummary

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

type exportDocument struct {
	Title    string
	FileStem string
	Metadata [][]string
	Sections []exportSection
}

type exportSection struct {
	Title     string
	FieldRows [][]exportFieldCell
	Headers   []string
	Rows      [][]string
}

type exportFieldCell struct {
	Label string
	Value string
	Span  int
}

func buildExportDocument(detail *workflowapp.InstanceDetail) exportDocument {
	if detail == nil {
		return exportDocument{Title: "流程记录", FileStem: "workflow"}
	}
	instance := detail.Instance
	title := firstText(instance.DefinitionName, instance.DefinitionKey, "流程记录")
	fileStem := safeFileStem(firstText(instance.BusinessKey, instance.ID, title))
	document := exportDocument{
		Title:    title,
		FileStem: fileStem,
		Metadata: [][]string{
			{"流程名称", title},
			{"申请编号", firstText(instance.BusinessKey, instance.ID)},
			{"流程版本", fmt.Sprintf("v%d", instance.DefinitionVersion)},
			{"发起人", firstText(instance.StarterName, instance.StarterID)},
			{"操作人", firstText(instance.OperatorName, instance.OperatorID)},
			{"流程状态", instanceStatusLabel(instance.Status)},
			{"发起时间", formatExportTime(instance.StartTime)},
			{"完成时间", formatExportTime(instance.EndTime)},
		},
	}
	document.Sections = append(document.Sections, exportFormSections(detail.Form, detail.FormData, detail.UserNames)...)
	document.Sections = append(document.Sections, exportHistorySection(detail.Tasks))
	return document
}

func exportFormSections(fields []workflowcore.FormField, data map[string]interface{}, userNames map[string]string) []exportSection {
	sections := make([]exportSection, 0)
	fieldCells := make([]exportFieldCell, 0)
	flushFields := func(title string) {
		if len(fieldCells) == 0 {
			return
		}
		sections = append(sections, exportSection{Title: title, FieldRows: packExportFieldRows(fieldCells)})
		fieldCells = nil
	}
	var appendFields func([]workflowcore.FormField, string)
	appendFields = func(current []workflowcore.FormField, sectionTitle string) {
		for _, field := range current {
			label := firstText(field.Label, field.Key)
			switch field.Type {
			case workflowcore.FormFieldTypeGroup:
				flushFields(sectionTitle)
				appendFields(field.Fields, label)
				sectionTitle = "表单内容"
			case workflowcore.FormFieldTypeDetailList:
				flushFields(sectionTitle)
				sections = append(sections, exportDetailSection(field, data[field.Key], userNames))
			case workflowcore.FormFieldTypeLabel:
				fieldCells = append(fieldCells, exportFieldCell{Label: label, Span: exportFieldSpan(field)})
			case workflowcore.FormFieldTypeDescription:
				fieldCells = append(fieldCells, exportFieldCell{Label: label, Value: strings.TrimSpace(field.Content), Span: exportFieldSpan(field)})
			case workflowcore.FormFieldTypeButton:
				content := ""
				if field.Help != nil {
					content = strings.TrimSpace(field.Help.Content)
				}
				fieldCells = append(fieldCells, exportFieldCell{Label: label, Value: content, Span: exportFieldSpan(field)})
			default:
				fieldCells = append(fieldCells, exportFieldCell{
					Label: label,
					Value: exportFieldValue(field, data[field.Key], userNames),
					Span:  exportFieldSpan(field),
				})
			}
		}
		flushFields(sectionTitle)
	}
	appendFields(fields, "表单内容")
	if len(sections) == 0 {
		sections = append(sections, exportSection{
			Title:     "表单内容",
			FieldRows: [][]exportFieldCell{{{Label: "暂无表单内容", Span: 24}}},
		})
	}
	return sections
}

func exportFieldSpan(field workflowcore.FormField) int {
	if field.Span <= 0 || field.Span > 24 {
		return 24
	}
	return field.Span
}

func packExportFieldRows(fields []exportFieldCell) [][]exportFieldCell {
	rows := make([][]exportFieldCell, 0)
	current := make([]exportFieldCell, 0)
	used := 0
	for _, field := range fields {
		if field.Span <= 0 || field.Span > 24 {
			field.Span = 24
		}
		if used > 0 && used+field.Span > 24 {
			rows = append(rows, current)
			current = nil
			used = 0
		}
		current = append(current, field)
		used += field.Span
		if used == 24 {
			rows = append(rows, current)
			current = nil
			used = 0
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	return rows
}

func exportDetailSection(field workflowcore.FormField, value interface{}, userNames map[string]string) exportSection {
	headers := make([]string, 0, len(field.Columns))
	for _, column := range field.Columns {
		headers = append(headers, firstText(column.Label, column.Key))
	}
	rows := make([][]string, 0)
	for _, item := range exportObjectRows(value) {
		row := make([]string, 0, len(field.Columns))
		for _, column := range field.Columns {
			row = append(row, exportFieldValue(column, item[column.Key], userNames))
		}
		rows = append(rows, row)
	}
	if len(headers) == 0 {
		headers = []string{"内容"}
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"暂无明细"})
	}
	return exportSection{Title: firstText(field.Label, field.Key, "明细列表"), Headers: headers, Rows: rows}
}

func exportHistorySection(tasks []workflowapp.TaskSummary) exportSection {
	rows := make([][]string, 0, len(tasks))
	for _, task := range tasks {
		handler := firstText(task.HandledByName, task.AssigneeName, task.HandledBy, task.AssigneeID)
		rows = append(rows, []string{
			firstText(task.NodeName, task.NodeID),
			handler,
			taskActionLabel(firstText(task.Action, task.Status)),
			strings.TrimSpace(task.Comment),
			formatExportTime(task.HandledAt),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"-", "-", "-", "", "-"})
	}
	return exportSection{
		Title:   "审批记录",
		Headers: []string{"节点", "处理人", "动作", "意见", "处理时间"},
		Rows:    rows,
	}
}

func exportFieldValue(field workflowcore.FormField, value interface{}, userNames map[string]string) string {
	if value == nil {
		return ""
	}
	switch field.Type {
	case workflowcore.FormFieldTypeBoolean:
		if enabled, ok := value.(bool); ok {
			if enabled {
				return "是"
			}
			return "否"
		}
	case workflowcore.FormFieldTypeUser:
		id := strings.TrimSpace(fmt.Sprint(value))
		return firstText(userNames[id], id)
	case workflowcore.FormFieldTypeUserMulti:
		values := exportStringSlice(value)
		for index, id := range values {
			values[index] = firstText(userNames[id], id)
		}
		return strings.Join(values, "、")
	case workflowcore.FormFieldTypeSelect, workflowcore.FormFieldTypeRadio:
		return exportOptionLabel(field.Options, strings.TrimSpace(fmt.Sprint(value)))
	case workflowcore.FormFieldTypeMultiSelect, workflowcore.FormFieldTypeCheckbox:
		values := exportStringSlice(value)
		for index, item := range values {
			values[index] = exportOptionLabel(field.Options, item)
		}
		return strings.Join(values, "、")
	case workflowcore.FormFieldTypeAttachment:
		values := exportStringSlice(value)
		for index, item := range values {
			values[index] = exportAttachment(item)
		}
		return strings.Join(values, "\n")
	}
	if values := exportStringSlice(value); len(values) > 0 {
		return strings.Join(values, "、")
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case json.Number:
		return typed.String()
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return string(encoded)
}

func exportOptionLabel(options []workflowcore.FormOption, value string) string {
	for _, option := range options {
		if option.Value == value {
			return firstText(option.Label, option.Value)
		}
		if label := exportOptionLabel(option.Children, value); label != value {
			return label
		}
	}
	return value
}

func exportAttachment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	name := path.Base(strings.SplitN(value, "?", 2)[0])
	if name == "." || name == "/" || name == "" {
		return value
	}
	return name + " " + value
}

func exportStringSlice(value interface{}) []string {
	result := make([]string, 0)
	switch values := value.(type) {
	case []string:
		for _, item := range values {
			if item = strings.TrimSpace(item); item != "" {
				result = append(result, item)
			}
		}
	case []interface{}:
		for _, item := range values {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
				result = append(result, text)
			}
		}
	}
	return result
}

func exportObjectRows(value interface{}) []map[string]interface{} {
	switch rows := value.(type) {
	case []map[string]interface{}:
		return rows
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			if item, ok := row.(map[string]interface{}); ok {
				result = append(result, item)
			}
		}
		return result
	default:
		return nil
	}
}

func instanceStatusLabel(status string) string {
	return map[string]string{
		"running": "审批中", "completed": "已完成", "rejected": "已驳回",
		"withdrawn": "已撤回", "cancelled": "已取消",
	}[status]
}

func taskActionLabel(action string) string {
	if label := map[string]string{
		"approve": "通过", "approved": "通过", "reject": "驳回", "rejected": "驳回",
		"submit": "提交办理", "submitted": "已办理", "pending": "待处理", "waiting": "等待中",
	}[action]; label != "" {
		return label
	}
	return action
}

func formatExportTime(timestamp int64) string {
	if timestamp <= 0 {
		return "-"
	}
	return time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
}

func firstText(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func safeFileStem(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\r', '\n', '\t':
			return '_'
		}
		if r < 32 {
			return -1
		}
		return r
	}, value)
	value = strings.Trim(value, ". ")
	if value == "" {
		return "workflow"
	}
	runes := []rune(value)
	if len(runes) > 80 {
		value = string(runes[:80])
	}
	return value
}

func uniqueFileStems(documents []exportDocument) []string {
	seen := make(map[string]int, len(documents))
	result := make([]string, 0, len(documents))
	for _, document := range documents {
		stem := safeFileStem(document.FileStem)
		seen[stem]++
		if seen[stem] > 1 {
			stem = fmt.Sprintf("%s-%d", stem, seen[stem])
		}
		result = append(result, stem)
	}
	return result
}
