package infrastructure

import (
	"context"
	"strings"

	"gorm.io/gorm"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

const workflowOverviewSQL = `
SELECT
	(
		SELECT COUNT(*)
		FROM workflow_process_tasks overview_task
		WHERE overview_task.task_assignee_id = ?
			AND overview_task.task_status = ?
			AND EXISTS (
				SELECT 1
				FROM workflow_process_instances pending_instance
				WHERE pending_instance.id = overview_task.instance_id
					AND pending_instance.admin_deleted_at = 0
			)
	) AS pending,
	(
		SELECT COUNT(*)
		FROM workflow_form_revision_tasks overview_revision_task
		INNER JOIN workflow_form_revision_requests overview_revision
			ON overview_revision.id = overview_revision_task.revision_request_id
		INNER JOIN workflow_process_instances overview_revision_instance
			ON overview_revision_instance.id = overview_revision.source_instance_id
		WHERE overview_revision_task.assignee_id = ?
			AND overview_revision_task.task_status = ?
			AND overview_revision_instance.admin_deleted_at = 0
	) AS revision_pending,
	(
		SELECT COUNT(*)
		FROM (
			SELECT handled_task.instance_id
			FROM workflow_process_tasks handled_task
			INNER JOIN workflow_process_instances handled_instance
				ON handled_instance.id = handled_task.instance_id
			WHERE handled_instance.admin_deleted_at = 0
				AND handled_task.task_status IN (?, ?, ?, ?)
				AND (handled_task.task_assignee_id = ? OR handled_task.handled_by = ?)
			UNION
			SELECT handled_revision.source_instance_id
			FROM workflow_form_revision_tasks handled_revision_task
			INNER JOIN workflow_form_revision_requests handled_revision
				ON handled_revision.id = handled_revision_task.revision_request_id
			INNER JOIN workflow_process_instances handled_revision_instance
				ON handled_revision_instance.id = handled_revision.source_instance_id
			WHERE handled_revision_instance.admin_deleted_at = 0
				AND handled_revision_task.task_status IN (?, ?)
				AND (handled_revision_task.assignee_id = ? OR handled_revision_task.handled_by = ?)
		) handled_instances
	) AS handled,
	(
		SELECT COUNT(*)
		FROM workflow_process_instances started_instance
		WHERE started_instance.admin_deleted_at = 0
			AND started_instance.starter_id = ?
			AND started_instance.starter_deleted_at = 0
	) AS started,
	(
		SELECT COUNT(*)
		FROM workflow_process_instances copied_instance
		WHERE copied_instance.admin_deleted_at = 0
			AND EXISTS (
				SELECT 1
				FROM workflow_instance_participants copied_participant
				WHERE copied_participant.instance_id = copied_instance.id
					AND copied_participant.user_id = ?
					AND copied_participant.participant_role = ?
			)
	) AS copied
`

func workflowOverviewQuery(db *gorm.DB, actorID string) *gorm.DB {
	actorID = strings.TrimSpace(actorID)
	return db.Raw(
		"SELECT (overview.pending + overview.revision_pending) AS pending, overview.handled, overview.started, overview.copied FROM ("+workflowOverviewSQL+") overview",
		actorID,
		workflowmodel.TaskStatusPending,
		actorID,
		workflowmodel.FormRevisionTaskStatusPending,
		workflowmodel.TaskStatusCompleted,
		workflowmodel.TaskStatusApproved,
		workflowmodel.TaskStatusRejected,
		workflowmodel.TaskStatusReturned,
		actorID,
		actorID,
		workflowmodel.FormRevisionTaskStatusApproved,
		workflowmodel.FormRevisionTaskStatusRejected,
		actorID,
		actorID,
		actorID,
		actorID,
		workflowmodel.ParticipantRoleCC,
	)
}

func (store *GormStore) GetWorkflowOverview(ctx context.Context, actorID string) (*application.WorkflowOverview, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	overview := &application.WorkflowOverview{}
	if err := workflowOverviewQuery(db, actorID).Scan(overview).Error; err != nil {
		return nil, err
	}
	return overview, nil
}
