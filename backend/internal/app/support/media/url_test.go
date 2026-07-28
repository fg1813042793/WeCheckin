package media

import (
	"os"
	"strings"
	"testing"
)

func TestFullURLKeepsAbsoluteURL(t *testing.T) {
	if got := FullURL("https://example.com/a.png", "https://cdn.local"); got != "https://example.com/a.png" {
		t.Fatalf("FullURL absolute = %q", got)
	}
}

func TestFullURLJoinsDomainAndRelativePath(t *testing.T) {
	got := FullURL("/uploads/a.png", "https://cdn.local")
	want := "https://cdn.local/uploads/a.png"
	if got != want {
		t.Fatalf("FullURL relative = %q, want %q", got, want)
	}
}

func TestFullURLHandlesEmptyPath(t *testing.T) {
	if got := FullURL("", "https://cdn.local"); got != "" {
		t.Fatalf("empty path should stay empty, got %q", got)
	}
}

func TestStaticDomainUsesDefaultWhenDatabaseUnavailable(t *testing.T) {
	if got := StaticDomain(); got == "" {
		t.Fatalf("StaticDomain should return a fallback when database is unavailable")
	}
}

func TestStaticDomainUsesQueryContext(t *testing.T) {
	src, err := os.ReadFile("static.go")
	if err != nil {
		t.Fatalf("read static.go: %v", err)
	}
	if strings.Contains(string(src), "database.DB.") {
		t.Fatalf("StaticDomain must use database.WithContext instead of direct database.DB calls")
	}
}
