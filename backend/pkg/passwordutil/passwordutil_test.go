package passwordutil

import "testing"

func TestHashPasswordUsesBcryptAndVerifies(t *testing.T) {
	hash, err := Hash("secret123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "" || hash == "secret123" {
		t.Fatalf("hash must be non-empty and not equal to plaintext")
	}
	if !Verify(hash, "secret123") {
		t.Fatalf("bcrypt hash should verify the original password")
	}
	if Verify(hash, "wrong-password") {
		t.Fatalf("bcrypt hash must reject a wrong password")
	}
	if NeedsRehash(hash) {
		t.Fatalf("fresh bcrypt hash should not need rehash")
	}
}

func TestVerifySupportsLegacyMD5AndRequiresRehash(t *testing.T) {
	const legacyMD5 = "e10adc3949ba59abbe56e057f20f883e" // 123456
	if !Verify(legacyMD5, "123456") {
		t.Fatalf("legacy MD5 hash should verify during migration window")
	}
	if Verify(legacyMD5, "wrong-password") {
		t.Fatalf("legacy MD5 hash must reject a wrong password")
	}
	if !NeedsRehash(legacyMD5) {
		t.Fatalf("legacy MD5 hash should need bcrypt rehash")
	}
}
