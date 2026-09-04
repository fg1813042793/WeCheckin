package workflowservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"

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
	Description string          `json:"description"`
	Category    string          `json:"category"`
	LogoURL     string          `json:"-"`
	Draft       json.RawMessage `json:"draft"`
}

type CopyRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	LogoURL     string `json:"-"`
}

type UpdateRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	LogoURL     *string         `json:"-"`
	Status      *int            `json:"status"`
	Draft       json.RawMessage `json:"draft"`
}

type DefinitionSummary struct {
	ID             uint   `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	LogoURL        string `json:"logoUrl"`
	Status         int    `json:"status"`
	CurrentVersion int    `json:"currentVersion"`
	AddUserID      uint   `json:"addUserId"`
	EditUserID     uint   `json:"editUserId"`
	AddTime        int64  `json:"addTime"`
	EditTime       int64  `json:"editTime"`
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
	query := db.Model(&model.WorkflowDefinition{})
	if text := strings.TrimSpace(keyword); text != "" {
		like := "%" + text + "%"
		query = query.Where("definition_name LIKE ? OR definition_key LIKE ?", like, like)
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
	if err := db.First(&item, id).Error; err != nil {
		return nil, definitionError(err)
	}
	draft, _, err := normalizeDraft(json.RawMessage(item.DraftJSON), item.Key, item.Name)
	if err != nil {
		return nil, err
	}
	return &DefinitionDetail{DefinitionSummary: summaryFromModel(ctx, item), Draft: draft}, nil
}

func CreateContext(ctx context.Context, adminID uint, request CreateRequest) (*DefinitionDetail, error) {
	request.Key = strings.TrimSpace(request.Key)
	request.Name = strings.TrimSpace(request.Name)
	if request.Key == "" || request.Name == "" {
		return nil, errors.New("流程编码和流程名称不能为空")
	}
	logoURL, err := normalizeDefinitionLogoURL(request.LogoURL)
	if err != nil {
		return nil, err
	}
	draftInput := request.Draft
	if len(bytes.TrimSpace(draftInput)) == 0 {
		initial := newDefaultDefinition(request.Key, request.Name)
		encoded, err := json.Marshal(initial)
		if err != nil {
			return nil, err
		}
		draftInput = encoded
	}
	draft, encoded, err := normalizeDraft(draftInput, request.Key, request.Name)
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

	db, cancel := database.WithContext(ctx)
	defer cancel()
	var duplicate int64
	if err := db.Model(&model.WorkflowDefinition{}).Where("definition_key = ?", request.Key).Count(&duplicate).Error; err != nil {
		return nil, err
	}
	if duplicate > 0 {
		return nil, errors.New("流程编码已存在")
	}
	item := definitionModelForCreate(adminID, request, encoded, logoURL, database.Now())
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
	if err := db.First(&source, sourceID).Error; err != nil {
		return nil, definitionError(err)
	}
	if len(bytes.TrimSpace([]byte(source.DraftJSON))) == 0 {
		return nil, errors.New("源流程设计数据为空")
	}
	return CreateContext(ctx, adminID, createRequestForCopy(source, request))
}

func createRequestForCopy(source model.WorkflowDefinition, request CopyRequest) CreateRequest {
	return CreateRequest{
		Key:         request.Key,
		Name:        request.Name,
		Description: request.Description,
		Category:    request.Category,
		LogoURL:     request.LogoURL,
		Draft:       json.RawMessage(source.DraftJSON),
	}
}

func definitionModelForCreate(adminID uint, request CreateRequest, draftJSON, logoURL string, now int64) model.WorkflowDefinition {
	return model.WorkflowDefinition{
		Key:            request.Key,
		Name:           request.Name,
		Description:    strings.TrimSpace(request.Description),
		Category:       strings.TrimSpace(request.Category),
		LogoURL:        logoURL,
		Status:         model.DefinitionStatusDraft,
		CurrentVersion: 0,
		DraftJSON:      draftJSON,
		AddUserID:      adminID,
		EditUserID:     adminID,
		AddTime:        now,
		EditTime:       now,
	}
}

func UpdateContext(ctx context.Context, adminID, id uint, request UpdateRequest) (*DefinitionDetail, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var item model.WorkflowDefinition
	if err := db.First(&item, id).Error; err != nil {
		return nil, definitionError(err)
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = item.Name
	}
	draftInput := request.Draft
	if len(bytes.TrimSpace(draftInput)) == 0 {
		draftInput = json.RawMessage(item.DraftJSON)
	}
	draft, encoded, err := normalizeDraft(draftInput, item.Key, name)
	if err != nil {
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
	updates := definitionContentUpdates(item, name, description, category, status, encoded, logoURL)
	if len(updates) > 0 {
		now := database.Now()
		updates["definition_edit_user_id"] = adminID
		updates["definition_edit_time"] = now
		updates["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")
		if err := definitionUpdateSession(db).Model(&model.WorkflowDefinition{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return nil, err
		}
		item.EditUserID = adminID
		item.EditTime = now
	}
	item.Name = name
	item.Description = description
	item.Category = category
	if logoURL != nil {
		item.LogoURL = *logoURL
	}
	item.Status = status
	item.DraftJSON = encoded
	return &DefinitionDetail{DefinitionSummary: summaryFromModel(ctx, item), Draft: draft}, nil
}

func definitionContentUpdates(item model.WorkflowDefinition, name, description, category string, status int, draftJSON string, logoURL *string) map[string]interface{} {
	updates := make(map[string]interface{}, 6)
	if name != item.Name {
		updates["definition_name"] = name
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
	if logoURL != nil && *logoURL != item.LogoURL {
		updates["definition_logo_url"] = *logoURL
	}
	return updates
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
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return definitionError(err)
		}
		draft, _, err := normalizeDraft(json.RawMessage(item.DraftJSON), item.Key, item.Name)
		if err != nil {
			return err
		}
		applyPublishInitiator(&draft, request.Initiator)
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
		if err := tx.Model(&model.WorkflowDefinition{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
			"definition_current_version": version,
			"definition_status":          model.DefinitionStatusPublished,
			"definition_draft_json":      encoded,
			"definition_edit_user_id":    adminID,
			"definition_edit_time":       now,
			"updated_at":                 gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
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

func DeleteContext(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var item model.WorkflowDefinition
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return definitionError(err)
		}
		if item.CurrentVersion > 0 {
			return errors.New("已发布的流程定义不能删除，可将其停用")
		}
		return tx.Delete(&item).Error
	})
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

func normalizeDraft(raw json.RawMessage, key, name string) (workflowcore.Definition, string, error) {
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
	encoded, err := json.Marshal(definition)
	if err != nil {
		return workflowcore.Definition{}, "", err
	}
	return definition, string(encoded), nil
}

func summaryFromModel(ctx context.Context, item model.WorkflowDefinition) DefinitionSummary {
	return DefinitionSummary{
		ID: item.ID, Key: item.Key, Name: item.Name, Description: item.Description,
		Category: item.Category, LogoURL: media.FullURLWithStaticDomainContext(ctx, item.LogoURL), Status: item.Status, CurrentVersion: item.CurrentVersion,
		AddUserID: item.AddUserID, EditUserID: item.EditUserID, AddTime: item.AddTime, EditTime: item.EditTime,
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
