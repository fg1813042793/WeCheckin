package infrastructure

import (
	"context"
	"strings"

	userfeedbackmodel "wecheckin/backend/internal/model/userfeedback"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
	"wecheckin/backend/internal/support/media"
)

const feedbackSummaryMaxRunes = 100

type feedbackSnapshotRow struct {
	userfeedbackmodel.Feedback
	AttachmentCount int64 `gorm:"column:attachment_count"`
}

func feedbackModel(record application.FeedbackRecord) userfeedbackmodel.Feedback {
	return userfeedbackmodel.Feedback{
		FeedbackNo:      record.FeedbackNo,
		SubmitterID:     record.SubmitterID,
		CreateRequestID: record.CreateRequestID,
		Status:          string(record.Status),
		Version:         record.Version,
		LastActivityAt:  record.LastActivityAt,
		ResolvedAt:      record.ResolvedAt,
		ClosedAt:        record.ClosedAt,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

func messageModel(record application.MessageRecord) userfeedbackmodel.Message {
	return userfeedbackmodel.Message{
		FeedbackID:  record.FeedbackID,
		MessageType: string(record.MessageType),
		AuthorType:  string(record.AuthorType),
		AuthorID:    record.AuthorID,
		Content:     record.Content,
		FromStatus:  string(record.FromStatus),
		ToStatus:    string(record.ToStatus),
		RequestID:   record.RequestID,
		CreatedAt:   record.CreatedAt,
	}
}

func attachmentModels(records []application.AttachmentRecord) []userfeedbackmodel.Attachment {
	rows := make([]userfeedbackmodel.Attachment, 0, len(records))
	for _, record := range records {
		rows = append(rows, userfeedbackmodel.Attachment{
			FeedbackID:      record.FeedbackID,
			MessageID:       record.MessageID,
			StorageProvider: record.StorageProvider,
			ObjectKey:       record.ObjectKey,
			OriginalName:    record.OriginalName,
			ContentType:     record.ContentType,
			SizeBytes:       record.SizeBytes,
			SortOrder:       record.SortOrder,
			CreatedAt:       record.CreatedAt,
		})
	}
	return rows
}

func feedbackSnapshot(row feedbackSnapshotRow) application.FeedbackSnapshot {
	return application.FeedbackSnapshot{
		ID:              row.ID,
		FeedbackNo:      row.FeedbackNo,
		SubmitterID:     row.SubmitterID,
		Status:          domain.Status(row.Status),
		HandlerID:       row.HandlerID,
		Version:         row.Version,
		AttachmentCount: row.AttachmentCount,
		LastActivityAt:  row.LastActivityAt,
		ResolvedAt:      row.ResolvedAt,
		ClosedAt:        row.ClosedAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func feedbackSummary(row userfeedbackmodel.Feedback) application.FeedbackSummary {
	return application.FeedbackSummary{
		ID:             row.ID,
		FeedbackNo:     row.FeedbackNo,
		SubmitterID:    row.SubmitterID,
		Status:         domain.Status(row.Status),
		HandlerID:      row.HandlerID,
		Version:        row.Version,
		LastActivityAt: row.LastActivityAt,
		ResolvedAt:     row.ResolvedAt,
		ClosedAt:       row.ClosedAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func buildFeedbackDetail(
	ctx context.Context,
	feedback userfeedbackmodel.Feedback,
	messages []userfeedbackmodel.Message,
	attachments []userfeedbackmodel.Attachment,
	names map[uint]string,
) *application.FeedbackDetail {
	return buildFeedbackDetailWithURL(ctx, feedback, messages, attachments, names, media.FullURLWithStaticDomainContext)
}

func buildFeedbackDetailWithURL(
	ctx context.Context,
	feedback userfeedbackmodel.Feedback,
	messages []userfeedbackmodel.Message,
	attachments []userfeedbackmodel.Attachment,
	names map[uint]string,
	fullURL func(context.Context, string) string,
) *application.FeedbackDetail {
	summary := feedbackSummary(feedback)
	summary.SubmitterName = names[feedback.SubmitterID]
	if feedback.HandlerID != nil {
		summary.HandlerName = names[*feedback.HandlerID]
	}

	attachmentsByMessage := make(map[uint64][]application.Attachment)
	for _, row := range attachments {
		path := "/" + row.ObjectKey
		url := ""
		if fullURL != nil {
			url = fullURL(ctx, path)
		}
		attachmentsByMessage[row.MessageID] = append(attachmentsByMessage[row.MessageID], application.Attachment{
			ID:           row.ID,
			ObjectKey:    row.ObjectKey,
			OriginalName: row.OriginalName,
			ContentType:  row.ContentType,
			SizeBytes:    row.SizeBytes,
			URL:          url,
			SortOrder:    row.SortOrder,
			CreatedAt:    row.CreatedAt,
		})
	}

	detailMessages := make([]application.Message, 0, len(messages))
	for _, row := range messages {
		messageAttachments := attachmentsByMessage[row.ID]
		if messageAttachments == nil {
			messageAttachments = []application.Attachment{}
		}
		if summary.Summary == "" && domain.MessageType(row.MessageType) == domain.MessageTypeInitial {
			summary.Summary = compactSummary(row.Content)
		}
		detailMessages = append(detailMessages, application.Message{
			ID:          row.ID,
			MessageType: domain.MessageType(row.MessageType),
			AuthorType:  domain.AuthorType(row.AuthorType),
			AuthorID:    row.AuthorID,
			AuthorName:  names[row.AuthorID],
			Content:     row.Content,
			FromStatus:  domain.Status(row.FromStatus),
			ToStatus:    domain.Status(row.ToStatus),
			Attachments: messageAttachments,
			CreatedAt:   row.CreatedAt,
		})
	}
	return &application.FeedbackDetail{
		FeedbackSummary: summary,
		Messages:        detailMessages,
		AllowsSupplement: domain.Status(feedback.Status).
			AllowsSupplement(),
	}
}

func compactSummary(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > feedbackSummaryMaxRunes {
		runes = runes[:feedbackSummaryMaxRunes]
	}
	return string(runes)
}
