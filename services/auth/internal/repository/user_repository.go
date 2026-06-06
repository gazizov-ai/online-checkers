package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

func mapUserInsertError(err error) error {
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

	return nil
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
		if mappedErr := mapUserInsertError(err); mappedErr != nil {
			return mappedErr
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

func (r *PostgresUserRepository) CreateUserWithOutboxEvent(
	ctx context.Context,
	user domain.User,
	event domain.OutboxEvent,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const createUserQuery = `
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

	userRow := userRow{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}

	if _, err := tx.NamedExecContext(ctx, createUserQuery, userRow); err != nil {
		if mappedErr := mapUserInsertError(err); mappedErr != nil {
			return mappedErr
		}

		return fmt.Errorf("insert user: %w", err)
	}

	headersJSON, err := json.Marshal(event.Headers)
	if err != nil {
		return fmt.Errorf("marshal outbox headers: %w", err)
	}

	const createOutboxEventQuery = `
		INSERT INTO auth_outbox_events (
			id,
			event_type,
			aggregate_type,
			aggregate_id,
			topic,
			kafka_key,
			payload,
			headers,
			status,
			attempts,
			created_at
		)
		VALUES (
			:id,
			:event_type,
			:aggregate_type,
			:aggregate_id,
			:topic,
			:kafka_key,
			:payload,
			:headers,
			:status,
			:attempts,
			:created_at
		)
	`

	outboxRow := outboxEventRow{
		ID:            event.ID,
		EventType:     event.EventType,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Topic:         event.Topic,
		KafkaKey:      event.KafkaKey,
		Payload:       event.Payload,
		Headers:       headersJSON,
		Status:        domain.OutboxStatusPending,
		Attempts:      0,
		CreatedAt:     event.CreatedAt,
	}

	if _, err := tx.NamedExecContext(ctx, createOutboxEventQuery, outboxRow); err != nil {
		return fmt.Errorf("insert auth outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create user with outbox event: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) ListPendingOutboxEvents(
	ctx context.Context,
	limit int,
) ([]domain.OutboxEvent, error) {
	const query = `
		SELECT
			id,
			event_type,
			aggregate_type,
			aggregate_id,
			topic,
			kafka_key,
			payload,
			headers,
			status,
			attempts,
			last_error,
			created_at,
			published_at
		FROM auth_outbox_events
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	var rows []outboxEventRow
	if err := r.db.SelectContext(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("list pending auth outbox events: %w", err)
	}

	events := make([]domain.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		event, err := outboxEventRowToDomain(row)
		if err != nil {
			return nil, fmt.Errorf("map auth outbox event row: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *PostgresUserRepository) MarkOutboxEventPublished(
	ctx context.Context,
	eventID uuid.UUID,
) error {
	const query = `
		UPDATE auth_outbox_events
		SET
			status = 'published',
			published_at = now(),
			last_error = NULL
		WHERE id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, eventID); err != nil {
		return fmt.Errorf("mark auth outbox event published: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) MarkOutboxEventFailedAttempt(
	ctx context.Context,
	eventID uuid.UUID,
	errMessage string,
) error {
	const query = `
		UPDATE auth_outbox_events
		SET
			attempts = attempts + 1,
			last_error = $2
		WHERE id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, eventID, errMessage); err != nil {
		return fmt.Errorf("mark auth outbox event failed attempt: %w", err)
	}

	return nil
}
