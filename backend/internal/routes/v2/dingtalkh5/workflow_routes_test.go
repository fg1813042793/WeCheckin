package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5WorkflowRoutesUseProtectedGroup(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		`auth.GET("/workflows/definitions", workflowHandler.ListDefinitions)`,
		`auth.GET("/workflows/definitions/:id", workflowHandler.GetDefinition)`,
		`auth.GET("/workflows/drafts/:definitionId", workflowHandler.GetStartDraft)`,
		`auth.PUT("/workflows/drafts/:definitionId", workflowHandler.SaveStartDraft)`,
		`auth.POST("/workflows/instances", workflowHandler.StartInstance)`,
		`auth.GET("/workflows/instances", workflowHandler.ListMyInstances)`,
		`auth.GET("/workflows/instances/:id", workflowHandler.GetMyInstance)`,
		`auth.POST("/workflows/instances/:id/withdraw", workflowHandler.WithdrawInstance)`,
		`auth.GET("/workflows/tasks", workflowHandler.ListMyTasks)`,
		`auth.POST("/workflows/tasks/:id/complete", workflowHandler.CompleteTask)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 routes missing protected workflow registration %q", snippet)
		}
	}
}
