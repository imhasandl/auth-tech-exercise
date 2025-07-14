package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateToken(userID string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"expires_at": time.Now().Add(time.Hour * 1).Unix(),
		"created_at": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CreateRefreshToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", "", err
	}

	base64Token := base64.StdEncoding.EncodeToString(tokenBytes)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(base64Token), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return base64Token, string(hashedToken), nil
}

func CheckRefreshToken(providedToken, hashedToken string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(providedToken))
}
