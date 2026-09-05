package infrastructure

import (
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestInstanceFormRevisionRoundTripsThroughPersistenceMapping(t *testing.T) {
	domainInstance := workflowdomain.ProcessInstance{
		ID: "instance-1", DefinitionID: 9, DefinitionVersion: 2,
		Status: workflowdomain.InstanceStatusRunning, FormRevision: 4,
	}
	row, err := instanceToModel(domainInstance, map[string]interface{}{"summary": "内容"}, 100)
	if err != nil {
		t.Fatalf("instanceToModel() error = %v", err)
	}
	if row.FormRevision != 4 {
		t.Fatalf("model form revision = %d, want 4", row.FormRevision)
	}

	state, err := stateFromModels(workflowmodel.ProcessInstance{
		ID: "instance-1", Status: string(workflowdomain.InstanceStatusRunning),
		FormDataJSON: `{}`, FormRevision: 6,
	}, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if state.Instance.FormRevision != 6 {
		t.Fatalf("domain form revision = %d, want 6", state.Instance.FormRevision)
	}
}
