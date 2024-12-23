package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTServiceInterface interface {
	GenerateJWT(userID int, expirationTime time.Time) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

var secretKey = []byte("your-secret-key")

type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

type JWTService struct{}

func (s *JWTService) GenerateJWT(userID int, expirationTime time.Time) (string, error) {
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "gofermart",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID == 0 || claims.Issuer != "gofermart" {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
