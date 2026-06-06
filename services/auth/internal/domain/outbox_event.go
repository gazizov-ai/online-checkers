package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"

	OutboxAggregateUser = "user"

	EventTypeUserRegistered = "user.registered"
)

type OutboxEvent struct {
	ID            uuid.UUID
	EventType     string
	AggregateType string
	AggregateID   uuid.UUID
	Topic         string
	KafkaKey      string
	Payload       []byte
	Headers       map[string]string
	Status        string
	Attempts      int
	LastError     *string
	CreatedAt     time.Time
	PublishedAt   *time.Time
}
