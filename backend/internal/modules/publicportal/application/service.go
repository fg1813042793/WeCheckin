package application

import (
	"context"
	"errors"
	"sort"
	"time"

	pm "wecheckin/backend/internal/model/publicportal"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) actor(ctx context.Context, store Store, id uint) (Actor, error) {
	if id == 0 {
		return Actor{}, ErrForbidden
	}
	if store == nil {
		return Actor{}, errors.New("public portal store unavailable")
	}
	actor, err := store.Actor(ctx, id)
	if err != nil {
		return Actor{}, err
	}
	if actor.ID != id {
		return Actor{}, ErrForbidden
	}
	return actor, nil
}
func (s *Service) Get(ctx context.Context, adminID uint) (AdminResponse, error) {
	actor, err := s.actor(ctx, s.store, adminID)
	if err != nil {
		return AdminResponse{}, err
	}
	settings, err := s.store.Settings(ctx)
	if err != nil {
		return AdminResponse{}, err
	}
	items, err := s.store.Items(ctx)
	if err != nil {
		return AdminResponse{}, err
	}
	sources, err := loadSources(ctx, s.store, items, adminID)
	if err != nil {
		return AdminResponse{}, err
	}
	result := AdminResponse{Settings: settingsDTO(settings), Items: make([]AdminItem, 0, len(items))}
	sortItems(items)
	for _, item := range items {
		source, exists := sources[sourceKey{Kind(item.Kind), item.SourceID}]
		result.Items = append(result.Items, adminItem(item, source, exists, actor))
	}
	return result, nil
}
func (s *Service) SaveSettings(ctx context.Context, adminID uint, input Settings) (Settings, error) {
	if err := validateSettings(&input); err != nil {
		return Settings{}, err
	}
	var saved pm.Settings
	err := s.mutate(ctx, adminID, func(tx Store, actor Actor, current pm.Settings) error {
		if current.Version != input.Version {
			return ErrVersionConflict
		}
		saved = pm.Settings{ID: 1, Enabled: input.Enabled, Title: input.Title, Description: input.Description, LogoURL: input.LogoURL, ShowEnded: input.ShowEnded, Version: current.Version + 1, UpdatedBy: actor.ID, UpdatedAt: time.Now().UnixMilli()}
		return tx.SaveSettings(ctx, saved, current.Version)
	})
	if err != nil {
		return Settings{}, err
	}
	return settingsDTO(saved), nil
}
func (s *Service) Options(ctx context.Context, adminID uint, query OptionsQuery) (OptionsResponse, error) {
	if err := ValidateOptions(&query); err != nil {
		return OptionsResponse{}, err
	}
	if _, err := s.actor(ctx, s.store, adminID); err != nil {
		return OptionsResponse{}, err
	}
	page, err := s.store.Options(ctx, adminID, query)
	if err != nil {
		return OptionsResponse{}, err
	}
	result := OptionsResponse{List: make([]SourceOption, 0, len(page.List)), Total: page.Total, Page: query.Page, PageSize: query.PageSize}
	for _, source := range page.List {
		result.List = append(result.List, SourceOption{Kind: source.Kind, SourceID: source.ID, ManagementTitle: source.ManagementTitle, Category: source.Category, DisplayTitle: sourceTitle(source), StartTime: source.StartTime, EndTime: source.EndTime, LoginRequired: source.LoginRequired, PasswordRequired: source.PasswordRequired})
	}
	return result, nil
}
func (s *Service) mutate(ctx context.Context, adminID uint, fn func(Store, Actor, pm.Settings) error) error {
	if adminID == 0 {
		return ErrForbidden
	}
	if s.store == nil {
		return errors.New("public portal store unavailable")
	}
	return s.store.Transaction(ctx, func(tx Store) error {
		actor, err := s.actor(ctx, tx, adminID)
		if err != nil {
			return err
		}
		settings, err := tx.LockSettings(ctx)
		if err != nil {
			return err
		}
		return fn(tx, actor, settings)
	})
}

type sourceKey struct {
	kind Kind
	id   uint
}

func loadSources(ctx context.Context, store Store, items []pm.Item, adminID uint) (map[sourceKey]Source, error) {
	result := make(map[sourceKey]Source, len(items))
	for _, kind := range []Kind{Survey, Exam} {
		ids := make([]uint, 0, len(items))
		for _, item := range items {
			if Kind(item.Kind) == kind {
				ids = append(ids, item.SourceID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		list, err := store.Sources(ctx, kind, ids, adminID)
		if err != nil {
			return nil, err
		}
		for _, source := range list {
			result[sourceKey{kind, source.ID}] = source
		}
	}
	return result, nil
}
func sortItems(items []pm.Item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].ID < items[j].ID
		}
		return items[i].Sort < items[j].Sort
	})
}
func settingsDTO(m pm.Settings) Settings {
	return Settings{Enabled: m.Enabled, Title: m.Title, Description: m.Description, LogoURL: m.LogoURL, ShowEnded: m.ShowEnded, Version: m.Version}
}
func sourceTitle(s Source) string {
	if s.DisplayTitle != "" {
		return s.DisplayTitle
	}
	return s.ManagementTitle
}
func eligible(s Source) bool {
	return s.ID > 0 && s.Status == 1 && s.Visibility == 0 && s.SettingsValid
}
func itemInput(m pm.Item) ItemInput {
	return ItemInput{Kind: Kind(m.Kind), SourceID: m.SourceID, Title: m.Title, Description: m.Description, Enabled: m.Enabled, Sort: m.Sort}
}
func adminItem(m pm.Item, s Source, exists bool, actor Actor) AdminItem {
	r := AdminItem{ItemInput: itemInput(m), ID: m.ID, Version: m.Version}
	if !exists {
		r.CanManage = m.CreatedBy == actor.ID || actor.ReservedSuperAdmin
		r.UnavailableReason = "来源已删除"
		return r
	}
	if !s.CanManage {
		r.UnavailableReason = "无权管理该来源"
		return r
	}
	r.CanManage = true
	r.ManagementTitle = s.ManagementTitle
	r.Category = s.Category
	r.DisplayTitle = sourceTitle(s)
	r.Available = eligible(s)
	if !r.Available {
		r.UnavailableReason = "来源未发布、非公开或配置无效"
	}
	return r
}
