package application

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

const formRevisionNotificationMaxRecipients = 100

func (service *Service) ReviseInstanceForm(ctx context.Context, request ReviseInstanceFormRequest) (*workflowdomain.State, error) {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Reason = strings.TrimSpace(request.Reason)
	if request.InstanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.ExpectedRevision <= 0 {
		return nil, ErrInstanceFormRevisionRequired
	}
	if len(request.FormData) == 0 {
		return nil, ErrInstanceFormRevisionDataRequired
	}
	if request.Reason == "" {
		return nil, ErrInstanceFormRevisionReasonRequired
	}
	if utf8.RuneCountInString(request.Reason) > 500 {
		return nil, ErrInstanceFormRevisionReasonTooLong
	}
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	notification, err := normalizeFormRevisionNotification(request.Notification)
	if err != nil {
		return nil, err
	}

	var state *workflowdomain.State
	var outboxIDs []string
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, loaded, err := store.LoadDefinitionAndStateByInstanceForUpdate(ctx, request.InstanceID)
		if err != nil {
			return err
		}
		if loaded.Instance.Status != workflowdomain.InstanceStatusRunning {
			return workflowdomain.ErrInstanceNotRunning
		}
		if loaded.Instance.FormRevision != request.ExpectedRevision {
			return ErrInstanceFormRevisionChanged
		}
		handledNodeIDs := formRevisionHandledNodeIDs(loaded, request.ActorID)
		permissions := workflowcore.PostHandleEditablePermissions(definition, handledNodeIDs)
		if len(permissions) == 0 {
			return ErrInstanceFormRevisionNotAllowed
		}
		patch := changedFormDataPatch(loaded.FormData, request.FormData)
		if len(patch) == 0 {
			return ErrInstanceFormRevisionDataRequired
		}
		if err := workflowcore.ValidatePostHandleFormPatch(definition, handledNodeIDs, loaded.FormData, patch); err != nil {
			return err
		}
		calculatedData, err := workflowcore.ApplyFormCalculations(definition.Form, workflowcore.MergeFormData(loaded.FormData, patch))
		if err != nil {
			return err
		}
		patch = changedFormDataPatch(loaded.FormData, calculatedData)
		if notification != nil {
			if err := validateFormRevisionRecipients(ctx, store, loaded, request.ActorID, notification.UserIDs); err != nil {
				return err
			}
		}

		loaded.FormData = workflowcore.MergeFormData(loaded.FormData, patch)
		loaded.Instance.FormRevision++
		fieldLabels := revisedFieldLabels(definition.Form, patch)
		nodeID, nodeName := formRevisionSourceNode(definition, handledNodeIDs)
		event := workflowdomain.HistoryEvent{
			ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevised,
			NodeID: nodeID, ActorID: request.ActorID,
			Message:   fmt.Sprintf("修改字段：%s；修改原因：%s", strings.Join(fieldLabels, "、"), request.Reason),
			EventTime: service.currentTime().UnixMilli(),
		}
		loaded.History = append(loaded.History, event)
		if notification != nil {
			appendFormRevisionNotificationIntents(loaded, definition.Name, nodeID, nodeName, event, fieldLabels, request.Reason, *notification, service.ids)
		}
		state = loaded
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		outboxIDs, err = store.PersistEffects(ctx, state)
		return err
	})
	if err != nil {
		return nil, err
	}
	service.dispatchNotifications(ctx, outboxIDs)
	return state, nil
}

func formRevisionHandledNodeIDs(state *workflowdomain.State, actorID string) []string {
	actorID = strings.TrimSpace(actorID)
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, event := range state.History {
		if strings.TrimSpace(event.ActorID) != actorID {
			continue
		}
		switch event.Type {
		case workflowdomain.HistoryTaskApproved, workflowdomain.HistoryTaskSubmitted, workflowdomain.HistoryTaskReturned:
		default:
			continue
		}
		nodeID := strings.TrimSpace(event.NodeID)
		if nodeID == "" {
			continue
		}
		if _, exists := seen[nodeID]; exists {
			continue
		}
		seen[nodeID] = struct{}{}
		result = append(result, nodeID)
	}
	return result
}

func changedFormDataPatch(current, submitted map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for field, value := range submitted {
		if currentValue, exists := current[field]; !exists || !reflect.DeepEqual(currentValue, value) {
			result[field] = value
		}
	}
	return result
}

func normalizeFormRevisionNotification(request *FormRevisionNotificationRequest) (*FormRevisionNotificationRequest, error) {
	if request == nil {
		return nil, nil
	}
	userIDs := uniqueTrimmedStrings(request.UserIDs)
	if len(userIDs) == 0 {
		return nil, nil
	}
	if len(userIDs) > formRevisionNotificationMaxRecipients {
		return nil, ErrFormRevisionNotificationTooMany
	}
	channels := uniqueTrimmedStrings(request.Channels)
	if len(channels) == 0 {
		return nil, ErrFormRevisionNotificationChannels
	}
	for _, channel := range channels {
		if channel != workflowcore.NotificationChannelInApp && channel != workflowcore.NotificationChannelDingTalkOA {
			return nil, ErrFormRevisionNotificationChannel
		}
	}
	return &FormRevisionNotificationRequest{UserIDs: userIDs, Channels: channels}, nil
}

func validateFormRevisionRecipients(ctx context.Context, store TransactionStore, state *workflowdomain.State, actorID string, recipients []string) error {
	allowed := workflowParticipantUserIDs(state)
	for _, recipientID := range recipients {
		if recipientID == strings.TrimSpace(actorID) {
			return ErrFormRevisionNotificationRecipient
		}
		if allowed[recipientID] {
			continue
		}
		copied, err := store.HasParticipant(ctx, state.Instance.ID, recipientID, string(workflowdomain.ParticipantRoleCC))
		if err != nil {
			return err
		}
		if !copied {
			return ErrFormRevisionNotificationRecipient
		}
	}
	return nil
}

func workflowParticipantUserIDs(state *workflowdomain.State) map[string]bool {
	result := make(map[string]bool)
	if state == nil {
		return result
	}
	for _, userID := range []string{state.Instance.StarterID, state.Instance.OperatorID} {
		if userID = strings.TrimSpace(userID); userID != "" {
			result[userID] = true
		}
	}
	for _, task := range state.Tasks {
		if userID := strings.TrimSpace(task.AssigneeID); userID != "" {
			result[userID] = true
		}
	}
	for _, participant := range state.Participants {
		if userID := strings.TrimSpace(participant.UserID); userID != "" {
			result[userID] = true
		}
	}
	for _, event := range state.History {
		if userID := strings.TrimSpace(event.ActorID); userID != "" {
			result[userID] = true
		}
	}
	return result
}

func revisedFieldLabels(fields []workflowcore.FormField, patch map[string]interface{}) []string {
	labels := make([]string, 0, len(patch))
	for _, field := range fields {
		if field.Type == workflowcore.FormFieldTypeGroup {
			labels = append(labels, revisedFieldLabels(field.Fields, patch)...)
			continue
		}
		if _, changed := patch[field.Key]; changed {
			label := strings.TrimSpace(field.Label)
			if label == "" {
				label = field.Key
			}
			labels = append(labels, label)
		}
	}
	if len(labels) < len(patch) {
		known := make(map[string]struct{}, len(labels))
		for _, label := range labels {
			known[label] = struct{}{}
		}
		unknown := make([]string, 0)
		for field := range patch {
			if _, exists := known[field]; !exists {
				unknown = append(unknown, field)
			}
		}
		sort.Strings(unknown)
		labels = append(labels, unknown...)
	}
	return labels
}

func formRevisionSourceNode(definition workflowcore.Definition, handledNodeIDs []string) (string, string) {
	handled := make(map[string]struct{}, len(handledNodeIDs))
	for _, nodeID := range handledNodeIDs {
		handled[nodeID] = struct{}{}
	}
	for index := len(definition.Nodes) - 1; index >= 0; index-- {
		node := definition.Nodes[index]
		if _, ok := handled[node.ID]; ok && node.PostHandleEdit != nil && node.PostHandleEdit.Enabled {
			return node.ID, node.Name
		}
	}
	return "", ""
}

func appendFormRevisionNotificationIntents(
	state *workflowdomain.State,
	workflowName, nodeID, nodeName string,
	event workflowdomain.HistoryEvent,
	fieldLabels []string,
	reason string,
	notification FormRevisionNotificationRequest,
	ids workflowdomain.IDGenerator,
) {
	workflowName = strings.TrimSpace(workflowName)
	if workflowName == "" {
		workflowName = state.Instance.DefinitionKey
	}
	config := workflowcore.NotificationConfig{
		Enabled: true, Channels: append([]string(nil), notification.Channels...),
		Title:   fmt.Sprintf("《%s》表单有更新", workflowName),
		Content: fmt.Sprintf("流程参与人 %s 修改了字段：%s。修改原因：%s", event.ActorID, strings.Join(fieldLabels, "、"), reason),
	}
	for _, recipientID := range notification.UserIDs {
		state.NotificationIntents = append(state.NotificationIntents, workflowdomain.NotificationIntent{
			ID: ids.NewID("notification"), Kind: workflowdomain.NotificationKindInstanceFormRevised,
			NodeID: nodeID, NodeName: nodeName, RecipientUserID: recipientID,
			WorkflowName: workflowName, Config: config, DedupeKeySuffix: event.ID,
		})
	}
}
