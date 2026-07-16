package randutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// HexString returns a cryptographically secure lowercase hex string.
func HexString(length int) string {
	if length <= 0 || length%2 != 0 {
		panic(fmt.Sprintf("HexString: length must be a positive even number, got %d", length))
	}
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("HexString: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}
