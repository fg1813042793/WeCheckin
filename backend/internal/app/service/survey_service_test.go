package service

import (
	"testing"
)

// 不依赖 DB 的纯函数测试

func TestSplitComma(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"1", []string{"1"}},
		{"1,2,3", []string{"1", "2", "3"}},
		{"1,2,", []string{"1", "2"}},
		{",", []string{""}},
	}
	for _, tt := range tests {
		got := splitComma(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("splitComma(%q) len=%d, want %d", tt.in, len(got), len(tt.want))
			continue
		}
		for i, v := range got {
			if v != tt.want[i] {
				t.Errorf("splitComma(%q)[%d] = %q, want %q", tt.in, i, v, tt.want[i])
			}
		}
	}
}

func TestContains(t *testing.T) {
	if !contains("1,2,3", 2) {
		t.Error("should contain 2")
	}
	if contains("1,2,3", 4) {
		t.Error("should not contain 4")
	}
	if contains("", 1) {
		t.Error("empty should not contain")
	}
}

func TestUserIDToStr(t *testing.T) {
	if userIDToStr(123, true) != "" {
		t.Error("anonymous should be empty")
	}
	if userIDToStr(123, false) != "123" {
		t.Error("non-anonymous should be string id")
	}
	if userIDToStr(0, false) != "0" {
		t.Error("0 should be '0'")
	}
}
