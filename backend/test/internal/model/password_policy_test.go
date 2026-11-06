package model_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordPolicyModelsShareUserCredentialColumns(t *testing.T) {
	for _, file := range []string{filepath.Join("account", "user.go"), filepath.Join("admin", "admin.go")} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(data)
		for _, snippet := range []string{
			"PasswordChangedAt",
			"column:user_password_changed_at",
			"PasswordMustChange",
			"column:user_password_must_change",
		} {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s missing %q", file, snippet)
			}
		}
	}

	history, err := os.ReadFile(filepath.Join("account", "password_history.go"))
	if err != nil {
		t.Fatalf("read password history model: %v", err)
	}
	for _, snippet := range []string{"type PasswordHistory struct", "user_password_histories", "PasswordHash", "ChangeSource"} {
		if !strings.Contains(string(history), snippet) {
			t.Fatalf("password history model missing %q", snippet)
		}
	}
}
