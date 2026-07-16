package randutil

import "testing"

func TestHexStringReturnsRequestedLength(t *testing.T) {
	got := HexString(32)
	if len(got) != 32 {
		t.Fatalf("HexString(32) length = %d, want 32", len(got))
	}
	for _, r := range got {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("HexString returned non-hex character %q in %q", r, got)
		}
	}
}

func TestHexStringRejectsInvalidLength(t *testing.T) {
	for _, length := range []int{0, -2, 31} {
		t.Run("length", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("HexString(%d) should panic", length)
				}
			}()
			_ = HexString(length)
		})
	}
}
