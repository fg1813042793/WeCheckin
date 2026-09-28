package workflowservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
)

type CreateRequest struct {
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	LogoURL     string          `json:"-"`
	Draft       json.RawMessage `json:"draft"`
}

type CopyRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Category    string `json:"category"`
	LogoURL     string `json:"-"`
}

type UpdateRequest struct {
	Name        string          `json:"name"`
	DisplayName *string         `json:"displayName"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	LogoURL     *string         `json:"-"`
	Status      *int            `json:"status"`
	Draft       json.RawMessage `json:"draft"`
}

type DefinitionSummary struct {
	ID                  uint   `json:"id"`
	Key                 string `json:"key"`
	Name                string `json:"name"`
	DisplayName         string `json:"displayName"`
	Description         string `json:"description"`
	Category            string `json:"category"`
	LogoURL             string `json:"logoUrl"`
	Status              int    `json:"status"`
	CurrentVersion      int    `json:"currentVersion"`
	StartConfigRevision int    `json:"startConfigRevision"`
	AddUserID           uint   `json:"addUserId"`
	EditUserID          uint   `json:"editUserId"`
	AddTime             int64  `json:"addTime"`
	EditTime            int64  `json:"editTime"`
}

type DefinitionDetail struct {
	DefinitionSummary
	Draft workflowcore.Definition `json:"draft"`
}

type ListResponse struct {
	List     []DefinitionSummary `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type ValidationResponse struct {
	Valid  bool                           `json:"valid"`
	Errors []workflowcore.ValidationError `json:"errors"`
}

type PublishResponse struct {
	DefinitionID uint   `json:"definitionId"`
	Version      int    `json:"version"`
	BPMNXML      string `json:"bpmnXml"`
}

type PublishRequest struct {
	Initiator *workflowcore.InitiatorConfig `json:"initiator,omitempty"`
	Note      string                        `json:"note"`
}

func GetListContext(ctx context.Context, keyword, category string, status, page, pageSize int) (*ListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := activeWorkflowDefinitionQuery(db).Model(&model.WorkflowDefinition{})
	if text := strings.TrimSpace(keyword); text != "" {
		like := "%" + text + "%"
		query = query.Where("definition_name LIKE ? OR definition_display_name LIKE ? OR definition_key LIKE ?", like, like, like)
	}
	if text := strings.TrimSpace(category); text != "" {
		query = query.Where("definition_category = ?", text)
	}
	if status >= 0 {
		query = query.Where("definition_status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.WorkflowDefinition
	if err := query.Order("definition_edit_time DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]DefinitionSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, summaryFromModel(ctx, row))
	}
	return &ListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func GetDetailContext(ctx context.Context, id uint) (*DefinitionDetail, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var item model.WorkflowDefinition
	if err := activeWorkflowDefinitionQuery(db).First(&item, id).Error; err != nil {
		return nil, definitionError(err)
	}
	draft, _, err := normalizeDraft(json.RawMessage(item.DraftJSON), item.Key, item.Name, item.DisplayName)
	if err != nil {
		return nil, err
	}
	if err := workflowcore.ApplyRuntimeStartConfigJSON(&draft, item.StartConfigJSON); err != nil {
		return nil, errors.New("流程运行配置格式无效：" + err.Error())
	}
	return &DefinitionDetail{DefinitionSummary: summaryFromModel(ctx, item), Draft: draft}, nil
}

func CreateContext(ctx context.Context, adminID uint, request CreateRequest) (*DefinitionDetail, error) {
	request.Key = strings.TrimSpace(request.Key)
	request.Name = strings.TrimSpace(request.Name)
	if request.Key == "" || request.Name == "" {
		return nil, errors.New("流程编码和流程名称不能为空")
	}
	displayName, err := normalizeDefinitionDisplayName(request.DisplayName)
	if err != nil {
		return nil, err
	}
	request.DisplayName = displayName
	logoURL, err := normalizeDefinitionLogoURL(request.LogoURL)
	if err != nil {
		return nil, err
	}
	draftInput := request.Draft
	if len(bytes.TrimSpace(draftInput)) == 0 {
		initial := newDefaultDefinition(request.Key, request.Name)
		initial.DisplayName = request.DisplayName
		encoded, err := json.Marshal(initial)
		if err != nil {
			return nil, err
		}
		draftInput = encoded
	}
	draft, encoded, err := normalizeDraft(draftInput, request.Key, request.Name, request.DisplayName)
	if err != nil {
		return nil, err
	}
	if validationErrors := workflowcore.ValidateDefinition(draft); len(validationErrors) > 0 {
		for _, validationError := range validationErrors {
			if validationError.Code == workflowcore.ValidationDefinitionKey || validationError.Code == workflowcore.ValidationSchemaVersion {
				return nil, validationErrorAsError(validationError)
			}
		}
	}
	startConfigJSON, err := workflowcore.EncodeRuntimeStartConfig(draft)
	if err != nil {
		return nil, err
	}
	if validationErrors := workflowcore.ValidateRuntimeStartConfig(draft); len(validationErrors) > 0 {
		return nil, validationErrorAsError(validationErrors[0])
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var duplicate int64
	if err := db.Model(&model.WorkflowDefinition{}).Where("definition_key = ?", request.Key).Count(&duplicate).Error; err != nil {
		return nil, err
	}
	if duplicate > 0 {
		return nil, errors.New("流程编码已存在")
	}
	item := definitionModelForCreate(adminID, request, encoded, startConfigJSON, logoURL, database.Now())
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &DefinitionDetail{DefinitionSummary: summaryFromModel(ctx, item), Draft: draft}, nil
}

func CopyContext(ctx context.Context, adminID, sourceID uint, request CopyRequest) (*DefinitionDetail, error) {
	if sourceID == 0 {
		return nil, errors.New("源流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var source model.WorkflowDefinition
	if err := activeWorkflowDefinitionQuery(db).First(&source, sourceID).Error; err != nil {
		return nil, definitionError(err)
	}
	if len(bytes.TrimSpace([]byte(source.DraftJSON))) == 0 {
		return nil, errors.New("源流程设计数据为空")
	}
	sourceDraft, _, err := normalizeDraft(json.RawMessage(source.DraftJSON), source.Key, source.Name, source.DisplayName)
	if err != nil {
		return nil, err
	}
	if err := workflowcore.ApplyRuntimeStartConfigJSON(&sourceDraft, source.StartConfigJSON); err != nil {
		return nil, errors.New("流程运行配置格式无效：" + err.Error())
	}
	sourceDraftJSON, err := json.Marshal(sourceDraft)
	if err != nil {
		return nil, err
	}
	source.DraftJSON = string(sourceDraftJSON)
	return CreateContext(ctx, adminID, createRequestForCopy(source, request))
}

func createRequestForCopy(source model.WorkflowDefinition, request CopyRequest) CreateRequest {
	return CreateRequest{
		Key:         request.Key,
		Name:        request.Name,
		DisplayName: request.DisplayName,
		Description: request.Description,
		Category:    request.Category,
		LogoURL:     request.LogoURL,
		Draft:       json.RawMessage(source.DraftJSON),
	}
}

func definitionModelForCreate(adminID uint, request CreateRequest, draftJSON, startConfigJSON, logoURL string, now int64) model.WorkflowDefinition {
	return model.WorkflowDefinition{
		Key:                 request.Key,
		Name:                request.Name,
		DisplayName:         request.DisplayName,
		Description:         strings.TrimSpace(request.Description),
		Category:            strings.TrimSpace(request.Category),
		LogoURL:             logoURL,
		Status:              model.DefinitionStatusDraft,
		CurrentVersion:      0,
		DraftJSON:           draftJSON,
		StartConfigJSON:     startConfigJSON,
		StartConfigRevision: 1,
		AddUserID:           adminID,
		EditUserID:          adminID,
		AddTime:             now,
		EditTime:            now,
	}
}

func UpdateContext(ctx context.Context, adminID, id uint, request UpdateRequest) (*DefinitionDetail, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var item model.WorkflowDefinition
	if err := activeWorkflowDefinitionQuery(db).First(&item, id).Error; err != nil {
		return nil, definitionError(err)
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = item.Name
	}
	displayName := item.DisplayName
	if request.DisplayName != nil {
		normalizedDisplayName, err := normalizeDefinitionDisplayName(*request.DisplayName)
		if err != nil {
			return nil, err
		}
		displayName = normalizedDisplayName
	}
	draftInput := request.Draft
	draftProvided := len(bytes.TrimSpace(draftInput)) > 0
	if !draftProvided {
		draftInput = json.RawMessage(item.DraftJSON)
	}
	draft, encoded, err := normalizeDraft(draftInput, item.Key, name, displayName)
	if err != nil {
		return nil, err
	}
	if !draftProvided {
		if err := workflowcore.ApplyRuntimeStartConfigJSON(&draft, item.StartConfigJSON); err != nil {
			return nil, errors.New("流程运行配置格式无效：" + err.Error())
		}
		encodedBytes, err := json.Marshal(draft)
		if err != nil {
			return nil, err
		}
		encoded = string(encodedBytes)
	}
	startConfigJSON, err := workflowcore.EncodeRuntimeStartConfig(draft)
	if err != nil {
		return nil, err
	}
	if validationErrors := workflowcore.ValidateRuntimeStartConfig(draft); len(validationErrors) > 0 {
		return nil, validationErrorAsError(validationErrors[0])
	}
	if err := validatePublishedRuntimeStartConfig(db, item, startConfigJSON); err != nil {
		return nil, err
	}
	status := item.Status
	if request.Status != nil {
		if *request.Status != model.DefinitionStatusDisabled && *request.Status != model.DefinitionStatusDraft && *request.Status != model.DefinitionStatusPublished {
			return nil, errors.New("流程状态无效")
		}
		if *request.Status == model.DefinitionStatusPublished && item.CurrentVersion == 0 {
			return nil, errors.New("未发布版本的流程不能直接设为已发布")
		}
		status = *request.Status
	}
	description := strings.TrimSpace(request.Description)
	category := strings.TrimSpace(request.Category)
	var logoURL *string
	if request.LogoURL != nil {
		normalized, err := normalizeDefinitionLogoURL(*request.LogoURL)
		if err != nil {
			return nil, err
		}
		logoURL = &normalized
	}
	updates := definitionContentUpdates(item, name, displayName, description, category, status, encoded, startConfigJSON, logoURL)
	if len(updates) > 0 {
		now := database.Now()
		updates["definition_edit_user_id"] = adminID
		updates["definition_edit_time"] = now
		updates["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")
		if err := definitionUpdateSession(db).Model(&model.WorkflowDefinition{}).Where("id = ? AND definition_deleted_at = ?", id, int64(0)).Updates(updates).Error; err != nil {
			return nil, err
		}
		item.EditUserID = adminID
		item.EditTime = now
	}
	item.Name = name
	item.DisplayName = displayName
	item.Description = description
	item.Category = category
	if logoURL != nil {
		item.LogoURL = *logoURL
	}
	item.Status = status
	item.DraftJSON = encoded
	if startConfigJSON != item.StartConfigJSON {
		item.StartConfigJSON = startConfigJSON
		item.StartConfigRevision++
	}
	return &DefinitionDetail{DefinitionSummary: summaryFromModel(ctx, item), Draft: draft}, nil
}

func definitionContentUpdates(item model.WorkflowDefinition, name, displayName, description, category string, status int, draftJSON, startConfigJSON string, logoURL *string) map[string]interface{} {
	updates := make(map[string]interface{}, 7)
	if name != item.Name {
		updates["definition_name"] = name
	}
	if displayName != item.DisplayName {
		updates["definition_display_name"] = displayName
	}
	if description != item.Description {
		updates["definition_description"] = description
	}
	if category != item.Category {
		updates["definition_category"] = category
	}
	if status != item.Status {
		updates["definition_status"] = status
	}
	if draftJSON != item.DraftJSON {
		updates["definition_draft_json"] = draftJSON
	}
	if startConfigJSON != item.StartConfigJSON {
		updates["definition_start_config_json"] = startConfigJSON
		updates["definition_start_config_revision"] = item.StartConfigRevision + 1
	}
	if logoURL != nil && *logoURL != item.LogoURL {
		updates["definition_logo_url"] = *logoURL
	}
	return updates
}

func normalizeDefinitionDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > 200 {
		return "", errors.New("用户显示名称不能超过 200 个字符")
	}
	return value, nil
}

func normalizeDefinitionLogoURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, "/uploads/workflow-logos/") {
		return "", errors.New("流程 Logo 地址无效")
	}
	return value, nil
}

func definitionUpdateSession(db *gorm.DB) *gorm.DB {
	// The primary-key update is already atomic; an outer transaction only adds round trips.
	return db.Session(&gorm.Session{SkipDefaultTransaction: true})
}

func activeWorkflowDefinitionQuery(db *gorm.DB) *gorm.DB {
	return db.Where("definition_deleted_at = ?", int64(0))
}

func UpdateStatusContext(ctx context.Context, adminID, id uint, status int) (*DefinitionSummary, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var result DefinitionSummary
	err := db.Transaction(func(tx *gorm.DB) error {
		var item model.WorkflowDefinition
		if err := activeWorkflowDefinitionQuery(tx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return definitionError(err)
		}
		if err := validateDefinitionStatusChange(item, status); err != nil {
			return err
		}
		if item.Status != status {
			now := database.Now()
			if err := tx.Model(&model.WorkflowDefinition{}).
				Where("id = ? AND definition_deleted_at = ?", id, int64(0)).
				Updates(map[string]interface{}{
					"definition_status":       status,
					"definition_edit_user_id": adminID,
					"definition_edit_time":    now,
					"updated_at":              gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error; err != nil {
				return err
			}
			item.Status = status
			item.EditUserID = adminID
			item.EditTime = now
		}
		result = summaryFromModel(ctx, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func validateDefinitionStatusChange(item model.WorkflowDefinition, status int) error {
	if status != model.DefinitionStatusDisabled && status != model.DefinitionStatusPublished {
		return errors.New("流程状态只能设置为启用或关闭")
	}
	if item.CurrentVersion < 1 {
		return errors.New("未发布流程不能启用或关闭")
	}
	return nil
}

func validatePublishedRuntimeStartConfig(db *gorm.DB, item model.WorkflowDefinition, startConfigJSON string) error {
	if item.CurrentVersion < 1 {
		return nil
	}
	var version model.WorkflowDefinitionVersion
	if err := db.First(&version, "definition_id = ? AND definition_version = ?", item.ID, item.CurrentVersion).Error; err != nil {
		return versionError(err)
	}
	published, _, err := normalizeDraft(json.RawMessage(version.SourceJSON), item.Key, item.Name, item.DisplayName)
	if err != nil {
		return err
	}
	if err := workflowcore.ApplyRuntimeStartConfigJSON(&published, startConfigJSON); err != nil {
		return errors.New("流程运行配置格式无效：" + err.Error())
	}
	if validationErrors := workflowcore.ValidateRuntimeStartConfig(published); len(validationErrors) > 0 {
		return errors.New("流程配置无法应用到当前发布版本：" + validationErrors[0].Message)
	}
	return nil
}

func ValidateContext(ctx context.Context, id uint) (*ValidationResponse, error) {
	detail, err := GetDetailContext(ctx, id)
	if err != nil {
		return nil, err
	}
	errors := validateDesignDraft(detail.Draft)
	if errors == nil {
		errors = make([]workflowcore.ValidationError, 0)
	}
	return &ValidationResponse{Valid: len(errors) == 0, Errors: errors}, nil
}

func validateDesignDraft(definition workflowcore.Definition) []workflowcore.ValidationError {
	return workflowcore.ValidateDefinition(definition)
}

func PublishContext(ctx context.Context, adminID, id uint, request PublishRequest) (*PublishResponse, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	note, err := normalizeVersionPublishNote(request.Note)
	if err != nil {
		return nil, err
	}
	var result PublishResponse
	err = db.Transaction(func(tx *gorm.DB) error {
		var item model.WorkflowDefinition
		if err := activeWorkflowDefinitionQuery(tx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return definitionError(err)
		}
		draft, _, err := normalizeDraft(json.RawMessage(item.DraftJSON), item.Key, item.Name, item.DisplayName)
		if err != nil {
			return err
		}
		if err := workflowcore.ApplyRuntimeStartConfigJSON(&draft, item.StartConfigJSON); err != nil {
			return errors.New("流程运行配置格式无效：" + err.Error())
		}
		applyPublishInitiator(&draft, request.Initiator)
		startConfigJSON, err := workflowcore.EncodeRuntimeStartConfig(draft)
		if err != nil {
			return err
		}
		encodedJSON, err := json.Marshal(draft)
		if err != nil {
			return err
		}
		encoded := string(encodedJSON)
		bpmn, err := workflowcore.CompileBPMN(draft)
		if err != nil {
			return err
		}
		version := item.CurrentVersion + 1
		now := database.Now()
		currentSnapshot := versionSnapshot{Metadata: metadataFromDefinition(item), Definition: draft}
		var previousSnapshot versionSnapshot
		if item.CurrentVersion > 0 {
			var previous model.WorkflowDefinitionVersion
			if err := tx.First(&previous, "definition_id = ? AND definition_version = ?", item.ID, item.CurrentVersion).Error; err != nil {
				return versionError(err)
			}
			var metadataRecorded bool
			previousSnapshot, metadataRecorded, err = versionSnapshotFromModel(previous, metadataFromDefinition(item))
			if err != nil {
				return err
			}
			if err := workflowcore.ApplyRuntimeStartConfigJSON(&previousSnapshot.Definition, startConfigJSON); err != nil {
				return errors.New("流程运行配置格式无效：" + err.Error())
			}
			if metadataRecorded && jsonEqual(previousSnapshot, currentSnapshot) {
				return errors.New("流程内容没有变化，无需重复发布")
			}
		}
		summary := buildVersionChangeSummary(item.CurrentVersion, previousSnapshot, currentSnapshot)
		versionItem, err := newDefinitionVersionModel(item.ID, version, adminID, now, encoded, string(bpmn), currentSnapshot, summary, note, 0)
		if err != nil {
			return err
		}
		if err := tx.Create(&versionItem).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"definition_current_version": version,
			"definition_status":          model.DefinitionStatusPublished,
			"definition_draft_json":      encoded,
			"definition_edit_user_id":    adminID,
			"definition_edit_time":       now,
			"updated_at":                 gorm.Expr("CURRENT_TIMESTAMP"),
		}
		if startConfigJSON != item.StartConfigJSON {
			updates["definition_start_config_json"] = startConfigJSON
			updates["definition_start_config_revision"] = item.StartConfigRevision + 1
		}
		if err := tx.Model(&model.WorkflowDefinition{}).Where("id = ? AND definition_deleted_at = ?", item.ID, int64(0)).Updates(updates).Error; err != nil {
			return err
		}
		result = PublishResponse{DefinitionID: item.ID, Version: version, BPMNXML: string(bpmn)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func applyPublishInitiator(definition *workflowcore.Definition, requested *workflowcore.InitiatorConfig) {
	if definition == nil {
		return
	}
	for index := range definition.Nodes {
		if definition.Nodes[index].Type != workflowcore.NodeTypeStart {
			continue
		}
		initiator := requested
		if initiator == nil {
			initiator = definition.Nodes[index].Initiator
		}
		if initiator == nil {
			initiator = &workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}
		}
		definition.Nodes[index].Initiator = &workflowcore.InitiatorConfig{
			Scope:           strings.TrimSpace(initiator.Scope),
			UserIDs:         append([]uint(nil), initiator.UserIDs...),
			DepartmentIDs:   append([]uint(nil), initiator.DepartmentIDs...),
			ExcludedUserIDs: append([]uint(nil), initiator.ExcludedUserIDs...),
		}
		return
	}
}

func DeleteContext(ctx context.Context, adminID, id uint) error {
	if id == 0 {
		return errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var item model.WorkflowDefinition
		if err := activeWorkflowDefinitionQuery(tx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return definitionError(err)
		}
		return tx.Model(&model.WorkflowDefinition{}).
			Where("id = ? AND definition_deleted_at = ?", id, int64(0)).
			Updates(workflowDefinitionDeleteUpdates(adminID, database.Now())).Error
	})
}

func workflowDefinitionDeleteUpdates(adminID uint, deletedAt int64) map[string]interface{} {
	return map[string]interface{}{
		"definition_status":       model.DefinitionStatusDisabled,
		"definition_deleted_at":   deletedAt,
		"definition_deleted_by":   adminID,
		"definition_edit_user_id": adminID,
		"definition_edit_time":    deletedAt,
		"updated_at":              gorm.Expr("CURRENT_TIMESTAMP"),
	}
}

func newDefaultDefinition(key, name string) workflowcore.Definition {
	return workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           strings.TrimSpace(key),
		Name:          strings.TrimSpace(name),
		Nodes: []workflowcore.Node{
			{
				ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始",
				Initiator: &workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
				Availability: &workflowcore.StartAvailabilityConfig{
					Mode: workflowcore.StartAvailabilityAlways, Timezone: workflowcore.DefaultStartAvailabilityTimezone,
				},
				StartLimit: &workflowcore.StartLimitConfig{Mode: workflowcore.StartLimitModeUnlimited},
			},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{{ID: "flow_start_end", Source: "start", Target: "end"}},
	}
}

func normalizeDraft(raw json.RawMessage, key, name string, displayNames ...string) (workflowcore.Definition, string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	var input struct {
		workflowcore.Definition
		LegacyInitiator *workflowcore.InitiatorConfig `json:"initiator,omitempty"`
	}
	if err := decoder.Decode(&input); err != nil {
		return workflowcore.Definition{}, "", errors.New("流程设计数据格式无效：" + err.Error())
	}
	definition := input.Definition
	if input.LegacyInitiator != nil {
		for index := range definition.Nodes {
			if definition.Nodes[index].Type != workflowcore.NodeTypeStart {
				continue
			}
			if definition.Nodes[index].Initiator == nil {
				initiator := *input.LegacyInitiator
				initiator.UserIDs = append([]uint(nil), input.LegacyInitiator.UserIDs...)
				initiator.DepartmentIDs = append([]uint(nil), input.LegacyInitiator.DepartmentIDs...)
				initiator.ExcludedUserIDs = append([]uint(nil), input.LegacyInitiator.ExcludedUserIDs...)
				definition.Nodes[index].Initiator = &initiator
			}
			break
		}
	}
	definition.Key = strings.TrimSpace(key)
	definition.Name = strings.TrimSpace(name)
	if len(displayNames) > 0 {
		definition.DisplayName = strings.TrimSpace(displayNames[0])
	}
	encoded, err := json.Marshal(definition)
	if err != nil {
		return workflowcore.Definition{}, "", err
	}
	return definition, string(encoded), nil
}

func summaryFromModel(ctx context.Context, item model.WorkflowDefinition) DefinitionSummary {
	return DefinitionSummary{
		ID: item.ID, Key: item.Key, Name: item.Name, DisplayName: item.DisplayName, Description: item.Description,
		Category: item.Category, LogoURL: media.FullURLWithStaticDomainContext(ctx, item.LogoURL), Status: item.Status, CurrentVersion: item.CurrentVersion,
		StartConfigRevision: item.StartConfigRevision,
		AddUserID:           item.AddUserID, EditUserID: item.EditUserID, AddTime: item.AddTime, EditTime: item.EditTime,
	}
}

func definitionError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("流程定义不存在")
	}
	return err
}

func validationErrorAsError(validationError workflowcore.ValidationError) error {
	return errors.New(validationError.Message)
}
