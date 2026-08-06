package online

import "testing"

func TestStringSliceToInterfaceKeepsOrderAndValues(t *testing.T) {
	got := stringSliceToInterface([]string{"a", "b"})

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected converted values: %#v", got)
	}
}
