package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Şifre hash'leme maliyeti
const BcryptCost = 12

// HashPassword şifre hash'leme
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword şifre doğrulama
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
