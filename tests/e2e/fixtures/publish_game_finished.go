//go:build e2e

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	eventsv1 "github.com/gazizov-ai/online-checkers/gen/events/v1"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	var brokersRaw string
	var whiteRaw string
	var blackRaw string

	flag.StringVar(&brokersRaw, "brokers", "localhost:19092", "comma-separated Kafka brokers")
	flag.StringVar(&whiteRaw, "white", "", "white player UUID")
	flag.StringVar(&blackRaw, "black", "", "black player UUID")
	flag.Parse()

	white := mustUUID("white", whiteRaw)
	black := mustUUID("black", blackRaw)
	if white == black {
		fatalf("players must be different")
	}

	event := &eventsv1.GameFinished{
		EventId:       uuid.NewString(),
		GameId:        uuid.NewString(),
		WhitePlayerId: white.String(),
		BlackPlayerId: black.String(),
		WinnerId:      white.String(),
		Result:        "white_win",
		Reason:        "resignation",
		FinishedAt:    timestamppb.Now(),
	}
	payload, err := proto.Marshal(event)
	if err != nil {
		fatalf("marshal event: %v", err)
	}

	producer := appkafka.NewProducer(appkafka.ProducerConfig{
		Brokers: splitCSV(brokersRaw),
		Topic:   "game.finished",
	})
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := producer.Publish(ctx, []byte(event.GameId), payload, map[string]string{
		"event_type":     "game.finished",
		"content_type":   "application/protobuf",
		"schema":         "online_checkers.events.v1.GameFinished",
		"schema_version": "v1",
	}); err != nil {
		fatalf("publish event: %v", err)
	}
}

func mustUUID(name, value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		fatalf("invalid %s UUID: %v", name, err)
	}
	return id
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
