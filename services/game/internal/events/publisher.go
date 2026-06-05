package events

import (
	"context"

	"github.com/gazizov-ai/online-checkers/gen/events/v1"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	gameservice "github.com/gazizov-ai/online-checkers/services/game/internal/service"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GameFinishedPublisher struct {
	producer *appkafka.Producer
}

func NewGameFinishedPublisher(producer *appkafka.Producer) *GameFinishedPublisher {
	return &GameFinishedPublisher{
		producer: producer,
	}
}

func (p *GameFinishedPublisher) PublishGameFinished(
	ctx context.Context,
	event gameservice.GameFinishedEvent,
) error {
	msg := &eventsv1.GameFinished{
		EventId:       event.EventID.String(),
		GameId:        event.GameID.String(),
		WhitePlayerId: event.WhitePlayerID.String(),
		BlackPlayerId: event.BlackPlayerID.String(),
		WinnerId:      event.WinnerID.String(),
		FinishedAt:    timestamppb.New(event.FinishedAt),
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return p.producer.Publish(
		ctx,
		[]byte(event.GameID.String()),
		payload,
		map[string]string{
			"event_type":     "game.finished",
			"content_type":   "application/protobuf",
			"schema":         "online_checkers.events.v1.GameFinished",
			"schema_version": "v1",
		},
	)
}
