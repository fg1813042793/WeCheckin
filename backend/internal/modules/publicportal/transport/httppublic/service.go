package httppublic

import (
	"context"

	"wecheckin/backend/internal/modules/publicportal/application"
)

type Service interface {
	List(context.Context) (application.PublicResponse, error)
}
