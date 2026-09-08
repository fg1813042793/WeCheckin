package swagger

import "testing"

func TestWorkflowNotificationSwaggerDocumentsDelete(t *testing.T) {
	for document, doc := range userFeedbackSwaggerDocuments(t) {
		t.Run(document, func(t *testing.T) {
			operation := requireSwaggerOperation(t, doc, "/api/v2/admin/workflow-notifications/{id}", "delete")
			requireSwaggerParameter(t, operation, "path", "id")
			assertSwaggerSecurity(t, operation, "AdminToken")
		})
	}
}
