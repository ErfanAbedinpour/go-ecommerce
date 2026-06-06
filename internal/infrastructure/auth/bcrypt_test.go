package auth

import (
	"testing"
)

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	h := NewPasswordHasher()
	password := "Admin@123456"

	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == password {
		t.Error("hash should not equal plaintext password")
	}
	if !h.Verify(hash, password) {
		t.Error("Verify() should return true for correct password")
	}
	if h.Verify(hash, "wrong-password") {
		t.Error("Verify() should return false for wrong password")
	}
}
