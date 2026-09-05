package application

import (
	"context"
	"strings"
)

func (service *Service) ListNotifications(ctx context.Context, query NotificationQuery) (*NotificationList, error) {
	if service == nil || service.notifications == nil {
		return nil, ErrNotificationUnavailable
	}
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.notifications.List(ctx, query)
}

func (service *Service) RetryNotification(ctx context.Context, id string) error {
	if service == nil || service.notifications == nil {
		return ErrNotificationUnavailable
	}
	return service.notifications.Retry(ctx, strings.TrimSpace(id))
}

func (service *Service) DispatchDueNotifications(ctx context.Context, limit int) (int, error) {
	if service == nil || service.notifications == nil {
		return 0, ErrNotificationUnavailable
	}
	return service.notifications.DispatchDue(ctx, limit)
}

func (service *Service) ListMyTasks(ctx context.Context, actorID string, query TaskQuery) (*TaskList, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	query.AssigneeID = actorID
	return service.ListTasks(ctx, query)
}

func (service *Service) publish(ctx context.Context, event LifecycleEvent) {
	if service != nil && service.publisher != nil {
		service.publisher.Publish(ctx, event)
	}
}

func (service *Service) dispatchNotifications(ctx context.Context, ids []string) {
	if service == nil || service.notifications == nil || len(ids) == 0 {
		return
	}
	_, _ = service.notifications.Dispatch(ctx, ids)
}
