package infrastructure

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
)

func TestInstanceSummariesIncludeCurrentPendingTaskContext(t *testing.T) {
	rows := []workflowmodel.ProcessInstance{
		{ID: "running", Status: workflowmodel.InstanceStatusRunning},
		{ID: "completed", Status: workflowmodel.InstanceStatusCompleted},
	}
	users := []model.User{
		{ID: 7, Name: "张三"},
		{ID: 8, Name: "李四"},
		{ID: 9, Name: "王五"},
	}
	tasks := []workflowmodel.ProcessTask{
		{InstanceID: "running", NodeID: "manager", NodeName: "上级评分", AssigneeID: "7", Status: workflowmodel.TaskStatusPending},
		{InstanceID: "running", NodeID: "manager", NodeName: "上级评分", AssigneeID: "8", Status: workflowmodel.TaskStatusPending},
		{InstanceID: "running", NodeID: "manager", NodeName: "上级评分", AssigneeID: "9", Status: workflowmodel.TaskStatusWaiting},
		{InstanceID: "completed", NodeID: "archive", NodeName: "归档", AssigneeID: "7", Status: workflowmodel.TaskStatusPending},
	}

	summaries := instanceSummariesWithCurrentTasks(rows, users, nil, tasks)
	if !reflect.DeepEqual(summaries[0].CurrentNodeNames, []string{"上级评分"}) {
		t.Fatalf("current node names = %#v", summaries[0].CurrentNodeNames)
	}
	if !reflect.DeepEqual(summaries[0].CurrentAssigneeNames, []string{"张三", "李四"}) {
		t.Fatalf("current assignee names = %#v", summaries[0].CurrentAssigneeNames)
	}
	if len(summaries[1].CurrentNodeNames) != 0 || len(summaries[1].CurrentAssigneeNames) != 0 {
		t.Fatalf("completed instance current context = %#v", summaries[1])
	}
}
