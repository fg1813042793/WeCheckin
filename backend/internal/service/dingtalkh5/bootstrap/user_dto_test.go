package bootstrap

import (
	"testing"

	"wecheckin/backend/internal/model"
)

func TestUserDTOExposesWorkflowActorIDWithoutChangingAccountID(t *testing.T) {
	dto := userDTO(&model.DingTalkH5PerfUser{ID: 66, Account: "David", Name: "David"})

	if dto.ID != "David" {
		t.Fatalf("id = %q, want David", dto.ID)
	}
	if dto.WorkflowActorID != "66" {
		t.Fatalf("workflow actor id = %q, want 66", dto.WorkflowActorID)
	}
}
