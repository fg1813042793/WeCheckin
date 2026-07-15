package service

import (
	"os"
	"strings"
	"testing"
)

func TestPasswordLoginDoesNotFilterByStoredHashInSQL(t *testing.T) {
	for _, file := range []string{"admin_auth.go", "passport.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, forbidden := range []string{
			"`admin_password` = ?",
			"`user_password` = ?",
			"passwordMD5",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s must verify passwords in service code, not via SQL/hash snippet %q", file, forbidden)
			}
		}
	}
}

func TestPasswordLoginUpgradesLegacyHashes(t *testing.T) {
	for _, file := range []string{"admin_auth.go", "passport.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		if !strings.Contains(text, "passwordutil.NeedsRehash") {
			t.Fatalf("%s must upgrade legacy password hashes after successful login", file)
		}
	}
}
