package passwordsession

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthenticationMiddlewareEnforcesPasswordRestriction(t *testing.T) {
	root := filepath.Join("..", "..", "middleware")
	for _, file := range []string{
		filepath.Join(root, "admin", "auth.go"),
		filepath.Join(root, "client", "auth.go"),
		filepath.Join(root, "dingtalk_h5", "auth.go"),
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(data)
		for _, snippet := range []string{"EvaluatePasswordStateContext", "passwordsession.Allowed", "passwordsession.ErrorCode", `"errorCode"`} {
			if !strings.Contains(text, snippet) {
				t.Errorf("%s missing restriction enforcement %q", file, snippet)
			}
		}
	}
}
