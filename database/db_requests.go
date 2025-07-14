package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-tech-exercise/cmd/auth"
)

type RegisterParams struct {
	ID       uuid.UUID
	Email    string
	Password string
}

func (db *DB) RegisterUser(params RegisterParams) (User, error) {
	query := `
	INSERT INTO users (id, email, password, created_at)
	VALUES ($1, $2, $3, NOW())
	RETURNING id, email, password, created_at`

	row := db.QueryRow(query, params.ID, params.Email, params.Password)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

type SaveRefreshTokenParams struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Token         string
	ExpiresAt     time.Time
	AccessTokenID string
	UserAgent     string
	IPAddress     string
}

func (db *DB) SaveRefreshToken(params SaveRefreshTokenParams) error {
	query := `
    INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, access_token_id, user_agent, ip_address)
    VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7);`

	_, err := db.Exec(query, params.ID, params.UserID, params.Token, params.ExpiresAt, params.AccessTokenID, params.UserAgent, params.IPAddress)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	query := `
	SELECT id, email, password, created_at
   FROM users
   WHERE email = $1`

	row := db.QueryRow(query, email)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func (db *DB) GetRefreshToken(providedToken string) (RefreshToken, error) {
	rows, err := db.Query(`
    SELECT id, user_id, token, expires_at, created_at, access_token_id, user_agent, ip_address
    FROM refresh_tokens 
    WHERE expires_at > NOW()
		`)
	if err != nil {
		return RefreshToken{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var token RefreshToken
		err := rows.Scan(&token.ID, &token.UserID, &token.Token, &token.ExpiresAt, &token.CreatedAt,
			&token.AccessTokenID, &token.UserAgent, &token.IPAddress)
		if err != nil {
			continue
		}

		if auth.CheckRefreshToken(providedToken, token.Token) == nil {
			return token, nil
		}
	}

	return RefreshToken{}, fmt.Errorf("refresh token not found or expired")
}

func (db *DB) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := db.Exec(query, token)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) DeleteRefreshTokensByUserID(userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := db.Exec(query, userID)
	return err
}
