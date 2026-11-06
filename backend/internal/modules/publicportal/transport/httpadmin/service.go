package httpadmin

import (
	"context"

	"wecheckin/backend/internal/modules/publicportal/application"
)

type Service interface {
	Get(context.Context, uint) (application.AdminResponse, error)
	SaveSettings(context.Context, uint, application.Settings) (application.Settings, error)
	Options(context.Context, uint, application.OptionsQuery) (application.OptionsResponse, error)
	CreateItem(context.Context, uint, application.ItemInput) (application.AdminItem, error)
	UpdateItem(context.Context, uint, uint64, application.ItemUpdate) (application.AdminItem, error)
	DeleteItem(context.Context, uint, uint64, uint64) (application.DeleteResponse, error)
	Preview(context.Context, uint) (application.PublicResponse, error)
}
