package query

import "testing"

func TestParseSortAllowsOnlyKnownFields(t *testing.T) {
	got := ParseSort("name:desc, id:asc, bad:desc", map[string]string{
		"name": "user_name",
		"id":   "id",
	})
	want := "`user_name` DESC, `id` ASC"
	if got != want {
		t.Fatalf("ParseSort() = %q, want %q", got, want)
	}
}

func TestParseSortDefaultsToAscendingAndIgnoresBlankInput(t *testing.T) {
	allowed := map[string]string{"createdAt": "created_at"}
	if got := ParseSort("createdAt", allowed); got != "`created_at` ASC" {
		t.Fatalf("ParseSort() = %q", got)
	}
	if got := ParseSort("", allowed); got != "" {
		t.Fatalf("empty sort should stay empty, got %q", got)
	}
}
