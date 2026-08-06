package position

import (
	"os"
	"strings"
	"testing"
)

func TestPositionServiceProtectsUserReferencedPositions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"GetListContext",
		"AddContext",
		"EditContext",
		"DeleteContext",
		"model.User{}",
		"`user_position_id` = ?",
		"岗位已被用户使用，不能删除",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("position service missing required snippet %s", snippet)
		}
	}
}
