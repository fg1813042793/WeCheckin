package auth

import (
	"testing"

	bootstrapsvc "wecheckin/backend/internal/service/dingtalkh5/bootstrap"
)

func TestBootstrapUserDTOPreservesWorkflowActorID(t *testing.T) {
	dto := bootstrapUserDTO(&bootstrapsvc.BootstrapResponse{
		User: bootstrapsvc.UserDTO{
			ID:              "David",
			Account:         "David",
			WorkflowActorID: "66",
		},
	})

	if dto.ID != "David" {
		t.Fatalf("id = %q, want David", dto.ID)
	}
	if dto.WorkflowActorID != "66" {
		t.Fatalf("workflow actor id = %q, want 66", dto.WorkflowActorID)
	}
}
