package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-tech-exercise/cmd/auth"
	"github.com/imhasandl/auth-tech-exercise/database"
	"github.com/imhasandl/auth-tech-exercise/utils"
	"github.com/imhasandl/auth-tech-exercise/webhook"
)

type Config struct {
	db          *database.DB
	tokenSecret string
	webhookUrl  string
}

func NewConfig(db *database.DB, tokenSecret, webhook_url string) *Config {
	return &Config{
		db:          db,
		tokenSecret: tokenSecret,
		webhookUrl:  webhook_url,
	}
}

func (cfg *Config) Register(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Email    string
		Password string
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't decode body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(requestData.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't hash password", err)
		return
	}

	user, err := cfg.db.RegisterUser(database.RegisterParams{
		ID:       uuid.New(),
		Email:    requestData.Email,
		Password: hashedPassword,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't register user", err)
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't generate auth token", err)
		return
	}

	base64Token, hashedToken, err := auth.CreateRefreshToken()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't create refresh token", err)
		return
	}

	accessTokenID := auth.CreateAccessTokenID(accessToken)
	userAgent := r.Header.Get("User-Agent")
	ipAddress := auth.GetClientIP(r)

	err = cfg.db.SaveRefreshToken(database.SaveRefreshTokenParams{
		ID:            uuid.New(),
		UserID:        user.ID,
		Token:         hashedToken,
		ExpiresAt:     time.Now().Add(time.Hour * 7 * 24),
		AccessTokenID: accessTokenID,
		UserAgent:     userAgent,
		IPAddress:     ipAddress,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't save refresh token", err)
		return
	}

	response := map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": base64Token,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *Config) GetUserID(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Bearer" {
		utils.RespondWithError(w, http.StatusUnauthorized, "no authorization header provided", nil)
		return
	}

	token := authHeader[1]

	userID, err := auth.ValidateToken(token, cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	response := map[string]string{
		"user_id": userID,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *Config) Login(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Email    string
		Password string
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't decode body", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(requestData.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't get user using email", err)
		return
	}

	err = auth.CheckPassword(user.Password, requestData.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "incorrect password", err)
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't generate token", err)
		return
	}

	base64Token, hashedToken, err := auth.CreateRefreshToken()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't create refresh token", err)
		return
	}

	accessTokenID := auth.CreateAccessTokenID(accessToken)
	userAgent := r.Header.Get("User-Agent")
	ipAddress := auth.GetClientIP(r)

	err = cfg.db.SaveRefreshToken(database.SaveRefreshTokenParams{
		ID:            uuid.New(),
		UserID:        user.ID,
		Token:         hashedToken,
		ExpiresAt:     time.Now().Add(time.Hour * 7 * 24),
		AccessTokenID: accessTokenID,
		UserAgent:     userAgent,
		IPAddress:     ipAddress,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't save refresh token", err)
		return
	}

	response := map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": base64Token,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *Config) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		RefreshToken string
		AccessToken  string
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't decode body", err)
		return
	}

	token, err := cfg.db.GetRefreshToken(requestData.RefreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid or expired refresh token", err)
		return
	}

	if time.Now().After(token.ExpiresAt) {
		utils.RespondWithError(w, http.StatusUnauthorized, "refresh token expired", nil)
		return
	}

	accessTokenID := auth.CreateAccessTokenID(requestData.AccessToken)
	if token.AccessTokenID != accessTokenID {
		utils.RespondWithError(w, http.StatusUnauthorized, "token pair mismatch", nil)
		return
	}

	currentUserAgent := r.Header.Get("User-Agent")
	if token.UserAgent != currentUserAgent {
		err = cfg.db.DeleteRefreshTokensByUserID(token.UserID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to deauthorize user", err)
			return
		}
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized: user-Agent changed", nil)
		return
	}

	currentIP := auth.GetClientIP(r)
	if token.IPAddress != currentIP {
		webhookURL := os.Getenv("WEBHOOK_URL")
		if webhookURL != "" {
			notificationData := map[string]interface{}{
				"user_id":   token.UserID.String(),
				"event":     "ip_change",
				"old_ip":    token.IPAddress,
				"new_ip":    currentIP,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			webhook.SendWebhookNotification(webhookURL, notificationData)
		}
	}

	err = cfg.db.DeleteRefreshToken(token.Token)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't delete refresh token", err)
		return
	}

	accessToken, err := auth.GenerateToken(token.UserID, cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't generate access token", err)
		return
	}

	base64Token, hashedToken, err := auth.CreateRefreshToken()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't create refresh token", err)
		return
	}

	newAccessTokenID := auth.CreateAccessTokenID(accessToken)
	err = cfg.db.SaveRefreshToken(database.SaveRefreshTokenParams{
		ID:            uuid.New(),
		UserID:        token.UserID,
		Token:         hashedToken,
		ExpiresAt:     time.Now().Add(time.Hour * 7 * 24),
		AccessTokenID: newAccessTokenID,
		UserAgent:     currentUserAgent,
		IPAddress:     currentIP,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't save refresh token", err)
		return
	}

	response := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": base64Token,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *Config) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Bearer" {
		utils.RespondWithError(w, http.StatusUnauthorized, "no authorization header provided", nil)
		return
	}

	token := authHeader[1]

	userIDStr, err := auth.ValidateToken(token, cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid user_id in token", err)
		return
	}

	err = cfg.db.DeleteRefreshTokensByUserID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "can't delete refresh tokens", err)
		return
	}

	response := map[string]string{
		"message": "successfully logged out",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
