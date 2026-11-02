package workflowcore

const maxHistoryImportMappings = 100

// CloneHistoryImportConfig preserves disabled editor drafts and normalizes omitted modes.
func CloneHistoryImportConfig(config *HistoryImportConfig) *HistoryImportConfig {
	if config == nil {
		return nil
	}
	cloned := *config
	if cloned.SourceMode == "" {
		cloned.SourceMode = HistoryImportSourceSelectHistory
	}
	if cloned.SourceStatus == "" {
		cloned.SourceStatus = HistoryImportStatusCompleted
	}
	if cloned.MappingMode == "" {
		cloned.MappingMode = HistoryImportMappingSameKey
	}
	cloned.Mappings = make([]HistoryImportMapping, len(config.Mappings))
	for index, mapping := range config.Mappings {
		cloned.Mappings[index] = mapping
		if mapping.Columns != nil {
			cloned.Mappings[index].Columns = make([]HistoryImportColumnMapping, len(mapping.Columns))
			copy(cloned.Mappings[index].Columns, mapping.Columns)
		}
	}
	return &cloned
}

func validateHistoryImport(config *HistoryImportConfig, fields map[string]FormField) []ValidationError {
	if config == nil {
		return nil
	}
	// Even inactive drafts are bounded, but may reference fields removed since editing.
	if len(config.Mappings) > maxHistoryImportMappings || len(config.SourceMode) > 100 || len(config.SourceStatus) > 100 || len(config.MappingMode) > 100 {
		return historyImportError("历史导入规则配置过大，字段映射最多 100 项")
	}
	for _, mapping := range config.Mappings {
		if len(mapping.SourceKey) > 100 || len(mapping.TargetKey) > 100 || len(mapping.Columns) > maxHistoryImportMappings {
			return historyImportError("历史导入字段编码最多 100 字符，明细列映射最多 100 项")
		}
		for _, column := range mapping.Columns {
			if len(column.SourceKey) > 100 || len(column.TargetKey) > 100 {
				return historyImportError("历史导入明细列编码最多 100 字符")
			}
		}
	}
	if !config.Enabled {
		return nil
	}
	config = CloneHistoryImportConfig(config)
	if config.SourceMode != HistoryImportSourceSelectHistory && config.SourceMode != HistoryImportSourceLatestHistory {
		return historyImportError("历史导入来源方式无效")
	}
	if config.SourceStatus != HistoryImportStatusCompleted && config.SourceStatus != HistoryImportStatusAll {
		return historyImportError("历史导入来源状态无效")
	}
	if config.MappingMode != HistoryImportMappingSameKey && config.MappingMode != HistoryImportMappingExplicit {
		return historyImportError("历史导入字段映射方式无效")
	}
	// Switching modes retains the editor draft; same-key imports never use its mappings.
	if config.MappingMode == HistoryImportMappingSameKey {
		return nil
	}
	if len(config.Mappings) == 0 {
		return historyImportError("历史导入字段映射需要 1–100 项")
	}
	sources, targets := make(map[string]bool), make(map[string]bool)
	for _, mapping := range config.Mappings {
		if !validHistoryImportPair(mapping.SourceKey, mapping.TargetKey, sources, targets) {
			return historyImportError("历史导入字段编码无效或源/目标字段重复")
		}
		source, sourceExists := fields[mapping.SourceKey]
		target, targetExists := fields[mapping.TargetKey]
		if !sourceExists || !targetExists {
			return historyImportError("历史导入映射字段不存在：" + mapping.SourceKey + " → " + mapping.TargetKey)
		}
		if !historyImportTypesCompatible(source, target) {
			return historyImportError("历史导入字段类型不兼容或包含计算/展示字段：" + mapping.SourceKey + " → " + mapping.TargetKey)
		}
		if target.Type != FormFieldTypeDetailList {
			if mapping.Columns != nil {
				return historyImportError("历史导入普通字段不能配置明细列映射")
			}
			continue
		}
		if errors := validateHistoryImportColumns(source, target, mapping.Columns); len(errors) > 0 {
			return errors
		}
	}
	return nil
}

func validateHistoryImportColumns(source, target FormField, columns []HistoryImportColumnMapping) []ValidationError {
	if columns != nil && len(columns) == 0 {
		return historyImportError("历史导入明细列映射需要 1–100 项")
	}
	sourceColumns, targetColumns := make(map[string]FormField), make(map[string]FormField)
	for _, column := range source.Columns {
		sourceColumns[column.Key] = column
	}
	for _, column := range target.Columns {
		targetColumns[column.Key] = column
	}
	if columns == nil {
		// The existing importer requires every target column to have a compatible same-key source.
		for _, column := range target.Columns {
			columns = append(columns, HistoryImportColumnMapping{SourceKey: column.Key, TargetKey: column.Key})
		}
	}
	sources, targets := make(map[string]bool), make(map[string]bool)
	for _, mapping := range columns {
		if !validHistoryImportPair(mapping.SourceKey, mapping.TargetKey, sources, targets) {
			return historyImportError("历史导入明细列编码无效或源/目标列重复")
		}
		sourceColumn, sourceExists := sourceColumns[mapping.SourceKey]
		targetColumn, targetExists := targetColumns[mapping.TargetKey]
		if !sourceExists || !targetExists {
			return historyImportError("历史导入明细列不存在：" + mapping.SourceKey + " → " + mapping.TargetKey)
		}
		if sourceColumn.Type == FormFieldTypeDetailList || targetColumn.Type == FormFieldTypeDetailList || !historyImportTypesCompatible(sourceColumn, targetColumn) {
			return historyImportError("历史导入明细列类型不兼容，不支持嵌套明细或计算/展示列")
		}
	}
	return nil
}

func validHistoryImportPair(source, target string, sources, targets map[string]bool) bool {
	if !formFieldKeyPattern.MatchString(source) || !formFieldKeyPattern.MatchString(target) || sources[source] || targets[target] {
		return false
	}
	sources[source], targets[target] = true, true
	return true
}

func historyImportTypesCompatible(source, target FormField) bool {
	if source.Calculation != nil || target.Calculation != nil {
		return false
	}
	sourceFamily, targetFamily := historyImportTypeFamily(source.Type), historyImportTypeFamily(target.Type)
	return sourceFamily != "" && sourceFamily == targetFamily
}

func historyImportTypeFamily(fieldType string) string {
	switch fieldType {
	case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypePhone, FormFieldTypeEmail:
		return FormFieldTypeText
	case FormFieldTypeNumber, FormFieldTypeAmount:
		return FormFieldTypeNumber
	case FormFieldTypeSelect, FormFieldTypeRadio:
		return FormFieldTypeSelect
	case FormFieldTypeMultiSelect, FormFieldTypeCheckbox:
		return FormFieldTypeMultiSelect
	case FormFieldTypeDate, FormFieldTypeDateTime, FormFieldTypeUser, FormFieldTypeDepartment,
		FormFieldTypeAttachment, FormFieldTypeBoolean, FormFieldTypeTime, FormFieldTypeDateRange,
		FormFieldTypeUserMulti, FormFieldTypeDepartmentMulti, FormFieldTypeDetailList:
		return fieldType
	default:
		return ""
	}
}

func historyImportError(message string) []ValidationError {
	return []ValidationError{{Code: ValidationHistoryImport, Message: message}}
}
