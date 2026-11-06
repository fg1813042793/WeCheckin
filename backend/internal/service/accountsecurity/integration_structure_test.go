package accountsecurity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordWritePathsUseAccountSecurityService(t *testing.T) {
	serviceRoot := filepath.Join("..")
	checks := []struct {
		path     string
		required []string
	}{
		{filepath.Join(serviceRoot, "admin", "adminmgr", "password.go"), []string{"accountsecurity.ChangeOwnPasswordContext"}},
		{filepath.Join(serviceRoot, "admin", "adminmgr", "service.go"), []string{"accountsecurity.SetInitialPassword", "accountsecurity.ResetTemporaryPassword"}},
		{filepath.Join(serviceRoot, "admin", "adminuser", "service.go"), []string{"accountsecurity.SetInitialPassword", "accountsecurity.ResetTemporaryPassword"}},
		{filepath.Join(serviceRoot, "client", "passport", "register.go"), []string{"accountsecurity.SetInitialPassword"}},
		{filepath.Join(serviceRoot, "dingtalkh5", "account", "service.go"), []string{"accountsecurity.ChangeOwnPasswordContext"}},
		{filepath.Join(serviceRoot, "dingtalkh5", "performance", "user", "service.go"), []string{"accountsecurity.SetInitialPassword", "accountsecurity.ResetTemporaryPassword"}},
		{filepath.Join(serviceRoot, "dingtalkh5", "performance", "template", "service.go"), []string{"accountsecurity.SetInitialPassword"}},
	}
	for _, check := range checks {
		data, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		text := string(data)
		for _, snippet := range check.required {
			if !strings.Contains(text, snippet) {
				t.Errorf("%s missing centralized password call %q", check.path, snippet)
			}
		}
	}
}
