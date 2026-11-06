package infrastructure

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/publicportal/application"
	"wecheckin/backend/internal/support/access"
)

func scopedSource(ctx context.Context, db *gorm.DB, adminID uint, kind application.Kind) (*gorm.DB, error) {
	if kind == application.Survey {
		return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Survey{}, access.SurveyAuditFields)
	}
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Exam{}, access.ExamAuditFields)
}
func sourceQuery(db *gorm.DB, kind application.Kind) *gorm.DB {
	if kind == application.Survey {
		return db.Model(&model.Survey{})
	}
	return db.Model(&model.Exam{})
}
func (s *GormStore) Sources(ctx context.Context, kind application.Kind, ids []uint, adminID uint) ([]application.Source, error) {
	if !kind.Valid() || len(ids) > application.MaxItems+1 {
		return nil, application.ErrInvalidArgument
	}
	if len(ids) == 0 {
		return []application.Source{}, nil
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	ctx = db.Statement.Context
	query := sourceQuery(db, kind)
	idColumn := string(kind) + "_id"
	// Only IDs are read outside the administrator's scope, to distinguish deleted
	// sources from inaccessible ones without loading their names/settings.
	existing := []uint{}
	if adminID > 0 {
		existenceQuery := query
		if s.inTx {
			// The actor read may precede a wait on the singleton lock. Use a
			// current read here so a REPEATABLE READ snapshot cannot mistake a
			// restored source for an orphan eligible for creator cleanup.
			existenceQuery = existenceQuery.Clauses(clause.Locking{Strength: "SHARE"})
		}
		if err = existenceQuery.Where(idColumn+" IN ?", ids).Pluck(idColumn, &existing).Error; err != nil {
			return nil, storageError(err)
		}
		query, err = scopedSource(ctx, db, adminID, kind)
		if err != nil {
			return nil, err
		}
	}
	if s.inTx {
		query = query.Clauses(clause.Locking{Strength: "SHARE"})
	}
	rows, err := readSources(query.Where(idColumn+" IN ?", ids).Order(idColumn+" ASC"), kind, adminID > 0)
	if err != nil {
		return nil, storageError(err)
	}
	if adminID == 0 {
		return rows, nil
	}
	byID := make(map[uint]application.Source, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	result := make([]application.Source, 0, len(existing))
	for _, id := range existing {
		row, ok := byID[id]
		if !ok {
			row = application.Source{Kind: kind, ID: id}
		}
		result = append(result, row)
	}
	return result, nil
}
func (s *GormStore) Options(ctx context.Context, adminID uint, q application.OptionsQuery) (application.SourcePage, error) {
	if adminID == 0 {
		return application.SourcePage{}, application.ErrForbidden
	}
	if err := application.ValidateOptions(&q); err != nil {
		return application.SourcePage{}, err
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return application.SourcePage{}, err
	}
	defer cancel()
	ctx = db.Statement.Context
	query, err := scopedSource(ctx, db, adminID, q.Kind)
	if err != nil {
		return application.SourcePage{}, err
	}
	prefix := string(q.Kind)
	query = query.Where(prefix+"_status = ? AND "+prefix+"_visibility = ?", 1, 0)
	if q.Keyword != "" {
		// CASE prevents JSON_EXTRACT from evaluating malformed legacy settings.
		// The OR stays grouped under both publication and administrator scope.
		settingsColumn := prefix + "_settings"
		displayName := "CASE WHEN JSON_VALID(" + settingsColumn + ") THEN " +
			"CASE WHEN JSON_TYPE(JSON_EXTRACT(" + settingsColumn + ", '$.displayTitle')) = 'STRING' " +
			"THEN JSON_UNQUOTE(JSON_EXTRACT(" + settingsColumn + ", '$.displayTitle')) ELSE '' END ELSE '' END"
		pattern := "%" + q.Keyword + "%"
		query = query.Where("("+prefix+"_title LIKE ? OR ("+displayName+") LIKE ?)", pattern, pattern)
	}
	var total int64
	if err = query.Count(&total).Error; err != nil {
		return application.SourcePage{}, storageError(err)
	}
	rows, err := readSources(query.Order(prefix+"_id ASC").Offset((q.Page-1)*q.PageSize).Limit(q.PageSize), q.Kind, true)
	return application.SourcePage{List: rows, Total: total}, storageError(err)
}
func readSources(query *gorm.DB, kind application.Kind, canManage bool) ([]application.Source, error) {
	result := []application.Source{}
	if kind == application.Survey {
		var rows []model.Survey
		err := query.Select("survey_id", "survey_title", "survey_category", "survey_desc", "survey_status", "survey_visibility", "survey_start_time", "survey_end_time", "survey_settings").Find(&rows).Error
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			result = append(result, parseSourceSettings(application.Source{Kind: kind, ID: row.ID, ManagementTitle: row.Title, Category: strings.TrimSpace(row.Category), Description: row.Desc, Status: row.Status, Visibility: row.Visibility, StartTime: row.StartTime, EndTime: row.EndTime, CanManage: canManage}, row.Settings))
		}
	} else {
		var rows []model.Exam
		err := query.Select("exam_id", "exam_title", "exam_category", "exam_desc", "exam_status", "exam_visibility", "exam_start_time", "exam_end_time", "exam_settings").Find(&rows).Error
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			result = append(result, parseSourceSettings(application.Source{Kind: kind, ID: row.ID, ManagementTitle: row.Title, Category: strings.TrimSpace(row.Category), Description: row.Description, Status: row.Status, Visibility: row.Visibility, StartTime: row.StartTime, EndTime: row.EndTime, CanManage: canManage}, row.Settings))
		}
	}
	return result, nil
}
func parseSourceSettings(source application.Source, raw string) application.Source {
	source.Description = plainDescription(source.Description)
	var settings struct {
		DisplayTitle  string          `json:"displayTitle"`
		FillTemplate  string          `json:"fillTemplate"`
		Password      string          `json:"password"`
		LoginRequired json.RawMessage `json:"loginRequired"`
	}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			return source
		}
	}
	source.SettingsValid = true
	source.DisplayTitle = strings.TrimSpace(settings.DisplayTitle)
	source.FillTemplate = settings.FillTemplate
	source.PasswordRequired = settings.Password != ""
	// Public surveys do not gain a synthetic login gate from their visibility.
	// Exam filling explicitly interprets the settings flag as a boolean.
	if source.Kind == application.Exam {
		_ = json.Unmarshal(settings.LoginRequired, &source.LoginRequired)
	}
	return source
}
