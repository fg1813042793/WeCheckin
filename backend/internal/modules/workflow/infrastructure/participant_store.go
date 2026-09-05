package infrastructure

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"unicode/utf8"
	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func (store *GormStore) HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	var count int64
	err = db.Model(&workflowmodel.InstanceParticipant{}).
		Where("instance_id = ? AND user_id = ? AND participant_role = ?", strings.TrimSpace(instanceID), strings.TrimSpace(userID), strings.TrimSpace(role)).
		Count(&count).Error
	return count > 0, err
}

func workflowUserDisplayName(db *gorm.DB, userID string) string {
	fallback := strings.TrimSpace(userID)
	id, err := strconv.ParseUint(fallback, 10, 64)
	if err != nil || id == 0 {
		return fallback
	}
	var user model.User
	if err := db.Select("id", "user_name", "user_account").First(&user, uint(id)).Error; err != nil {
		return fallback
	}
	if name := workflowUserName(user); name != "" {
		return name
	}
	return fallback
}

func workflowUserName(user model.User) string {
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	return strings.TrimSpace(user.Account)
}

func renderNotificationPayload(state *workflowdomain.State, intent workflowdomain.NotificationIntent, starterName string) application.NotificationPayload {
	result := ""
	switch intent.Kind {
	case workflowdomain.NotificationKindApprovalResultApproved:
		result = "已通过"
	case workflowdomain.NotificationKindApprovalResultRejected:
		result = "已驳回"
	case workflowdomain.NotificationKindApprovalResultReturned:
		result = "已退回"
		if targetNodeName := strings.TrimSpace(intent.TargetNodeName); targetNodeName != "" {
			result += fmt.Sprintf("至“%s”", targetNodeName)
		}
	}
	replacements := map[string]string{
		"{{workflowName}}": intent.WorkflowName,
		"{{nodeName}}":     intent.NodeName,
		"{{starterName}}":  starterName,
		"{{instanceId}}":   state.Instance.ID,
		"{{taskId}}":       intent.TaskID,
		"{{result}}":       result,
	}
	render := func(value string, limit int) string {
		for token, replacement := range replacements {
			value = strings.ReplaceAll(value, token, replacement)
		}
		value = strings.TrimSpace(value)
		if utf8.RuneCountInString(value) <= limit {
			return value
		}
		return string([]rune(value)[:limit])
	}
	payload := application.NotificationPayload{
		Title: render(intent.Config.Title, 64), Content: render(intent.Config.Content, 1000),
		WorkflowName: intent.WorkflowName, NodeName: intent.NodeName,
		StarterID: state.Instance.StarterID, StarterName: starterName,
		InstanceID: state.Instance.ID, TaskID: intent.TaskID,
		RecipientUserID: intent.RecipientUserID, Kind: string(intent.Kind),
	}
	if intent.Kind == workflowdomain.NotificationKindInstanceCommented || intent.Kind == workflowdomain.NotificationKindInstanceFormRevised {
		payload.MessageType = application.NotificationMessageTypeActionCard
	}
	return payload
}
