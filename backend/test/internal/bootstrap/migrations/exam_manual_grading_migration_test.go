package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamManualGradingMigrationSchemaAndPermissions(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20261031140000_add_exam_manual_grading.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(source)
	for column, definition := range map[string]string{
		"exam_r_review_snapshot": "MEDIUMTEXT NULL",
		"exam_r_graded_by":       "BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"exam_r_graded_at":       "BIGINT NOT NULL DEFAULT 0",
		"exam_r_grading_comment": "VARCHAR(1000) NOT NULL DEFAULT ''''",
	} {
		for _, snippet := range []string{"COLUMN_NAME = '" + column + "'", "ADD COLUMN `" + column + "` " + definition} {
			if !strings.Contains(sql, snippet) {
				t.Errorf("missing idempotent column definition %q", snippet)
			}
		}
	}
	if strings.Count(sql, "FROM information_schema.COLUMNS") != 4 || strings.Count(sql, "TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record'") != 4 || strings.Count(sql, "DEALLOCATE PREPARE") != 4 {
		t.Fatal("each added column requires its own schema guard and prepared DDL cleanup")
	}
	for _, snippet := range []string{
		"'admin:menu:exam:record-grade', '考试答卷批改', 'admin', 'button'",
		"'admin:menu:exam', '', '', 'exam-record:grade',\n   7, 1",
		"'admin:api:exam-record:grade', '考试答卷批改接口', 'admin', 'api'",
		"'admin:api-category:exam', '/api/v2/admin/exams/:id/records/:recordId/grading'",
		"SELECT 'admin:menu:exam:edit' AS source_key, 'admin:menu:exam:record-grade' AS target_key",
		"UNION ALL SELECT 'admin:api:exam:edit', 'admin:api:exam-record:grade'",
		"source_grant.`grant_effect` = 'allow'", "source_grant.`grant_status` = 1",
		"source_grant.`grant_scope_value`", "NOT EXISTS",
		"existing_grant.`grant_subject_type` = source_grant.`grant_subject_type`",
		"existing_grant.`grant_subject_id` = source_grant.`grant_subject_id`",
		"existing_grant.`grant_permission_key` = target_perm.`permission_key`",
	} {
		if !strings.Contains(sql, snippet) {
			t.Errorf("missing safe grading permission backfill %q", snippet)
		}
	}
	_, grants, ok := strings.Cut(sql, "INSERT INTO `permission_grants`")
	if !ok {
		t.Fatal("grading grant backfill missing")
	}
	for _, forbidden := range []string{"exam:list", "exam:list-page", "ON DUPLICATE KEY UPDATE", "existing_grant.`grant_effect`", "existing_grant.`grant_status`", "UPDATE `permission_grants`", "DELETE FROM"} {
		if strings.Contains(grants, forbidden) {
			t.Errorf("backfill must preserve existing target grants and exclude viewers: %q", forbidden)
		}
	}
}
