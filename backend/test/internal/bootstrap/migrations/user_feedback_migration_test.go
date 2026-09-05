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

func TestUserFeedbackSchemaMigrationDefinesEachTableContract(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905100000_create_user_feedback.sql")
	expectedTables := []string{
		"user_feedbacks",
		"user_feedback_messages",
		"user_feedback_attachments",
		"user_feedback_daily_sequences",
	}
	assertStringSet(t, "feedback tables", extractFeedbackTableNames(text), expectedTables)

	contracts := []struct {
		name        string
		definitions []string
		indexes     []string
	}{
		{
			name: "user_feedbacks",
			indexes: []string{
				"PRIMARY KEY (`id`)",
				"UNIQUE KEY `uk_user_feedbacks_feedback_no` (`feedback_no`)",
				"UNIQUE KEY `uk_user_feedbacks_submitter_request` (`submitter_id`,`create_request_id`)",
				"KEY `idx_user_feedbacks_status_activity` (`feedback_status`,`last_activity_at`,`id`)",
				"KEY `idx_user_feedbacks_submitter_activity` (`submitter_id`,`last_activity_at`,`id`)",
				"KEY `idx_user_feedbacks_handler_activity` (`handler_id`,`last_activity_at`,`id`)",
				"KEY `idx_user_feedbacks_last_activity` (`last_activity_at`,`id`)",
			},
			definitions: []string{
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
			},
		},
		{
			name: "user_feedback_messages",
			indexes: []string{
				"PRIMARY KEY (`id`)",
				"UNIQUE KEY `uk_user_feedback_messages_request` (`feedback_id`,`author_type`,`author_id`,`request_id`)",
				"KEY `idx_user_feedback_messages_timeline` (`feedback_id`,`created_at`,`id`)",
			},
			definitions: []string{
				"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
				"`feedback_id` BIGINT UNSIGNED NOT NULL",
				"`message_type` VARCHAR(24) NOT NULL COMMENT 'initial, supplement or status'",
				"`author_type` VARCHAR(16) NOT NULL COMMENT 'user or admin'",
				"`author_id` BIGINT UNSIGNED NOT NULL",
				"`content` TEXT NOT NULL",
				"`from_status` VARCHAR(24) NOT NULL DEFAULT ''",
				"`to_status` VARCHAR(24) NOT NULL DEFAULT ''",
				"`request_id` VARCHAR(64) NOT NULL",
				"`created_at` BIGINT NOT NULL DEFAULT 0",
			},
		},
		{
			name: "user_feedback_attachments",
			indexes: []string{
				"PRIMARY KEY (`id`)",
				"KEY `idx_user_feedback_attachments_feedback_message` (`feedback_id`,`message_id`,`sort_order`,`id`)",
			},
			definitions: []string{
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
			},
		},
		{
			name:    "user_feedback_daily_sequences",
			indexes: []string{"PRIMARY KEY (`sequence_date`)"},
			definitions: []string{
				"`sequence_date` DATE NOT NULL",
				"`current_value` BIGINT UNSIGNED NOT NULL DEFAULT 0",
				"`updated_at` BIGINT NOT NULL DEFAULT 0",
			},
		},
	}

	for _, contract := range contracts {
		contract := contract
		t.Run(contract.name, func(t *testing.T) {
			block := extractCreateTableBlock(t, text, contract.name)
			expectedColumns := extractBacktickNames(strings.Join(contract.definitions, "\n"))
			assertStringSet(t, contract.name+" columns", extractTableColumns(block), expectedColumns)
			assertStringSet(t, contract.name+" indexes", extractTableIndexes(block), contract.indexes)
			for _, detail := range contract.definitions {
				if !strings.Contains(block, detail) {
					t.Fatalf("%s table block missing %q", contract.name, detail)
				}
			}
		})
	}

	attachments := extractCreateTableBlock(t, text, "user_feedback_attachments")
	if strings.Contains(strings.ToLower(attachments), "`url`") {
		t.Fatal("user feedback attachments must store object keys and metadata, not full URLs")
	}
	if !strings.Contains(text, "UTC milliseconds") {
		t.Fatal("user feedback migration must document UTC millisecond time semantics")
	}
}

func TestUserFeedbackPermissionRecordsAreExact(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	definitions := parsePermissionDefinitions(t, text)

	expected := map[string]permissionDefinition{
		"admin:menu:user-feedback": {
			Platform: "admin", Type: "menu", Path: "/user-feedbacks",
			Icon: "ChatDotRound", Perms: "user-feedback:list", Status: 1,
		},
		"admin:menu:user-feedback:list": {
			Platform: "admin", Type: "button",
			Parent: "admin:menu:user-feedback", Perms: "user-feedback:list", Status: 1,
		},
		"admin:menu:user-feedback:handle": {
			Platform: "admin", Type: "button",
			Parent: "admin:menu:user-feedback", Perms: "user-feedback:handle", Status: 1,
		},
		"admin:api-category:user-feedback": {
			Platform: "admin", Type: "api_category", Status: 1,
		},
		"admin:api:user-feedback:list": {
			Platform: "admin", Type: "api",
			Parent: "admin:api-category:user-feedback", Path: "/api/v2/admin/user-feedbacks", Perms: "user-feedback:list", Status: 1,
		},
		"admin:api:user-feedback:handle": {
			Platform: "admin", Type: "api",
			Parent: "admin:api-category:user-feedback", Path: "/api/v2/admin/user-feedbacks/:id/status", Perms: "user-feedback:handle", Status: 1,
		},
		"dingtalk_h5:menu:feedback": {
			Platform: "dingtalk_h5", Type: "menu", Path: "feedback", Icon: "mine", Status: 1,
		},
		"dingtalk_h5:api-category:feedback": {
			Platform: "dingtalk_h5", Type: "api_category", Status: 1,
		},
		"dingtalk_h5:api:feedback:list": {
			Platform: "dingtalk_h5", Type: "api",
			Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks", Perms: "feedback:list", Status: 1,
		},
		"dingtalk_h5:api:feedback:create": {
			Platform: "dingtalk_h5", Type: "api",
			Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks", Perms: "feedback:create", Status: 1,
		},
		"dingtalk_h5:api:feedback:detail": {
			Platform: "dingtalk_h5", Type: "api",
			Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks/:id", Perms: "feedback:detail", Status: 1,
		},
		"dingtalk_h5:api:feedback:supplement": {
			Platform: "dingtalk_h5", Type: "api",
			Parent: "dingtalk_h5:api-category:feedback", Path: "/api/v2/dingtalk/h5/user-feedbacks/:id/supplements", Perms: "feedback:supplement", Status: 1,
		},
	}
	if !reflect.DeepEqual(definitions, expected) {
		t.Fatalf("feedback permission definitions = %#v, want %#v", definitions, expected)
	}
}

func TestUserFeedbackRoleBackfillFiltersAndRepeatExecution(t *testing.T) {
	text := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	rules := []backfillRule{
		parseBackfillRule(t, text, "user-feedback-admin-role-backfill"),
		parseBackfillRule(t, text, "user-feedback-h5-role-backfill"),
	}
	sources := []grantFixture{
		{SubjectType: "role", SubjectID: 1, PermissionKey: "admin:login", Effect: "allow", Status: 1},
		{SubjectType: "role", SubjectID: 2, PermissionKey: "admin:login", Effect: "deny", Status: 1},
		{SubjectType: "role", SubjectID: 3, PermissionKey: "admin:login", Effect: "allow", Status: 0},
		{SubjectType: "user", SubjectID: 4, PermissionKey: "admin:login", Effect: "allow", Status: 1},
		{SubjectType: "user", SubjectID: 10, PermissionKey: "admin:login", Effect: "deny", Status: 1},
		{SubjectType: "user", SubjectID: 11, PermissionKey: "admin:login", Effect: "allow", Status: 0},
		{SubjectType: "role", SubjectID: 5, PermissionKey: "dingtalk_h5:menu:dashboard", Effect: "allow", Status: 1},
		{SubjectType: "role", SubjectID: 9, PermissionKey: "dingtalk_h5:api:bootstrap:view", Effect: "allow", Status: 1},
		{SubjectType: "role", SubjectID: 6, PermissionKey: "dingtalk_h5:menu:dashboard", Effect: "deny", Status: 1},
		{SubjectType: "role", SubjectID: 7, PermissionKey: "dingtalk_h5:api:bootstrap:view", Effect: "allow", Status: 0},
		{SubjectType: "user", SubjectID: 8, PermissionKey: "dingtalk_h5:menu:dashboard", Effect: "allow", Status: 1},
		{SubjectType: "user", SubjectID: 12, PermissionKey: "dingtalk_h5:api:bootstrap:view", Effect: "deny", Status: 1},
		{SubjectType: "user", SubjectID: 13, PermissionKey: "dingtalk_h5:api:bootstrap:view", Effect: "allow", Status: 0},
	}

	store := map[string]grantFixture{
		grantIdentity("role", 1, "admin:menu:user-feedback"): {
			SubjectType: "role", SubjectID: 1, PermissionKey: "admin:menu:user-feedback", Effect: "deny", Status: 0,
		},
		grantIdentity("role", 1, "admin:menu:user-feedback:list"): {
			SubjectType: "role", SubjectID: 1, PermissionKey: "admin:menu:user-feedback:list", Effect: "allow", Status: 0,
		},
		grantIdentity("role", 1, "admin:menu:user-feedback:handle"): {
			SubjectType: "role", SubjectID: 1, PermissionKey: "admin:menu:user-feedback:handle", Effect: "deny", Status: 1,
		},
		grantIdentity("role", 1, "admin:api:user-feedback:list"): {
			SubjectType: "role", SubjectID: 1, PermissionKey: "admin:api:user-feedback:list", Effect: "allow", Status: 1,
		},
		grantIdentity("role", 5, "dingtalk_h5:menu:feedback"): {
			SubjectType: "role", SubjectID: 5, PermissionKey: "dingtalk_h5:menu:feedback", Effect: "allow", Status: 0,
		},
		grantIdentity("role", 9, "dingtalk_h5:api:feedback:detail"): {
			SubjectType: "role", SubjectID: 9, PermissionKey: "dingtalk_h5:api:feedback:detail", Effect: "deny", Status: 0,
		},
		grantIdentity("user", 4, "admin:menu:user-feedback"): {
			SubjectType: "user", SubjectID: 4, PermissionKey: "admin:menu:user-feedback", Effect: "deny", Status: 0,
		},
		grantIdentity("user", 4, "admin:menu:user-feedback:list"): {
			SubjectType: "user", SubjectID: 4, PermissionKey: "admin:menu:user-feedback:list", Effect: "allow", Status: 0,
		},
		grantIdentity("user", 10, "admin:menu:user-feedback:handle"): {
			SubjectType: "user", SubjectID: 10, PermissionKey: "admin:menu:user-feedback:handle", Effect: "deny", Status: 1,
		},
		grantIdentity("user", 11, "admin:api:user-feedback:list"): {
			SubjectType: "user", SubjectID: 11, PermissionKey: "admin:api:user-feedback:list", Effect: "allow", Status: 1,
		},
		grantIdentity("user", 8, "dingtalk_h5:api:feedback:detail"): {
			SubjectType: "user", SubjectID: 8, PermissionKey: "dingtalk_h5:api:feedback:detail", Effect: "deny", Status: 0,
		},
	}
	untouchedKeys := []string{
		grantIdentity("user", 4, "admin:menu:user-feedback"),
		grantIdentity("user", 4, "admin:menu:user-feedback:list"),
		grantIdentity("user", 10, "admin:menu:user-feedback:handle"),
		grantIdentity("user", 11, "admin:api:user-feedback:list"),
		grantIdentity("user", 8, "dingtalk_h5:api:feedback:detail"),
	}
	original := cloneGrantStore(store)
	applyBackfillRules(store, rules, sources)

	backfilledKeys := []string{
		grantIdentity("role", 1, "admin:menu:user-feedback"),
		grantIdentity("role", 1, "admin:menu:user-feedback:list"),
		grantIdentity("role", 1, "admin:menu:user-feedback:handle"),
		grantIdentity("role", 1, "admin:api:user-feedback:list"),
		grantIdentity("role", 1, "admin:api:user-feedback:handle"),
		grantIdentity("role", 5, "dingtalk_h5:menu:feedback"),
		grantIdentity("role", 9, "dingtalk_h5:api:feedback:list"),
		grantIdentity("role", 9, "dingtalk_h5:api:feedback:create"),
		grantIdentity("role", 9, "dingtalk_h5:api:feedback:detail"),
		grantIdentity("role", 9, "dingtalk_h5:api:feedback:supplement"),
	}
	expectedKeys := append(append([]string(nil), backfilledKeys...), untouchedKeys...)
	assertStringSet(t, "first backfill grants", grantStoreKeys(store), expectedKeys)
	for _, identity := range backfilledKeys {
		grant := store[identity]
		if grant.SubjectType != "role" || grant.Effect != "allow" || grant.Status != 1 {
			t.Fatalf("backfilled grant %s = %#v, want active role allow", identity, grant)
		}
	}
	for _, identity := range untouchedKeys {
		if !reflect.DeepEqual(store[identity], original[identity]) {
			t.Fatalf("non-role grant %s changed from %#v to %#v", identity, original[identity], store[identity])
		}
	}

	firstExecution := cloneGrantStore(store)
	applyBackfillRules(store, rules, sources)
	if !reflect.DeepEqual(store, firstExecution) {
		t.Fatalf("repeated backfill changed grant set: first=%#v second=%#v", firstExecution, store)
	}
}

type permissionDefinition struct {
	Platform string
	Type     string
	Parent   string
	Path     string
	Icon     string
	Perms    string
	Status   int
}

type backfillRule struct {
	SourceSubjectType string
	SourceEffect      string
	SourceStatus      int
	TargetsBySource   map[string][]string
	InsertEffect      string
	InsertStatus      int
	UpsertEffect      string
	UpsertStatus      int
}

type grantFixture struct {
	SubjectType   string
	SubjectID     uint
	PermissionKey string
	Effect        string
	Status        int
}

func extractFeedbackTableNames(text string) []string {
	pattern := regexp.MustCompile("(?i)CREATE\\s+TABLE\\s+IF\\s+NOT\\s+EXISTS\\s+`(user_feedback[^`]*)`")
	matches := pattern.FindAllStringSubmatch(text, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	return result
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

func extractTableColumns(block string) []string {
	pattern := regexp.MustCompile("(?m)^\\s*`([^`]+)`\\s+")
	matches := pattern.FindAllStringSubmatch(block, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	return result
}

func extractTableIndexes(block string) []string {
	pattern := regexp.MustCompile("(?m)^\\s*((?:PRIMARY KEY|UNIQUE KEY `[^`]+`|KEY `[^`]+`)\\s*\\([^\\n)]+\\))")
	matches := pattern.FindAllStringSubmatch(block, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	return result
}

func parsePermissionDefinitions(t *testing.T, text string) map[string]permissionDefinition {
	t.Helper()
	insertPattern := regexp.MustCompile("(?im)^[ \\t]*INSERT(?:[ \\t]+IGNORE)?[ \\t]+INTO[ \\t]+`permissions`[ \\t]*")
	headerPattern := regexp.MustCompile("(?is)^[ \\t]*INSERT(?:[ \\t]+IGNORE)?[ \\t]+INTO[ \\t]+`permissions`[ \\t]*\\((.*?)\\)[ \\t\\r\\n]+VALUES\\b")
	upsertPattern := regexp.MustCompile("(?is)\\bON[ \\t\\r\\n]+DUPLICATE[ \\t\\r\\n]+KEY[ \\t\\r\\n]+UPDATE\\b")
	expectedColumns := []string{
		"permission_key", "permission_name", "permission_platform", "permission_type",
		"permission_parent_key", "permission_resource_path", "permission_icon", "permission_perms",
		"permission_sort", "permission_status", "permission_add_time", "permission_edit_time",
		"created_at", "updated_at",
	}
	result := make(map[string]permissionDefinition)
	insertLocations := insertPattern.FindAllStringIndex(text, -1)
	for _, location := range insertLocations {
		statementLength := sqlStatementLength(t, text[location[0]:])
		statement := text[location[0] : location[0]+statementLength]
		header := headerPattern.FindStringSubmatchIndex(statement)
		if len(header) != 4 {
			t.Fatal("permissions INSERT must use an explicit column list and VALUES clause")
		}
		columnText := statement[header[2]:header[3]]
		actualColumns := extractBacktickNames(columnText)
		if !reflect.DeepEqual(actualColumns, expectedColumns) {
			t.Fatalf("permissions INSERT columns = %v, want ordered columns %v", actualColumns, expectedColumns)
		}

		valuesText := statement[header[1] : len(statement)-1]
		upsertLocation := upsertPattern.FindStringIndex(valuesText)
		if upsertLocation == nil {
			t.Fatal("permissions INSERT must use ON DUPLICATE KEY UPDATE")
		}
		upsertText := valuesText[upsertLocation[0]:]
		if !regexp.MustCompile("(?i)`permission_status`\\s*=\\s*1").MatchString(upsertText) {
			t.Fatal("permissions upsert must reactivate existing records")
		}
		valuesText = valuesText[:upsertLocation[0]]
		tuples := parseSQLTuples(t, valuesText)
		if len(tuples) == 0 {
			t.Fatal("permissions INSERT contains no VALUES tuples")
		}
		for _, tuple := range tuples {
			fields := splitTopLevelCSV(t, tuple)
			if len(fields) != len(expectedColumns) {
				t.Fatalf("permission tuple has %d fields, want %d: %s", len(fields), len(expectedColumns), tuple)
			}
			key := parseSQLString(t, fields[0])
			definition := permissionDefinition{
				Platform: parseSQLString(t, fields[2]),
				Type:     parseSQLString(t, fields[3]),
				Parent:   parseSQLString(t, fields[4]),
				Path:     parseSQLString(t, fields[5]),
				Icon:     parseSQLString(t, fields[6]),
				Perms:    parseSQLString(t, fields[7]),
				Status:   parseSQLInt(t, fields[9]),
			}
			if _, exists := result[key]; exists {
				t.Fatalf("duplicate permission definition %s", key)
			}
			result[key] = definition
		}
	}
	if len(insertLocations) == 0 {
		t.Fatal("permissions INSERT not found")
	}
	return result
}

func sqlStatementLength(t *testing.T, text string) int {
	t.Helper()
	inString := false
	for i := 0; i < len(text); i++ {
		if text[i] == '\'' {
			if inString && i+1 < len(text) && text[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if text[i] == ';' && !inString {
			return i + 1
		}
	}
	t.Fatal("SQL INSERT statement has no terminator")
	return 0
}

func extractBacktickNames(text string) []string {
	pattern := regexp.MustCompile("`([^`]+)`")
	matches := pattern.FindAllStringSubmatch(text, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	return result
}

func parseBackfillRule(t *testing.T, text, source string) backfillRule {
	t.Helper()
	block := extractBackfillBlock(t, text, source)
	expectedColumns := []string{
		"grant_subject_type", "grant_subject_id", "grant_permission_key", "grant_permission_id",
		"grant_effect", "grant_scope_value", "grant_source", "grant_status",
		"grant_add_time", "grant_edit_time", "created_at", "updated_at",
	}
	headerPattern := regexp.MustCompile("(?is)^INSERT\\s+INTO\\s+`permission_grants`\\s*\\((.*?)\\)\\s*SELECT\\s+DISTINCT\\b")
	header := headerPattern.FindStringSubmatch(block)
	if len(header) != 2 {
		t.Fatalf("backfill %s must use INSERT columns followed by SELECT DISTINCT", source)
	}
	if columns := extractBacktickNames(header[1]); !reflect.DeepEqual(columns, expectedColumns) {
		t.Fatalf("backfill %s columns = %v, want %v", source, columns, expectedColumns)
	}
	projectionPattern := regexp.MustCompile("(?is)SELECT\\s+DISTINCT\\s+source_grant\\.`grant_subject_type`\\s*,\\s*source_grant\\.`grant_subject_id`\\s*,\\s*target_perm\\.`permission_key`\\s*,\\s*target_perm\\.`id`\\s*,")
	if !projectionPattern.MatchString(block) {
		t.Fatalf("backfill %s must project target identity from the source grant and target permission", source)
	}
	if !strings.Contains(block, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("backfill %s must use ON DUPLICATE KEY UPDATE", source)
	}
	rule := backfillRule{
		SourceSubjectType: captureString(t, block, "source subject type", "source_grant\\.`grant_subject_type`\\s*=\\s*'([^']+)'"),
		SourceEffect:      captureString(t, block, "source effect", "source_grant\\.`grant_effect`\\s*=\\s*'([^']+)'"),
		SourceStatus:      captureInt(t, block, "source status", "source_grant\\.`grant_status`\\s*=\\s*([0-9]+)"),
		TargetsBySource:   make(map[string][]string),
		InsertEffect:      captureString(t, block, "insert effect", "target_perm\\.`id`,\\s*'([^']+)',\\s*'',\\s*'"+regexp.QuoteMeta(source)+"',\\s*[0-9]+"),
		InsertStatus:      captureInt(t, block, "insert status", "target_perm\\.`id`,\\s*'[^']+',\\s*'',\\s*'"+regexp.QuoteMeta(source)+"',\\s*([0-9]+)"),
		UpsertEffect:      captureString(t, block, "upsert effect", "`grant_effect`\\s*=\\s*'([^']+)'"),
		UpsertStatus:      captureInt(t, block, "upsert status", "`grant_status`\\s*=\\s*([0-9]+)"),
	}

	sourceKeys := parseBackfillSourceKeys(t, block)
	mappingPattern := regexp.MustCompile("(?m)(?:SELECT|UNION ALL SELECT)\\s+'([^']+)'(?:\\s+AS source_key)?,\\s*'([^']+)'(?:\\s+AS target_key)?")
	mappings := mappingPattern.FindAllStringSubmatch(block, -1)
	if len(mappings) > 0 {
		for _, mapping := range mappings {
			rule.TargetsBySource[mapping[1]] = append(rule.TargetsBySource[mapping[1]], mapping[2])
		}
	} else {
		targetBlockPattern := regexp.MustCompile("(?s)target_perm\\.`permission_key`\\s+IN\\s*\\((.*?)\\)\\s*WHERE")
		match := targetBlockPattern.FindStringSubmatch(block)
		if len(match) != 2 {
			t.Fatal("admin backfill target permission list not found")
		}
		targets := extractQuotedStrings(match[1])
		for _, sourceKey := range sourceKeys {
			rule.TargetsBySource[sourceKey] = append([]string(nil), targets...)
		}
	}
	assertStringSet(t, source+" mapping sources", mapKeys(rule.TargetsBySource), sourceKeys)
	return rule
}

func extractBackfillBlock(t *testing.T, text, source string) string {
	t.Helper()
	marker := "'" + source + "'"
	markerIndex := strings.Index(text, marker)
	if markerIndex < 0 {
		t.Fatalf("backfill source %s not found", source)
	}
	start := strings.LastIndex(text[:markerIndex], "INSERT INTO `permission_grants`")
	if start < 0 {
		t.Fatalf("permission_grants INSERT for %s not found", source)
	}
	endOffset := strings.Index(text[markerIndex:], ";")
	if endOffset < 0 {
		t.Fatalf("permission_grants INSERT for %s has no terminator", source)
	}
	return text[start : markerIndex+endOffset+1]
}

func parseBackfillSourceKeys(t *testing.T, block string) []string {
	t.Helper()
	equalsPattern := regexp.MustCompile("source_grant\\.`grant_permission_key`\\s*=\\s*'([^']+)'")
	if match := equalsPattern.FindStringSubmatch(block); len(match) == 2 {
		return []string{match[1]}
	}
	inPattern := regexp.MustCompile("(?s)source_grant\\.`grant_permission_key`\\s+IN\\s*\\((.*?)\\)")
	match := inPattern.FindStringSubmatch(block)
	if len(match) != 2 {
		t.Fatal("backfill source permission filter not found")
	}
	return extractQuotedStrings(match[1])
}

func applyBackfillRules(store map[string]grantFixture, rules []backfillRule, sources []grantFixture) {
	for _, rule := range rules {
		for _, source := range sources {
			if source.SubjectType != rule.SourceSubjectType || source.Effect != rule.SourceEffect || source.Status != rule.SourceStatus {
				continue
			}
			for _, target := range rule.TargetsBySource[source.PermissionKey] {
				identity := grantIdentity(source.SubjectType, source.SubjectID, target)
				if existing, found := store[identity]; found {
					existing.Effect = rule.UpsertEffect
					existing.Status = rule.UpsertStatus
					store[identity] = existing
					continue
				}
				store[identity] = grantFixture{
					SubjectType: source.SubjectType, SubjectID: source.SubjectID, PermissionKey: target,
					Effect: rule.InsertEffect, Status: rule.InsertStatus,
				}
			}
		}
	}
}

func parseSQLTuples(t *testing.T, text string) []string {
	t.Helper()
	var tuples []string
	depth, start, lastEnd := 0, -1, 0
	inString := false
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\'':
			if inString && i+1 < len(text) && text[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
		case '(':
			if !inString {
				if depth == 0 {
					separator := strings.TrimSpace(text[lastEnd:i])
					if len(tuples) == 0 && separator != "" {
						t.Fatalf("unexpected text before first SQL tuple: %q", separator)
					}
					if len(tuples) > 0 && separator != "," {
						t.Fatalf("SQL tuples must be comma-separated, found %q", separator)
					}
					start = i + 1
				}
				depth++
			}
		case ')':
			if !inString {
				if depth == 0 {
					t.Fatal("unexpected closing parenthesis in SQL tuples")
				}
				depth--
				if depth == 0 && start >= 0 {
					tuples = append(tuples, text[start:i])
					start = -1
					lastEnd = i + 1
				}
			}
		}
	}
	if inString || depth != 0 {
		t.Fatal("unbalanced permission VALUES tuples")
	}
	if trailing := strings.TrimSpace(text[lastEnd:]); trailing != "" {
		t.Fatalf("unexpected text after SQL tuples: %q", trailing)
	}
	return tuples
}

func splitTopLevelCSV(t *testing.T, text string) []string {
	t.Helper()
	var result []string
	depth, start := 0, 0
	inString := false
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\'':
			if inString && i+1 < len(text) && text[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
		case '(':
			if !inString {
				depth++
			}
		case ')':
			if !inString {
				depth--
			}
		case ',':
			if !inString && depth == 0 {
				result = append(result, strings.TrimSpace(text[start:i]))
				start = i + 1
			}
		}
	}
	if inString || depth != 0 {
		t.Fatal("unbalanced SQL tuple fields")
	}
	return append(result, strings.TrimSpace(text[start:]))
}

func parseSQLString(t *testing.T, value string) string {
	t.Helper()
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '\'' || value[len(value)-1] != '\'' {
		t.Fatalf("SQL value %q is not a string literal", value)
	}
	return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
}

func parseSQLInt(t *testing.T, value string) int {
	t.Helper()
	result, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		t.Fatalf("SQL value %q is not an integer literal: %v", value, err)
	}
	return result
}

func captureString(t *testing.T, text, label, expression string) string {
	t.Helper()
	match := regexp.MustCompile(expression).FindStringSubmatch(text)
	if len(match) != 2 {
		t.Fatalf("%s not found", label)
	}
	return match[1]
}

func captureInt(t *testing.T, text, label, expression string) int {
	t.Helper()
	value := captureString(t, text, label, expression)
	result, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("parse %s %q: %v", label, value, err)
	}
	return result
}

func extractQuotedStrings(text string) []string {
	pattern := regexp.MustCompile("'([^']+)'")
	matches := pattern.FindAllStringSubmatch(text, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	return result
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

func mapKeys(values map[string][]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}

func grantIdentity(subjectType string, subjectID uint, permissionKey string) string {
	return subjectType + ":" + strconv.FormatUint(uint64(subjectID), 10) + ":" + permissionKey
}

func grantStoreKeys(store map[string]grantFixture) []string {
	result := make([]string, 0, len(store))
	for key := range store {
		result = append(result, key)
	}
	return result
}

func cloneGrantStore(store map[string]grantFixture) map[string]grantFixture {
	result := make(map[string]grantFixture, len(store))
	for key, grant := range store {
		result[key] = grant
	}
	return result
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
