package infrastructure

import (
	"context"
	"errors"
	"math"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	userfeedbackmodel "wecheckin/backend/internal/model/userfeedback"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
	"wecheckin/backend/internal/support/media"
)

const feedbackOverviewSelect = "COALESCE(SUM(feedback_status = 'pending'), 0) AS pending, " +
	"COALESCE(SUM(feedback_status = 'processing'), 0) AS processing, " +
	"COALESCE(SUM(feedback_status = 'resolved'), 0) AS resolved, " +
	"COALESCE(SUM(feedback_status = 'closed'), 0) AS closed"

type feedbackInitialMessageRow struct {
	FeedbackID uint64 `gorm:"column:feedback_id"`
	Content    string `gorm:"column:content"`
}

type feedbackImageSummaryRow struct {
	FeedbackID     uint64 `gorm:"column:feedback_id"`
	ImageCount     int64  `gorm:"column:image_count"`
	FirstObjectKey string `gorm:"column:first_object_key"`
}

func (store *GormStore) FindCreateReplay(ctx context.Context, key application.CreateReplayKey) (*application.FeedbackDetail, bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, false, err
	}
	defer cancel()
	var feedback userfeedbackmodel.Feedback
	err = db.Where("submitter_id = ? AND create_request_id = ?", key.SubmitterID, key.RequestID).Take(&feedback).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	detail, err := loadFeedbackDetail(ctx, db, feedback)
	return detail, err == nil, err
}

func (store *GormStore) FindMessageReplay(ctx context.Context, key application.MessageReplayKey) (*application.FeedbackDetail, bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, false, err
	}
	defer cancel()
	var replay struct {
		FeedbackID uint64 `gorm:"column:feedback_id"`
	}
	err = db.Model(&userfeedbackmodel.Message{}).
		Select("feedback_id").
		Where("feedback_id = ? AND author_type = ? AND author_id = ? AND request_id = ?", key.FeedbackID, key.AuthorType, key.AuthorID, key.RequestID).
		Take(&replay).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	feedback, err := findFeedback(db, "id = ?", replay.FeedbackID)
	if err != nil {
		return nil, false, err
	}
	detail, err := loadFeedbackDetail(ctx, db, feedback)
	return detail, err == nil, err
}

func (store *GormStore) GetUserOverview(ctx context.Context, userID uint) (application.Overview, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return application.Overview{}, err
	}
	defer cancel()
	var overview application.Overview
	err = db.Model(&userfeedbackmodel.Feedback{}).
		Select(feedbackOverviewSelect).
		Where("submitter_id = ?", userID).
		Scan(&overview).Error
	return overview, err
}

func (store *GormStore) ListUserFeedbacks(ctx context.Context, userID uint, query application.UserListQuery) (application.FeedbackList, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return application.FeedbackList{}, err
	}
	defer cancel()
	query.Page, query.PageSize = queryPage(query.Page, query.PageSize)
	offset, err := safeFeedbackOffset(query.Page, query.PageSize)
	if err != nil {
		return application.FeedbackList{}, err
	}
	statement := applyUserFeedbackFilters(db.Model(&userfeedbackmodel.Feedback{}), userID, query)
	return listFeedbacks(ctx, statement, db, query.Page, query.PageSize, offset)
}

func (store *GormStore) GetUserFeedback(ctx context.Context, id uint64, userID uint) (*application.FeedbackDetail, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	feedback, err := findFeedback(db, "id = ? AND submitter_id = ?", id, userID)
	if err != nil {
		return nil, err
	}
	return loadFeedbackDetail(ctx, db, feedback)
}

func (store *GormStore) GetAdminOverview(ctx context.Context, query application.AdminListQuery) (application.Overview, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return application.Overview{}, err
	}
	defer cancel()
	var overview application.Overview
	statement := applyAdminFeedbackFilters(db.Model(&userfeedbackmodel.Feedback{}), query, false)
	err = statement.Select(feedbackOverviewSelect).Scan(&overview).Error
	return overview, err
}

func (store *GormStore) ListAdminFeedbacks(ctx context.Context, query application.AdminListQuery) (application.FeedbackList, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return application.FeedbackList{}, err
	}
	defer cancel()
	query.Page, query.PageSize = queryPage(query.Page, query.PageSize)
	offset, err := safeFeedbackOffset(query.Page, query.PageSize)
	if err != nil {
		return application.FeedbackList{}, err
	}
	statement := applyAdminFeedbackFilters(db.Model(&userfeedbackmodel.Feedback{}), query, true)
	return listFeedbacks(ctx, statement, db, query.Page, query.PageSize, offset)
}

func (store *GormStore) GetAdminFeedback(ctx context.Context, id uint64) (*application.FeedbackDetail, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	feedback, err := findFeedback(db, "id = ?", id)
	if err != nil {
		return nil, err
	}
	return loadFeedbackDetail(ctx, db, feedback)
}

func findFeedback(db *gorm.DB, condition string, values ...any) (userfeedbackmodel.Feedback, error) {
	var feedback userfeedbackmodel.Feedback
	err := db.Where(condition, values...).Take(&feedback).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return userfeedbackmodel.Feedback{}, application.ErrFeedbackNotFound
	}
	return feedback, err
}

func loadFeedbackDetail(ctx context.Context, db *gorm.DB, feedback userfeedbackmodel.Feedback) (*application.FeedbackDetail, error) {
	messages := make([]userfeedbackmodel.Message, 0)
	if err := db.Where("feedback_id = ?", feedback.ID).
		Order("created_at ASC").Order("id ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	attachments := make([]userfeedbackmodel.Attachment, 0)
	if err := db.Where("feedback_id = ?", feedback.ID).
		Order("sort_order ASC").Order("id ASC").
		Find(&attachments).Error; err != nil {
		return nil, err
	}
	names, err := loadFeedbackDisplayNames(db, feedback, messages)
	if err != nil {
		return nil, err
	}
	return buildFeedbackDetail(ctx, feedback, messages, attachments, names), nil
}

func loadFeedbackDisplayNames(db *gorm.DB, feedback userfeedbackmodel.Feedback, messages []userfeedbackmodel.Message) (map[uint]string, error) {
	ids := make([]uint, 0, len(messages)+2)
	seen := make(map[uint]struct{}, len(messages)+2)
	appendID := func(id uint) {
		if id == 0 {
			return
		}
		if _, exists := seen[id]; exists {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	appendID(feedback.SubmitterID)
	if feedback.HandlerID != nil {
		appendID(*feedback.HandlerID)
	}
	for _, message := range messages {
		appendID(message.AuthorID)
	}
	return loadUserNames(db, ids)
}

func loadUserNames(db *gorm.DB, ids []uint) (map[uint]string, error) {
	names := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	users := make([]model.User, 0)
	if err := db.Select("id", "user_name").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	for _, user := range users {
		names[user.ID] = strings.TrimSpace(user.Name)
	}
	return names, nil
}

func applyUserFeedbackFilters(db *gorm.DB, userID uint, query application.UserListQuery) *gorm.DB {
	db = db.Where("user_feedbacks.submitter_id = ?", userID)
	if query.Status != "" {
		db = db.Where("user_feedbacks.feedback_status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := feedbackContainsLikePattern(keyword)
		db = db.Where(`(
			user_feedbacks.feedback_no LIKE ? ESCAPE '!'
			OR EXISTS (
				SELECT 1 FROM user_feedback_messages keyword_message
				WHERE keyword_message.feedback_id = user_feedbacks.id
				AND keyword_message.content LIKE ? ESCAPE '!'
			)
		)`, pattern, pattern)
	}
	return db
}

func applyAdminFeedbackFilters(db *gorm.DB, query application.AdminListQuery, includeStatus bool) *gorm.DB {
	if includeStatus && query.Status != "" {
		db = db.Where("user_feedbacks.feedback_status = ?", query.Status)
	}
	if query.SubmitterID > 0 {
		db = db.Where("user_feedbacks.submitter_id = ?", query.SubmitterID)
	}
	if query.HandlerID != nil {
		db = db.Where("user_feedbacks.handler_id = ?", *query.HandlerID)
	}
	if query.SubmittedFrom > 0 {
		db = db.Where("user_feedbacks.created_at >= ?", query.SubmittedFrom)
	}
	if query.SubmittedTo > 0 {
		db = db.Where("user_feedbacks.created_at <= ?", query.SubmittedTo)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		pattern := feedbackContainsLikePattern(keyword)
		db = db.Where(`(
			user_feedbacks.feedback_no LIKE ? ESCAPE '!'
			OR EXISTS (
				SELECT 1 FROM user_feedback_messages keyword_message
				WHERE keyword_message.feedback_id = user_feedbacks.id
				AND keyword_message.content LIKE ? ESCAPE '!'
			)
			OR EXISTS (
				SELECT 1 FROM users keyword_submitter
				WHERE keyword_submitter.id = user_feedbacks.submitter_id
				AND keyword_submitter.user_name LIKE ? ESCAPE '!'
			)
		)`, pattern, pattern, pattern)
	}
	return db
}

func feedbackContainsLikePattern(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return "%" + replacer.Replace(strings.TrimSpace(value)) + "%"
}

func listFeedbacks(ctx context.Context, statement, db *gorm.DB, page, pageSize, offset int) (application.FeedbackList, error) {
	result := application.FeedbackList{List: []application.FeedbackSummary{}, Page: page, PageSize: pageSize}
	if err := statement.Session(&gorm.Session{}).Count(&result.Total).Error; err != nil {
		return application.FeedbackList{}, err
	}
	rows := make([]userfeedbackmodel.Feedback, 0)
	if err := statement.Session(&gorm.Session{}).
		Order("last_activity_at DESC,id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return application.FeedbackList{}, err
	}
	if len(rows) == 0 {
		return result, nil
	}
	summaries, err := feedbackSummaries(ctx, db, rows)
	if err != nil {
		return application.FeedbackList{}, err
	}
	result.List = summaries
	return result, nil
}

func feedbackSummaries(ctx context.Context, db *gorm.DB, rows []userfeedbackmodel.Feedback) ([]application.FeedbackSummary, error) {
	feedbackIDs := make([]uint64, 0, len(rows))
	userIDs := make([]uint, 0, len(rows)*2)
	seenUsers := make(map[uint]struct{}, len(rows)*2)
	appendUserID := func(id uint) {
		if id == 0 {
			return
		}
		if _, exists := seenUsers[id]; exists {
			return
		}
		seenUsers[id] = struct{}{}
		userIDs = append(userIDs, id)
	}
	for _, row := range rows {
		feedbackIDs = append(feedbackIDs, row.ID)
		appendUserID(row.SubmitterID)
		if row.HandlerID != nil {
			appendUserID(*row.HandlerID)
		}
	}

	initialRows := make([]feedbackInitialMessageRow, 0)
	if err := db.Model(&userfeedbackmodel.Message{}).
		Select("feedback_id", "content").
		Where("feedback_id IN ? AND message_type = ?", feedbackIDs, domain.MessageTypeInitial).
		Order("feedback_id ASC").Order("created_at ASC").Order("id ASC").
		Find(&initialRows).Error; err != nil {
		return nil, err
	}
	initialContent := make(map[uint64]string, len(initialRows))
	for _, row := range initialRows {
		if _, exists := initialContent[row.FeedbackID]; !exists {
			initialContent[row.FeedbackID] = compactSummary(row.Content)
		}
	}

	imageRows := make([]feedbackImageSummaryRow, 0)
	if err := db.Model(&userfeedbackmodel.Attachment{}).
		Select(`user_feedback_attachments.feedback_id, COUNT(*) AS image_count,
			COALESCE((
				SELECT first_attachment.object_key
				FROM user_feedback_attachments AS first_attachment
				WHERE first_attachment.feedback_id = user_feedback_attachments.feedback_id
				ORDER BY first_attachment.sort_order ASC, first_attachment.id ASC
				LIMIT 1
			), '') AS first_object_key`).
		Where("feedback_id IN ?", feedbackIDs).
		Group("user_feedback_attachments.feedback_id").
		Find(&imageRows).Error; err != nil {
		return nil, err
	}
	imageCounts := make(map[uint64]int64, len(imageRows))
	firstObjectKeys := make(map[uint64]string, len(imageRows))
	for _, row := range imageRows {
		imageCounts[row.FeedbackID] = row.ImageCount
		firstObjectKeys[row.FeedbackID] = row.FirstObjectKey
	}

	names, err := loadUserNames(db, userIDs)
	if err != nil {
		return nil, err
	}
	summaries := make([]application.FeedbackSummary, 0, len(rows))
	for _, row := range rows {
		summary := feedbackSummary(row)
		summary.SubmitterName = names[row.SubmitterID]
		if row.HandlerID != nil {
			summary.HandlerName = names[*row.HandlerID]
		}
		summary.Summary = initialContent[row.ID]
		summary.ImageCount = imageCounts[row.ID]
		summary.FirstImageURL = feedbackAttachmentURL(ctx, firstObjectKeys[row.ID], media.FullURLWithStaticDomainContext)
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func queryPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = application.DefaultPageSize
	}
	if pageSize > application.MaxPageSize {
		pageSize = application.MaxPageSize
	}
	return page, pageSize
}

func safeFeedbackOffset(page, pageSize int) (int, error) {
	if page < 1 || pageSize < 1 || page-1 > math.MaxInt/pageSize {
		return 0, application.ErrInvalidArgument
	}
	return (page - 1) * pageSize, nil
}
