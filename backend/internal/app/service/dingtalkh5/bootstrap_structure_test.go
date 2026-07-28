package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestBootstrapResponseIsLightweight(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	typesText := string(typesSrc)
	reviewsText := string(reviewsSrc)
	bootstrapBody := reviewsText
	if start := strings.Index(reviewsText, "func BootstrapContext"); start >= 0 {
		bootstrapBody = reviewsText[start:]
		if end := strings.Index(bootstrapBody, "\n}\n\nfunc "); end >= 0 {
			bootstrapBody = bootstrapBody[:end+3]
		}
	}

	for _, snippet := range []string{
		"type BootstrapResponse struct",
		"User  UserDTO",
		"Menus []AppMenuDTO",
		"DingTalkH5MenusForUserContext(ctx, user)",
		"if user.RoleID == 0",
		"return dingTalkH5DefaultMenusByRole(user.Role)",
	} {
		if !strings.Contains(typesText+reviewsText, snippet) {
			t.Fatalf("bootstrap response must keep lightweight identity/menu snippet %q", snippet)
		}
	}
	for _, snippet := range []string{
		"Users    []UserDTO",
		"Reviews  []ReviewDTO",
		"Template TemplateDTO",
	} {
		if strings.Contains(typesText, snippet) {
			t.Fatalf("bootstrap response must not embed bulk payload field %q", snippet)
		}
	}
	for _, snippet := range []string{
		"ListUsersContext(ctx, user)",
		"ListReviewsContext(ctx, user",
		"LoadTemplateContext(ctx)",
		"EnsureSeedContext(ctx)",
	} {
		if strings.Contains(bootstrapBody, snippet) {
			t.Fatalf("BootstrapContext must not load bulk data with %q", snippet)
		}
	}
}
