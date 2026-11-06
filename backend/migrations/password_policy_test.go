package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestPasswordPolicyMigration(t *testing.T) {
	data, err := os.ReadFile("20261105100000_add_password_policy.sql")
	if err != nil {
		t.Fatalf("read password policy migration: %v", err)
	}
	sql := string(data)
	for _, snippet := range []string{
		"user_password_changed_at",
		"user_password_must_change",
		"CREATE TABLE IF NOT EXISTS `user_password_histories`",
		"`user_id` BIGINT UNSIGNED NOT NULL",
		"`password_hash` VARCHAR(255) NOT NULL",
		"`change_source` VARCHAR(32) NOT NULL",
		"KEY `idx_user_password_history_user_time` (`user_id`,`changed_at`)",
	} {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("password policy migration missing %q", snippet)
		}
	}
	if strings.Contains(sql, "UPDATE `users` SET `user_password_must_change` = 1") {
		t.Fatal("migration must not force existing accounts to change passwords")
	}
}
