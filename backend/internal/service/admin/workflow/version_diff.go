package workflowservice

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/workflowcore"
)

const (
	versionChangeCategoryBasic        = "basic"
	versionChangeCategoryForm         = "form"
	versionChangeCategoryNode         = "node"
	versionChangeCategoryRoute        = "route"
	versionChangeCategoryStart        = "start"
	versionChangeCategoryNotification = "notification"
	versionChangeCategoryAutomation   = "automation"
)

type VersionChangeItem struct {
	Category string `json:"category"`
	Action   string `json:"action"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type VersionChangeSummary struct {
	BaseVersion int                 `json:"baseVersion"`
	Headline    string              `json:"headline"`
	ChangeCount int                 `json:"changeCount"`
	Items       []VersionChangeItem `json:"items"`
}

type versionMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	LogoURL     string `json:"logoUrl"`
}

type versionSnapshot struct {
	Metadata   versionMetadata
	Definition workflowcore.Definition
}

func buildVersionChangeSummary(baseVersion int, before, after versionSnapshot) VersionChangeSummary {
	if baseVersion <= 0 {
		item := VersionChangeItem{Category: versionChangeCategoryBasic, Action: "add", Title: "首次发布", Detail: "创建流程的首个可用版本"}
		return VersionChangeSummary{Headline: "首次发布", ChangeCount: 1, Items: []VersionChangeItem{item}}
	}

	items := make([]VersionChangeItem, 0)
	items = appendMetadataChanges(items, before.Metadata, after.Metadata)
	items = appendFormChanges(items, before.Definition.Form, after.Definition.Form)
	items = appendNodeChanges(items, before.Definition.Nodes, after.Definition.Nodes)
	items = appendEdgeChanges(items, before.Definition.Edges, after.Definition.Edges)
	if len(items) == 0 {
		items = append(items, VersionChangeItem{Category: versionChangeCategoryBasic, Action: "update", Title: "重新发布", Detail: "流程内容未发生可识别的语义变化"})
	}
	return VersionChangeSummary{
		BaseVersion: baseVersion,
		Headline:    versionChangeHeadline(items),
		ChangeCount: len(items),
		Items:       items,
	}
}

func appendMetadataChanges(items []VersionChangeItem, before, after versionMetadata) []VersionChangeItem {
	fields := []struct {
		label  string
		before string
		after  string
	}{
		{label: "流程名称", before: before.Name, after: after.Name},
		{label: "流程分类", before: before.Category, after: after.Category},
		{label: "流程说明", before: before.Description, after: after.Description},
		{label: "流程 Logo", before: before.LogoURL, after: after.LogoURL},
	}
	for _, field := range fields {
		if field.before == field.after {
			continue
		}
		items = append(items, VersionChangeItem{
			Category: versionChangeCategoryBasic,
			Action:   "update",
			Title:    field.label + "已修改",
			Detail:   fmt.Sprintf("%s -> %s", displayVersionValue(field.before), displayVersionValue(field.after)),
		})
	}
	return items
}

type flatVersionFormField struct {
	Key   string
	Label string
	Field workflowcore.FormField
	Order int
}

func appendFormChanges(items []VersionChangeItem, before, after []workflowcore.FormField) []VersionChangeItem {
	beforeFields := flattenVersionFormFields(before)
	afterFields := flattenVersionFormFields(after)
	beforeByKey := make(map[string]flatVersionFormField, len(beforeFields))
	afterByKey := make(map[string]flatVersionFormField, len(afterFields))
	for _, field := range beforeFields {
		beforeByKey[field.Key] = field
	}
	for _, field := range afterFields {
		afterByKey[field.Key] = field
	}
	for _, field := range afterFields {
		previous, exists := beforeByKey[field.Key]
		if !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryForm, Action: "add", Title: "新增字段", Detail: field.Label})
			continue
		}
		if jsonEqual(versionComparableFormField(previous.Field), versionComparableFormField(field.Field)) {
			continue
		}
		detail := field.Label
		if previous.Label != field.Label {
			detail = fmt.Sprintf("%s -> %s", previous.Label, field.Label)
		} else if previous.Field.Type != field.Field.Type {
			detail = fmt.Sprintf("%s：类型 %s -> %s", field.Label, previous.Field.Type, field.Field.Type)
		} else {
			detail = field.Label + "的配置已调整"
		}
		items = append(items, VersionChangeItem{Category: versionChangeCategoryForm, Action: "update", Title: "修改字段", Detail: detail})
	}
	for _, field := range beforeFields {
		if _, exists := afterByKey[field.Key]; !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryForm, Action: "delete", Title: "删除字段", Detail: field.Label})
		}
	}
	if commonFieldOrderChanged(beforeFields, afterFields, beforeByKey, afterByKey) {
		items = append(items, VersionChangeItem{Category: versionChangeCategoryForm, Action: "reorder", Title: "调整字段顺序", Detail: "表单字段排列顺序已修改"})
	}
	return items
}

func versionComparableFormField(field workflowcore.FormField) workflowcore.FormField {
	field.Fields = nil
	field.Columns = nil
	return field
}

func flattenVersionFormFields(fields []workflowcore.FormField) []flatVersionFormField {
	result := make([]flatVersionFormField, 0)
	var visit func([]workflowcore.FormField)
	visit = func(current []workflowcore.FormField) {
		for _, field := range current {
			result = append(result, flatVersionFormField{Key: field.Key, Label: firstNonEmpty(field.Label, field.Key), Field: field, Order: len(result)})
			visit(field.Fields)
			visit(field.Columns)
		}
	}
	visit(fields)
	return result
}

func commonFieldOrderChanged(before, after []flatVersionFormField, beforeByKey, afterByKey map[string]flatVersionFormField) bool {
	beforeOrder := make([]string, 0)
	afterOrder := make([]string, 0)
	for _, field := range before {
		if _, exists := afterByKey[field.Key]; exists {
			beforeOrder = append(beforeOrder, field.Key)
		}
	}
	for _, field := range after {
		if _, exists := beforeByKey[field.Key]; exists {
			afterOrder = append(afterOrder, field.Key)
		}
	}
	return !jsonEqual(beforeOrder, afterOrder)
}

func appendNodeChanges(items []VersionChangeItem, before, after []workflowcore.Node) []VersionChangeItem {
	beforeByID := make(map[string]workflowcore.Node, len(before))
	afterByID := make(map[string]workflowcore.Node, len(after))
	for _, node := range before {
		beforeByID[node.ID] = node
	}
	for _, node := range after {
		afterByID[node.ID] = node
	}
	for _, node := range after {
		previous, exists := beforeByID[node.ID]
		if !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryNode, Action: "add", Title: "新增节点", Detail: node.Name})
			continue
		}
		if previous.Name != node.Name || previous.Type != node.Type {
			items = append(items, VersionChangeItem{
				Category: versionChangeCategoryNode, Action: "update", Title: "修改节点",
				Detail: fmt.Sprintf("%s -> %s", firstNonEmpty(previous.Name, previous.ID), firstNonEmpty(node.Name, node.ID)),
			})
		}
		if !jsonEqual(previous.Assignee, node.Assignee) || previous.ApprovalMode != node.ApprovalMode ||
			previous.CompletionRate != node.CompletionRate || !jsonEqual(previous.FormPermissions, node.FormPermissions) {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryNode, Action: "update", Title: "调整节点配置", Detail: firstNonEmpty(node.Name, node.ID) + "的处理人、审批方式或字段权限已修改"})
		}
		if !jsonEqual(previous.Initiator, node.Initiator) || !jsonEqual(previous.Availability, node.Availability) || !jsonEqual(previous.StartLimit, node.StartLimit) {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryStart, Action: "update", Title: "调整发起配置", Detail: "允许发起范围、可用时间或次数限制已修改"})
		}
		if !jsonEqual(previous.Notification, node.Notification) || !jsonEqual(previous.ResultNotification, node.ResultNotification) {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryNotification, Action: "update", Title: "调整通知配置", Detail: firstNonEmpty(node.Name, node.ID) + "的通知渠道或消息模板已修改"})
		}
		if !jsonEqual(previous.Automation, node.Automation) || !jsonEqual(previous.Timer, node.Timer) {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryAutomation, Action: "update", Title: "调整自动动作", Detail: firstNonEmpty(node.Name, node.ID) + "的变量或定时配置已修改"})
		}
	}
	for _, node := range before {
		if _, exists := afterByID[node.ID]; !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryNode, Action: "delete", Title: "删除节点", Detail: firstNonEmpty(node.Name, node.ID)})
		}
	}
	return items
}

func appendEdgeChanges(items []VersionChangeItem, before, after []workflowcore.Edge) []VersionChangeItem {
	beforeByID := make(map[string]workflowcore.Edge, len(before))
	afterByID := make(map[string]workflowcore.Edge, len(after))
	for _, edge := range before {
		beforeByID[edge.ID] = edge
	}
	for _, edge := range after {
		afterByID[edge.ID] = edge
		previous, exists := beforeByID[edge.ID]
		if !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryRoute, Action: "add", Title: "新增流转路径", Detail: edgeDisplay(edge)})
			continue
		}
		if !jsonEqual(previous, edge) {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryRoute, Action: "update", Title: "修改流转条件", Detail: edgeDisplay(edge)})
		}
	}
	for _, edge := range before {
		if _, exists := afterByID[edge.ID]; !exists {
			items = append(items, VersionChangeItem{Category: versionChangeCategoryRoute, Action: "delete", Title: "删除流转路径", Detail: edgeDisplay(edge)})
		}
	}
	return items
}

func versionChangeHeadline(items []VersionChangeItem) string {
	labels := map[string]string{
		versionChangeCategoryBasic: "基础信息", versionChangeCategoryForm: "表单", versionChangeCategoryNode: "节点",
		versionChangeCategoryRoute: "流转", versionChangeCategoryStart: "发起配置",
		versionChangeCategoryNotification: "通知", versionChangeCategoryAutomation: "自动动作",
	}
	counts := make(map[string]int)
	for _, item := range items {
		counts[item.Category]++
	}
	order := []string{versionChangeCategoryBasic, versionChangeCategoryForm, versionChangeCategoryNode, versionChangeCategoryRoute, versionChangeCategoryStart, versionChangeCategoryNotification, versionChangeCategoryAutomation}
	parts := make([]string, 0, len(counts))
	for _, category := range order {
		if counts[category] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d 项", labels[category], counts[category]))
		}
	}
	return strings.Join(parts, " · ")
}

func versionSnapshotFromModel(row model.WorkflowDefinitionVersion, fallback versionMetadata) (versionSnapshot, bool, error) {
	var definition workflowcore.Definition
	if err := json.Unmarshal([]byte(row.SourceJSON), &definition); err != nil {
		return versionSnapshot{}, false, fmt.Errorf("解析流程定义 v%d 失败: %w", row.Version, err)
	}
	metadata, recorded, err := decodeVersionMetadata(row.MetadataJSON)
	if err != nil {
		return versionSnapshot{}, false, fmt.Errorf("解析流程定义 v%d 元数据失败: %w", row.Version, err)
	}
	if !recorded {
		metadata = fallback
		metadata.Name = definition.Name
	}
	return versionSnapshot{Metadata: metadata, Definition: definition}, recorded, nil
}

func metadataFromDefinition(item model.WorkflowDefinition) versionMetadata {
	return versionMetadata{Name: item.Name, Description: item.Description, Category: item.Category, LogoURL: item.LogoURL}
}

func encodeVersionMetadata(metadata versionMetadata) (string, error) {
	encoded, err := json.Marshal(metadata)
	return string(encoded), err
}

func decodeVersionMetadata(raw string) (versionMetadata, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return versionMetadata{}, false, nil
	}
	var metadata versionMetadata
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return versionMetadata{}, false, err
	}
	return metadata, true, nil
}

func encodeVersionChangeSummary(summary VersionChangeSummary) (string, error) {
	encoded, err := json.Marshal(summary)
	return string(encoded), err
}

func decodeVersionChangeSummary(raw string) (VersionChangeSummary, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return VersionChangeSummary{}, false, nil
	}
	var summary VersionChangeSummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return VersionChangeSummary{}, false, err
	}
	if summary.Items == nil {
		summary.Items = make([]VersionChangeItem, 0)
	}
	return summary, true, nil
}

func versionContentHash(snapshot versionSnapshot) (string, error) {
	encoded, err := json.Marshal(struct {
		Metadata   versionMetadata         `json:"metadata"`
		Definition workflowcore.Definition `json:"definition"`
	}{Metadata: snapshot.Metadata, Definition: snapshot.Definition})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func jsonEqual(left, right interface{}) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func displayVersionValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "未设置"
	}
	const maxRunes = 60
	runes := []rune(value)
	if len(runes) > maxRunes {
		value = string(runes[:maxRunes]) + "..."
	}
	return "“" + value + "”"
}

func edgeDisplay(edge workflowcore.Edge) string {
	if name := strings.TrimSpace(edge.Name); name != "" {
		return name
	}
	return firstNonEmpty(edge.Source, "?") + " -> " + firstNonEmpty(edge.Target, "?")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return "-"
}

func sortedVersionNumbers(rows []model.WorkflowDefinitionVersion) []int {
	versions := make([]int, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, row.Version)
	}
	sort.Ints(versions)
	return versions
}
