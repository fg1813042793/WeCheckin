package domain

import (
	"fmt"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestRevisionApproveActivatesNextStageAndAppliesAtEnd(t *testing.T) {
	revision := pendingRevisionWithStages(1, 1)

	result, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil || result.Applied || revision.Tasks[1].Status != RevisionTaskStatusPending {
		t.Fatalf("first stage result=%#v revision=%#v err=%v", result, revision, err)
	}
	result, err = revision.CompleteTask("task-2-1", "user-2-1", RevisionActionApprove, "同意", nil, 2000)
	if err != nil || !result.Applied || revision.Status != FormRevisionStatusApplied || revision.CompletedAt != 2000 {
		t.Fatalf("final result=%#v revision=%#v err=%v", result, revision, err)
	}
}

func TestRevisionRejectCancelsRemainingTasks(t *testing.T) {
	revision := pendingRevisionWithStages(2, 1)

	_, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionReject, "数据不正确", nil, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if revision.Status != FormRevisionStatusRejected {
		t.Fatalf("status=%s", revision.Status)
	}
	for _, task := range revision.Tasks {
		if task.ID != "task-1-1" && task.Status != RevisionTaskStatusCancelled {
			t.Fatalf("remaining task=%#v", task)
		}
	}
}

func TestRevisionCompleteTaskIsIdempotent(t *testing.T) {
	revision := pendingRevisionWithStages(1)

	first, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil {
		t.Fatal(err)
	}
	second, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Applied || !second.Idempotent || revision.Status != FormRevisionStatusApplied {
		t.Fatalf("first=%#v second=%#v revision=%#v", first, second, revision)
	}
}

func TestRevisionSequentialApprovalActivatesOneAssigneeAtATime(t *testing.T) {
	revision := pendingRevisionWithTasks([]FormRevisionTask{
		revisionTask("task-1", 1, "node-1", "user-1", workflowcore.ApprovalModeSequential, 1, 2, RevisionTaskStatusPending),
		revisionTask("task-2", 1, "node-1", "user-2", workflowcore.ApprovalModeSequential, 2, 2, RevisionTaskStatusWaiting),
	})

	result, err := revision.CompleteTask("task-1", "user-1", RevisionActionApprove, "", nil, 1000)
	if err != nil || result.Applied || revision.Tasks[1].Status != RevisionTaskStatusPending {
		t.Fatalf("first result=%#v revision=%#v err=%v", result, revision, err)
	}
	result, err = revision.CompleteTask("task-2", "user-2", RevisionActionApprove, "", nil, 2000)
	if err != nil || !result.Applied {
		t.Fatalf("second result=%#v revision=%#v err=%v", result, revision, err)
	}
}

func TestRevisionCountersignCancelsRemainingTasksWhenThresholdReached(t *testing.T) {
	revision := pendingRevisionWithTasks([]FormRevisionTask{
		revisionTask("task-1", 1, "node-1", "user-1", workflowcore.ApprovalModeCountersign, 1, 3, RevisionTaskStatusPending),
		revisionTask("task-2", 1, "node-1", "user-2", workflowcore.ApprovalModeCountersign, 2, 3, RevisionTaskStatusPending),
		revisionTask("task-3", 1, "node-1", "user-3", workflowcore.ApprovalModeCountersign, 3, 3, RevisionTaskStatusPending),
	})
	for index := range revision.Tasks {
		revision.Tasks[index].CompletionRate = 60
	}

	result, err := revision.CompleteTask("task-1", "user-1", RevisionActionApprove, "", nil, 1000)
	if err != nil || result.Applied {
		t.Fatalf("first result=%#v err=%v", result, err)
	}
	result, err = revision.CompleteTask("task-2", "user-2", RevisionActionApprove, "", nil, 2000)
	if err != nil || !result.Applied || revision.Tasks[2].Status != RevisionTaskStatusCancelled {
		t.Fatalf("threshold result=%#v revision=%#v err=%v", result, revision, err)
	}
}

func TestRevisionCompleteTaskRejectsWrongActorAndAction(t *testing.T) {
	revision := pendingRevisionWithStages(1)
	if _, err := revision.CompleteTask("task-1-1", "other", RevisionActionApprove, "", nil, 1000); err != ErrRevisionTaskActorMismatch {
		t.Fatalf("actor error=%v", err)
	}
	if _, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionAction("submit"), "", nil, 1000); err != ErrInvalidRevisionTaskAction {
		t.Fatalf("action error=%v", err)
	}
}

func pendingRevisionWithStages(taskCounts ...int) *FormRevisionRequest {
	tasks := make([]FormRevisionTask, 0)
	for stageIndex, count := range taskCounts {
		for taskIndex := 0; taskIndex < count; taskIndex++ {
			status := RevisionTaskStatusWaiting
			if stageIndex == 0 {
				status = RevisionTaskStatusPending
			}
			tasks = append(tasks, revisionTask(
				fmt.Sprintf("task-%d-%d", stageIndex+1, taskIndex+1),
				stageIndex+1,
				fmt.Sprintf("node-%d-%d", stageIndex+1, taskIndex+1),
				fmt.Sprintf("user-%d-%d", stageIndex+1, taskIndex+1),
				workflowcore.ApprovalModeParallel,
				1,
				1,
				status,
			))
		}
	}
	return pendingRevisionWithTasks(tasks)
}

func pendingRevisionWithTasks(tasks []FormRevisionTask) *FormRevisionRequest {
	return &FormRevisionRequest{ID: "revision-1", Status: FormRevisionStatusPending, Tasks: tasks}
}

func revisionTask(id string, stage int, nodeID, assigneeID, mode string, sequence, total int, status RevisionTaskStatus) FormRevisionTask {
	return FormRevisionTask{
		ID: id, Stage: stage, NodeID: nodeID, AssigneeID: assigneeID,
		ApprovalMode: mode, CompletionRate: 100, Sequence: sequence, Total: total, Status: status,
	}
}
