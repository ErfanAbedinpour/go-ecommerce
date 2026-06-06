package auth

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// PasswordHasher provides bcrypt password hashing and verification.
type PasswordHasher struct{}

// NewPasswordHasher creates a new PasswordHasher.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// Hash generates a bcrypt hash from a plaintext password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify checks a plaintext password against a bcrypt hash.
func (h *PasswordHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
