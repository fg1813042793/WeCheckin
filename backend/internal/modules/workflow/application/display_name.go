package application

import (
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func workflowNameForState(definition workflowcore.Definition, state *workflowdomain.State) string {
	if state != nil {
		if name := strings.TrimSpace(state.Instance.DefinitionName); name != "" {
			return name
		}
	}
	return definition.EffectiveName()
}
