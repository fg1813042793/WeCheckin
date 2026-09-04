package workflowcore

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

var definitionKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,99}$`)
var formFieldKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,99}$`)
var notificationTemplateTokenPattern = regexp.MustCompile(`\{\{([^{}]+)\}\}`)

func ValidateDefinition(definition Definition) []ValidationError {
	errors := make([]ValidationError, 0)
	if definition.SchemaVersion != CurrentSchemaVersion {
		errors = append(errors, ValidationError{Code: ValidationSchemaVersion, Message: "流程定义版本不受支持"})
	}
	if !definitionKeyPattern.MatchString(strings.TrimSpace(definition.Key)) {
		errors = append(errors, ValidationError{Code: ValidationDefinitionKey, Message: "流程编码需以字母开头，仅允许字母、数字、下划线和中划线"})
	}
	formFields, formErrors := validateFormSchema(definition.Form)
	errors = append(errors, formErrors...)

	nodes := make(map[string]Node, len(definition.Nodes))
	incoming := make(map[string][]Edge, len(definition.Nodes))
	outgoing := make(map[string][]Edge, len(definition.Nodes))
	startIDs := make([]string, 0, 1)
	endIDs := make([]string, 0, 1)
	for _, node := range definition.Nodes {
		node.ID = strings.TrimSpace(node.ID)
		if node.ID == "" {
			errors = append(errors, ValidationError{Code: ValidationDuplicateNode, Message: "节点 ID 不能为空"})
			continue
		}
		if _, exists := nodes[node.ID]; exists {
			errors = append(errors, ValidationError{Code: ValidationDuplicateNode, Message: "节点 ID 重复", NodeID: node.ID})
			continue
		}
		nodes[node.ID] = node
		switch node.Type {
		case NodeTypeStart:
			startIDs = append(startIDs, node.ID)
			errors = append(errors, validateInitiator(node)...)
			errors = append(errors, validateStartAvailability(node)...)
			errors = append(errors, validateStartLimit(node)...)
			errors = append(errors, validateFieldPermissions(node, formFields)...)
		case NodeTypeEnd:
			endIDs = append(endIDs, node.ID)
		case NodeTypeApproval:
			errors = append(errors, validateApproval(node)...)
			errors = append(errors, validateOptionalNotification(node)...)
			errors = append(errors, validateOptionalResultNotification(node)...)
			errors = append(errors, validateFieldPermissions(node, formFields)...)
		case NodeTypeHandle:
			errors = append(errors, validateNodeAssignee(node, "办理节点")...)
			errors = append(errors, validateOptionalNotification(node)...)
			errors = append(errors, validateFieldPermissions(node, formFields)...)
		case NodeTypeCC:
			errors = append(errors, validateNodeAssignee(node, "抄送节点")...)
			errors = append(errors, validateOptionalNotification(node)...)
		case NodeTypeNotify:
			errors = append(errors, validateNodeAssignee(node, "通知节点")...)
			errors = append(errors, validateRequiredNotification(node)...)
		case NodeTypeAutomation:
			errors = append(errors, validateAutomation(node)...)
		case NodeTypeTimer:
			errors = append(errors, validateTimer(node)...)
		case NodeTypeExclusive, NodeTypeParallel:
			if node.GatewayMode != GatewayModeSplit && node.GatewayMode != GatewayModeJoin {
				errors = append(errors, ValidationError{Code: ValidationGatewayMode, Message: "网关必须设置为分支或汇聚", NodeID: node.ID})
			}
		default:
			errors = append(errors, ValidationError{Code: ValidationGatewayMode, Message: "不支持的节点类型", NodeID: node.ID})
		}
	}

	if len(startIDs) == 0 {
		errors = append(errors, ValidationError{Code: ValidationMissingStart, Message: "流程必须包含一个开始节点"})
	} else if len(startIDs) > 1 {
		errors = append(errors, ValidationError{Code: ValidationMultipleStarts, Message: "流程只能包含一个开始节点"})
	}
	if len(endIDs) == 0 {
		errors = append(errors, ValidationError{Code: ValidationMissingEnd, Message: "流程必须至少包含一个结束节点"})
	}

	edgeIDs := make(map[string]struct{}, len(definition.Edges))
	for _, edge := range definition.Edges {
		edge.ID = strings.TrimSpace(edge.ID)
		if edge.ID == "" {
			errors = append(errors, ValidationError{Code: ValidationDuplicateEdge, Message: "连线 ID 不能为空"})
			continue
		}
		if _, exists := edgeIDs[edge.ID]; exists {
			errors = append(errors, ValidationError{Code: ValidationDuplicateEdge, Message: "连线 ID 重复", EdgeID: edge.ID})
			continue
		}
		edgeIDs[edge.ID] = struct{}{}
		if _, ok := nodes[edge.Source]; !ok {
			errors = append(errors, ValidationError{Code: ValidationEdgeEndpoint, Message: "连线起点不存在", EdgeID: edge.ID})
			continue
		}
		if _, ok := nodes[edge.Target]; !ok {
			errors = append(errors, ValidationError{Code: ValidationEdgeEndpoint, Message: "连线终点不存在", EdgeID: edge.ID})
			continue
		}
		outgoing[edge.Source] = append(outgoing[edge.Source], edge)
		incoming[edge.Target] = append(incoming[edge.Target], edge)
	}

	for nodeID, node := range nodes {
		if node.Type != NodeTypeStart && len(incoming[nodeID]) == 0 {
			errors = append(errors, ValidationError{Code: ValidationIncomingRequired, Message: "节点缺少进入连线", NodeID: nodeID})
		}
		if node.Type != NodeTypeEnd && len(outgoing[nodeID]) == 0 {
			errors = append(errors, ValidationError{Code: ValidationOutgoingRequired, Message: "节点缺少离开连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeStart && len(incoming[nodeID]) > 0 {
			errors = append(errors, ValidationError{Code: ValidationIncomingRequired, Message: "开始节点不能有进入连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeEnd && len(outgoing[nodeID]) > 0 {
			errors = append(errors, ValidationError{Code: ValidationOutgoingRequired, Message: "结束节点不能有离开连线", NodeID: nodeID})
		}
		if node.Type == NodeTypeExclusive || node.Type == NodeTypeParallel {
			errors = append(errors, validateGateway(node, incoming[nodeID], outgoing[nodeID])...)
		}
	}

	if len(startIDs) == 1 {
		reachable := visitForward(startIDs[0], outgoing)
		for nodeID := range nodes {
			if !reachable[nodeID] {
				errors = append(errors, ValidationError{Code: ValidationUnreachableNode, Message: "节点无法从开始节点到达", NodeID: nodeID})
			}
		}
	}
	if len(endIDs) > 0 {
		canReachEnd := make(map[string]bool, len(nodes))
		for _, endID := range endIDs {
			for nodeID := range visitBackward(endID, incoming) {
				canReachEnd[nodeID] = true
			}
		}
		for nodeID := range nodes {
			if !canReachEnd[nodeID] {
				errors = append(errors, ValidationError{Code: ValidationNoPathToEnd, Message: "节点不存在通往结束节点的路径", NodeID: nodeID})
			}
		}
	}
	return errors
}

func validateInitiator(node Node) []ValidationError {
	if node.Initiator == nil {
		return nil
	}
	switch strings.TrimSpace(node.Initiator.Scope) {
	case InitiatorScopeAll:
		if len(node.Initiator.UserIDs) == 0 && len(node.Initiator.DepartmentIDs) == 0 &&
			validInitiatorIDs(node.Initiator.ExcludedUserIDs) {
			return nil
		}
	case InitiatorScopeSpecified:
		if len(node.Initiator.UserIDs)+len(node.Initiator.DepartmentIDs) > 0 &&
			validInitiatorIDs(node.Initiator.UserIDs) &&
			validInitiatorIDs(node.Initiator.DepartmentIDs) &&
			validInitiatorIDs(node.Initiator.ExcludedUserIDs) {
			return nil
		}
	}
	return []ValidationError{{Code: ValidationInitiator, Message: "开始节点发起人范围无效", NodeID: node.ID}}
}

func validInitiatorIDs(ids []uint) bool {
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return false
		}
		if _, exists := seen[id]; exists {
			return false
		}
		seen[id] = struct{}{}
	}
	return true
}

func validateStartAvailability(node Node) []ValidationError {
	if validStartAvailabilityConfig(node.Availability) {
		return nil
	}
	return []ValidationError{{Code: ValidationStartAvailability, Message: "开始节点允许发起时间配置无效", NodeID: node.ID}}
}

func validateStartLimit(node Node) []ValidationError {
	if validStartLimitConfig(node.StartLimit, node.Availability) {
		return nil
	}
	return []ValidationError{{Code: ValidationStartLimit, Message: "开始节点发起次数限制配置无效", NodeID: node.ID}}
}

func validateFormSchema(fields []FormField) (map[string]FormField, []ValidationError) {
	fieldByKey := make(map[string]FormField)
	layoutKeys := make(map[string]struct{})
	errors := validateFormSchemaFields(fields, formSchemaRoot, layoutKeys, fieldByKey)
	errors = append(errors, validateFormRuleSchemas(fields, fieldByKey)...)
	return fieldByKey, errors
}

type formSchemaContext uint8

const (
	formSchemaRoot formSchemaContext = iota
	formSchemaGroup
	formSchemaDetail
)

func validateFormSchemaFields(fields []FormField, context formSchemaContext, layoutKeys map[string]struct{}, fieldByKey map[string]FormField) []ValidationError {
	localKeys := make(map[string]struct{}, len(fields))
	errors := make([]ValidationError, 0)
	for _, field := range fields {
		field.Key = strings.TrimSpace(field.Key)
		field.Label = strings.TrimSpace(field.Label)
		if !formFieldKeyPattern.MatchString(field.Key) || field.Label == "" {
			errors = append(errors, ValidationError{Code: ValidationFormFieldKey, Message: "表单字段编码或名称无效"})
			continue
		}
		keySet := layoutKeys
		if context == formSchemaDetail {
			keySet = localKeys
		}
		if _, exists := keySet[field.Key]; exists {
			errors = append(errors, ValidationError{Code: ValidationFormFieldDuplicate, Message: "表单字段编码重复：" + field.Key})
			continue
		}
		keySet[field.Key] = struct{}{}

		if !validFormFieldSpan(field.Span) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldSpan, Message: "表单字段宽度无效：" + field.Key})
		}

		switch field.Type {
		case FormFieldTypeGroup:
			if context != formSchemaRoot || len(field.Fields) == 0 || strings.TrimSpace(field.Content) != "" {
				errors = append(errors, ValidationError{Code: ValidationFormFieldLayout, Message: "表单分组配置无效：" + field.Key})
			}
			if !validFormHelp(field.Help) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldHelp, Message: "表单说明配置无效：" + field.Key})
			}
			errors = append(errors, validateFormSchemaFields(field.Fields, formSchemaGroup, layoutKeys, fieldByKey)...)
			continue
		case FormFieldTypeLabel:
			if context == formSchemaDetail || len(field.Fields) > 0 || strings.TrimSpace(field.Content) != "" || field.Help != nil || utf8.RuneCountInString(field.Label) > 100 {
				errors = append(errors, ValidationError{Code: ValidationFormFieldLayout, Message: "表单标签配置无效：" + field.Key})
			}
			continue
		case FormFieldTypeDescription:
			content := strings.TrimSpace(field.Content)
			if context == formSchemaDetail || len(field.Fields) > 0 || field.Help != nil || content == "" || utf8.RuneCountInString(content) > 2000 {
				errors = append(errors, ValidationError{Code: ValidationFormFieldLayout, Message: "表单说明文字配置无效：" + field.Key})
			}
			continue
		case FormFieldTypeButton:
			if context == formSchemaDetail || len(field.Fields) > 0 || strings.TrimSpace(field.Content) != "" || utf8.RuneCountInString(field.Label) > 30 {
				errors = append(errors, ValidationError{Code: ValidationFormFieldLayout, Message: "表单按钮配置无效：" + field.Key})
			}
			if field.Help == nil || !validFormHelp(field.Help) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldHelp, Message: "表单按钮说明配置无效：" + field.Key})
			}
			continue
		case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypeNumber, FormFieldTypeSelect,
			FormFieldTypeMultiSelect, FormFieldTypeDate, FormFieldTypeDateTime, FormFieldTypeUser,
			FormFieldTypeDepartment, FormFieldTypeAttachment, FormFieldTypeBoolean, FormFieldTypeAmount,
			FormFieldTypePhone, FormFieldTypeEmail, FormFieldTypeRadio, FormFieldTypeCheckbox,
			FormFieldTypeTime, FormFieldTypeDateRange, FormFieldTypeUserMulti, FormFieldTypeDepartmentMulti:
		case FormFieldTypeDetailList:
			if context == formSchemaDetail {
				errors = append(errors, ValidationError{Code: ValidationFormFieldColumns, Message: "明细列表不支持嵌套：" + field.Key})
			}
		default:
			errors = append(errors, ValidationError{Code: ValidationFormFieldType, Message: "表单字段类型无效：" + field.Key})
			continue
		}
		if len(field.Fields) > 0 || strings.TrimSpace(field.Content) != "" {
			errors = append(errors, ValidationError{Code: ValidationFormFieldLayout, Message: "业务字段不能配置子组件或说明文字：" + field.Key})
		}
		if !validFormHelp(field.Help) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldHelp, Message: "表单说明配置无效：" + field.Key})
		}
		fieldByKey[field.Key] = field
		if isFormOptionFieldType(field.Type) {
			if !validFormFieldOptionSource(field) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldOptionSource, Message: "选择字段选项来源配置无效：" + field.Key})
			}
			if formFieldUsesAPIOptions(field) {
				if len(field.Options) > 0 && !validFormOptions(field.Options) {
					errors = append(errors, ValidationError{Code: ValidationFormFieldOptions, Message: "选择字段必须配置不重复的选项：" + field.Key})
				}
			} else if !validFormOptions(field.Options) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldOptions, Message: "选择字段必须配置不重复的选项：" + field.Key})
			}
		} else if field.OptionSource != nil {
			errors = append(errors, ValidationError{Code: ValidationFormFieldOptionSource, Message: "非选择字段不能配置选项来源：" + field.Key})
		}
		if field.MaxLength < 0 || (field.Min != nil && field.Max != nil && *field.Min > *field.Max) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldRange, Message: "表单字段约束无效：" + field.Key})
		}
		visibleRowsConfigured := field.MinVisibleRows != 0 || field.MaxVisibleRows != 0
		if visibleRowsConfigured && (field.Type != FormFieldTypeTextarea ||
			field.MinVisibleRows < 1 || field.MaxVisibleRows < 1 ||
			field.MinVisibleRows > field.MaxVisibleRows || field.MaxVisibleRows > 30) {
			errors = append(errors, ValidationError{Code: ValidationFormFieldVisibleRows, Message: "多行文本显示行数配置无效：" + field.Key})
		}
		if field.Type == FormFieldTypeDetailList {
			rowKey := detailListRowKey(field)
			if !formFieldKeyPattern.MatchString(rowKey) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldColumns, Message: "明细列表行标识无效：" + field.Key})
			}
			if len(field.Columns) == 0 {
				errors = append(errors, ValidationError{Code: ValidationFormFieldColumns, Message: "明细列表必须配置列：" + field.Key})
			} else {
				columnByKey := make(map[string]FormField, len(field.Columns))
				errors = append(errors, validateFormSchemaFields(field.Columns, formSchemaDetail, nil, columnByKey)...)
				if _, exists := columnByKey[rowKey]; exists {
					errors = append(errors, ValidationError{Code: ValidationFormFieldColumns, Message: "明细列表行标识不能与列编码重复：" + field.Key})
				}
			}
			if field.MinRows < 0 || field.MaxRows < 0 || (field.MaxRows > 0 && field.MinRows > field.MaxRows) {
				errors = append(errors, ValidationError{Code: ValidationFormFieldRows, Message: "明细列表行数约束无效：" + field.Key})
			}
		}
		if field.Default != nil {
			if err := validateFieldValue(field, field.Default, false); err != nil {
				errors = append(errors, ValidationError{Code: ValidationFormFieldType, Message: "表单字段默认值无效：" + field.Key})
			}
		}
	}
	return errors
}

func validFormHelp(help *FormHelp) bool {
	if help == nil {
		return true
	}
	buttonText := strings.TrimSpace(help.ButtonText)
	title := strings.TrimSpace(help.Title)
	content := strings.TrimSpace(help.Content)
	return (buttonText == "" || utf8.RuneCountInString(buttonText) <= 30) &&
		title != "" && utf8.RuneCountInString(title) <= 100 &&
		content != "" && utf8.RuneCountInString(content) <= 2000
}

func validateFormRuleSchemas(fields []FormField, fieldByKey map[string]FormField) []ValidationError {
	errors := make([]ValidationError, 0)
	for _, field := range fields {
		if field.Type == FormFieldTypeGroup {
			errors = append(errors, validateFormRuleSchemas(field.Fields, fieldByKey)...)
			continue
		}
		if isFormLayoutFieldType(field.Type) {
			continue
		}
		errors = append(errors, validateFieldRuleSchema(field, fieldByKey)...)
		if field.Type == FormFieldTypeDetailList {
			columns := make(map[string]FormField, len(field.Columns))
			for _, column := range field.Columns {
				columns[strings.TrimSpace(column.Key)] = column
			}
			for _, column := range field.Columns {
				errors = append(errors, validateFieldRuleSchema(column, columns)...)
			}
		}
	}
	return errors
}

func validateFieldRuleSchema(field FormField, fieldByKey map[string]FormField) []ValidationError {
	seen := make(map[string]struct{}, len(field.Rules))
	errors := make([]ValidationError, 0)
	for _, rule := range field.Rules {
		rule.ID = strings.TrimSpace(rule.ID)
		rule.Type = strings.TrimSpace(rule.Type)
		rule.Field = strings.TrimSpace(rule.Field)
		rule.Column = strings.TrimSpace(rule.Column)
		rule.Operator = strings.TrimSpace(rule.Operator)
		rule.Message = strings.TrimSpace(rule.Message)
		valid := formFieldKeyPattern.MatchString(rule.ID) && utf8.RuneCountInString(rule.Message) <= 200
		if _, exists := seen[rule.ID]; exists {
			valid = false
		}
		seen[rule.ID] = struct{}{}
		if rule.When != nil && !validFormRuleCondition(*rule.When, field.Key, fieldByKey) {
			valid = false
		}
		switch rule.Type {
		case FormRuleMinLength, FormRuleMaxLength:
			valid = valid && isTextRuleField(field.Type) && validRuleIntegerBounds(rule.Min, rule.Max, rule.Type == FormRuleMinLength, rule.Type == FormRuleMaxLength)
		case FormRulePattern:
			_, err := regexp.Compile(rule.Pattern)
			valid = valid && isTextRuleField(field.Type) && strings.TrimSpace(rule.Pattern) != "" && len(rule.Pattern) <= 500 && err == nil
		case FormRuleNumberRange:
			valid = valid && isNumberRuleField(field.Type) && validRuleBounds(rule.Min, rule.Max)
		case FormRuleDecimalPlaces:
			valid = valid && isNumberRuleField(field.Type) && rule.Precision != nil && *rule.Precision >= 0 && *rule.Precision <= 10
		case FormRuleSelectionCount:
			valid = valid && isSelectionRuleField(field.Type) && validRuleIntegerBounds(rule.Min, rule.Max, false, false)
		case FormRuleCompareField:
			target, exists := fieldByKey[rule.Field]
			valid = valid && exists && rule.Field != field.Key && isCompareRuleField(field.Type) &&
				comparableRuleFields(field.Type, target.Type) && validCompareOperatorForFields(field.Type, target.Type, rule.Operator)
		case FormRuleColumnSum:
			column, exists := detailListColumn(field, rule.Column)
			valid = valid && field.Type == FormFieldTypeDetailList && exists && isNumberRuleField(column.Type) &&
				rule.Value != nil && !math.IsNaN(*rule.Value) && !math.IsInf(*rule.Value, 0) &&
				validCompareOperator(rule.Operator)
		case FormRuleConditionalRequired:
			valid = valid && rule.When != nil
		default:
			valid = false
		}
		if !valid {
			errors = append(errors, ValidationError{Code: ValidationFormFieldRules, Message: "表单字段校验规则无效：" + field.Key})
		}
	}
	return errors
}

func validFormRuleCondition(condition FormRuleCondition, ownerKey string, fieldByKey map[string]FormField) bool {
	condition.Field = strings.TrimSpace(condition.Field)
	condition.Operator = strings.TrimSpace(condition.Operator)
	if condition.Field == "" || condition.Field == ownerKey {
		return false
	}
	if _, exists := fieldByKey[condition.Field]; !exists {
		return false
	}
	if condition.Operator == FormRuleOperatorEmpty || condition.Operator == FormRuleOperatorNotEmpty {
		return true
	}
	return validCompareOperator(condition.Operator)
}

func validCompareOperator(operator string) bool {
	switch operator {
	case FormRuleOperatorEQ, FormRuleOperatorNE, FormRuleOperatorGT, FormRuleOperatorGTE, FormRuleOperatorLT, FormRuleOperatorLTE:
		return true
	default:
		return false
	}
}

func validRuleBounds(minimum, maximum *float64) bool {
	return (minimum != nil || maximum != nil) && (minimum == nil || maximum == nil || *minimum <= *maximum)
}

func validRuleIntegerBounds(minimum, maximum *float64, requireMin, requireMax bool) bool {
	if (requireMin && minimum == nil) || (requireMax && maximum == nil) || !validRuleBounds(minimum, maximum) {
		return false
	}
	for _, value := range []*float64{minimum, maximum} {
		if value != nil && (*value < 0 || *value != math.Trunc(*value)) {
			return false
		}
	}
	return true
}

func isTextRuleField(fieldType string) bool {
	return fieldType == FormFieldTypeText || fieldType == FormFieldTypeTextarea || fieldType == FormFieldTypePhone || fieldType == FormFieldTypeEmail
}

func isNumberRuleField(fieldType string) bool {
	return fieldType == FormFieldTypeNumber || fieldType == FormFieldTypeAmount
}

func isSelectionRuleField(fieldType string) bool {
	return fieldType == FormFieldTypeMultiSelect || fieldType == FormFieldTypeCheckbox || fieldType == FormFieldTypeAttachment ||
		fieldType == FormFieldTypeUserMulti || fieldType == FormFieldTypeDepartmentMulti
}

func isCompareRuleField(fieldType string) bool {
	return compareRuleFieldFamily(fieldType) != ""
}

func comparableRuleFields(left, right string) bool {
	leftFamily := compareRuleFieldFamily(left)
	return leftFamily != "" && leftFamily == compareRuleFieldFamily(right)
}

func validCompareOperatorForFields(left, right, operator string) bool {
	if !comparableRuleFields(left, right) || !validCompareOperator(operator) {
		return false
	}
	if operator == FormRuleOperatorEQ || operator == FormRuleOperatorNE {
		return true
	}
	return isOrderedCompareRuleField(left) && isOrderedCompareRuleField(right)
}

func compareRuleFieldFamily(fieldType string) string {
	switch {
	case isNumberRuleField(fieldType):
		return "number"
	case isTextRuleField(fieldType):
		return "text"
	case fieldType == FormFieldTypeSelect || fieldType == FormFieldTypeRadio:
		return "choice"
	case fieldType == FormFieldTypeDate || fieldType == FormFieldTypeDateTime || fieldType == FormFieldTypeTime:
		return "temporal:" + fieldType
	case fieldType == FormFieldTypeBoolean || fieldType == FormFieldTypeUser || fieldType == FormFieldTypeDepartment:
		return fieldType
	default:
		return ""
	}
}

func isOrderedCompareRuleField(fieldType string) bool {
	return isNumberRuleField(fieldType) || fieldType == FormFieldTypeDate ||
		fieldType == FormFieldTypeDateTime || fieldType == FormFieldTypeTime
}

func detailListColumn(field FormField, key string) (FormField, bool) {
	key = strings.TrimSpace(key)
	for _, column := range field.Columns {
		if strings.TrimSpace(column.Key) == key {
			return column, true
		}
	}
	return FormField{}, false
}

func isFormLayoutFieldType(fieldType string) bool {
	return fieldType == FormFieldTypeGroup ||
		fieldType == FormFieldTypeLabel ||
		fieldType == FormFieldTypeDescription ||
		fieldType == FormFieldTypeButton
}

func validFormFieldSpan(span int) bool {
	return span == 0 || span == 6 || span == 8 || span == 12 || span == 24
}

func validFormOptions(options []FormOption) bool {
	if len(options) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(options))
	return validFormOptionNodes(options, seen)
}

func validFormOptionNodes(options []FormOption, seen map[string]struct{}) bool {
	for _, option := range options {
		value := strings.TrimSpace(option.Value)
		if strings.TrimSpace(option.Label) == "" || value == "" {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
		if len(option.Children) > 0 && !validFormOptionNodes(option.Children, seen) {
			return false
		}
	}
	return true
}

func isFormOptionFieldType(fieldType string) bool {
	return fieldType == FormFieldTypeSelect ||
		fieldType == FormFieldTypeMultiSelect ||
		fieldType == FormFieldTypeRadio ||
		fieldType == FormFieldTypeCheckbox
}

func validFormFieldOptionSource(field FormField) bool {
	if field.OptionSource == nil {
		return true
	}
	if !isFormOptionFieldType(field.Type) {
		return false
	}
	switch normalizedOptionSourceType(field.OptionSource.Type) {
	case OptionSourceStatic:
		return true
	case OptionSourceAPI:
		return validBackendOptionSourceURL(field.OptionSource.URL) &&
			validOptionSourceMethod(field.OptionSource.Method) &&
			strings.TrimSpace(field.OptionSource.LabelField) != "" &&
			strings.TrimSpace(field.OptionSource.ValueField) != ""
	default:
		return false
	}
}

func formFieldUsesAPIOptions(field FormField) bool {
	return field.OptionSource != nil && normalizedOptionSourceType(field.OptionSource.Type) == OptionSourceAPI
}

func normalizedOptionSourceType(sourceType string) string {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	if sourceType == "" {
		return OptionSourceStatic
	}
	return sourceType
}

func validBackendOptionSourceURL(rawURL string) bool {
	optionURL := strings.TrimSpace(rawURL)
	if optionURL == "" || strings.Contains(optionURL, "://") || strings.HasPrefix(optionURL, "//") {
		return false
	}
	if strings.ContainsAny(optionURL, " \t\r\n") {
		return false
	}
	return strings.HasPrefix(optionURL, "/api/")
}

func validOptionSourceMethod(method string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	return method == "" || method == "GET" || method == "POST"
}

func validateFieldPermissions(node Node, fields map[string]FormField) []ValidationError {
	errors := make([]ValidationError, 0)
	seen := make(map[string]struct{}, len(node.FormPermissions))
	for _, permission := range node.FormPermissions {
		field := strings.TrimSpace(permission.Field)
		formField, ok := fields[field]
		if !ok {
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionField, Message: "节点字段权限引用了不存在的字段", NodeID: node.ID})
			continue
		}
		if _, exists := seen[field]; exists {
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionDuplicate, Message: "节点字段权限重复", NodeID: node.ID})
			continue
		}
		seen[field] = struct{}{}
		switch permission.Access {
		case FieldAccessHidden, FieldAccessRead, FieldAccessWrite:
		default:
			errors = append(errors, ValidationError{Code: ValidationFieldPermissionAccess, Message: "节点字段权限无效", NodeID: node.ID})
		}
		if len(permission.Actions) > 0 {
			if formField.Type != FormFieldTypeDetailList || permission.Access != FieldAccessWrite {
				errors = append(errors, ValidationError{Code: ValidationFieldPermissionAction, Message: "行级动作权限只能配置在可编辑明细列表字段上", NodeID: node.ID})
				continue
			}
			if !validFieldPermissionActions(permission.Actions) {
				errors = append(errors, ValidationError{Code: ValidationFieldPermissionAction, Message: "节点字段行级动作权限无效", NodeID: node.ID})
			}
		}
	}
	return errors
}

func validFieldPermissionActions(actions []string) bool {
	seen := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		action = strings.TrimSpace(action)
		switch action {
		case FieldActionAdd, FieldActionDelete:
			if _, exists := seen[action]; exists {
				return false
			}
			seen[action] = struct{}{}
		default:
			return false
		}
	}
	return true
}

func validateApproval(node Node) []ValidationError {
	errors := make([]ValidationError, 0, 2)
	switch node.ApprovalMode {
	case ApprovalModeSingle, ApprovalModeSequential, ApprovalModeParallel, ApprovalModeCountersign:
	default:
		errors = append(errors, ValidationError{Code: ValidationApprovalMode, Message: "审批方式无效", NodeID: node.ID})
	}
	errors = append(errors, validateNodeAssignee(node, "审批节点")...)
	errors = append(errors, validateDepartmentApprovalChain(node)...)
	if node.ApprovalMode == ApprovalModeCountersign && (node.CompletionRate < 1 || node.CompletionRate > 100) {
		errors = append(errors, ValidationError{Code: ValidationCompletionRate, Message: "会签通过比例必须在 1 到 100 之间", NodeID: node.ID})
	}
	return errors
}

func validateDepartmentApprovalChain(node Node) []ValidationError {
	config := node.DepartmentApprovalChain
	if config == nil || !config.Enabled {
		return nil
	}
	if node.Assignee == nil || node.Assignee.Type != AssigneeTypeOrgIdentity {
		return []ValidationError{{Code: ValidationDepartmentApprovalChain, Message: "逐级部门审批必须使用组织审批身份", NodeID: node.ID}}
	}
	if !usesStarterDepartmentOrgIdentity(node.Assignee.Value) {
		return []ValidationError{{Code: ValidationDepartmentApprovalChain, Message: "逐级部门审批必须从发起人部门开始解析", NodeID: node.ID}}
	}
	if config.StopMode != DepartmentApprovalChainStopRoot && config.StopMode != DepartmentApprovalChainStopDepartment {
		return []ValidationError{{Code: ValidationDepartmentApprovalChain, Message: "逐级部门审批终止范围无效", NodeID: node.ID}}
	}
	if config.StopMode == DepartmentApprovalChainStopDepartment && config.StopDepartmentID == 0 {
		return []ValidationError{{Code: ValidationDepartmentApprovalChain, Message: "逐级部门审批必须选择终止部门", NodeID: node.ID}}
	}
	if config.MissingAssigneePolicy != DepartmentApprovalChainMissingSkip && config.MissingAssigneePolicy != DepartmentApprovalChainMissingError {
		return []ValidationError{{Code: ValidationDepartmentApprovalChain, Message: "逐级部门审批的负责人缺失策略无效", NodeID: node.ID}}
	}
	return nil
}

func usesStarterDepartmentOrgIdentity(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]) != ""
	}
	return len(parts) == 2 && strings.TrimSpace(parts[0]) == "starter_department" && strings.TrimSpace(parts[1]) != ""
}

func validateNodeAssignee(node Node, label string) []ValidationError {
	if node.Assignee == nil || strings.TrimSpace(node.Assignee.Type) == "" {
		return []ValidationError{{Code: ValidationAssigneeRequired, Message: label + "必须配置处理人", NodeID: node.ID}}
	}
	if node.Assignee.Type == AssigneeTypeInitiator {
		return nil
	}
	if strings.TrimSpace(node.Assignee.Value) == "" {
		return []ValidationError{{Code: ValidationAssigneeRequired, Message: label + "必须配置处理人", NodeID: node.ID}}
	}
	switch node.Assignee.Type {
	case AssigneeTypeUser, AssigneeTypeRole, AssigneeTypeDepartmentLeader, AssigneeTypeManager, AssigneeTypeVariable, AssigneeTypeOrgIdentity:
		return nil
	default:
		return []ValidationError{{Code: ValidationAssigneeRequired, Message: label + "处理人类型无效", NodeID: node.ID}}
	}
}

func validateAutomation(node Node) []ValidationError {
	if node.Automation == nil || node.Automation.Type != AutomationTypeSetVariables || len(node.Automation.Variables) == 0 {
		return []ValidationError{{Code: ValidationAutomation, Message: "自动动作必须配置变量写入", NodeID: node.ID}}
	}
	for key := range node.Automation.Variables {
		if !formFieldKeyPattern.MatchString(strings.TrimSpace(key)) {
			return []ValidationError{{Code: ValidationAutomation, Message: "自动动作变量名无效", NodeID: node.ID}}
		}
	}
	return nil
}

func validateTimer(node Node) []ValidationError {
	if node.Timer == nil || node.Timer.DelaySeconds < 1 || node.Timer.DelaySeconds > 31536000 {
		return []ValidationError{{Code: ValidationTimer, Message: "定时等待时长必须在 1 到 31536000 秒之间", NodeID: node.ID}}
	}
	return nil
}

func validateOptionalNotification(node Node) []ValidationError {
	if node.Notification == nil || !node.Notification.Enabled {
		return nil
	}
	return validateNotificationConfig(node, node.Notification, false)
}

func validateOptionalResultNotification(node Node) []ValidationError {
	if node.ResultNotification == nil || !node.ResultNotification.Enabled {
		return nil
	}
	return validateNotificationConfig(node, node.ResultNotification, true)
}

func validateRequiredNotification(node Node) []ValidationError {
	if node.Notification == nil || !node.Notification.Enabled {
		return []ValidationError{{Code: ValidationNotification, Message: "通知节点必须启用通知", NodeID: node.ID}}
	}
	return validateNotification(node)
}

func validateNotification(node Node) []ValidationError {
	return validateNotificationConfig(node, node.Notification, false)
}

func validateNotificationConfig(node Node, config *NotificationConfig, allowResult bool) []ValidationError {
	if config == nil {
		return []ValidationError{{Code: ValidationNotification, Message: "通知配置不能为空", NodeID: node.ID}}
	}
	seen := make(map[string]struct{}, len(config.Channels))
	for _, raw := range config.Channels {
		channel := strings.TrimSpace(raw)
		if channel != NotificationChannelInApp && channel != NotificationChannelDingTalkOA {
			return []ValidationError{{Code: ValidationNotification, Message: "通知渠道无效", NodeID: node.ID}}
		}
		if _, exists := seen[channel]; exists {
			return []ValidationError{{Code: ValidationNotification, Message: "通知渠道不能重复", NodeID: node.ID}}
		}
		seen[channel] = struct{}{}
	}
	if len(seen) == 0 {
		return []ValidationError{{Code: ValidationNotification, Message: "通知渠道不能为空", NodeID: node.ID}}
	}
	title := strings.TrimSpace(config.Title)
	content := strings.TrimSpace(config.Content)
	if title == "" || utf8.RuneCountInString(title) > 256 || content == "" || utf8.RuneCountInString(content) > 2000 {
		return []ValidationError{{Code: ValidationNotification, Message: "通知标题或正文长度无效", NodeID: node.ID}}
	}
	extraTokens := []string(nil)
	if allowResult {
		extraTokens = append(extraTokens, "result")
	}
	if !validNotificationTemplate(title, extraTokens...) || !validNotificationTemplate(content, extraTokens...) {
		return []ValidationError{{Code: ValidationNotification, Message: "通知模板包含不支持的占位符", NodeID: node.ID}}
	}
	if !allowResult && len(config.ResultTypes) > 0 {
		return []ValidationError{{Code: ValidationNotification, Message: "非结果通知不能配置结果类型", NodeID: node.ID}}
	}
	if allowResult && config.ResultTypes != nil {
		if len(config.ResultTypes) == 0 {
			return []ValidationError{{Code: ValidationNotification, Message: "结果通知至少选择一种通知结果", NodeID: node.ID}}
		}
		seenResults := make(map[string]struct{}, len(config.ResultTypes))
		for _, resultType := range config.ResultTypes {
			if resultType != NotificationResultApproved && resultType != NotificationResultRejected && resultType != NotificationResultReturned {
				return []ValidationError{{Code: ValidationNotification, Message: "结果通知类型无效", NodeID: node.ID}}
			}
			if _, exists := seenResults[resultType]; exists {
				return []ValidationError{{Code: ValidationNotification, Message: "结果通知类型不能重复", NodeID: node.ID}}
			}
			seenResults[resultType] = struct{}{}
		}
	}
	return nil
}

func validNotificationTemplate(value string, extraTokens ...string) bool {
	allowed := map[string]struct{}{
		"workflowName": {}, "nodeName": {}, "starterName": {}, "instanceId": {}, "taskId": {},
	}
	for _, token := range extraTokens {
		allowed[token] = struct{}{}
	}
	valid := true
	remainder := notificationTemplateTokenPattern.ReplaceAllStringFunc(value, func(token string) string {
		matches := notificationTemplateTokenPattern.FindStringSubmatch(token)
		if len(matches) != 2 {
			valid = false
			return ""
		}
		if _, ok := allowed[strings.TrimSpace(matches[1])]; !ok || strings.TrimSpace(matches[1]) != matches[1] {
			valid = false
		}
		return ""
	})
	return valid && !strings.Contains(remainder, "{{") && !strings.Contains(remainder, "}}")
}

func validateGateway(node Node, incoming, outgoing []Edge) []ValidationError {
	errors := make([]ValidationError, 0)
	if node.GatewayMode == GatewayModeSplit && len(outgoing) < 2 {
		errors = append(errors, ValidationError{Code: ValidationBranchCount, Message: "分支网关至少需要两条分支", NodeID: node.ID})
	}
	if node.GatewayMode == GatewayModeJoin && len(incoming) < 2 {
		errors = append(errors, ValidationError{Code: ValidationBranchCount, Message: "汇聚网关至少需要两条进入连线", NodeID: node.ID})
	}
	if node.Type != NodeTypeExclusive || node.GatewayMode != GatewayModeSplit {
		return errors
	}
	defaultCount := 0
	for _, edge := range outgoing {
		if edge.Default {
			defaultCount++
			continue
		}
		if edge.Condition == nil {
			errors = append(errors, ValidationError{Code: ValidationBranchConditionRequired, Message: "条件分支必须配置条件或设为默认分支", NodeID: node.ID, EdgeID: edge.ID})
			continue
		}
		if !validCondition(*edge.Condition) {
			errors = append(errors, ValidationError{Code: ValidationConditionInvalid, Message: "条件表达式不完整", NodeID: node.ID, EdgeID: edge.ID})
		}
	}
	if defaultCount > 1 {
		errors = append(errors, ValidationError{Code: ValidationMultipleDefaultBranches, Message: "条件网关只能有一条默认分支", NodeID: node.ID})
	}
	return errors
}

func validCondition(condition Condition) bool {
	if strings.TrimSpace(condition.Field) == "" || condition.Value == nil {
		return false
	}
	switch condition.Operator {
	case ConditionEQ, ConditionNE, ConditionGT, ConditionGTE, ConditionLT, ConditionLTE:
		return true
	default:
		return false
	}
}

func visitForward(start string, outgoing map[string][]Edge) map[string]bool {
	visited := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, edge := range outgoing[current] {
			stack = append(stack, edge.Target)
		}
	}
	return visited
}

func visitBackward(end string, incoming map[string][]Edge) map[string]bool {
	visited := map[string]bool{}
	stack := []string{end}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, edge := range incoming[current] {
			stack = append(stack, edge.Source)
		}
	}
	return visited
}
