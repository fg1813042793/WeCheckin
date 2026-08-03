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
		`h.Group("/passport", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/fav", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/news", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/enroll", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/event", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/survey", middleware.ClientAuth(), middleware.ClientPerm())`,
		`h.Group("/exam", middleware.ClientAuth(), middleware.ClientPerm())`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("legacy authenticated client routes must use API permission middleware with %s", want)
		}
	}
}
