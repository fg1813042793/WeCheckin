package workflowsummary

import (
	"context"
	"strings"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/pkg/database"
)

func (service *Service) exportManagementNames(ctx context.Context, instances []workflowapp.InstanceSummary) (map[uint]string, error) {
	names := make(map[uint]string)
	// Runtime-only consumers can still export their provided historical snapshots.
	if service.db == nil {
		return names, nil
	}
	ids := make([]uint, 0, len(instances))
	seen := make(map[uint]bool)
	for _, instance := range instances {
		if instance.DefinitionID > 0 && !seen[instance.DefinitionID] {
			seen[instance.DefinitionID] = true
			ids = append(ids, instance.DefinitionID)
		}
	}
	if len(ids) == 0 {
		return names, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	var definitions []workflowmodel.Definition
	// Disabled/draft definitions may still have authorized historical records to export.
	if err := service.db.WithContext(queryCtx).Select("id", "definition_name").
		Where("id IN ? AND definition_deleted_at = ?", ids, 0).Find(&definitions).Error; err != nil {
		return nil, err
	}
	for _, definition := range definitions {
		if name := strings.TrimSpace(definition.Name); name != "" {
			names[definition.ID] = name
		}
	}
	return names, nil
}
