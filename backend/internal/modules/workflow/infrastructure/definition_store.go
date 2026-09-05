package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/workflowcore"
)

func (store *GormStore) LoadPublishedDefinition(ctx context.Context, definitionID uint, version int) (workflowcore.Definition, int, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return workflowcore.Definition{}, 0, err
	}
	defer cancel()

	var definitionModel workflowmodel.Definition
	if err := db.Clauses(clause.Locking{Strength: "SHARE"}).First(&definitionModel, definitionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, 0, fmt.Errorf("流程定义不存在: %w", err)
		}
		return workflowcore.Definition{}, 0, err
	}
	if definitionModel.Status != workflowmodel.DefinitionStatusPublished || definitionModel.CurrentVersion < 1 {
		return workflowcore.Definition{}, 0, ErrDefinitionNotPublished
	}
	if version <= 0 {
		version = definitionModel.CurrentVersion
	}
	return loadDefinitionVersion(db, definitionID, version)
}

func (store *GormStore) ListPublishedDefinitions(ctx context.Context) ([]application.PublishedDefinition, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var rows []workflowmodel.Definition
	if err := db.Where("definition_status = ? AND definition_current_version > 0", workflowmodel.DefinitionStatusPublished).
		Order("definition_category ASC").Order("definition_name ASC").Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]application.PublishedDefinition, 0, len(rows))
	for _, row := range rows {
		row.LogoURL = media.FullURLWithStaticDomainContext(ctx, row.LogoURL)
		definition, version, err := loadDefinitionVersion(db, row.ID, row.CurrentVersion)
		if err != nil {
			return nil, err
		}
		result = append(result, publishedDefinition(row, definition, version, publishedAssigneeLabels{}, false))
	}
	return result, nil
}

func (store *GormStore) GetPublishedDefinition(ctx context.Context, definitionID uint) (*application.PublishedDefinition, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row workflowmodel.Definition
	if err := db.First(&row, "id = ? AND definition_status = ? AND definition_current_version > 0", definitionID, workflowmodel.DefinitionStatusPublished).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefinitionNotPublished
		}
		return nil, err
	}
	definition, version, err := loadDefinitionVersion(db, row.ID, row.CurrentVersion)
	if err != nil {
		return nil, err
	}
	row.LogoURL = media.FullURLWithStaticDomainContext(ctx, row.LogoURL)
	labels, err := loadPublishedAssigneeLabels(db, definition.Nodes)
	if err != nil {
		return nil, err
	}
	result := publishedDefinition(row, definition, version, labels, true)
	return &result, nil
}

func loadDefinitionVersion(db *gorm.DB, definitionID uint, version int) (workflowcore.Definition, int, error) {
	var versionModel workflowmodel.DefinitionVersion
	if err := db.First(&versionModel, "definition_id = ? AND definition_version = ?", definitionID, version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, 0, fmt.Errorf("流程定义版本不存在: %w", err)
		}
		return workflowcore.Definition{}, 0, err
	}
	var definition workflowcore.Definition
	if err := json.Unmarshal([]byte(versionModel.SourceJSON), &definition); err != nil {
		return workflowcore.Definition{}, 0, fmt.Errorf("解析流程定义版本失败: %w", err)
	}
	if validationErrors := workflowcore.ValidateDefinition(definition); len(validationErrors) > 0 {
		return workflowcore.Definition{}, 0, workflowcore.ValidationErrors(validationErrors)
	}
	return definition, versionModel.Version, nil
}

func publishedDefinition(row workflowmodel.Definition, definition workflowcore.Definition, version int, labels publishedAssigneeLabels, includeGraph bool) application.PublishedDefinition {
	result := application.PublishedDefinition{
		ID: row.ID, Key: row.Key, Name: row.Name, Description: row.Description,
		Category: row.Category, LogoURL: row.LogoURL, Version: version, Form: cloneDefinitionForm(definition),
		FieldPermissions: definitionFieldPermissions(definition), StartNodeID: definitionStartNodeID(definition),
		Initiator: definitionInitiator(definition), Availability: definitionStartAvailability(definition),
		StartLimit: definitionStartLimit(definition), StartLimitStatus: application.StartLimitStatus{Allowed: true},
	}
	if includeGraph {
		result.Nodes, result.Edges = buildPublishedWorkflowGraph(definition, labels)
	}
	return result
}

func definitionInitiator(definition workflowcore.Definition) workflowcore.InitiatorConfig {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart && node.Initiator != nil {
			return workflowcore.InitiatorConfig{
				Scope:           node.Initiator.Scope,
				UserIDs:         append([]uint(nil), node.Initiator.UserIDs...),
				DepartmentIDs:   append([]uint(nil), node.Initiator.DepartmentIDs...),
				ExcludedUserIDs: append([]uint(nil), node.Initiator.ExcludedUserIDs...),
			}
		}
	}
	return workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}
}

func definitionStartAvailability(definition workflowcore.Definition) workflowcore.StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartAvailability(definition.Nodes[index].Availability)
		}
	}
	return workflowcore.DefaultStartAvailability()
}

func definitionStartLimit(definition workflowcore.Definition) workflowcore.StartLimitConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartLimit(definition.Nodes[index].StartLimit)
		}
	}
	return workflowcore.DefaultStartLimit()
}

func cloneDefinitionForm(definition workflowcore.Definition) []workflowcore.FormField {
	return cloneFormFields(definition.Form)
}

func cloneFormFields(fields []workflowcore.FormField) []workflowcore.FormField {
	if len(fields) == 0 {
		return nil
	}
	result := make([]workflowcore.FormField, 0, len(fields))
	for _, field := range fields {
		field.Options = cloneFormOptions(field.Options)
		if field.OptionSource != nil {
			optionSource := *field.OptionSource
			field.OptionSource = &optionSource
		}
		field.Columns = cloneFormFields(field.Columns)
		field.Fields = cloneFormFields(field.Fields)
		if field.Help != nil {
			help := *field.Help
			field.Help = &help
		}
		result = append(result, field)
	}
	return result
}

func cloneFormOptions(options []workflowcore.FormOption) []workflowcore.FormOption {
	if len(options) == 0 {
		return nil
	}
	result := make([]workflowcore.FormOption, 0, len(options))
	for _, option := range options {
		option.Children = cloneFormOptions(option.Children)
		result = append(result, option)
	}
	return result
}

func definitionFieldPermissions(definition workflowcore.Definition) map[string][]workflowcore.FieldPermission {
	result := make(map[string][]workflowcore.FieldPermission)
	for _, node := range definition.Nodes {
		if len(node.FormPermissions) == 0 {
			continue
		}
		permissions := make([]workflowcore.FieldPermission, 0, len(node.FormPermissions))
		for _, permission := range node.FormPermissions {
			permission.Actions = append([]string(nil), permission.Actions...)
			permissions = append(permissions, permission)
		}
		result[node.ID] = permissions
	}
	if result == nil {
		return map[string][]workflowcore.FieldPermission{}
	}
	return result
}

func definitionNodeTypes(definition workflowcore.Definition) map[string]string {
	result := make(map[string]string, len(definition.Nodes))
	for _, node := range definition.Nodes {
		result[node.ID] = node.Type
	}
	return result
}

func definitionStartNodeID(definition workflowcore.Definition) string {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart {
			return node.ID
		}
	}
	return ""
}
