package infrastructure

import (
	"gorm.io/gorm"
	"strings"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

func applyInstanceFilters(db *gorm.DB, query application.InstanceQuery) *gorm.DB {
	db = db.Where("admin_deleted_at = 0")
	if query.DefinitionID > 0 {
		db = db.Where("definition_id = ?", query.DefinitionID)
	}
	if query.DefinitionVersion > 0 {
		db = db.Where("definition_version = ?", query.DefinitionVersion)
	}
	if len(query.InstanceIDs) > 0 {
		db = db.Where("id IN ?", query.InstanceIDs)
	}
	if value := strings.TrimSpace(query.DefinitionName); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM workflow_definitions name_definition
			WHERE name_definition.id = workflow_process_instances.definition_id
			AND name_definition.definition_name LIKE ? ESCAPE '!'
		)`, containsLikePattern(value))
	}
	if value := strings.TrimSpace(query.DefinitionCategory); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM workflow_definitions scope_definition
			WHERE scope_definition.id = workflow_process_instances.definition_id
			AND scope_definition.definition_category = ?
		)`, value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		db = db.Where("instance_status = ?", value)
	}
	if value := strings.TrimSpace(query.BusinessType); value != "" {
		db = db.Where("business_type = ?", value)
	}
	if value := strings.TrimSpace(query.BusinessKey); value != "" {
		db = db.Where("business_key = ?", value)
	}
	if value := strings.TrimSpace(query.StarterID); value != "" {
		db = db.Where("starter_id = ?", value)
	}
	if value := strings.TrimSpace(query.StarterName); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM users starter_user
			WHERE starter_user.id = CAST(workflow_process_instances.starter_id AS UNSIGNED)
			AND starter_user.user_name LIKE ? ESCAPE '!'
		)`, containsLikePattern(value))
	}
	if query.StartTimeFrom > 0 {
		db = db.Where("start_time >= ?", query.StartTimeFrom)
	}
	if query.StartTimeTo > 0 {
		db = db.Where("start_time <= ?", query.StartTimeTo)
	}
	if query.EndTimeFrom > 0 {
		db = db.Where("end_time >= ?", query.EndTimeFrom)
	}
	if query.EndTimeTo > 0 {
		db = db.Where("end_time <= ?", query.EndTimeTo)
	}
	if userID := strings.TrimSpace(query.ScopeUserID); userID != "" {
		switch query.Scope {
		case application.InstanceScopeStarted:
			db = db.Where("starter_id = ? AND starter_deleted_at = ?", userID, int64(0))
		case application.InstanceScopeHandled:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_process_tasks scope_task
				WHERE scope_task.instance_id = workflow_process_instances.id
				AND scope_task.task_status IN (?, ?, ?, ?)
				AND (scope_task.task_assignee_id = ? OR scope_task.handled_by = ?)
			)`,
				workflowmodel.TaskStatusCompleted,
				workflowmodel.TaskStatusApproved,
				workflowmodel.TaskStatusRejected,
				workflowmodel.TaskStatusReturned,
				userID,
				userID,
			)
		case application.InstanceScopeCopied:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_instance_participants scope_participant
				WHERE scope_participant.instance_id = workflow_process_instances.id
				AND scope_participant.user_id = ? AND scope_participant.participant_role = ?
			)`, userID, workflowmodel.ParticipantRoleCC)
		}
	}
	if where, args := instanceVisibilityWhere(query.Visibility); where != "" {
		db = db.Where(where, args...)
	}
	return db
}

func instanceVisibilityWhere(visibility *application.InstanceVisibility) (string, []interface{}) {
	if visibility == nil || (visibility.Ready && visibility.All) {
		return "", nil
	}
	if !visibility.Ready {
		return "1 = 0", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if userIDs := normalizeVisibilityUintIDs(visibility.UserIDs); len(userIDs) > 0 {
		clauses = append(clauses, "CAST(starter_id AS UNSIGNED) IN ?")
		args = append(args, userIDs)
	}
	if departmentIDs := normalizeVisibilityUintIDs(visibility.DepartmentIDs); len(departmentIDs) > 0 {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM user_depts summary_user_dept
			WHERE summary_user_dept.user_dept_user_id = CAST(workflow_process_instances.starter_id AS UNSIGNED)
			AND summary_user_dept.user_dept_dept_id IN ?
		)`)
		args = append(args, departmentIDs)
	}
	if len(clauses) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(clauses, " OR ") + ")", args
}

func normalizeVisibilityUintIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func applyTaskFilters(db *gorm.DB, query application.TaskQuery) *gorm.DB {
	if query.HideAdminDeleted {
		db = db.Where("workflow_process_tasks.admin_deleted_at = ?", int64(0))
	}
	db = db.Where(`EXISTS (
		SELECT 1 FROM workflow_process_instances active_instance
		WHERE active_instance.id = workflow_process_tasks.instance_id
		AND active_instance.admin_deleted_at = 0
	)`)
	if value := strings.TrimSpace(query.InstanceID); value != "" {
		db = db.Where("instance_id = ?", value)
	}
	if value := strings.TrimSpace(query.AssigneeID); value != "" {
		db = db.Where("task_assignee_id = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		db = db.Where("task_status = ?", value)
	}
	if hasTaskInstanceFilters(query) {
		instances := db.Session(&gorm.Session{NewDB: true}).
			Model(&workflowmodel.ProcessInstance{}).
			Select("id")
		instances = applyInstanceFilters(instances, application.InstanceQuery{
			DefinitionName:     query.DefinitionName,
			DefinitionCategory: query.DefinitionCategory,
			StarterName:        query.StarterName,
			StartTimeFrom:      query.StartTimeFrom,
			StartTimeTo:        query.StartTimeTo,
		})
		db = db.Where("instance_id IN (?)", instances)
	}
	return db
}

func hasTaskInstanceFilters(query application.TaskQuery) bool {
	return strings.TrimSpace(query.DefinitionName) != "" ||
		strings.TrimSpace(query.DefinitionCategory) != "" ||
		strings.TrimSpace(query.StarterName) != "" ||
		query.StartTimeFrom > 0 || query.StartTimeTo > 0
}

func containsLikePattern(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return "%" + replacer.Replace(strings.TrimSpace(value)) + "%"
}
