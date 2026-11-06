package accountsecurity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoginServicesExposeAndPersistPasswordState(t *testing.T) {
	checks := []struct {
		path     string
		required []string
	}{
		{filepath.Join("..", "admin", "adminauth", "service.go"), []string{"PasswordChangeRequired", "PasswordExpired", "PasswordExpiresAt", "PasswordExpiryWarning", `"passwordChangedAt"`, `"passwordMustChange"`}},
		{filepath.Join("..", "client", "passport", "login.go"), []string{"PasswordChangeRequired", "PasswordExpired", "PasswordExpiresAt", "PasswordExpiryWarning", `"passwordChangedAt"`, `"passwordMustChange"`}},
		{filepath.Join("..", "dingtalkh5", "auth", "service.go"), []string{"PasswordChangeRequired", "PasswordExpired", "PasswordExpiresAt", "PasswordExpiryWarning", "EvaluatePasswordState"}},
	}
	for _, check := range checks {
		data, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		text := string(data)
		for _, snippet := range check.required {
			if !strings.Contains(text, snippet) {
				t.Errorf("%s missing password state contract %q", check.path, snippet)
			}
		}
	}
}
