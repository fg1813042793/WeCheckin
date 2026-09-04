package httpclient

import (
	"context"
	"strings"
	"testing"
)

func (stub *runtimeServiceStub) ListPublishedDefinitionCategories(context.Context) ([]string, error) {
	return []string{"finance", "hr"}, nil
}

func TestListDefinitionCategoriesReturnsPublishedCategories(t *testing.T) {
	handler := NewRuntimeHandler(&runtimeServiceStub{})
	requestContext := newUserContext(42)

	handler.ListDefinitionCategories(context.Background(), requestContext)

	body := string(requestContext.Response.Body())
	if !strings.Contains(body, `"data":["finance","hr"]`) {
		t.Fatalf("category response = %s", body)
	}
}

func TestWorkflowListQueriesParseSearchFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)

	instanceContext := newUserContext(42)
	instanceContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/instances?scope=handled&definitionName=%E7%BB%A9%E6%95%88&definitionCategory=performance&starterName=%E5%BC%A0&startTimeFrom=1000&startTimeTo=1999&page=2&pageSize=10")
	handler.ListMyInstances(context.Background(), instanceContext)
	if stub.instanceQuery.DefinitionName != "绩效" ||
		stub.instanceQuery.DefinitionCategory != "performance" || stub.instanceQuery.StarterName != "张" ||
		stub.instanceQuery.StartTimeFrom != 1000 || stub.instanceQuery.StartTimeTo != 1999 {
		t.Fatalf("instance search filters = %+v", stub.instanceQuery)
	}

	taskContext := newUserContext(42)
	taskContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/tasks?assigneeId=999&status=pending&definitionName=%E7%BB%A9%E6%95%88&definitionCategory=performance&starterName=%E6%9D%8E&startTimeFrom=2000&startTimeTo=2999&page=3&pageSize=15")
	handler.ListMyTasks(context.Background(), taskContext)
	if stub.taskQuery.DefinitionName != "绩效" ||
		stub.taskQuery.AssigneeID != "" || stub.taskQuery.Status != "pending" ||
		stub.taskQuery.DefinitionCategory != "performance" || stub.taskQuery.StarterName != "李" ||
		stub.taskQuery.StartTimeFrom != 2000 || stub.taskQuery.StartTimeTo != 2999 {
		t.Fatalf("task search filters = %+v", stub.taskQuery)
	}
}
