package migrations_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestUserFeedbackSchemaMigrationKeepsColumnsAndIndexesInTheirTables(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905100000_create_user_feedback.sql")
	contracts := map[string][]string{
		"user_feedbacks": {
			"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
			"`feedback_no` VARCHAR(32) NOT NULL",
			"`submitter_id` BIGINT UNSIGNED NOT NULL",
			"`create_request_id` VARCHAR(64) NOT NULL",
			"`feedback_status` VARCHAR(24) NOT NULL DEFAULT 'pending'",
			"`handler_id` BIGINT UNSIGNED NULL COMMENT",
			"`version` BIGINT UNSIGNED NOT NULL DEFAULT 1",
			"`last_activity_at` BIGINT NOT NULL DEFAULT 0",
			"`resolved_at` BIGINT NULL COMMENT",
			"`closed_at` BIGINT NULL COMMENT",
			"`created_at` BIGINT NOT NULL DEFAULT 0",
			"`updated_at` BIGINT NOT NULL DEFAULT 0",
			"PRIMARY KEY (`id`)",
			"UNIQUE KEY `uk_user_feedbacks_feedback_no` (`feedback_no`)",
			"UNIQUE KEY `uk_user_feedbacks_submitter_request` (`submitter_id`,`create_request_id`)",
			"KEY `idx_user_feedbacks_status_activity` (`feedback_status`,`last_activity_at`,`id`)",
			"KEY `idx_user_feedbacks_submitter_activity` (`submitter_id`,`last_activity_at`,`id`)",
			"KEY `idx_user_feedbacks_handler_activity` (`handler_id`,`last_activity_at`,`id`)",
			"KEY `idx_user_feedbacks_last_activity` (`last_activity_at`,`id`)",
		},
		"user_feedback_messages": {
			"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
			"`feedback_id` BIGINT UNSIGNED NOT NULL",
			"`message_type` VARCHAR(24) NOT NULL",
			"`author_type` VARCHAR(16) NOT NULL",
			"`author_id` BIGINT UNSIGNED NOT NULL",
			"`content` TEXT NOT NULL",
			"`from_status` VARCHAR(24) NOT NULL DEFAULT ''",
			"`to_status` VARCHAR(24) NOT NULL DEFAULT ''",
			"`request_id` VARCHAR(64) NOT NULL",
			"UNIQUE KEY `uk_user_feedback_messages_request` (`feedback_id`,`author_type`,`author_id`,`request_id`)",
			"KEY `idx_user_feedback_messages_timeline` (`feedback_id`,`created_at`,`id`)",
		},
		"user_feedback_attachments": {
			"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
			"`feedback_id` BIGINT UNSIGNED NOT NULL",
			"`message_id` BIGINT UNSIGNED NOT NULL",
			"`storage_provider` VARCHAR(32) NOT NULL DEFAULT ''",
			"`object_key` VARCHAR(512) NOT NULL",
			"`original_name` VARCHAR(255) NOT NULL",
			"`content_type` VARCHAR(127) NOT NULL DEFAULT ''",
			"`size_bytes` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`sort_order` INT NOT NULL DEFAULT 0",
			"KEY `idx_user_feedback_attachments_feedback_message` (`feedback_id`,`message_id`,`sort_order`,`id`)",
		},
		"user_feedback_daily_sequences": {
			"`sequence_date` DATE NOT NULL",
			"`current_value` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`updated_at` BIGINT NOT NULL DEFAULT 0",
			"PRIMARY KEY (`sequence_date`)",
		},
	}

	for table, fragments := range contracts {
		t.Run(table, func(t *testing.T) {
			requireFragments(t, extractCreateTableBlock(t, text, table), fragments)
		})
	}
}

func TestUserFeedbackPermissionRecordsAreExact(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	inserts := findInsertStatements(text, "permissions")
	if len(inserts) == 0 {
		t.Fatal("permissions INSERT not found")
	}
	rowPattern := regexp.MustCompile(`(?s)\(\s*'([^']+)'\s*,\s*'[^']*'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*'[^']*'\s*,\s*'([^']*)'`)
	var keys []string
	apiPaths := make(map[string]string)
	for _, insert := range inserts {
		rows := rowPattern.FindAllStringSubmatch(insert, -1)
		if len(rows) == 0 {
			t.Fatal("permissions INSERT contains no recognizable VALUES rows")
		}
		for _, row := range rows {
			keys = append(keys, row[1])
			if row[3] == "api" {
				apiPaths[row[1]] = row[4]
			}
		}
	}
	expectedKeys := []string{
		"admin:menu:user-feedback", "admin:menu:user-feedback:list", "admin:menu:user-feedback:handle",
		"admin:api-category:user-feedback", "admin:api:user-feedback:list", "admin:api:user-feedback:handle",
		"dingtalk_h5:menu:feedback", "dingtalk_h5:api-category:feedback", "dingtalk_h5:api:feedback:list",
		"dingtalk_h5:api:feedback:create", "dingtalk_h5:api:feedback:detail", "dingtalk_h5:api:feedback:supplement",
	}
	assertStringSet(t, "feedback permission keys", keys, expectedKeys)
	expectedPaths := map[string]string{
		"admin:api:user-feedback:list":        "/api/v2/admin/user-feedbacks",
		"admin:api:user-feedback:handle":      "/api/v2/admin/user-feedbacks/:id/status",
		"dingtalk_h5:api:feedback:list":       "/api/v2/dingtalk/h5/user-feedbacks",
		"dingtalk_h5:api:feedback:create":     "/api/v2/dingtalk/h5/user-feedbacks",
		"dingtalk_h5:api:feedback:detail":     "/api/v2/dingtalk/h5/user-feedbacks/:id",
		"dingtalk_h5:api:feedback:supplement": "/api/v2/dingtalk/h5/user-feedbacks/:id/supplements",
	}
	if !reflect.DeepEqual(apiPaths, expectedPaths) {
		t.Fatalf("feedback API resource paths = %#v, want %#v", apiPaths, expectedPaths)
	}
	if regexp.MustCompile(`(?m)^--\s*(GET|POST|PATCH|PUT|DELETE)\s+/api/v2/`).MatchString(text) {
		t.Fatal("permission migration must not duplicate the Go route catalog as method/path comments")
	}
}

func TestUserFeedbackGrantBackfillsAreScopedAndInsertOnly(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	grants := findInsertStatements(text, "permission_grants")
	if len(grants) != 2 {
		t.Fatalf("permission_grants INSERT count = %d, want 2", len(grants))
	}
	for _, grant := range grants {
		if !strings.Contains(grant, "NOT EXISTS") && !strings.Contains(grant, "LEFT JOIN") {
			t.Fatal("grant backfill must insert only missing target grants")
		}
		if strings.Contains(grant, "ON DUPLICATE KEY UPDATE") {
			t.Fatal("grant backfill must not overwrite existing effect, status, scope, or source")
		}
		requireFragments(t, grant, []string{
			"existing_grant.`grant_permission_key` = target_perm.`permission_key`",
		})
	}

	admin := statementContaining(t, grants, "user-feedback-admin-role-backfill")
	requireFragments(t, admin, []string{
		"FROM `roles`", "`role_name` = '超级管理员'", "ORDER BY `id`", "LIMIT 1",
		"existing_grant.`grant_subject_type` = 'role'",
		"existing_grant.`grant_subject_id` = super_admin.`id`",
		"'admin:menu:user-feedback'", "'admin:api:user-feedback:list'", "'admin:api:user-feedback:handle'",
	})
	if strings.Contains(admin, "admin:login") || strings.Contains(admin, "source_grant") {
		t.Fatal("admin feedback grants must be scoped by the built-in super administrator role, not login permission")
	}

	h5 := statementContaining(t, grants, "user-feedback-h5-role-backfill")
	requireFragments(t, h5, []string{
		"source_grant.`grant_subject_type` = 'role'",
		"source_grant.`grant_effect` = 'allow'",
		"source_grant.`grant_status` = 1",
		"existing_grant.`grant_subject_type` = source_grant.`grant_subject_type`",
		"existing_grant.`grant_subject_id` = source_grant.`grant_subject_id`",
		"'dingtalk_h5:menu:dashboard' AS source_key, 'dingtalk_h5:menu:feedback' AS target_key",
		"'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:list'",
		"'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:supplement'",
	})
}

func TestApplyMissingFeedbackGrantsPreservesExistingAndIsRepeatable(t *testing.T) {
	existing := []grantFixture{
		{Subject: "role:1", Permission: "admin:menu:user-feedback", Effect: "deny", Status: 0, Scope: "department:7", Source: "manual"},
		{Subject: "role:1", Permission: "admin:api:user-feedback:list", Effect: "allow", Status: 0, Scope: "self", Source: "legacy"},
		{Subject: "role:2", Permission: "dingtalk_h5:menu:feedback", Effect: "deny", Status: 1, Scope: "team", Source: "custom"},
		{Subject: "role:2", Permission: "dingtalk_h5:api:feedback:list", Effect: "allow", Status: 1, Scope: "all", Source: "seed"},
	}
	missing := []grantFixture{
		{Subject: "role:1", Permission: "admin:api:user-feedback:handle", Effect: "allow", Status: 1, Source: "feedback-admin"},
		{Subject: "role:2", Permission: "dingtalk_h5:api:feedback:create", Effect: "allow", Status: 1, Source: "feedback-h5"},
	}
	candidates := append(append([]grantFixture(nil), existing...), missing...)
	for i := range existing {
		candidates[i].Effect = "allow"
		candidates[i].Status = 1
		candidates[i].Scope = ""
		candidates[i].Source = "feedback-backfill"
	}

	store := grantStore(existing)
	applyMissingGrants(store, candidates)
	expected := grantStore(append(append([]grantFixture(nil), existing...), missing...))
	if !reflect.DeepEqual(store, expected) {
		t.Fatalf("first application = %#v, want %#v", store, expected)
	}
	applyMissingGrants(store, candidates)
	if !reflect.DeepEqual(store, expected) {
		t.Fatalf("repeated application = %#v, want stable %#v", store, expected)
	}
}

type grantFixture struct {
	Subject    string
	Permission string
	Effect     string
	Status     int
	Scope      string
	Source     string
}

func applyMissingGrants(store map[string]grantFixture, candidates []grantFixture) {
	for _, candidate := range candidates {
		key := candidate.Subject + ":" + candidate.Permission
		if _, exists := store[key]; !exists {
			store[key] = candidate
		}
	}
}

func grantStore(grants []grantFixture) map[string]grantFixture {
	store := make(map[string]grantFixture, len(grants))
	applyMissingGrants(store, grants)
	return store
}

func extractCreateTableBlock(t *testing.T, text, table string) string {
	t.Helper()
	pattern := regexp.MustCompile("(?is)CREATE\\s+TABLE\\s+IF\\s+NOT\\s+EXISTS\\s+`" + regexp.QuoteMeta(table) + "`\\s*\\((.*?)\\)\\s*ENGINE=")
	match := pattern.FindStringSubmatch(text)
	if len(match) != 2 {
		t.Fatalf("CREATE TABLE block for %s not found", table)
	}
	return match[1]
}

func findInsertStatements(text, table string) []string {
	pattern := regexp.MustCompile("(?is)INSERT\\s+INTO\\s+`" + regexp.QuoteMeta(table) + "`\\s*\\(.*?;")
	return pattern.FindAllString(text, -1)
}

func statementContaining(t *testing.T, statements []string, marker string) string {
	t.Helper()
	for _, statement := range statements {
		if strings.Contains(statement, marker) {
			return statement
		}
	}
	t.Fatalf("statement containing %q not found", marker)
	return ""
}

func requireFragments(t *testing.T, text string, fragments []string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("SQL block missing %q", fragment)
		}
	}
}

func assertStringSet(t *testing.T, label string, actual, expected []string) {
	t.Helper()
	actual = append([]string(nil), actual...)
	expected = append([]string(nil), expected...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s = %v, want %v", label, actual, expected)
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
