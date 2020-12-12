package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// IPassword - password manager interface.
type IPassword interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) error
}

// PasswordManager - IPassword implementation.
type PasswordManager struct{}

// HashPassword - hashes given password (salt is used for improving security).
func (p *PasswordManager) HashPassword(password string) (string, error) {
	// bcrypt automatically handles hash generating.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// CheckPassword - checks if hash of given password matches the given hash.
func (p *PasswordManager) CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
