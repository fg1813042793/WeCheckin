package application

import (
	"context"
	"errors"
	"strings"
)

func (service *Service) GetMyOverview(ctx context.Context, actorID string) (*WorkflowOverview, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	return service.store.GetWorkflowOverview(ctx, actorID)
}
