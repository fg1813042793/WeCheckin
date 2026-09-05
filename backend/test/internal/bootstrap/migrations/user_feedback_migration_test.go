package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserFeedbackSchemaMigrationCreatesTablesAndIndexes(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905100000_create_user_feedback.sql")

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `user_feedbacks`",
		"CREATE TABLE IF NOT EXISTS `user_feedback_messages`",
		"CREATE TABLE IF NOT EXISTS `user_feedback_attachments`",
		"CREATE TABLE IF NOT EXISTS `user_feedback_daily_sequences`",
		"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		"`feedback_no` VARCHAR(32) NOT NULL",
		"`submitter_id` BIGINT UNSIGNED NOT NULL",
		"`create_request_id` VARCHAR(64) NOT NULL",
		"`feedback_status` VARCHAR(24) NOT NULL DEFAULT 'pending'",
		"`handler_id` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"`version` BIGINT UNSIGNED NOT NULL DEFAULT 1",
		"`last_activity_at` BIGINT NOT NULL DEFAULT 0",
		"`resolved_at` BIGINT NOT NULL DEFAULT 0",
		"`closed_at` BIGINT NOT NULL DEFAULT 0",
		"`created_at` BIGINT NOT NULL DEFAULT 0",
		"`updated_at` BIGINT NOT NULL DEFAULT 0",
		"UNIQUE KEY `uk_user_feedbacks_feedback_no` (`feedback_no`)",
		"UNIQUE KEY `uk_user_feedbacks_submitter_request` (`submitter_id`,`create_request_id`)",
		"KEY `idx_user_feedbacks_status_activity` (`feedback_status`,`last_activity_at`,`id`)",
		"KEY `idx_user_feedbacks_submitter_activity` (`submitter_id`,`last_activity_at`,`id`)",
		"KEY `idx_user_feedbacks_handler_activity` (`handler_id`,`last_activity_at`,`id`)",
		"KEY `idx_user_feedbacks_last_activity` (`last_activity_at`,`id`)",
		"`message_type` VARCHAR(24) NOT NULL",
		"`author_type` VARCHAR(16) NOT NULL",
		"`content` TEXT NOT NULL",
		"`from_status` VARCHAR(24) NOT NULL DEFAULT ''",
		"`to_status` VARCHAR(24) NOT NULL DEFAULT ''",
		"`request_id` VARCHAR(64) NOT NULL",
		"UNIQUE KEY `uk_user_feedback_messages_request` (`feedback_id`,`author_type`,`author_id`,`request_id`)",
		"KEY `idx_user_feedback_messages_timeline` (`feedback_id`,`created_at`,`id`)",
		"`storage_provider` VARCHAR(32) NOT NULL DEFAULT ''",
		"`object_key` VARCHAR(512) NOT NULL",
		"`original_name` VARCHAR(255) NOT NULL",
		"`content_type` VARCHAR(127) NOT NULL DEFAULT ''",
		"`size_bytes` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"`sort_order` INT NOT NULL DEFAULT 0",
		"KEY `idx_user_feedback_attachments_feedback_message` (`feedback_id`,`message_id`,`sort_order`,`id`)",
		"`sequence_date` DATE NOT NULL",
		"`current_value` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"PRIMARY KEY (`sequence_date`)",
		"UTC milliseconds",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("user feedback schema migration missing %q", required)
		}
	}

	if strings.Contains(strings.ToLower(text), "`url`") {
		t.Fatal("user feedback attachments must store object keys and metadata, not full URLs")
	}
}

func TestUserFeedbackPermissionMigrationRegistersRoutesAndBackfillsRoles(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")

	for _, required := range []string{
		"'admin:menu:user-feedback'",
		"'/user-feedbacks'",
		"'ChatDotRound'",
		"'admin:menu:user-feedback:list'",
		"'admin:menu:user-feedback:handle'",
		"'user-feedback:list'",
		"'user-feedback:handle'",
		"'admin:api-category:user-feedback'",
		"'admin:api:user-feedback:list'",
		"'admin:api:user-feedback:handle'",
		"GET /api/v2/admin/user-feedbacks/overview -> admin:api:user-feedback:list (user-feedback:list)",
		"GET /api/v2/admin/user-feedbacks -> admin:api:user-feedback:list (user-feedback:list)",
		"GET /api/v2/admin/user-feedbacks/:id -> admin:api:user-feedback:list (user-feedback:list)",
		"PATCH /api/v2/admin/user-feedbacks/:id/status -> admin:api:user-feedback:handle (user-feedback:handle)",
		"'dingtalk_h5:menu:feedback'",
		"'dingtalk_h5:api-category:feedback'",
		"'dingtalk_h5:api:feedback:list'",
		"'dingtalk_h5:api:feedback:create'",
		"'dingtalk_h5:api:feedback:detail'",
		"'dingtalk_h5:api:feedback:supplement'",
		"GET /api/v2/dingtalk/h5/user-feedbacks -> dingtalk_h5:api:feedback:list",
		"POST /api/v2/dingtalk/h5/user-feedbacks -> dingtalk_h5:api:feedback:create",
		"GET /api/v2/dingtalk/h5/user-feedbacks/:id -> dingtalk_h5:api:feedback:detail",
		"POST /api/v2/dingtalk/h5/user-feedbacks/:id/supplements -> dingtalk_h5:api:feedback:supplement",
		"FROM `permission_grants` source_grant",
		"source_grant.`grant_subject_type` = 'role'",
		"source_grant.`grant_permission_key` = 'admin:login'",
		"source_grant.`grant_permission_key` IN (\n    'dingtalk_h5:menu:dashboard',\n    'dingtalk_h5:api:bootstrap:view'\n  )",
		"user-feedback-admin-role-backfill",
		"user-feedback-h5-role-backfill",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("user feedback permission migration missing %q", required)
		}
	}

	for _, forbidden := range []string{
		"admin:api:user-feedback:overview",
		"dingtalk_h5:api:feedback:overview",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("overview must reuse the list API permission record, found %q", forbidden)
		}
	}

	if count := strings.Count(text, "ON DUPLICATE KEY UPDATE"); count < 3 {
		t.Fatalf("user feedback permission migration must be repeatable, ON DUPLICATE KEY UPDATE count = %d", count)
	}
}

func readUserFeedbackMigration(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read user feedback migration %s: %v", name, err)
	}
	return string(source)
}
