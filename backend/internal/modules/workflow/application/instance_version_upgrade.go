package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

var (
	ErrInstanceVersionAlreadyLatest = errors.New("流程实例已是最新版本")
	ErrInstanceVersionIncompatible  = errors.New("流程实例与最新版本不兼容")
	ErrInstanceVersionConflict      = errors.New("流程实例版本已变化，请刷新后重试")
)

type InstanceVersionUpgradeBlocker struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type InstanceVersionUpgradePreview struct {
	InstanceID     string                          `json:"instanceId"`
	DefinitionID   uint                            `json:"definitionId"`
	CurrentVersion int                             `json:"currentVersion"`
	TargetVersion  int                             `json:"targetVersion"`
	Compatible     bool                            `json:"compatible"`
	AlreadyLatest  bool                            `json:"alreadyLatest"`
	Blockers       []InstanceVersionUpgradeBlocker `json:"blockers"`
}

type UpgradeInstanceVersionRequest struct {
	InstanceID          string `json:"-"`
	ActorID             string `json:"-"`
	ExpectedFromVersion int    `json:"expectedFromVersion"`
}

type UpgradeInstanceVersionResult struct {
	InstanceID      string `json:"instanceId"`
	DefinitionID    uint   `json:"definitionId"`
	PreviousVersion int    `json:"previousVersion"`
	CurrentVersion  int    `json:"currentVersion"`
}

func (service *Service) PreviewInstanceVersionUpgrade(ctx context.Context, instanceID string) (*InstanceVersionUpgradePreview, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	var preview InstanceVersionUpgradePreview
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		loaded, err := loadInstanceVersionUpgradeContext(ctx, store, instanceID)
		if err != nil {
			return err
		}
		preview = analyzeInstanceVersionUpgrade(loaded.source, loaded.target, loaded.state, loaded.activeRevision, loaded.targetVersion)
		return nil
	})
	return &preview, err
}

func (service *Service) UpgradeInstanceVersion(ctx context.Context, request UpgradeInstanceVersionRequest) (*UpgradeInstanceVersionResult, error) {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	if request.InstanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.ExpectedFromVersion <= 0 {
		return nil, ErrInstanceVersionConflict
	}
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	var result UpgradeInstanceVersionResult
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		loaded, err := loadInstanceVersionUpgradeContext(ctx, store, request.InstanceID)
		if err != nil {
			return err
		}
		if loaded.state.Instance.DefinitionVersion != request.ExpectedFromVersion {
			return ErrInstanceVersionConflict
		}
		preview := analyzeInstanceVersionUpgrade(loaded.source, loaded.target, loaded.state, loaded.activeRevision, loaded.targetVersion)
		if preview.AlreadyLatest {
			return ErrInstanceVersionAlreadyLatest
		}
		if !preview.Compatible {
			return fmt.Errorf("%w: %s", ErrInstanceVersionIncompatible, joinUpgradeBlockers(preview.Blockers))
		}
		if err := store.UpdateInstanceDefinitionVersion(ctx, request.InstanceID, preview.CurrentVersion, preview.TargetVersion); err != nil {
			return err
		}
		now := service.currentTime().UnixMilli()
		event := workflowdomain.HistoryEvent{
			ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceVersionUpgraded,
			ActorID:   request.ActorID,
			Message:   fmt.Sprintf("管理员将流程实例版本由 v%d 升级至 v%d", preview.CurrentVersion, preview.TargetVersion),
			EventTime: now,
		}
		if err := store.AppendInstanceHistory(ctx, request.InstanceID, event, now); err != nil {
			return err
		}
		result = UpgradeInstanceVersionResult{
			InstanceID: request.InstanceID, DefinitionID: preview.DefinitionID,
			PreviousVersion: preview.CurrentVersion, CurrentVersion: preview.TargetVersion,
		}
		return nil
	})
	return &result, err
}

type instanceVersionUpgradeContext struct {
	source         workflowcore.Definition
	target         workflowcore.Definition
	state          *workflowdomain.State
	activeRevision *workflowdomain.FormRevisionRequest
	targetVersion  int
}

func loadInstanceVersionUpgradeContext(ctx context.Context, store TransactionStore, instanceID string) (instanceVersionUpgradeContext, error) {
	source, state, err := store.LoadDefinitionAndStateByInstanceForUpdate(ctx, instanceID)
	if err != nil {
		return instanceVersionUpgradeContext{}, err
	}
	if state == nil {
		return instanceVersionUpgradeContext{}, errors.New("流程实例不存在")
	}
	target, targetVersion, err := store.LoadPublishedDefinition(ctx, state.Instance.DefinitionID, 0)
	if err != nil {
		return instanceVersionUpgradeContext{}, err
	}
	activeRevision, err := store.LoadActiveFormRevisionForUpdate(ctx, instanceID)
	if err != nil {
		return instanceVersionUpgradeContext{}, err
	}
	return instanceVersionUpgradeContext{source: source, target: target, state: state, activeRevision: activeRevision, targetVersion: targetVersion}, nil
}

func analyzeInstanceVersionUpgrade(
	source workflowcore.Definition,
	target workflowcore.Definition,
	state *workflowdomain.State,
	activeRevision *workflowdomain.FormRevisionRequest,
	targetVersion int,
) InstanceVersionUpgradePreview {
	preview := InstanceVersionUpgradePreview{TargetVersion: targetVersion, Blockers: []InstanceVersionUpgradeBlocker{}}
	if state == nil {
		addUpgradeBlocker(&preview, "instance_missing", "流程实例不存在")
		return preview
	}
	preview.InstanceID = state.Instance.ID
	preview.DefinitionID = state.Instance.DefinitionID
	preview.CurrentVersion = state.Instance.DefinitionVersion
	if targetVersion <= state.Instance.DefinitionVersion {
		preview.AlreadyLatest = targetVersion == state.Instance.DefinitionVersion
		addUpgradeBlocker(&preview, "already_latest", "流程实例已是最新发布版本")
	}
	if state.Instance.Status != workflowdomain.InstanceStatusRunning {
		addUpgradeBlocker(&preview, "instance_not_running", "仅审批中的流程实例可以升级版本")
	}
	if activeRevision != nil && activeRevision.Status == workflowdomain.FormRevisionStatusPending {
		addUpgradeBlocker(&preview, "active_form_revision", "实例存在进行中的表单修订，完成或取消后才能升级")
	}
	sourceNodes := definitionNodeTypes(source)
	targetNodes := definitionNodeTypes(target)
	for _, token := range state.Tokens {
		if token.Status != workflowdomain.TokenStatusActive && token.Status != workflowdomain.TokenStatusWaiting {
			continue
		}
		checkUpgradeNode(&preview, token.NodeID, "active_node", sourceNodes, targetNodes)
	}
	for _, task := range state.Tasks {
		if task.Status != workflowdomain.TaskStatusPending && task.Status != workflowdomain.TaskStatusWaiting {
			continue
		}
		checkUpgradeNode(&preview, task.NodeID, "pending_task_node", sourceNodes, targetNodes)
	}
	targetFields := definitionFormKeys(target.Form)
	formKeys := make([]string, 0, len(state.FormData))
	for key := range state.FormData {
		formKeys = append(formKeys, key)
	}
	sort.Strings(formKeys)
	for _, key := range formKeys {
		if _, exists := targetFields[key]; !exists {
			addUpgradeBlocker(&preview, "form_field_missing", fmt.Sprintf("最新版表单缺少实例已有字段 %s", key))
		}
	}
	preview.Compatible = len(preview.Blockers) == 0
	return preview
}

func checkUpgradeNode(preview *InstanceVersionUpgradePreview, nodeID, prefix string, sourceNodes, targetNodes map[string]string) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return
	}
	sourceType, sourceExists := sourceNodes[nodeID]
	targetType, targetExists := targetNodes[nodeID]
	if !targetExists {
		addUpgradeBlocker(preview, prefix+"_missing", fmt.Sprintf("最新版流程缺少当前运行节点 %s", nodeID))
		return
	}
	if !sourceExists || sourceType != targetType {
		addUpgradeBlocker(preview, prefix+"_type_changed", fmt.Sprintf("节点 %s 的类型已变化", nodeID))
	}
}

func definitionNodeTypes(definition workflowcore.Definition) map[string]string {
	result := make(map[string]string, len(definition.Nodes))
	for _, node := range definition.Nodes {
		if id := strings.TrimSpace(node.ID); id != "" {
			result[id] = strings.TrimSpace(node.Type)
		}
	}
	return result
}

func definitionFormKeys(fields []workflowcore.FormField) map[string]struct{} {
	result := make(map[string]struct{})
	var collect func([]workflowcore.FormField)
	collect = func(items []workflowcore.FormField) {
		for _, field := range items {
			if key := strings.TrimSpace(field.Key); key != "" {
				result[key] = struct{}{}
			}
			collect(field.Fields)
			collect(field.Columns)
		}
	}
	collect(fields)
	return result
}

func addUpgradeBlocker(preview *InstanceVersionUpgradePreview, code, message string) {
	for _, blocker := range preview.Blockers {
		if blocker.Code == code && blocker.Message == message {
			return
		}
	}
	preview.Blockers = append(preview.Blockers, InstanceVersionUpgradeBlocker{Code: code, Message: message})
}

func joinUpgradeBlockers(blockers []InstanceVersionUpgradeBlocker) string {
	messages := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		if message := strings.TrimSpace(blocker.Message); message != "" {
			messages = append(messages, message)
		}
	}
	return strings.Join(messages, "；")
}
