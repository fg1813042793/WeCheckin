package application

import (
	"context"
	"errors"
	"strconv"
	"time"
)

func (s *Service) List(ctx context.Context) (PublicResponse, error) { return s.public(ctx, false) }
func (s *Service) Preview(ctx context.Context, adminID uint) (PublicResponse, error) {
	if _, err := s.actor(ctx, s.store, adminID); err != nil {
		return PublicResponse{}, err
	}
	return s.public(ctx, true)
}
func (s *Service) public(ctx context.Context, preview bool) (PublicResponse, error) {
	if s.store == nil {
		return PublicResponse{}, errors.New("public portal store unavailable")
	}
	settings, err := s.store.Settings(ctx)
	if err != nil {
		return PublicResponse{}, err
	}
	result := PublicResponse{Enabled: settings.Enabled || preview, Title: settings.Title, Description: settings.Description, LogoURL: settings.LogoURL, Items: []PublicItem{}}
	if !result.Enabled {
		return result, nil
	}
	items, err := s.store.Items(ctx)
	if err != nil {
		return PublicResponse{}, err
	}
	sources, err := loadSources(ctx, s.store, items, 0)
	if err != nil {
		return PublicResponse{}, err
	}
	sortItems(items)
	now := time.Now().UnixMilli()
	for _, item := range items {
		source, exists := sources[sourceKey{Kind(item.Kind), item.SourceID}]
		if !item.Enabled || !exists || !eligible(source) {
			continue
		}
		state := "active"
		if source.EndTime > 0 && now > source.EndTime {
			state = "ended"
		} else if source.StartTime > 0 && now < source.StartTime {
			state = "upcoming"
		}
		if state == "ended" && !settings.ShowEnded {
			continue
		}
		title := item.Title
		if title == "" {
			title = sourceTitle(source)
		}
		description := item.Description
		if description == "" {
			description = source.Description
		}
		template := "sf"
		if source.Kind == Exam {
			template = "ef"
		}
		if source.FillTemplate == template+"1" {
			template += "1"
		}
		result.Items = append(result.Items, PublicItem{ID: item.ID, Kind: Kind(item.Kind), SourceID: item.SourceID, Title: title, ManagementTitle: source.ManagementTitle, Category: source.Category, Description: description, StartTime: source.StartTime, EndTime: source.EndTime, State: state, LoginRequired: source.LoginRequired, PasswordRequired: source.PasswordRequired, URL: "/" + template + "/" + strconv.FormatUint(uint64(item.SourceID), 10)})
	}
	return result, nil
}
