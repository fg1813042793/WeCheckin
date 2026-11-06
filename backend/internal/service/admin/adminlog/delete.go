package adminlog

import (
	"context"
	"sort"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

const MaxDeleteSize = 100

type DeleteRequest struct {
	IDs []uint `json:"ids"`
}
type DeleteResult struct {
	DeletedCount int `json:"deletedCount"`
}

func DeleteForAdminContext(ctx context.Context, category string, ids []uint, adminID uint) (DeleteResult, error) {
	if len(ids) == 0 || len(ids) > MaxDeleteSize {
		return DeleteResult{}, invalid("ids", "每次必须选择 1–100 条日志")
	}
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			return DeleteResult{}, invalid("ids", "必须是不重复的正整数")
		}
		seen[id] = true
	}
	ids = append([]uint(nil), ids...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return DeleteResult{}, ErrUnavailable
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		scopedQuery, err := authorizedQueryContext(ctx, tx, category, adminID, false, true)
		if err != nil {
			return err
		}
		scopedQuery = scopedQuery.Where("`logs`.`id` IN ?", ids)
		if eventID := auditlog.CurrentEventID(ctx); eventID != "" {
			scopedQuery = scopedQuery.Where("(`log_event_id` <> ? OR `log_event_id` IS NULL)", eventID)
		}
		var matched []uint
		if err := scopedQuery.Session(&gorm.Session{}).Clauses(clause.Locking{Strength: "UPDATE"}).Order("`logs`.`id`").Pluck("id", &matched).Error; err != nil {
			return err
		}
		if len(matched) != len(ids) {
			return ErrNotFound
		}
		result := scopedQuery.Delete(&model.Log{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(ids)) {
			return ErrNotFound
		}
		return nil
	})
	if err != nil {
		return DeleteResult{}, err
	}
	auditlog.Update(ctx, func(d *auditlog.Details) {
		d.TargetType, d.AffectedCount = "audit_log", len(ids)
		d.TargetIDs = make([]string, len(ids))
		for i, id := range ids {
			d.TargetIDs[i] = strconv.FormatUint(uint64(id), 10)
		}
		if len(ids) == 1 {
			d.TargetID = d.TargetIDs[0]
		}
	})
	return DeleteResult{DeletedCount: len(ids)}, nil
}
