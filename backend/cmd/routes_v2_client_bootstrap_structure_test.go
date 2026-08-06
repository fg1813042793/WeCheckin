package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2ClientRoutesExposePermissionBootstrap(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`client.GET("/me/bootstrap", pp.Bootstrap)`,
		`client := h.Group("/api/v2", clientmw.ClientAuth(), clientmw.ClientPerm())`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("client v2 routes must expose permission-controlled bootstrap with %q", snippet)
		}
	}
}
