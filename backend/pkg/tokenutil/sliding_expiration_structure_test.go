package tokenutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllAuthenticationPathsHonorSlidingExpirationSwitch(t *testing.T) {
	files := map[string]string{
		filepath.Join("..", "..", "internal", "middleware", "admin", "auth.go"):        `tokenutil.IsTokenSlidingExpirationEnabledContext(ctx, "admin")`,
		filepath.Join("..", "..", "internal", "middleware", "client", "auth.go"):       `tokenutil.IsTokenSlidingExpirationEnabledContext(ctx, "user")`,
		filepath.Join("..", "..", "internal", "service", "admin", "online", "user.go"): `tokenutil.IsTokenSlidingExpirationEnabledContext(ctx, role)`,
	}
	for path, expected := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(data)
		if !strings.Contains(source, expected) {
			t.Fatalf("%s must guard token TTL refresh with the global sliding expiration switch", path)
		}
	}
}
