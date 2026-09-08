package handler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthenticatedClientHandlersUseSessionIdentity(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "enroll", "handler.go"),
		filepath.Join("client", "event", "handler.go"),
		filepath.Join("client", "event", "dynamics.go"),
		filepath.Join("client", "favorite", "handler.go"),
		filepath.Join("client", "news", "handler.go"),
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(source)
		for _, forbidden := range []string{
			`c.Query("user_id")`,
			`c.PostForm("user_id")`,
			`c.PostForm("token")`,
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s must use the authenticated session identity instead of %s", file, forbidden)
			}
		}
	}
}
