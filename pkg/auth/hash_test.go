package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	hashService := &HashService{}

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid Password",
			password:    "securepassword",
			expectError: false,
		},
		{
			name:        "Empty Password",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := hashService.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
			}
		})
	}
}

func TestComparePassword(t *testing.T) {
	hashService := &HashService{}

	tests := []struct {
		name           string
		password       string
		hashedPassword string
		setup          func() string
		expectMatch    bool
	}{
		{
			name:     "Matching Password",
			password: "securepassword",
			setup: func() string {
				hashedPassword, _ := hashService.HashPassword("securepassword")
				return hashedPassword
			},
			expectMatch: true,
		},
		{
			name:     "Non-Matching Password",
			password: "wrongpassword",
			setup: func() string {
				hashedPassword, _ := hashService.HashPassword("securepassword")
				return hashedPassword
			},
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hashedPassword string
			if tt.setup != nil {
				hashedPassword = tt.setup()
			} else {
				hashedPassword = tt.hashedPassword
			}

			match := hashService.ComparePassword(hashedPassword, tt.password)
			assert.Equal(t, tt.expectMatch, match)
		})
	}
}
