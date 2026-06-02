package repository

import (
	"time"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/google/uuid"
)

type userRow struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	Email        *string   `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

func userRowToDomain(row userRow) domain.User {
	return domain.User{
		ID:           row.ID,
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}

func userRowToDomainPtr(row userRow) *domain.User {
	user := userRowToDomain(row)
	return &user
}
