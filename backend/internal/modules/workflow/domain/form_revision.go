package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"wecheckin/backend/internal/workflowcore"
)

type FormRevisionStatus string

const (
	FormRevisionStatusPending   FormRevisionStatus = "pending"
	FormRevisionStatusApplied   FormRevisionStatus = "applied"
	FormRevisionStatusRejected  FormRevisionStatus = "rejected"
	FormRevisionStatusCancelled FormRevisionStatus = "cancelled"
	FormRevisionStatusConflict  FormRevisionStatus = "conflict"
)

type RevisionTaskStatus string

const (
	RevisionTaskStatusWaiting   RevisionTaskStatus = "waiting"
	RevisionTaskStatusPending   RevisionTaskStatus = "pending"
	RevisionTaskStatusApproved  RevisionTaskStatus = "approved"
	RevisionTaskStatusRejected  RevisionTaskStatus = "rejected"
	RevisionTaskStatusCancelled RevisionTaskStatus = "cancelled"
)

type RevisionAction string

const (
	RevisionActionApprove RevisionAction = "approve"
	RevisionActionReject  RevisionAction = "reject"
)

const (
	FormRevisionModeDirect           = workflowcore.CompletedRevisionModeDirect
	FormRevisionModeDownstreamReview = workflowcore.CompletedRevisionModeDownstreamReview
)

var (
	ErrRevisionNotPending         = errors.New("表单修订请求已结束")
	ErrRevisionTaskNotFound       = errors.New("表单修订任务不存在")
	ErrRevisionTaskAlreadyHandled = errors.New("表单修订任务已处理")
	ErrRevisionTaskActorMismatch  = errors.New("当前用户不是表单修订任务处理人")
	ErrInvalidRevisionTaskAction  = errors.New("表单修订任务操作无效")
	ErrInvalidRevisionTaskPlan    = errors.New("表单修订任务计划无效")
)

type FormRevisionRequest struct {
	ID                  string
	SourceInstanceID    string
	SourceNodeID        string
	SourceNodeName      string
	RequesterID         string
	BaseFormRevision    int64
	AppliedFormRevision int64
	BeforeFormData      map[string]interface{}
	Patch               map[string]interface{}
	ProposedFormData    map[string]interface{}
	ChangedFieldLabels  []string
	Reason              string
	Mode                string
	Status              FormRevisionStatus
	Tasks               []FormRevisionTask
	CreatedAt           int64
	CompletedAt         int64
}

type FormRevisionTask struct {
	ID                string
	RevisionRequestID string
	SourceTaskID      string
	NodeID            string
	NodeName          string
	Stage             int
	AssigneeID        string
	AssigneeName      string
	ApprovalMode      string
	CompletionRate    int
	Sequence          int
	Total             int
	Status            RevisionTaskStatus
	Action            RevisionAction
	Comment           string
	Images            []workflowcore.FormAttachment
	HandledBy         string
	HandledAt         int64
}

type CompleteRevisionTaskResult struct {
	Applied          bool
	Idempotent       bool
	ActivatedTaskIDs []string
}

func (revision *FormRevisionRequest) CompleteTask(
	taskID string,
	actorID string,
	action RevisionAction,
	comment string,
	images []workflowcore.FormAttachment,
	handledAt int64,
) (CompleteRevisionTaskResult, error) {
	if revision == nil {
		return CompleteRevisionTaskResult{}, ErrRevisionNotPending
	}
	taskIndex := revisionTaskIndex(revision.Tasks, taskID)
	if taskIndex < 0 {
		return CompleteRevisionTaskResult{}, ErrRevisionTaskNotFound
	}
	task := &revision.Tasks[taskIndex]
	if strings.TrimSpace(task.AssigneeID) != strings.TrimSpace(actorID) {
		return CompleteRevisionTaskResult{}, ErrRevisionTaskActorMismatch
	}
	if action != RevisionActionApprove && action != RevisionActionReject {
		return CompleteRevisionTaskResult{}, ErrInvalidRevisionTaskAction
	}
	if task.Status == RevisionTaskStatusApproved || task.Status == RevisionTaskStatusRejected {
		if task.Action == action && strings.TrimSpace(task.HandledBy) == strings.TrimSpace(actorID) {
			return CompleteRevisionTaskResult{Idempotent: true}, nil
		}
		return CompleteRevisionTaskResult{}, ErrRevisionTaskAlreadyHandled
	}
	if revision.Status != FormRevisionStatusPending {
		return CompleteRevisionTaskResult{}, ErrRevisionNotPending
	}
	if task.Status != RevisionTaskStatusPending {
		return CompleteRevisionTaskResult{}, ErrRevisionTaskAlreadyHandled
	}

	task.Action = action
	task.Comment = comment
	task.Images = cloneRevisionImages(images)
	task.HandledBy = strings.TrimSpace(actorID)
	task.HandledAt = handledAt
	if action == RevisionActionReject {
		task.Status = RevisionTaskStatusRejected
		revision.Status = FormRevisionStatusRejected
		revision.CompletedAt = handledAt
		cancelOpenRevisionTasks(revision.Tasks, task.ID)
		return CompleteRevisionTaskResult{}, nil
	}

	task.Status = RevisionTaskStatusApproved
	groupComplete, activated, err := advanceRevisionTaskGroup(revision.Tasks, taskIndex)
	if err != nil {
		return CompleteRevisionTaskResult{}, err
	}
	result := CompleteRevisionTaskResult{ActivatedTaskIDs: activated}
	if !groupComplete || !revisionStageComplete(revision.Tasks, task.Stage) {
		return result, nil
	}

	nextStage := nextRevisionStage(revision.Tasks, task.Stage)
	if nextStage > 0 {
		result.ActivatedTaskIDs = append(result.ActivatedTaskIDs, activateRevisionStage(revision.Tasks, nextStage)...)
		return result, nil
	}
	revision.Status = FormRevisionStatusApplied
	revision.CompletedAt = handledAt
	result.Applied = true
	return result, nil
}

func revisionTaskIndex(tasks []FormRevisionTask, taskID string) int {
	taskID = strings.TrimSpace(taskID)
	for index := range tasks {
		if strings.TrimSpace(tasks[index].ID) == taskID {
			return index
		}
	}
	return -1
}

func advanceRevisionTaskGroup(tasks []FormRevisionTask, currentIndex int) (bool, []string, error) {
	current := tasks[currentIndex]
	switch current.ApprovalMode {
	case workflowcore.ApprovalModeSingle:
		return true, nil, nil
	case workflowcore.ApprovalModeSequential:
		if next := nextWaitingRevisionTask(tasks, current.Stage, current.NodeID); next >= 0 {
			tasks[next].Status = RevisionTaskStatusPending
			return false, []string{tasks[next].ID}, nil
		}
		return revisionTaskGroupApproved(tasks, current.Stage, current.NodeID), nil, nil
	case workflowcore.ApprovalModeParallel:
		return revisionTaskGroupApproved(tasks, current.Stage, current.NodeID), nil, nil
	case workflowcore.ApprovalModeCountersign:
		if !revisionCountersignThresholdReached(tasks, current.Stage, current.NodeID, current.CompletionRate) {
			return false, nil, nil
		}
		cancelRevisionTaskGroup(tasks, current.Stage, current.NodeID)
		return true, nil, nil
	default:
		return false, nil, fmt.Errorf("%w：不支持的审批方式 %s", ErrInvalidRevisionTaskPlan, current.ApprovalMode)
	}
}

func revisionStageComplete(tasks []FormRevisionTask, stage int) bool {
	groups := make(map[string]struct{})
	for _, task := range tasks {
		if task.Stage == stage {
			groups[task.NodeID] = struct{}{}
		}
	}
	if len(groups) == 0 {
		return false
	}
	for nodeID := range groups {
		if !revisionTaskGroupComplete(tasks, stage, nodeID) {
			return false
		}
	}
	return true
}

func revisionTaskGroupComplete(tasks []FormRevisionTask, stage int, nodeID string) bool {
	mode := ""
	completionRate := 0
	for _, task := range tasks {
		if task.Stage == stage && task.NodeID == nodeID {
			mode = task.ApprovalMode
			completionRate = task.CompletionRate
			break
		}
	}
	switch mode {
	case workflowcore.ApprovalModeSingle:
		return revisionTaskGroupHasApproved(tasks, stage, nodeID)
	case workflowcore.ApprovalModeSequential, workflowcore.ApprovalModeParallel:
		return revisionTaskGroupApproved(tasks, stage, nodeID)
	case workflowcore.ApprovalModeCountersign:
		return revisionCountersignThresholdReached(tasks, stage, nodeID, completionRate)
	default:
		return false
	}
}

func revisionTaskGroupHasApproved(tasks []FormRevisionTask, stage int, nodeID string) bool {
	for _, task := range tasks {
		if task.Stage == stage && task.NodeID == nodeID && task.Status == RevisionTaskStatusApproved {
			return true
		}
	}
	return false
}

func revisionTaskGroupApproved(tasks []FormRevisionTask, stage int, nodeID string) bool {
	found := false
	for _, task := range tasks {
		if task.Stage != stage || task.NodeID != nodeID {
			continue
		}
		found = true
		if task.Status != RevisionTaskStatusApproved {
			return false
		}
	}
	return found
}

func revisionCountersignThresholdReached(tasks []FormRevisionTask, stage int, nodeID string, completionRate int) bool {
	total := 0
	approved := 0
	for _, task := range tasks {
		if task.Stage != stage || task.NodeID != nodeID {
			continue
		}
		total++
		if task.Status == RevisionTaskStatusApproved {
			approved++
		}
	}
	return total > 0 && completionRate >= 1 && completionRate <= 100 && approved*100 >= completionRate*total
}

func nextWaitingRevisionTask(tasks []FormRevisionTask, stage int, nodeID string) int {
	indexes := make([]int, 0)
	for index := range tasks {
		if tasks[index].Stage == stage && tasks[index].NodeID == nodeID && tasks[index].Status == RevisionTaskStatusWaiting {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) == 0 {
		return -1
	}
	sort.Slice(indexes, func(i, j int) bool { return tasks[indexes[i]].Sequence < tasks[indexes[j]].Sequence })
	return indexes[0]
}

func nextRevisionStage(tasks []FormRevisionTask, current int) int {
	next := 0
	for _, task := range tasks {
		if task.Stage <= current || task.Status != RevisionTaskStatusWaiting {
			continue
		}
		if next == 0 || task.Stage < next {
			next = task.Stage
		}
	}
	return next
}

func activateRevisionStage(tasks []FormRevisionTask, stage int) []string {
	groups := make(map[string][]int)
	for index := range tasks {
		if tasks[index].Stage == stage && tasks[index].Status == RevisionTaskStatusWaiting {
			groups[tasks[index].NodeID] = append(groups[tasks[index].NodeID], index)
		}
	}
	activated := make([]string, 0)
	for _, indexes := range groups {
		sort.Slice(indexes, func(i, j int) bool { return tasks[indexes[i]].Sequence < tasks[indexes[j]].Sequence })
		if len(indexes) == 0 {
			continue
		}
		if tasks[indexes[0]].ApprovalMode == workflowcore.ApprovalModeSequential {
			indexes = indexes[:1]
		}
		for _, index := range indexes {
			tasks[index].Status = RevisionTaskStatusPending
			activated = append(activated, tasks[index].ID)
		}
	}
	sort.Strings(activated)
	return activated
}

func cancelRevisionTaskGroup(tasks []FormRevisionTask, stage int, nodeID string) {
	for index := range tasks {
		if tasks[index].Stage == stage && tasks[index].NodeID == nodeID &&
			(tasks[index].Status == RevisionTaskStatusPending || tasks[index].Status == RevisionTaskStatusWaiting) {
			tasks[index].Status = RevisionTaskStatusCancelled
		}
	}
}

func cancelOpenRevisionTasks(tasks []FormRevisionTask, exceptTaskID string) {
	for index := range tasks {
		if tasks[index].ID == exceptTaskID {
			continue
		}
		if tasks[index].Status == RevisionTaskStatusPending || tasks[index].Status == RevisionTaskStatusWaiting {
			tasks[index].Status = RevisionTaskStatusCancelled
		}
	}
}

func cloneRevisionImages(source []workflowcore.FormAttachment) []workflowcore.FormAttachment {
	return append([]workflowcore.FormAttachment(nil), source...)
}
