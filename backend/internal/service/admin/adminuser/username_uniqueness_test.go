package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestNormalizeAdminUserName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{name: "trim surrounding spaces", input: "  Nick  ", want: "Nick"},
		{name: "keep internal spaces", input: "Nick Admin", want: "Nick Admin"},
		{name: "reject blank", input: " \t\n ", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAdminUserName(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalize username: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSameNormalizedAdminUserName(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{name: "same value", left: "Nick", right: "Nick", want: true},
		{name: "ignore surrounding spaces", left: " Nick ", right: "Nick", want: true},
		{name: "ignore letter case", left: "NICK", right: "nick", want: true},
		{name: "different value", left: "Nick", right: "Foster", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameNormalizedAdminUserName(tt.left, tt.right); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdminUserWritePathsCheckGlobalUserNameUniqueness(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	if !strings.Contains(text, "func ensureUserNameAvailableTx(tx *gorm.DB, name string, excludeUserID uint) error") {
		t.Fatal("user service must provide a shared transactional username uniqueness check")
	}
	if !strings.Contains(text, "用户名已存在，请更换后重试") {
		t.Fatal("duplicate usernames must return an actionable user-facing error")
	}
	if !strings.Contains(text, "sameNormalizedAdminUserName(current.Name, name)") {
		t.Fatal("editing a user without changing its normalized username must remain allowed")
	}
	if count := strings.Count(text, "ensureUserNameAvailableTx(tx, normalizedName,"); count < 3 {
		t.Fatalf("add and both edit paths must check username uniqueness, got %d checks", count)
	}
	if strings.Contains(text, "userVisibleQueryContext(ctx, tx, adminID).Where(\"`user_name`") {
		t.Fatal("username uniqueness must not be limited by the administrator data scope")
	}
}
