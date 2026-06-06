package outbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	ListPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, eventID uuid.UUID) error
	MarkOutboxEventFailedAttempt(ctx context.Context, eventID uuid.UUID, errMessage string) error
}

type Producer interface {
	Publish(
		ctx context.Context,
		key []byte,
		value []byte,
		headers map[string]string,
	) error
}

type Worker struct {
	repo     Repository
	producer Producer

	batchSize    int
	pollInterval time.Duration
}

func NewWorker(repo Repository, producer Producer) *Worker {
	return &Worker{
		repo:         repo,
		producer:     producer,
		batchSize:    10,
		pollInterval: 1 * time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	log.Println("auth outbox worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("auth outbox worker stopped")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				log.Printf("auth outbox process batch: %v", err)
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) error {
	events, err := w.repo.ListPendingOutboxEvents(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := w.processEvent(ctx, event); err != nil {
			log.Printf(
				"auth outbox event failed: event_id=%s event_type=%s aggregate_id=%s err=%v",
				event.ID,
				event.EventType,
				event.AggregateID,
				err,
			)
		}
	}

	return nil
}

func (w *Worker) processEvent(ctx context.Context, event domain.OutboxEvent) error {
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := w.producer.Publish(
		publishCtx,
		[]byte(event.KafkaKey),
		event.Payload,
		event.Headers,
	); err != nil {
		markErr := w.repo.MarkOutboxEventFailedAttempt(ctx, event.ID, err.Error())
		if markErr != nil {
			return fmt.Errorf("publish failed: %w; mark failed attempt: %v", err, markErr)
		}

		return fmt.Errorf("publish event: %w", err)
	}

	if err := w.repo.MarkOutboxEventPublished(ctx, event.ID); err != nil {
		return fmt.Errorf("mark event published: %w", err)
	}

	log.Printf(
		"auth outbox event published: event_id=%s event_type=%s aggregate_id=%s topic=%s",
		event.ID,
		event.EventType,
		event.AggregateID,
		event.Topic,
	)

	return nil
}
