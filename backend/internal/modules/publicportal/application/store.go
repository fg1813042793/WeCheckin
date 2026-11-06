package application

import (
	"context"

	portalmodel "wecheckin/backend/internal/model/publicportal"
)

type Actor struct {
	ID                 uint
	ReservedSuperAdmin bool
}

// Source is a bounded projection. Raw settings and schema never leave storage.
// An existing but unscoped source contains only Kind/ID and CanManage=false.
type Source struct {
	Kind             Kind
	ID               uint
	ManagementTitle  string
	Category         string
	DisplayTitle     string
	Description      string
	Status           int
	Visibility       int
	StartTime        int64
	EndTime          int64
	FillTemplate     string
	LoginRequired    bool
	PasswordRequired bool
	CanManage        bool
	SettingsValid    bool
}

type SourcePage struct {
	List  []Source
	Total int64
}

type Store interface {
	Transaction(context.Context, func(Store) error) error
	Actor(context.Context, uint) (Actor, error)
	Settings(context.Context) (portalmodel.Settings, error)
	// LockSettings creates the default singleton if absent and locks it for all
	// portal mutations, serializing item count/uniqueness and settings versions.
	LockSettings(context.Context) (portalmodel.Settings, error)
	SaveSettings(context.Context, portalmodel.Settings, uint64) error
	Items(context.Context) ([]portalmodel.Item, error)
	InsertItem(context.Context, *portalmodel.Item) error
	UpdateItem(context.Context, portalmodel.Item, uint64) error
	DeleteItem(context.Context, uint64, uint64) error
	Sources(context.Context, Kind, []uint, uint) ([]Source, error)
	Options(context.Context, uint, OptionsQuery) (SourcePage, error)
}
