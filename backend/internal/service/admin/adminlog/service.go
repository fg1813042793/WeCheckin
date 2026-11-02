package adminlog

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

var (
	ErrReadOnly    = errors.New("审计日志只读，不允许删除或清空")
	ErrForbidden   = errors.New("无权限访问审计日志")
	ErrNotFound    = errors.New("日志不存在或无权访问")
	ErrUnavailable = errors.New("审计日志服务暂不可用")
)

// GetList and GetListContext are retained for old callers, restricted to operations.
func GetList(keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	return GetListContext(context.Background(), keyword, page, pageSize, adminID)
}

func GetListContext(ctx context.Context, keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	result, err := ListContext(ctx, CategoryOperation, Query{Search: keyword, Page: page, PageSize: pageSize}, adminID)
	return result.List, result.Total, err
}

func ListContext(ctx context.Context, category string, filter Query, adminID uint) (Page, error) {
	return listContext(ctx, category, filter, adminID, false)
}

func ExportContext(ctx context.Context, category string, filter Query, adminID uint) (Page, error) {
	return listContext(ctx, category, filter, adminID, true)
}

func listContext(ctx context.Context, category string, filter Query, adminID uint, export bool) (Page, error) {
	if err := filter.Validate(export); err != nil {
		return Page{}, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := visibleQueryContext(ctx, db, category, adminID, export)
	if err != nil {
		return Page{}, err
	}
	query = applyFilters(query, filter)
	// The collector already persisted this request's unfinished attempt. Exclude
	// it from both the total and page without excluding legacy NULL event IDs.
	if eventID := auditlog.CurrentEventID(ctx); eventID != "" {
		query = query.Where("(`log_event_id` <> ? OR `log_event_id` IS NULL)", eventID)
	}
	result := Page{List: make([]model.Log, 0), Page: filter.Page, PageSize: filter.PageSize}
	if err := query.Count(&result.Total).Error; err != nil {
		return Page{}, err
	}
	// Large field diffs, selected IDs and user agents are detail-only; never load them for
	// a page/export merely to discard them while serializing the response.
	if err := query.Omit("log_changes", "log_target_ids", "log_user_agent").Order("`add_time` DESC, `id` DESC").Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&result.List).Error; err != nil {
		return Page{}, err
	}
	for i := range result.List {
		normalizeRead(&result.List[i])
	}
	return result, nil
}

func DetailContext(ctx context.Context, category string, id, adminID uint) (*model.Log, error) {
	if id == 0 {
		return nil, invalid("id", "必须为正整数")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := visibleQueryContext(ctx, db, category, adminID, false)
	if err != nil {
		return nil, err
	}
	var row model.Log
	if err := query.Where("`logs`.`id` = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	normalizeRead(&row)
	return &row, nil
}

func categoryQuery(db *gorm.DB, category string) *gorm.DB {
	// A rolling upgrade's legacy writer must never expose log_type=1 through log:list.
	legacyCondition := "`log_type` <> 1 OR `log_type` IS NULL"
	if category == CategoryLogin {
		legacyCondition = "`log_type` = 1"
	}
	return db.Model(&model.Log{}).Where("(`log_category` = ? OR ((`log_category` = '' OR `log_category` IS NULL) AND ("+legacyCondition+")))", category)
}

func applyFilters(query *gorm.DB, q Query) *gorm.DB {
	if q.Search != "" {
		// Bind values and escape LIKE metacharacters: searches cannot enlarge scope.
		search := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(q.Search) + "%"
		columns := []string{"log_content", "log_admin_name", "log_admin_desc", "log_attempted_account", "log_target_name", "log_target_id", "log_event_id", "log_request_id", "log_reason_code", "log_reason"}
		parts, args := make([]string, 0, len(columns)), make([]interface{}, 0, len(columns))
		for _, column := range columns {
			parts = append(parts, "`"+column+"` LIKE ? ESCAPE '!'")
			args = append(args, search)
		}
		query = query.Where("("+strings.Join(parts, " OR ")+")", args...)
	}
	if q.Source != "" {
		if q.Source == "admin" {
			query = query.Where("(`log_source` = ? OR `log_source` = '' OR `log_source` IS NULL)", q.Source)
		} else {
			query = query.Where("`log_source` = ?", q.Source)
		}
	}
	if q.Outcome != "" {
		if q.Outcome == "unknown" {
			query = query.Where("(`log_outcome` = ? OR `log_outcome` = '' OR `log_outcome` IS NULL)", q.Outcome)
		} else {
			query = query.Where("`log_outcome` = ?", q.Outcome)
		}
	}
	for _, item := range []struct{ column, value string }{{"log_module", q.Module}, {"log_action", q.Action}, {"log_add_ip", q.IP}} {
		if item.value != "" {
			query = query.Where("`"+item.column+"` = ?", item.value)
		}
	}
	if q.ActorID != nil {
		query = query.Where("`create_by` = ?", *q.ActorID)
	}
	if q.StartTime > 0 {
		query = query.Where("`add_time` >= ?", q.StartTime)
	}
	if q.EndTime > 0 {
		query = query.Where("`add_time` <= ?", q.EndTime)
	}
	return query
}

func Clear() error                                   { return ClearContext(context.Background()) }
func ClearContext(context.Context) error             { return ErrReadOnly }
func DeleteIDs(ids []uint) error                     { return DeleteIDsContext(context.Background(), ids) }
func DeleteIDsContext(context.Context, []uint) error { return ErrReadOnly }
