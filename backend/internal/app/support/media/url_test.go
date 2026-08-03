package media

import (
	"os"
	"strings"
	"testing"
	"time"
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

func TestStaticDomainUsesCacheAndUnorderedLookup(t *testing.T) {
	src, err := os.ReadFile("static.go")
	if err != nil {
		t.Fatalf("read static.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"getStaticDomainCache",
		"setStaticDomainCache",
		"InvalidateStaticDomainCache",
		"Take(&setup)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("StaticDomain must include %q", snippet)
		}
	}
	if strings.Contains(text, "First(&setup)") {
		t.Fatalf("StaticDomain should use Take to avoid unnecessary ORDER BY id")
	}
}

func TestFullURLWithStaticDomainSkipsLookupWhenDomainUnneeded(t *testing.T) {
	src, err := os.ReadFile("static.go")
	if err != nil {
		t.Fatalf("read static.go: %v", err)
	}
	body := string(src)
	if start := strings.Index(body, "func FullURLWithStaticDomain"); start >= 0 {
		body = body[start:]
	}
	for _, snippet := range []string{
		`if path == ""`,
		`strings.HasPrefix(path, "http://")`,
		`strings.HasPrefix(path, "https://")`,
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("FullURLWithStaticDomain should skip StaticDomain lookup with %s", snippet)
		}
	}
}

func TestStaticDomainCacheCopiesExpiresAndInvalidates(t *testing.T) {
	InvalidateStaticDomainCache()
	now := time.Now()
	if _, ok := getStaticDomainCache(now); ok {
		t.Fatalf("empty static domain cache should miss")
	}

	setStaticDomainCache("https://cdn.local", now)
	got, ok := getStaticDomainCache(now.Add(staticDomainCacheTTL / 2))
	if !ok || got != "https://cdn.local" {
		t.Fatalf("expected cached static domain before ttl, got domain=%q ok=%v", got, ok)
	}
	if _, ok := getStaticDomainCache(now.Add(staticDomainCacheTTL + time.Second)); ok {
		t.Fatalf("expired static domain cache should miss")
	}

	setStaticDomainCache("https://cdn.local", now)
	InvalidateStaticDomainCache()
	if _, ok := getStaticDomainCache(now); ok {
		t.Fatalf("invalidated static domain cache should miss")
	}
}
