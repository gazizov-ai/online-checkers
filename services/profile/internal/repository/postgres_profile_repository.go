package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresProfileRepository struct {
	db *sqlx.DB
}

func NewPostgresProfileRepository(db *sqlx.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{db: db}
}

type profileRow struct {
	UserID      uuid.UUID `db:"user_id"`
	Username    string    `db:"username"`
	DisplayName *string   `db:"display_name"`
	CountryCode *string   `db:"country_code"`
	AvatarURL   *string   `db:"avatar_url"`
	Bio         *string   `db:"bio"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

var _ ProfileRepository = (*PostgresProfileRepository)(nil)

func profileFromRow(row profileRow) domain.Profile {
	return domain.Profile{
		UserID:      row.UserID,
		Username:    row.Username,
		DisplayName: row.DisplayName,
		CountryCode: row.CountryCode,
		AvatarURL:   row.AvatarURL,
		Bio:         row.Bio,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (r *PostgresProfileRepository) GetProfileByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Profile, error) {
	const query = `
		SELECT 
			user_id,
			username,
			display_name,
			country_code,
			avatar_url,
			bio,
			created_at,
			updated_at
		FROM profiles
		WHERE user_id = $1
	`

	var row profileRow
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, ErrProfileNotFound
		}

		return domain.Profile{}, err
	}

	return profileFromRow(row), nil
}

func (r *PostgresProfileRepository) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	input domain.UpdateProfileInput,
) (domain.Profile, error) {
	const query = `
		UPDATE profiles
		SET
			display_name = COALESCE($2, display_name),
			country_code = COALESCE($3, country_code),
			avatar_url = COALESCE($4, avatar_url),
			bio = COALESCE($5, bio),
			updated_at = now()
		WHERE user_id = $1
		RETURNING
			user_id,
			username,
			display_name,
			country_code,
			avatar_url,
			bio,
			created_at,
			updated_at
	`

	var row profileRow
	if err := r.db.GetContext(
		ctx,
		&row,
		query,
		userID,
		input.DisplayName,
		input.CountryCode,
		input.AvatarURL,
		input.Bio,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, ErrProfileNotFound
		}

		return domain.Profile{}, err
	}

	return profileFromRow(row), nil
}

func (r *PostgresProfileRepository) GetProfilesByUserIDs(
	ctx context.Context,
	userIDs []uuid.UUID,
) ([]domain.Profile, error) {
	if len(userIDs) == 0 {
		return []domain.Profile{}, nil
	}

	const query = `
		SELECT
			user_id,
			username,
			display_name,
			country_code,
			avatar_url,
			bio,
			created_at,
			updated_at
		FROM profiles
		WHERE user_id = ANY($1)
	`

	var rows []profileRow
	if err := r.db.SelectContext(ctx, &rows, query, pq.Array(userIDs)); err != nil {
		return nil, err
	}

	profiles := make([]domain.Profile, 0, len(rows))
	for _, row := range rows {
		profiles = append(profiles, profileFromRow(row))
	}

	return profiles, nil
}

func (r *PostgresProfileRepository) CreateProfileFromUserRegistered(
	ctx context.Context,
	event domain.UserRegisteredEvent,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertEventQuery = `
		INSERT INTO profile_processed_events (event_id)
		VALUES ($1)
		ON CONFLICT (event_id) DO NOTHING
	`

	result, err := tx.ExecContext(ctx, insertEventQuery, event.EventID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return tx.Commit()
	}

	const insertProfileQuery = `
		INSERT INTO profiles (
			user_id,
			username,
			display_name,
			country_code,
			avatar_url,
			bio,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			NULL,
			NULL,
			NULL,
			NULL,
			$3,
			$3
		)
		ON CONFLICT (user_id) DO NOTHING
	`

	if _, err := tx.ExecContext(
		ctx,
		insertProfileQuery,
		event.UserID,
		event.Username,
		event.RegisteredAt,
	); err != nil {
		return err
	}

	return tx.Commit()
}
