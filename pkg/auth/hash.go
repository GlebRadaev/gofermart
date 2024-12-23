package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type HashServiceInterface interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) bool
}

type HashService struct{}

func (b *HashService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (b *HashService) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
