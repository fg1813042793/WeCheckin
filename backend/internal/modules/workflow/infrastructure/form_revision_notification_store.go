package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm/clause"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/pkg/database"
)

func (store *GormStore) PersistRevisionNotifications(
	ctx context.Context,
	state *workflowdomain.State,
	revision *workflowdomain.FormRevisionRequest,
	intents []application.FormRevisionNotificationIntent,
) ([]string, error) {
	if state == nil || revision == nil {
		return nil, errors.New("表单修订通知上下文不能为空")
	}
	if len(intents) == 0 {
		return nil, nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	now := database.Now()
	outboxIDs := make([]string, 0)
	for _, intent := range intents {
		payload := intent.Payload
		if strings.TrimSpace(payload.SourceType) == "" {
			payload.SourceType = "workflow_form_revision"
		}
		if strings.TrimSpace(payload.SourceID) == "" {
			payload.SourceID = revision.ID
		}
		if strings.TrimSpace(payload.View) == "" {
			payload.View = "workflow:form-revision-detail:" + revision.ID
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		channels := uniqueSortedRevisionChannels(intent.Channels)
		for _, channel := range channels {
			outboxID := intent.ID + "-" + channel
			row := workflowmodel.NotificationOutbox{
				ID: outboxID, InstanceID: state.Instance.ID, BusinessKey: state.Instance.BusinessKey,
				NodeID: intent.NodeID, TaskID: intent.TaskID, RecipientUserID: intent.RecipientUserID,
				Kind: intent.Kind, Channel: channel, Status: workflowmodel.NotificationStatusPending,
				DedupeKey: strings.Join([]string{
					revision.ID, intent.Kind, intent.RecipientUserID, channel, intent.DedupeKeySuffix,
				}, ":"),
				PayloadJSON: string(payloadJSON), NextRetryAt: now, AddTime: now, EditTime: now,
			}
			result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil {
				return nil, result.Error
			}
			if result.RowsAffected > 0 {
				outboxIDs = append(outboxIDs, outboxID)
			}
		}
	}
	return outboxIDs, nil
}

func uniqueSortedRevisionChannels(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
