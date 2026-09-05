package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) RemindInstance(ctx context.Context, request RemindInstanceRequest) (*RemindInstanceResult, error) {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.NodeID = strings.TrimSpace(request.NodeID)
	if request.InstanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.NodeID == "" {
		return nil, ErrReminderNodeRequired
	}
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var result *RemindInstanceResult
	var outboxIDs []string
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, state, err := store.LoadDefinitionAndStateByInstanceForUpdate(ctx, request.InstanceID)
		if err != nil {
			return err
		}
		if state == nil || state.Instance.Status != workflowdomain.InstanceStatusRunning {
			return workflowdomain.ErrInstanceNotRunning
		}
		if strings.TrimSpace(state.Instance.StarterID) != request.ActorID {
			return ErrReminderStarterOnly
		}
		node, ok := reminderDefinitionNode(definition, request.NodeID)
		if !ok || (node.Type != workflowcore.NodeTypeApproval && node.Type != workflowcore.NodeTypeHandle) {
			return ErrReminderNodeUnavailable
		}
		recipients := reminderRecipients(state.Tasks, request.NodeID, request.ActorID)
		if len(recipients) == 0 {
			return ErrReminderNodeUnavailable
		}

		now := service.currentTime()
		nowMillis := now.UnixMilli()
		lastRemindedAt, todayCount := reminderUsage(state.History, request.NodeID, now)
		if lastRemindedAt > 0 && nowMillis < lastRemindedAt+reminderCooldown.Milliseconds() {
			return ErrReminderCooldown
		}
		if todayCount >= reminderDailyLimit {
			return ErrReminderDailyLimit
		}

		reminderID := service.ids.NewID("reminder")
		state.History = append(state.History, workflowdomain.HistoryEvent{
			ID: reminderID, Type: workflowdomain.HistoryInstanceReminded, NodeID: node.ID,
			ActorID: request.ActorID, Message: fmt.Sprintf("已提醒 %d 位处理人处理“%s”", len(recipients), node.Name), EventTime: nowMillis,
		})
		config := reminderNotificationConfig(node)
		for _, recipient := range recipients {
			state.NotificationIntents = append(state.NotificationIntents, workflowdomain.NotificationIntent{
				ID: service.ids.NewID("notification"), Kind: workflowdomain.NotificationKindTaskReminder,
				NodeID: node.ID, NodeName: node.Name, TaskID: recipient.TaskID,
				RecipientUserID: recipient.UserID, WorkflowName: definition.Name,
				Config: config, DedupeKeySuffix: reminderID,
			})
		}
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		outboxIDs, err = store.PersistEffects(ctx, state)
		if err != nil {
			return err
		}
		result = &RemindInstanceResult{
			NodeID: node.ID, NodeName: node.Name, RemindedCount: len(recipients),
			RemindedAt: nowMillis, NextAllowedAt: now.Add(reminderCooldown).UnixMilli(),
			RemainingCount: reminderDailyLimit - todayCount - 1,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	service.dispatchNotifications(ctx, outboxIDs)
	return result, nil
}

type reminderRecipient struct {
	UserID string
	TaskID string
}

func reminderRecipients(tasks []workflowdomain.Task, nodeID, actorID string) []reminderRecipient {
	seen := make(map[string]struct{})
	result := make([]reminderRecipient, 0)
	for _, task := range tasks {
		userID := strings.TrimSpace(task.AssigneeID)
		if task.Status != workflowdomain.TaskStatusPending || task.NodeID != nodeID || userID == "" || userID == actorID {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		result = append(result, reminderRecipient{UserID: userID, TaskID: task.ID})
	}
	return result
}

func reminderDefinitionNode(definition workflowcore.Definition, nodeID string) (workflowcore.Node, bool) {
	for _, node := range definition.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return workflowcore.Node{}, false
}

func reminderUsage(history []workflowdomain.HistoryEvent, nodeID string, now time.Time) (int64, int) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).UnixMilli()
	lastRemindedAt := int64(0)
	todayCount := 0
	for _, event := range history {
		if event.Type != workflowdomain.HistoryInstanceReminded || event.NodeID != nodeID {
			continue
		}
		if event.EventTime > lastRemindedAt {
			lastRemindedAt = event.EventTime
		}
		if event.EventTime >= dayStart && event.EventTime < dayEnd {
			todayCount++
		}
	}
	return lastRemindedAt, todayCount
}

func reminderNotificationConfig(node workflowcore.Node) workflowcore.NotificationConfig {
	channels := make([]string, 0, 2)
	seen := make(map[string]struct{})
	if node.Notification != nil {
		for _, channel := range node.Notification.Channels {
			channel = strings.TrimSpace(channel)
			if channel == "" {
				continue
			}
			if _, exists := seen[channel]; exists {
				continue
			}
			seen[channel] = struct{}{}
			channels = append(channels, channel)
		}
	}
	if len(channels) == 0 {
		channels = append(channels, workflowcore.NotificationChannelDingTalkOA)
	}
	return workflowcore.NotificationConfig{
		Enabled: true, Channels: channels, Title: "流程处理提醒",
		Content: "{{starterName}} 提醒你处理《{{workflowName}}》中的“{{nodeName}}”",
	}
}

func decorateInstanceReminders(detail *InstanceDetail, actorID string, now time.Time) {
	detail.ReminderPolicy = ReminderPolicy{CooldownSeconds: int(reminderCooldown.Seconds()), DailyLimit: reminderDailyLimit}
	detail.ReminderNodes = []ReminderNodeSummary{}
	if detail.Instance.StarterID != actorID || detail.Instance.Status != string(workflowdomain.InstanceStatusRunning) {
		return
	}

	type reminderGroup struct {
		nodeName string
		userIDs  map[string]struct{}
		names    []string
	}
	groups := make(map[string]*reminderGroup)
	order := make([]string, 0)
	for _, task := range detail.Tasks {
		userID := strings.TrimSpace(task.AssigneeID)
		if task.Status != string(workflowdomain.TaskStatusPending) || userID == "" || userID == actorID {
			continue
		}
		group := groups[task.NodeID]
		if group == nil {
			group = &reminderGroup{nodeName: strings.TrimSpace(task.NodeName), userIDs: make(map[string]struct{})}
			groups[task.NodeID] = group
			order = append(order, task.NodeID)
		}
		if _, exists := group.userIDs[userID]; exists {
			continue
		}
		group.userIDs[userID] = struct{}{}
		name := strings.TrimSpace(task.AssigneeName)
		if name != "" {
			group.names = append(group.names, name)
		}
	}

	nowMillis := now.UnixMilli()
	for _, nodeID := range order {
		group := groups[nodeID]
		lastRemindedAt, todayCount := reminderSummaryUsage(detail.History, nodeID, now)
		nextAllowedAt := int64(0)
		blockedReason := ""
		canRemind := true
		if lastRemindedAt > 0 {
			nextAllowedAt = lastRemindedAt + reminderCooldown.Milliseconds()
			if nowMillis < nextAllowedAt {
				canRemind = false
				blockedReason = "cooldown"
			}
		}
		if todayCount >= reminderDailyLimit {
			canRemind = false
			blockedReason = "daily_limit"
		}
		remainingCount := reminderDailyLimit - todayCount
		if remainingCount < 0 {
			remainingCount = 0
		}
		nodeName := group.nodeName
		if nodeName == "" {
			nodeName = "当前节点"
		}
		detail.ReminderNodes = append(detail.ReminderNodes, ReminderNodeSummary{
			NodeID: nodeID, NodeName: nodeName, AssigneeNames: append([]string(nil), group.names...),
			AssigneeCount: len(group.userIDs), CanRemind: canRemind, BlockedReason: blockedReason,
			LastRemindedAt: lastRemindedAt, NextAllowedAt: nextAllowedAt,
			TodayCount: todayCount, RemainingCount: remainingCount,
		})
	}
}

func reminderSummaryUsage(history []HistorySummary, nodeID string, now time.Time) (int64, int) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).UnixMilli()
	lastRemindedAt := int64(0)
	todayCount := 0
	for _, event := range history {
		if event.EventType != string(workflowdomain.HistoryInstanceReminded) || event.NodeID != nodeID {
			continue
		}
		if event.EventTime > lastRemindedAt {
			lastRemindedAt = event.EventTime
		}
		if event.EventTime >= dayStart && event.EventTime < dayEnd {
			todayCount++
		}
	}
	return lastRemindedAt, todayCount
}
