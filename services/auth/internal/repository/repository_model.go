package repository

import (
	"encoding/json"
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

type outboxEventRow struct {
	ID            uuid.UUID  `db:"id"`
	EventType     string     `db:"event_type"`
	AggregateType string     `db:"aggregate_type"`
	AggregateID   uuid.UUID  `db:"aggregate_id"`
	Topic         string     `db:"topic"`
	KafkaKey      string     `db:"kafka_key"`
	Payload       []byte     `db:"payload"`
	Headers       []byte     `db:"headers"`
	Status        string     `db:"status"`
	Attempts      int        `db:"attempts"`
	LastError     *string    `db:"last_error"`
	CreatedAt     time.Time  `db:"created_at"`
	PublishedAt   *time.Time `db:"published_at"`
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

func outboxEventRowToDomain(row outboxEventRow) (domain.OutboxEvent, error) {
	var headers map[string]string
	if len(row.Headers) > 0 {
		if err := json.Unmarshal(row.Headers, &headers); err != nil {
			return domain.OutboxEvent{}, err
		}
	}

	if headers == nil {
		headers = map[string]string{}
	}

	return domain.OutboxEvent{
		ID:            row.ID,
		EventType:     row.EventType,
		AggregateType: row.AggregateType,
		AggregateID:   row.AggregateID,
		Topic:         row.Topic,
		KafkaKey:      row.KafkaKey,
		Payload:       row.Payload,
		Headers:       headers,
		Status:        row.Status,
		Attempts:      row.Attempts,
		LastError:     row.LastError,
		CreatedAt:     row.CreatedAt,
		PublishedAt:   row.PublishedAt,
	}, nil
}
