package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-tech-exercise/cmd/auth"
	"github.com/imhasandl/auth-tech-exercise/database"
	"github.com/imhasandl/auth-tech-exercise/utils"
)

type Config struct {
	db          *database.DB
	tokenSecret string
}

func NewConfig(db *database.DB, tokenSecret string) *Config {
	return &Config{
		db:          db,
		tokenSecret: tokenSecret,
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

	accessToken, err := auth.GenerateToken(user.ID.String(), cfg.tokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't generate auth token", err)
		return
	}

	base64Token, hashedToken, err := auth.CreateRefreshToken()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "can't create refresh token", err)
		return
	}

	err = cfg.db.SaveRefreshToken(database.SaveRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(time.Hour * 7 * 24),
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

func (cfg *Config) Login(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Email    string
		Password string
	}

	_ = requestData
}
