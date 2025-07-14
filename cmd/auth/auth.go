package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateToken(userID uuid.UUID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"expires_at": time.Now().Add(time.Hour * 1).Unix(),
		"created_at": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString([]byte(secret))
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

func ValidateToken(token, secret string) (string, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method used: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	exp, ok := claims["expires_at"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid expiration time")
	}

	if time.Now().Unix() > int64(exp) {
		return "", fmt.Errorf("token expired")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID")
	}

	return userID, nil
}

func CreateAccessTokenID(accessToken string) string {
	hash := sha256.Sum256([]byte(accessToken))
	return hex.EncodeToString(hash[:])
}

func GetClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
