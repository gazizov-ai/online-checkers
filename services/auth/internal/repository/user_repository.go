package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	const query = `
		INSERT INTO users (
			id,
			username,
			email,
			password_hash,
			created_at
		)
		VALUES (
			:id,
			:username,
			:email,
			:password_hash,
			:created_at
		)
	`

	row := userRow{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}

	_, err := r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				switch pqErr.Constraint {
				case "users_username_key":
					return ErrUsernameAlreadyExists
				case "users_email_key":
					return ErrEmailAlreadyExists
				}
			}
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT 
			id,
			username,
			email,
			password_hash,
			created_at
		FROM users
		WHERE id = $1
	`

	var row userRow

	err := r.db.GetContext(ctx, &row, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return userRowToDomainPtr(row), nil
}

func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	const query = `
		SELECT 
			id,
			username,
			email,
			password_hash,
			created_at
		FROM users
		WHERE username = $1
	`

	var row userRow

	err := r.db.GetContext(ctx, &row, query, username)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return userRowToDomainPtr(row), nil
}
