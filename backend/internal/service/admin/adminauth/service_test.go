package adminauth

import "testing"

func TestGenRandomStringReturnsRequestedLength(t *testing.T) {
	got := genRandomString(32)

	if len(got) != 32 {
		t.Fatalf("genRandomString length = %d, want 32", len(got))
	}
	for _, r := range got {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("genRandomString returned non-hex character %q in %q", r, got)
		}
	}
}
