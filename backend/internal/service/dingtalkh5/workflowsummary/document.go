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
	Layout   string
	Title    string
	FileStem string
	Metadata [][]string
	Sections []exportSection
}

type exportSection struct {
	Title        string
	FieldRows    [][]exportFieldCell
	Headers      []string
	Rows         [][]string
	AnswerBlocks []AnswerBlock
}

// DocumentLayoutAnswerSheet opts PDF and DOCX into a flowing answer-sheet layout.
// XLSX continues to use the section headers and rows.
const DocumentLayoutAnswerSheet = "answer_sheet"

// AnswerBlock contains display text supplied by the caller, including scoring
// and empty-answer wording. Renderers must not infer or recalculate these values.
type AnswerBlock struct {
	Title           string
	Answer          string
	ReferenceAnswer string
	Score           string
	Result          string
}

type exportFieldCell struct {
	Label string
	Value string
	Span  float64
}

// Document describes a structured file that can be rendered as PDF, DOCX or XLSX.
// The aliases keep the existing workflow export implementation compatible while
// allowing other bounded services to reuse the renderer without duplicating it.
type Document = exportDocument
type Section = exportSection
type FieldCell = exportFieldCell

func buildExportDocument(detail *workflowapp.InstanceDetail, content ExportContent) exportDocument {
	if detail == nil {
		return exportDocument{Title: "流程记录", FileStem: "workflow"}
	}
	instance := detail.Instance
	starterPrefix := ""
	if name := firstText(instance.StarterName, detail.UserNames[instance.StarterID]); name != "" {
		starterPrefix = name + "-"
	}
	if content == ExportContentForm {
		title := starterPrefix + firstText(instance.DefinitionName, "表单")
		return exportDocument{
			Title:    title,
			FileStem: safeFileStem(title),
			Sections: exportFormSections(detail.Form, detail.FormData, detail.UserNames, true),
		}
	}
	title := firstText(instance.DefinitionName, instance.DefinitionKey, "流程记录")
	fileStem := safeFileStem(firstText(instance.BusinessKey, instance.ID, title))
	document := exportDocument{
		Title:    starterPrefix + title,
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
	document.Sections = append(document.Sections, exportFormSections(detail.Form, detail.FormData, detail.UserNames, false)...)
	document.Sections = append(document.Sections, exportHistorySection(detail.Tasks))
	return document
}

func exportFormSections(fields []workflowcore.FormField, data map[string]interface{}, userNames map[string]string, preserveGroups bool) []exportSection {
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
				// Groups without leading scalar fields need their own heading.
				if preserveGroups && (len(field.Fields) == 0 || field.Fields[0].Type == workflowcore.FormFieldTypeGroup || field.Fields[0].Type == workflowcore.FormFieldTypeDetailList) {
					sections = append(sections, exportSection{Title: label})
				}
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

func exportFieldSpan(field workflowcore.FormField) float64 {
	return normalizedExportSpan(field.Span)
}

func packExportFieldRows(fields []exportFieldCell) [][]exportFieldCell {
	rows := make([][]exportFieldCell, 0)
	current := make([]exportFieldCell, 0)
	used := 0
	for _, field := range fields {
		field.Span = normalizedExportSpan(field.Span)
		// Integer units keep five fifths on one row without float accumulation.
		span := exportGridSpan(field.Span, exportFineGridColumns)
		if used > 0 && used+span > exportFineGridColumns {
			rows = append(rows, current)
			current = nil
			used = 0
		}
		current = append(current, field)
		used += span
		if used == exportFineGridColumns {
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
		"withdrawn": "已撤回", "cancelled": "已取消", "voided": "已作废",
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
	seen := make(map[string]bool, len(documents))
	result := make([]string, 0, len(documents))
	for _, document := range documents {
		stem := safeFileStem(document.FileStem)
		base := stem
		for ordinal := 2; seen[strings.ToLower(stem)]; ordinal++ {
			stem = fmt.Sprintf("%s-%d", base, ordinal)
		}
		seen[strings.ToLower(stem)] = true
		result = append(result, stem)
	}
	return result
}
