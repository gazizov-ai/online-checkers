package consumer

import (
	"context"
	"fmt"
	"log"

	eventsv1 "github.com/gazizov-ai/online-checkers/gen/events/v1"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/service"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type UserRegisteredHandler struct {
	service *service.ProfileService
}

func NewUserRegisteredHandler(service *service.ProfileService) *UserRegisteredHandler {
	return &UserRegisteredHandler{service: service}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, value []byte, headers map[string]string) error {
	if headers["event_type"] != "user.registered" {
		log.Printf("profile consumer: skip unknown event_type=%q", headers["event_type"])
		return nil
	}

	var msg eventsv1.UserRegistered
	if err := proto.Unmarshal(value, &msg); err != nil {
		log.Printf("profile consumer: invalid user.registered protobuf: %v", err)
		return nil
	}

	eventID, err := uuid.Parse(msg.EventId)
	if err != nil {
		log.Printf("profile consumer: invalid event_id=%q: %v", msg.EventId, err)
		return nil
	}

	userID, err := uuid.Parse(msg.UserId)
	if err != nil {
		log.Printf("profile consumer: invalid user_id=%q: %v", msg.UserId, err)
		return nil
	}

	if msg.RegisteredAt == nil {
		log.Printf("profile consumer: registered_at is nil event_id=%s", msg.EventId)
		return nil
	}

	event := domain.UserRegisteredEvent{
		EventID:      eventID,
		UserID:       userID,
		Username:     msg.Username,
		RegisteredAt: msg.RegisteredAt.AsTime(),
	}

	if err := h.service.HandleUserRegistered(ctx, event); err != nil {
		return fmt.Errorf("handle user.registered event_id=%s user_id=%s: %w", eventID, userID, err)
	}

	log.Printf("profile created from user.registered: event_id=%s user_id=%s username=%s", eventID, userID, msg.Username)

	return nil
}
