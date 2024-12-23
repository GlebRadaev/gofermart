package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	jwtService := &JWTService{}

	tests := []struct {
		name           string
		userID         int
		expirationTime time.Time
		expectError    bool
	}{
		{
			name:           "Valid Token",
			userID:         123,
			expirationTime: time.Now().Add(time.Hour),
			expectError:    false,
		},
		{
			name:           "Expired Token",
			userID:         123,
			expirationTime: time.Now().Add(-time.Hour),
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtService.GenerateJWT(tt.userID, tt.expirationTime)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	jwtService := &JWTService{}

	tests := []struct {
		name        string
		tokenString string
		setup       func() string
		expectError bool
	}{
		{
			name: "Valid Token",
			setup: func() string {
				token, _ := jwtService.GenerateJWT(123, time.Now().Add(time.Hour))
				return token
			},
			expectError: false,
		},
		{
			name:        "Invalid Token",
			tokenString: "invalid.token.string",
			expectError: true,
		},
		{
			name: "Expired Token",
			setup: func() string {
				token, _ := jwtService.GenerateJWT(123, time.Now().Add(-time.Hour))
				return token
			},
			expectError: true,
		},
		{
			name: "Invalid Claims Type",
			setup: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					Issuer:    "gofermart",
				})
				signedToken, _ := token.SignedString([]byte("your-secret-key"))
				return signedToken
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string
			if tt.setup != nil {
				tokenString = tt.setup()
			} else {
				tokenString = tt.tokenString
			}

			claims, err := jwtService.ValidateToken(tokenString)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
