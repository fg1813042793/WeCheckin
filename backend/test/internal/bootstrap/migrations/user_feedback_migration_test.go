package migrations_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
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
			"`created_at` BIGINT NOT NULL DEFAULT 0",
			"PRIMARY KEY (`id`)",
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
			"`created_at` BIGINT NOT NULL DEFAULT 0",
			"PRIMARY KEY (`id`)",
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
	expected := map[string]permissionRecord{
		"admin:menu:user-feedback":            {Platform: "admin", Type: "menu", Path: "/user-feedbacks", Perms: "user-feedback:list", Status: 1},
		"admin:menu:user-feedback:list":       {Platform: "admin", Type: "button", Parent: "admin:menu:user-feedback", Perms: "user-feedback:list", Status: 1},
		"admin:menu:user-feedback:handle":     {Platform: "admin", Type: "button", Parent: "admin:menu:user-feedback", Perms: "user-feedback:handle", Status: 1},
		"admin:api-category:user-feedback":    {Platform: "admin", Type: "api_category", Status: 1},
		"admin:api:user-feedback:list":        {Platform: "admin", Type: "api", Parent: "admin:api-category:user-feedback", Path: "/api/v2/admin/user-feedbacks", Perms: "user-feedback:list", Status: 1},
		"admin:api:user-feedback:handle":      {Platform: "admin", Type: "api", Parent: "admin:api-category:user-feedback", Path: "/api/v2/admin/user-feedbacks/:id/status", Perms: "user-feedback:handle", Status: 1},
		"dingtalk_h5:menu:feedback":           {Platform: "dingtalk_h5", Type: "menu", Path: "feedback", Status: 1},
		"dingtalk_h5:api-category:feedback":   {Platform: "dingtalk_h5", Type: "api_category", Status: 1},
		"dingtalk_h5:api:feedback:list":       {Platform: "dingtalk_h5", Type: "api", Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks", Perms: "feedback:list", Status: 1},
		"dingtalk_h5:api:feedback:create":     {Platform: "dingtalk_h5", Type: "api", Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks", Perms: "feedback:create", Status: 1},
		"dingtalk_h5:api:feedback:detail":     {Platform: "dingtalk_h5", Type: "api", Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks/:id", Perms: "feedback:detail", Status: 1},
		"dingtalk_h5:api:feedback:supplement": {Platform: "dingtalk_h5", Type: "api", Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks/:id/supplements", Perms: "feedback:supplement", Status: 1},
	}
	if actual := extractPermissionRecords(t, text); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("feedback permission records = %#v, want %#v", actual, expected)
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
	})
	if strings.Contains(admin, "admin:login") || strings.Contains(admin, "source_grant") {
		t.Fatal("admin feedback grants must be scoped by the built-in super administrator role, not login permission")
	}
	assertStringSet(t, "Admin feedback grant targets", extractINValues(t, admin, "target_perm.`permission_key`"), []string{
		"admin:menu:user-feedback",
		"admin:menu:user-feedback:list",
		"admin:menu:user-feedback:handle",
		"admin:api:user-feedback:list",
		"admin:api:user-feedback:handle",
	})

	h5 := statementContaining(t, grants, "user-feedback-h5-role-backfill")
	requireFragments(t, h5, []string{
		"source_grant.`grant_subject_type` = 'role'",
		"source_grant.`grant_effect` = 'allow'",
		"source_grant.`grant_status` = 1",
		"existing_grant.`grant_subject_type` = source_grant.`grant_subject_type`",
		"existing_grant.`grant_subject_id` = source_grant.`grant_subject_id`",
	})
	assertStringSet(t, "H5 feedback grant mappings", extractH5Mappings(h5), []string{
		"dingtalk_h5:menu:dashboard->dingtalk_h5:menu:feedback",
		"dingtalk_h5:api:bootstrap:view->dingtalk_h5:api:feedback:list",
		"dingtalk_h5:api:bootstrap:view->dingtalk_h5:api:feedback:create",
		"dingtalk_h5:api:bootstrap:view->dingtalk_h5:api:feedback:detail",
		"dingtalk_h5:api:bootstrap:view->dingtalk_h5:api:feedback:supplement",
	})
}

func TestUserFeedbackSourceFilterSpecification(t *testing.T) {
	targets := []string{"feedback:menu", "feedback:api"}
	tests := []struct {
		name   string
		source sourceGrantFixture
		want   []string
	}{
		{name: "active role allow", source: sourceGrantFixture{SubjectType: "role", Effect: "allow", Status: 1}, want: targets},
		{name: "ordinary user", source: sourceGrantFixture{SubjectType: "user", Effect: "allow", Status: 1}},
		{name: "role deny", source: sourceGrantFixture{SubjectType: "role", Effect: "deny", Status: 1}},
		{name: "disabled role allow", source: sourceGrantFixture{SubjectType: "role", Effect: "allow", Status: 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := specificationTargetsForSource(test.source, targets); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("targets = %v, want %v", got, test.want)
			}
		})
	}
}

func TestUserFeedbackInsertOnlyGrantSpecificationPreservesExistingAndIsRepeatable(t *testing.T) {
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
	applyInsertOnlyGrantSpecification(store, candidates)
	expected := grantStore(append(append([]grantFixture(nil), existing...), missing...))
	if !reflect.DeepEqual(store, expected) {
		t.Fatalf("first application = %#v, want %#v", store, expected)
	}
	applyInsertOnlyGrantSpecification(store, candidates)
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

type sourceGrantFixture struct {
	SubjectType string
	Effect      string
	Status      int
}

type permissionRecord struct {
	Platform string
	Type     string
	Parent   string
	Path     string
	Perms    string
	Status   int
}

// specificationTargetsForSource models source-filter examples only; it does not execute SQL.
func specificationTargetsForSource(source sourceGrantFixture, targets []string) []string {
	if source.SubjectType != "role" || source.Effect != "allow" || source.Status != 1 {
		return nil
	}
	return append([]string(nil), targets...)
}

// applyInsertOnlyGrantSpecification models required set semantics only; it does not execute SQL.
func applyInsertOnlyGrantSpecification(store map[string]grantFixture, candidates []grantFixture) {
	for _, candidate := range candidates {
		key := candidate.Subject + ":" + candidate.Permission
		if _, exists := store[key]; !exists {
			store[key] = candidate
		}
	}
}

func grantStore(grants []grantFixture) map[string]grantFixture {
	store := make(map[string]grantFixture, len(grants))
	applyInsertOnlyGrantSpecification(store, grants)
	return store
}

func extractPermissionRecords(t *testing.T, text string) map[string]permissionRecord {
	t.Helper()
	rowPattern := regexp.MustCompile(`(?s)\(\s*'([^']+)'\s*,\s*'[^']*'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*'[^']*'\s*,\s*'([^']*)'\s*,\s*[0-9]+\s*,\s*([0-9]+)\s*,`)
	records := make(map[string]permissionRecord)
	for _, insert := range findInsertStatements(text, "permissions") {
		rows := rowPattern.FindAllStringSubmatch(insert, -1)
		if len(rows) == 0 {
			t.Fatal("permissions INSERT contains no recognizable VALUES rows")
		}
		for _, row := range rows {
			status, err := strconv.Atoi(row[7])
			if err != nil {
				t.Fatalf("parse permission status %q: %v", row[7], err)
			}
			if _, exists := records[row[1]]; exists {
				t.Fatalf("duplicate permission record %s", row[1])
			}
			records[row[1]] = permissionRecord{
				Platform: row[2], Type: row[3], Parent: row[4], Path: row[5], Perms: row[6], Status: status,
			}
		}
	}
	if len(records) == 0 {
		t.Fatal("permissions INSERT not found")
	}
	return records
}

func extractINValues(t *testing.T, statement, leftHandSide string) []string {
	t.Helper()
	pattern := regexp.MustCompile("(?s)" + regexp.QuoteMeta(leftHandSide) + "\\s+IN\\s*\\((.*?)\\)")
	match := pattern.FindStringSubmatch(statement)
	if len(match) != 2 {
		t.Fatalf("IN list for %s not found", leftHandSide)
	}
	quoted := regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(match[1], -1)
	values := make([]string, 0, len(quoted))
	for _, value := range quoted {
		values = append(values, value[1])
	}
	return values
}

func extractH5Mappings(statement string) []string {
	pattern := regexp.MustCompile(`(?m)(?:SELECT|UNION ALL SELECT)\s+'([^']+)'(?:\s+AS source_key)?,\s*'([^']+)'(?:\s+AS target_key)?`)
	matches := pattern.FindAllStringSubmatch(statement, -1)
	mappings := make([]string, 0, len(matches))
	for _, match := range matches {
		mappings = append(mappings, match[1]+"->"+match[2])
	}
	return mappings
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
