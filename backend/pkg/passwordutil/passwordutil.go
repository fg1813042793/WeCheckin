package passwordutil

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Verify(stored, plain string) bool {
	if isBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
	}
	if isLegacyMD5(stored) {
		sum := md5.Sum([]byte(plain))
		return hex.EncodeToString(sum[:]) == strings.ToLower(stored)
	}
	return false
}

func NeedsRehash(stored string) bool {
	return !isBcryptHash(stored)
}

func isBcryptHash(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$")
}

func isLegacyMD5(stored string) bool {
	if len(stored) != 32 {
		return false
	}
	for _, r := range stored {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
