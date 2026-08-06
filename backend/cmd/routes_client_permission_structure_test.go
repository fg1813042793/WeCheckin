package main

import (
	"os"
	"strings"
	"testing"
)

func TestLegacyClientAuthenticatedRoutesUseClientAPIPermissionMiddleware(t *testing.T) {
	src, err := os.ReadFile("routes_client.go")
	if err != nil {
		t.Fatalf("read routes_client.go: %v", err)
	}
	text := string(src)
	required := []string{
		`h.Group("/passport", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/fav", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/news", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/enroll", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/event", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/survey", clientmw.ClientAuth(), clientmw.ClientPerm())`,
		`h.Group("/exam", clientmw.ClientAuth(), clientmw.ClientPerm())`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("legacy authenticated client routes must use API permission middleware with %s", want)
		}
	}
}
