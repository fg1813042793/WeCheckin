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
		`auth.GET("/workflows/categories", workflowHandler.ListDefinitionCategories)`,
		`auth.GET("/workflows/definitions", workflowHandler.ListDefinitions)`,
		`auth.GET("/workflows/definitions/:id", workflowHandler.GetDefinition)`,
		`auth.POST("/workflows/attachments", handler.WorkflowAttachment.Upload)`,
		`auth.GET("/workflows/drafts/:definitionId", workflowHandler.GetStartDraft)`,
		`auth.PUT("/workflows/drafts/:definitionId", workflowHandler.SaveStartDraft)`,
		`auth.DELETE("/workflows/drafts/:definitionId", workflowHandler.DeleteStartDraft)`,
		`auth.POST("/workflows/instances", workflowHandler.StartInstance)`,
		`auth.GET("/workflows/instances", workflowHandler.ListMyInstances)`,
		`auth.GET("/workflows/instances/:id", workflowHandler.GetMyInstance)`,
		`auth.DELETE("/workflows/instances/:id", workflowHandler.DeleteMyInstance)`,
		`auth.POST("/workflows/instances/:id/withdraw", workflowHandler.WithdrawInstance)`,
		`auth.POST("/workflows/instances/:id/comments", workflowHandler.CommentInstance)`,
		`auth.POST("/workflows/instances/:id/reminders", workflowHandler.RemindInstance)`,
		`auth.GET("/workflows/tasks", workflowHandler.ListMyTasks)`,
		`auth.POST("/workflows/tasks/:id/complete", workflowHandler.CompleteTask)`,
		`auth.GET("/workflows/summary/definitions", workflowSummaryHandler.ListDefinitions)`,
		`auth.GET("/workflows/summary/definitions/:id", workflowSummaryHandler.GetDefinition)`,
		`auth.GET("/workflows/summary/instances", workflowSummaryHandler.ListInstances)`,
		`auth.GET("/workflows/summary/instances/:id", workflowSummaryHandler.GetInstance)`,
		`auth.GET("/workflows/summary/export", workflowSummaryHandler.Export)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 routes missing protected workflow registration %q", snippet)
		}
	}
}
