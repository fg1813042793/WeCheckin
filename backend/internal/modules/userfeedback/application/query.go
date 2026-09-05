package application

import "context"

func (service *Service) GetUserOverview(ctx context.Context, userID uint) (Overview, error) {
	if err := service.requireStore(); err != nil {
		return Overview{}, err
	}
	if userID == 0 {
		return Overview{}, ErrInvalidArgument
	}
	return service.store.GetUserOverview(ctx, userID)
}

func (service *Service) ListUserFeedbacks(ctx context.Context, userID uint, query UserListQuery) (FeedbackList, error) {
	if err := service.requireStore(); err != nil {
		return FeedbackList{}, err
	}
	if userID == 0 {
		return FeedbackList{}, ErrInvalidArgument
	}
	normalized, err := normalizeUserListQuery(query)
	if err != nil {
		return FeedbackList{}, err
	}
	result, err := service.store.ListUserFeedbacks(ctx, userID, normalized)
	if err != nil {
		return FeedbackList{}, err
	}
	result.Page, result.PageSize = normalized.Page, normalized.PageSize
	return result, nil
}

func (service *Service) GetUserFeedback(ctx context.Context, id uint64, userID uint) (*FeedbackDetail, error) {
	if err := service.requireStore(); err != nil {
		return nil, err
	}
	if id == 0 || userID == 0 {
		return nil, ErrInvalidArgument
	}
	detail, err := service.store.GetUserFeedback(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.ID != id || detail.SubmitterID != userID {
		return nil, ErrFeedbackNotFound
	}
	return decorateDetail(detail), nil
}

func (service *Service) GetAdminOverview(ctx context.Context, query AdminListQuery) (Overview, error) {
	if err := service.requireStore(); err != nil {
		return Overview{}, err
	}
	normalized, err := normalizeAdminListQuery(query)
	if err != nil {
		return Overview{}, err
	}
	return service.store.GetAdminOverview(ctx, normalized)
}

func (service *Service) ListAdminFeedbacks(ctx context.Context, query AdminListQuery) (FeedbackList, error) {
	if err := service.requireStore(); err != nil {
		return FeedbackList{}, err
	}
	normalized, err := normalizeAdminListQuery(query)
	if err != nil {
		return FeedbackList{}, err
	}
	result, err := service.store.ListAdminFeedbacks(ctx, normalized)
	if err != nil {
		return FeedbackList{}, err
	}
	result.Page, result.PageSize = normalized.Page, normalized.PageSize
	return result, nil
}

func (service *Service) GetAdminFeedback(ctx context.Context, id uint64) (*FeedbackDetail, error) {
	if err := service.requireStore(); err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, ErrInvalidArgument
	}
	detail, err := service.store.GetAdminFeedback(ctx, id)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.ID != id {
		return nil, ErrFeedbackNotFound
	}
	return decorateDetail(detail), nil
}

func normalizeUserListQuery(query UserListQuery) (UserListQuery, error) {
	keyword, err := normalizeKeyword(query.Keyword)
	if err != nil {
		return UserListQuery{}, err
	}
	if query.Status != "" && !query.Status.Valid() {
		return UserListQuery{}, ErrInvalidArgument
	}
	page, pageSize, err := normalizePage(query.Page, query.PageSize)
	if err != nil {
		return UserListQuery{}, err
	}
	query.Keyword, query.Page, query.PageSize = keyword, page, pageSize
	return query, nil
}

func normalizeAdminListQuery(query AdminListQuery) (AdminListQuery, error) {
	keyword, err := normalizeKeyword(query.Keyword)
	if err != nil {
		return AdminListQuery{}, err
	}
	if query.Status != "" && !query.Status.Valid() {
		return AdminListQuery{}, ErrInvalidArgument
	}
	if query.HandlerID != nil && *query.HandlerID == 0 {
		return AdminListQuery{}, ErrInvalidArgument
	}
	if query.SubmittedFrom < 0 || query.SubmittedTo < 0 ||
		(query.SubmittedFrom > 0 && query.SubmittedTo > 0 && query.SubmittedFrom > query.SubmittedTo) {
		return AdminListQuery{}, ErrInvalidArgument
	}
	page, pageSize, err := normalizePage(query.Page, query.PageSize)
	if err != nil {
		return AdminListQuery{}, err
	}
	query.Keyword, query.Page, query.PageSize = keyword, page, pageSize
	return query, nil
}
