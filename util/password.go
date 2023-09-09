package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashedPassword - return the bcrypt hash of the password
func HashedPassword(password string) (string, error) {
	// get teh bcrypt of the password by specify the password + cost
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash the password: %w", err)
	}

	return string(hashPassword), nil
}

// CheckPassword - check if the provided password is correct or not
func CheckPassword(password, hashPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}
