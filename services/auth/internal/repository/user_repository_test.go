package repository

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestMapUserInsertError(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		want       error
	}{
		{name: "username", constraint: "users_username_key", want: ErrUsernameAlreadyExists},
		{name: "email", constraint: "users_email_key", want: ErrEmailAlreadyExists},
		{name: "unknown constraint", constraint: "users_pkey", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapUserInsertError(&pq.Error{Code: "23505", Constraint: tt.constraint})
			if !errors.Is(got, tt.want) {
				t.Fatalf("mapUserInsertError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutboxEventRowToDomain(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC()
	headers, _ := json.Marshal(map[string]string{"event_type": "user.registered"})

	got, err := outboxEventRowToDomain(outboxEventRow{
		ID:        id,
		Headers:   headers,
		Payload:   []byte("payload"),
		Status:    domain.OutboxStatusPending,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("outboxEventRowToDomain() error = %v", err)
	}
	if got.ID != id || got.Headers["event_type"] != "user.registered" {
		t.Fatalf("unexpected mapped event: %+v", got)
	}

	if _, err := outboxEventRowToDomain(outboxEventRow{Headers: []byte("{")}); err == nil {
		t.Fatal("expected invalid headers error")
	}
}
