package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/consumer"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/handler"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer database.Close()

	ratingRepo := repository.NewPostgresRatingRepository(database)
	ratingService := service.NewRatingService(ratingRepo)
	ratingHandler := handler.NewHandler(ratingService)

	brokers := splitCSV(cfg.KafkaBrokers)
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKERS is required")
	}

	gameFinishedKafkaConsumer := appkafka.NewConsumer(appkafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   cfg.GameFinishedTopic,
		GroupID: cfg.RatingConsumerGroup,
	})
	defer func() {
		if err := gameFinishedKafkaConsumer.Close(); err != nil {
			log.Printf("close game finished consumer: %v", err)
		}
	}()

	gameFinishedConsumer := consumer.NewGameFinishedConsumer(
		gameFinishedKafkaConsumer,
		ratingService,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("starting game.finished consumer")

		if err := gameFinishedConsumer.Run(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			log.Printf("game.finished consumer stopped: %v", err)
		}
	}()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	//r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": cfg.ServiceName,
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			_ = httpx.WriteError(
				w,
				http.StatusServiceUnavailable,
				"not_ready",
				"database can't be reached",
			)
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ready",
			"service": cfg.ServiceName,
		})
	})

	r.Get("/api/v1/ratings/{user_id}", ratingHandler.GetRating)
	r.Get("/api/v1/leaderboard", ratingHandler.GetLeaderboard)

	addr := ":" + cfg.HTTPPort

	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
