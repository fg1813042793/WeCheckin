package infrastructure

import (
	"context"
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/inappnotification/application"
)

func (store *GormStore) ListTemplates(ctx context.Context, query application.DingTalkTemplateQuery, enabledOnly bool) ([]notificationmodel.DingTalkTemplate, int64, error) {
	query, err := application.NormalizeDingTalkTemplateQuery(query)
	if err != nil {
		return nil, 0, err
	}
	if store == nil || store.db == nil {
		return nil, 0, application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	db = db.Model(&notificationmodel.DingTalkTemplate{})
	if enabledOnly {
		db = db.Where("enabled = ?", true)
	} else if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}
	if query.TargetKind != "" {
		db = db.Where("target_kind = ?", query.TargetKind)
	}
	if query.ID > 0 {
		db = db.Where("id = ?", query.ID)
	}
	if query.Keyword != "" {
		pattern := templateSearchPattern(query.Keyword)
		db = db.Where("(name LIKE ? ESCAPE '!' OR title LIKE ? ESCAPE '!')", pattern, pattern)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, templateStoreError(err)
	}
	rows := make([]notificationmodel.DingTalkTemplate, 0)
	err = db.Order("id DESC").Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error
	return rows, total, templateStoreError(err)
}

func (store *GormStore) GetTemplate(ctx context.Context, id uint) (notificationmodel.DingTalkTemplate, error) {
	if store == nil || store.db == nil {
		return notificationmodel.DingTalkTemplate{}, application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	var row notificationmodel.DingTalkTemplate
	err := db.Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, application.ErrTemplateNotFound
	}
	return row, templateStoreError(err)
}

func (store *GormStore) CreateTemplate(ctx context.Context, row *notificationmodel.DingTalkTemplate) error {
	if store == nil || store.db == nil {
		return application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	return templateStoreError(db.Create(row).Error)
}

func (store *GormStore) UpdateTemplate(ctx context.Context, row *notificationmodel.DingTalkTemplate, expectedVersion int) error {
	if store == nil || store.db == nil {
		return application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	// Explicit fields include false/empty values, while creation attribution stays immutable.
	result := db.Model(&notificationmodel.DingTalkTemplate{}).Where("id = ? AND version = ?", row.ID, expectedVersion).Updates(map[string]any{
		"name": row.Name, "enabled": row.Enabled, "title": row.Title, "content": row.Content, "message_type": row.MessageType,
		"target_kind": row.TargetKind, "definition_id": row.DefinitionID, "definition_name": row.DefinitionName, "target_url": row.URL,
		"button_title": row.ButtonTitle, "head_color": row.HeadColor, "version": row.Version, "updated_by": row.UpdatedBy, "updated_at": row.UpdatedAt,
	})
	if result.Error != nil {
		return templateStoreError(result.Error)
	}
	if result.RowsAffected == 0 {
		return application.ErrTemplateVersionConflict
	}
	return nil
}

func (store *GormStore) DeleteTemplate(ctx context.Context, id uint, version int) error {
	if store == nil || store.db == nil {
		return application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	result := db.Where("id = ? AND version = ?", id, version).Delete(&notificationmodel.DingTalkTemplate{})
	if result.Error != nil {
		return templateStoreError(result.Error)
	}
	if result.RowsAffected == 0 {
		return application.ErrTemplateVersionConflict
	}
	return nil
}

func publishedTemplateWorkflows(db *gorm.DB) *gorm.DB {
	return db.Model(&workflowmodel.Definition{}).Where("definition_status = ? AND definition_current_version > 0 AND definition_deleted_at = 0", workflowmodel.DefinitionStatusPublished)
}

func (store *GormStore) PublishedTemplateWorkflow(ctx context.Context, id uint) (application.DingTalkTemplateWorkflowOption, error) {
	if store == nil || store.db == nil {
		return application.DingTalkTemplateWorkflowOption{}, application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	var row workflowmodel.Definition
	err := publishedTemplateWorkflows(db).Select("id", "definition_name", "definition_display_name").Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.DingTalkTemplateWorkflowOption{}, application.ErrTemplateWorkflowUnavailable
	}
	return templateWorkflowOption(row), templateStoreError(err)
}

func (store *GormStore) TemplateWorkflowOptions(ctx context.Context, query application.DingTalkTemplateQuery) ([]application.DingTalkTemplateWorkflowOption, int64, error) {
	query, err := application.NormalizeDingTalkTemplateQuery(query)
	if err != nil {
		return nil, 0, err
	}
	if store == nil || store.db == nil {
		return nil, 0, application.ErrTemplateStoreUnavailable
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	db = publishedTemplateWorkflows(db)
	if query.ID > 0 {
		db = db.Where("id = ?", query.ID)
	}
	if query.Keyword != "" {
		pattern := templateSearchPattern(query.Keyword)
		db = db.Where("(definition_name LIKE ? ESCAPE '!' OR definition_display_name LIKE ? ESCAPE '!')", pattern, pattern)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, templateStoreError(err)
	}
	var rows []workflowmodel.Definition
	err = db.Select("id", "definition_name", "definition_display_name").Order("id DESC").Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&rows).Error
	if err != nil {
		return nil, 0, templateStoreError(err)
	}
	options := make([]application.DingTalkTemplateWorkflowOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, templateWorkflowOption(row))
	}
	return options, total, nil
}

func templateWorkflowOption(row workflowmodel.Definition) application.DingTalkTemplateWorkflowOption {
	name := strings.TrimSpace(row.Name)
	if name == "" {
		name = strings.TrimSpace(row.DisplayName)
	}
	return application.DingTalkTemplateWorkflowOption{ID: row.ID, Name: name}
}

func templateSearchPattern(value string) string {
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value) + "%"
}

func templateStoreError(err error) error {
	if isDuplicateKeyError(err) {
		return application.ErrTemplateNameExists
	}
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && (mysqlErr.Number == 1146 || mysqlErr.Number == 1054) {
		return application.ErrTemplateSchemaMissing
	}
	return err
}

var _ application.TemplateStore = (*GormStore)(nil)
