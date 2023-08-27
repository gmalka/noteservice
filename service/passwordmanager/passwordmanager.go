package passwordmanager

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(verifiable, wanted string) error
}

type passwordHasher struct {
}

func NewPasswordHasher() PasswordHasher {
	return passwordHasher{}
}

func (p passwordHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("can't hash password: %v", err)
	}

	return string(hashedPassword), nil
}

func (p passwordHasher) CheckPassword(verifiable, wanted string) error {
	err := bcrypt.CompareHashAndPassword([]byte(wanted), []byte(verifiable))
	if err != nil {
		return fmt.Errorf("invalid password: %v", err)
	}

	return nil
}
