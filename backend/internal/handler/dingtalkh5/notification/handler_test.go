package notification

import "testing"

func TestParseNotificationID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  uint
		ok    bool
	}{
		{name: "valid", value: "42", want: 42, ok: true},
		{name: "trim spaces", value: " 7 ", want: 7, ok: true},
		{name: "zero", value: "0", ok: false},
		{name: "negative", value: "-1", ok: false},
		{name: "text", value: "abc", ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseNotificationID(test.value)
			if got != test.want || ok != test.ok {
				t.Fatalf("parseNotificationID(%q) = (%d, %v), want (%d, %v)", test.value, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestParseUnreadOnly(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "1", want: true},
		{value: "true", want: true},
		{value: "TRUE", want: true},
		{value: "0", want: false},
		{value: "false", want: false},
		{value: "", want: false},
	}
	for _, test := range tests {
		if got := parseUnreadOnly(test.value); got != test.want {
			t.Fatalf("parseUnreadOnly(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}
