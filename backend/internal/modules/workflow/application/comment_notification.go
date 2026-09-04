package application

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

const commentNotificationMaxRecipients = 100

func (service *Service) normalizeCommentNotification(
	ctx context.Context,
	detail *InstanceDetail,
	actorID string,
	request *CommentNotificationRequest,
) (*CommentNotificationRequest, error) {
	if request == nil {
		return nil, nil
	}
	channels, err := normalizeCommentNotificationChannels(request.Channels)
	if err != nil {
		return nil, err
	}
	recipients := uniqueTrimmedStrings(request.UserIDs)
	if len(recipients) > commentNotificationMaxRecipients {
		return nil, ErrCommentNotificationRecipientsTooMany
	}
	actorID = strings.TrimSpace(actorID)
	filtered := make([]string, 0, len(recipients))
	for _, recipientID := range recipients {
		if recipientID != actorID {
			filtered = append(filtered, recipientID)
		}
	}
	if len(filtered) == 0 {
		return nil, ErrCommentNotificationRecipientsRequired
	}
	for _, recipientID := range filtered {
		allowed, err := service.isCommentNotificationRecipient(ctx, detail, recipientID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrCommentNotificationRecipientDenied
		}
	}
	return &CommentNotificationRequest{UserIDs: filtered, Channels: channels}, nil
}

func normalizeCommentNotificationChannels(values []string) ([]string, error) {
	channels := uniqueTrimmedStrings(values)
	if len(channels) == 0 {
		return nil, ErrCommentNotificationChannelsRequired
	}
	for _, channel := range channels {
		if channel != workflowcore.NotificationChannelInApp && channel != workflowcore.NotificationChannelDingTalkOA {
			return nil, ErrCommentNotificationChannelInvalid
		}
	}
	return channels, nil
}

func (service *Service) isCommentNotificationRecipient(ctx context.Context, detail *InstanceDetail, userID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	if detail == nil || userID == "" {
		return false, nil
	}
	if userID == strings.TrimSpace(detail.Instance.StarterID) {
		return true, nil
	}
	for _, task := range detail.Tasks {
		if userID == strings.TrimSpace(task.AssigneeID) || userID == strings.TrimSpace(task.HandledBy) {
			return true, nil
		}
	}
	return service.store.HasParticipant(ctx, detail.Instance.ID, userID, string(workflowdomain.ParticipantRoleCC))
}

func commentNotificationEffects(
	detail *InstanceDetail,
	event workflowdomain.HistoryEvent,
	notification CommentNotificationRequest,
	ids workflowdomain.IDGenerator,
) *workflowdomain.State {
	workflowName := strings.TrimSpace(detail.Instance.DefinitionName)
	if workflowName == "" {
		workflowName = strings.TrimSpace(detail.Instance.DefinitionKey)
	}
	if workflowName == "" {
		workflowName = "流程"
	}
	nodeName := commentNotificationNodeName(detail, event.NodeID)
	actorName := historyActorName(detail, event.ActorID, "")
	if actorName == "" {
		actorName = "流程参与人"
	}
	config := workflowcore.NotificationConfig{
		Enabled:  true,
		Channels: append([]string(nil), notification.Channels...),
		Title:    fmt.Sprintf("《%s》有新评论", workflowName),
		Content:  commentNotificationContent(actorName, nodeName, event.Message, len(event.Images)),
	}
	state := &workflowdomain.State{Instance: workflowdomain.ProcessInstance{
		ID:                detail.Instance.ID,
		DefinitionID:      detail.Instance.DefinitionID,
		DefinitionVersion: detail.Instance.DefinitionVersion,
		DefinitionKey:     detail.Instance.DefinitionKey,
		BusinessType:      detail.Instance.BusinessType,
		BusinessKey:       detail.Instance.BusinessKey,
		StarterID:         detail.Instance.StarterID,
		OperatorID:        detail.Instance.OperatorID,
		Status:            workflowdomain.InstanceStatus(detail.Instance.Status),
		StartTime:         detail.Instance.StartTime,
	}}
	for _, recipientID := range notification.UserIDs {
		state.NotificationIntents = append(state.NotificationIntents, workflowdomain.NotificationIntent{
			ID: ids.NewID("notification"), Kind: workflowdomain.NotificationKindInstanceCommented,
			NodeID: event.NodeID, NodeName: nodeName, RecipientUserID: recipientID,
			WorkflowName: workflowName, Config: config, DedupeKeySuffix: event.ID,
		})
	}
	return state
}

func commentNotificationNodeName(detail *InstanceDetail, nodeID string) string {
	nodeID = strings.TrimSpace(nodeID)
	if detail == nil || nodeID == "" {
		return ""
	}
	for _, task := range detail.Tasks {
		if strings.TrimSpace(task.NodeID) == nodeID && strings.TrimSpace(task.NodeName) != "" {
			return strings.TrimSpace(task.NodeName)
		}
	}
	for _, node := range detail.Nodes {
		if strings.TrimSpace(node.ID) == nodeID {
			return strings.TrimSpace(node.Name)
		}
	}
	for _, node := range detail.NodeProgress {
		if strings.TrimSpace(node.NodeID) == nodeID {
			return strings.TrimSpace(node.NodeName)
		}
	}
	return ""
}

func commentNotificationContent(actorName, nodeName, comment string, imageCount int) string {
	comment = strings.Join(strings.Fields(strings.TrimSpace(comment)), " ")
	if utf8.RuneCountInString(comment) > 160 {
		comment = string([]rune(comment)[:160]) + "..."
	}
	if comment == "" {
		comment = "发布了图片评论"
	}
	if imageCount > 0 {
		comment += fmt.Sprintf("（附 %d 张图片）", imageCount)
	}
	if strings.TrimSpace(nodeName) == "" {
		return fmt.Sprintf("%s 添加评论：%s", actorName, comment)
	}
	return fmt.Sprintf("%s 在“%s”添加评论：%s", actorName, nodeName, comment)
}
