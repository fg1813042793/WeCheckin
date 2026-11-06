package application

import (
	"context"
	"math"
	"time"

	pm "wecheckin/backend/internal/model/publicportal"
)

func (s *Service) CreateItem(ctx context.Context, adminID uint, input ItemInput) (AdminItem, error) {
	if err := validateItem(&input); err != nil {
		return AdminItem{}, err
	}
	var result AdminItem
	err := s.mutate(ctx, adminID, func(tx Store, actor Actor, _ pm.Settings) error {
		items, err := tx.Items(ctx)
		if err != nil {
			return err
		}
		if len(items) >= MaxItems {
			return ErrItemLimit
		}
		if duplicate(items, input, 0) {
			return ErrDuplicateSource
		}
		source, err := manageableSource(ctx, tx, adminID, input.Kind, input.SourceID)
		if err != nil {
			return err
		}
		if !eligible(source) {
			return ErrSourceUnavailable
		}
		now := time.Now().UnixMilli()
		item := pm.Item{Kind: string(input.Kind), SourceID: input.SourceID, Title: input.Title, Description: input.Description, Enabled: input.Enabled, Sort: input.Sort, Version: 1, CreatedBy: actor.ID, UpdatedBy: actor.ID, CreatedAt: now, UpdatedAt: now}
		if err = tx.InsertItem(ctx, &item); err != nil {
			return err
		}
		result = adminItem(item, source, true, actor)
		return nil
	})
	if err != nil {
		return AdminItem{}, err
	}
	return result, nil
}
func (s *Service) UpdateItem(ctx context.Context, adminID uint, id uint64, input ItemUpdate) (AdminItem, error) {
	if id == 0 || input.Version == 0 || input.Version == math.MaxUint64 {
		return AdminItem{}, ErrInvalidArgument
	}
	if err := validateItem(&input.ItemInput); err != nil {
		return AdminItem{}, err
	}
	var result AdminItem
	err := s.mutate(ctx, adminID, func(tx Store, actor Actor, _ pm.Settings) error {
		items, err := tx.Items(ctx)
		if err != nil {
			return err
		}
		item, err := findItem(items, id, input.Version)
		if err != nil {
			return err
		}
		if _, err = manageableSource(ctx, tx, adminID, Kind(item.Kind), item.SourceID); err != nil {
			return err
		}
		if duplicate(items, input.ItemInput, id) {
			return ErrDuplicateSource
		}
		source, err := manageableSource(ctx, tx, adminID, input.Kind, input.SourceID)
		if err != nil {
			return err
		}
		// Managers may disable the original source after publication is revoked.
		disablingOriginal := item.Kind == string(input.Kind) && item.SourceID == input.SourceID && !input.Enabled
		if !eligible(source) && !disablingOriginal {
			return ErrSourceUnavailable
		}
		item.Kind = string(input.Kind)
		item.SourceID = input.SourceID
		item.Title = input.Title
		item.Description = input.Description
		item.Enabled = input.Enabled
		item.Sort = input.Sort
		item.Version++
		item.UpdatedBy = actor.ID
		item.UpdatedAt = time.Now().UnixMilli()
		if err = tx.UpdateItem(ctx, item, input.Version); err != nil {
			return err
		}
		result = adminItem(item, source, true, actor)
		return nil
	})
	if err != nil {
		return AdminItem{}, err
	}
	return result, nil
}
func (s *Service) DeleteItem(ctx context.Context, adminID uint, id, version uint64) (DeleteResponse, error) {
	if id == 0 || version == 0 {
		return DeleteResponse{}, ErrInvalidArgument
	}
	err := s.mutate(ctx, adminID, func(tx Store, actor Actor, _ pm.Settings) error {
		items, err := tx.Items(ctx)
		if err != nil {
			return err
		}
		item, err := findItem(items, id, version)
		if err != nil {
			return err
		}
		sources, err := tx.Sources(ctx, Kind(item.Kind), []uint{item.SourceID}, adminID)
		if err != nil {
			return err
		}
		if len(sources) == 0 {
			if item.CreatedBy != actor.ID && !actor.ReservedSuperAdmin {
				return ErrForbidden
			}
		} else if !sources[0].CanManage {
			return ErrForbidden
		}
		return tx.DeleteItem(ctx, id, version)
	})
	if err != nil {
		return DeleteResponse{}, err
	}
	return DeleteResponse{ID: id}, nil
}
func manageableSource(ctx context.Context, tx Store, adminID uint, kind Kind, id uint) (Source, error) {
	rows, err := tx.Sources(ctx, kind, []uint{id}, adminID)
	if err != nil {
		return Source{}, err
	}
	if len(rows) == 0 {
		return Source{}, ErrSourceUnavailable
	}
	if !rows[0].CanManage {
		return Source{}, ErrForbidden
	}
	return rows[0], nil
}
func duplicate(items []pm.Item, input ItemInput, except uint64) bool {
	for _, item := range items {
		if item.ID != except && item.Kind == string(input.Kind) && item.SourceID == input.SourceID {
			return true
		}
	}
	return false
}
func findItem(items []pm.Item, id, version uint64) (pm.Item, error) {
	for _, item := range items {
		if item.ID == id {
			if item.Version != version {
				return pm.Item{}, ErrVersionConflict
			}
			return item, nil
		}
	}
	return pm.Item{}, ErrItemNotFound
}
