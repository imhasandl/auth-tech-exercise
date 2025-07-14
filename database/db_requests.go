package database

import (
	"time"

	"github.com/google/uuid"
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
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

type SaveRefreshTokenParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Token  string
	ExpiresAt time.Time
}

func (db *DB) SaveRefreshToken(params SaveRefreshTokenParams) error {
	query := `
	INSERT INTO refresh_token (id, user_id, token, expires_at, created_at)
	VALUES ($1, $2, $3, $4, NOW());`

	_, err := db.Exec(query, params.ID, params.UserID, params.Token, params.ExpiresAt)
	if err != nil {
		return err
	}

	return nil
}
