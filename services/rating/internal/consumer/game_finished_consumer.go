package consumer

import (
	"context"
	"fmt"
	"log"

	eventsv1 "github.com/gazizov-ai/online-checkers/gen/events/v1"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type RatingService interface {
	ProcessGameFinished(ctx context.Context, event domain.GameFinishedEvent) error
}

type GameFinishedConsumer struct {
	consumer      *appkafka.Consumer
	ratingService RatingService
}

func NewGameFinishedConsumer(
	consumer *appkafka.Consumer,
	ratingService RatingService,
) *GameFinishedConsumer {
	return &GameFinishedConsumer{
		consumer:      consumer,
		ratingService: ratingService,
	}
}

func (c *GameFinishedConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx, c.handleMessage)
}

func (c *GameFinishedConsumer) Close() error {
	return c.consumer.Close()
}

func (c *GameFinishedConsumer) handleMessage(ctx context.Context, msg appkafka.Message) error {
	eventType := msg.Headers["event_type"]
	if eventType != "game.finished" {
		log.Printf("skip message with event_type=%q", eventType)
		return nil
	}

	var pbEvent eventsv1.GameFinished
	if err := proto.Unmarshal(msg.Value, &pbEvent); err != nil {
		log.Printf("skip invalid game.finished protobuf: %v", err)
		return nil
	}

	event, err := mapGameFinishedEvent(&pbEvent)
	if err != nil {
		log.Printf("skip invalid game.finished event: %v", err)
		return nil
	}

	if err := c.ratingService.ProcessGameFinished(ctx, event); err != nil {
		return fmt.Errorf("process game finished: %w", err)
	}

	return nil
}

func mapGameFinishedEvent(pbEvent *eventsv1.GameFinished) (domain.GameFinishedEvent, error) {
	eventID, err := uuid.Parse(pbEvent.EventId)
	if err != nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("parse event_id: %w", err)
	}

	gameID, err := uuid.Parse(pbEvent.GameId)
	if err != nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("parse game_id: %w", err)
	}

	whitePlayerID, err := uuid.Parse(pbEvent.WhitePlayerId)
	if err != nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("parse white_player_id: %w", err)
	}

	blackPlayerID, err := uuid.Parse(pbEvent.BlackPlayerId)
	if err != nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("parse black_player_id: %w", err)
	}

	winnerID, err := uuid.Parse(pbEvent.WinnerId)
	if err != nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("parse winner_id: %w", err)
	}

	if pbEvent.FinishedAt == nil {
		return domain.GameFinishedEvent{}, fmt.Errorf("finished_at is required")
	}

	return domain.GameFinishedEvent{
		EventID:       eventID,
		GameID:        gameID,
		WhitePlayerID: whitePlayerID,
		BlackPlayerID: blackPlayerID,
		WinnerID:      winnerID,
		FinishedAt:    pbEvent.FinishedAt.AsTime(),
	}, nil
}
