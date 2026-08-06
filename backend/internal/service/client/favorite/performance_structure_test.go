package favorite

import (
	"os"
	"strings"
	"testing"
)

func TestFavoriteListBatchesReferencedEnrolls(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"db.Where(\"id = ?\", f.OID).First(&enroll)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("favorite list should not fetch referenced objects one by one: %q", snippet)
		}
	}
	required := []string{
		"favoriteOIDs(favs)",
		"Where(\"`id` IN ?\"",
		"enrollByID",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("favorite list should batch load referenced objects with %q", snippet)
		}
	}
}
