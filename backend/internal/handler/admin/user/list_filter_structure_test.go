package user

import (
	"os"
	"strings"
	"testing"
)

func TestUserListHandlerParsesStructuredFilters(t *testing.T) {
	source, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read user handler: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		`Mobile:`,
		`strings.TrimSpace(c.Query("mobile"))`,
		`parseOptionalUserStatus(c.Query("status"))`,
		`parseOptionalUserFilterID(c.Query("positionId"))`,
		`parseOptionalUserFilterID(c.Query("managerUserId"))`,
		`parseOptionalUserFilterID(c.Query("roleId"))`,
		"GetUserListByFilterContext",
		`response.Fail(c, "筛选参数错误")`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list handler missing filter contract %q", snippet)
		}
	}
}

func TestParseOptionalUserStatus(t *testing.T) {
	for _, test := range []struct {
		input string
		want  *int
		ok    bool
	}{
		{input: "", want: nil, ok: true},
		{input: "0", want: intPointer(0), ok: true},
		{input: "2", want: intPointer(2), ok: true},
		{input: "9", want: intPointer(9), ok: true},
		{input: "3", want: nil, ok: false},
		{input: "invalid", want: nil, ok: false},
	} {
		got, ok := parseOptionalUserStatus(test.input)
		if ok != test.ok || !equalOptionalInt(got, test.want) {
			t.Fatalf("parse status %q = %#v, %v; want %#v, %v", test.input, got, ok, test.want, test.ok)
		}
	}
}

func TestParseOptionalUserFilterID(t *testing.T) {
	for _, test := range []struct {
		input string
		want  uint
		ok    bool
	}{
		{input: "", want: 0, ok: true},
		{input: " 12 ", want: 12, ok: true},
		{input: "0", want: 0, ok: false},
		{input: "-1", want: 0, ok: false},
		{input: "invalid", want: 0, ok: false},
	} {
		got, ok := parseOptionalUserFilterID(test.input)
		if got != test.want || ok != test.ok {
			t.Fatalf("parse filter ID %q = %d, %v; want %d, %v", test.input, got, ok, test.want, test.ok)
		}
	}
}

func intPointer(value int) *int { return &value }

func equalOptionalInt(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
