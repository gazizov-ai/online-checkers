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

	if whitePlayerID == blackPlayerID {
		return domain.GameFinishedEvent{}, fmt.Errorf("white_player_id and black_player_id must be different")
	}

	result := domain.GameResult(pbEvent.Result)
	if result == "" {
		result = legacyResultFromWinnerID(pbEvent.WinnerId, whitePlayerID, blackPlayerID)
	}

	if err := validateGameResult(result); err != nil {
		return domain.GameFinishedEvent{}, err
	}

	reason := domain.FinishReason(pbEvent.Reason)
	if reason == "" {
		reason = domain.FinishReasonCheckersRules
	}

	if err := validateFinishReason(reason); err != nil {
		return domain.GameFinishedEvent{}, err
	}

	winnerID, err := parseOptionalWinnerID(pbEvent.WinnerId)
	if err != nil {
		return domain.GameFinishedEvent{}, err
	}

	if err := validateResultWinnerConsistency(result, winnerID, whitePlayerID, blackPlayerID); err != nil {
		return domain.GameFinishedEvent{}, err
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
		Result:        result,
		Reason:        reason,
		FinishedAt:    pbEvent.FinishedAt.AsTime(),
	}, nil
}

func parseOptionalWinnerID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}

	winnerID, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse winner_id: %w", err)
	}

	return &winnerID, nil
}

func validateGameResult(result domain.GameResult) error {
	switch result {
	case domain.GameResultWhiteWin, domain.GameResultBlackWin, domain.GameResultDraw:
		return nil
	default:
		return fmt.Errorf("invalid result: %q", result)
	}
}

func validateFinishReason(reason domain.FinishReason) error {
	switch reason {
	case domain.FinishReasonCheckersRules,
		domain.FinishReasonResignation,
		domain.FinishReasonDrawAgreement:
		return nil
	default:
		return fmt.Errorf("invalid finish_reason: %q", reason)
	}
}

func validateResultWinnerConsistency(
	result domain.GameResult,
	winnerID *uuid.UUID,
	whitePlayerID uuid.UUID,
	blackPlayerID uuid.UUID,
) error {
	switch result {
	case domain.GameResultDraw:
		if winnerID != nil {
			return fmt.Errorf("draw must not have winner_id")
		}
		return nil

	case domain.GameResultWhiteWin:
		if winnerID == nil {
			return fmt.Errorf("white_win requires winner_id")
		}
		if *winnerID != whitePlayerID {
			return fmt.Errorf("white_win winner_id must equal white_player_id")
		}
		return nil

	case domain.GameResultBlackWin:
		if winnerID == nil {
			return fmt.Errorf("black_win requires winner_id")
		}
		if *winnerID != blackPlayerID {
			return fmt.Errorf("black_win winner_id must equal black_player_id")
		}
		return nil

	default:
		return fmt.Errorf("invalid result: %q", result)
	}
}

func legacyResultFromWinnerID(
	winnerIDRaw string,
	whitePlayerID uuid.UUID,
	blackPlayerID uuid.UUID,
) domain.GameResult {
	winnerID, err := uuid.Parse(winnerIDRaw)
	if err != nil {
		return ""
	}

	switch winnerID {
	case whitePlayerID:
		return domain.GameResultWhiteWin
	case blackPlayerID:
		return domain.GameResultBlackWin
	default:
		return ""
	}
}
