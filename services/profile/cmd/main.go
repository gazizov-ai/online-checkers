package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	profileconsumer "github.com/gazizov-ai/online-checkers/services/profile/internal/consumer"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/handler"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
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

	profileRepo := repository.NewPostgresProfileRepository(database)
	profileService := service.NewProfileService(profileRepo)
	profileHandler := handler.NewProfileHandler(profileService)

	userRegisteredHandler := profileconsumer.NewUserRegisteredHandler(profileService)

	kafkaConsumer := appkafka.NewConsumer(appkafka.ConsumerConfig{
		Brokers: strings.Split(cfg.KafkaBrokers, ","),
		Topic:   cfg.UserRegisteredTopic,
		GroupID: cfg.ProfileConsumerGroup,
	})
	defer kafkaConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if err := kafkaConsumer.Run(ctx, func(ctx context.Context, msg appkafka.Message) error {
				return userRegisteredHandler.Handle(ctx, msg.Value, msg.Headers)
			}); err != nil {
				if ctx.Err() != nil {
					return
				}

				log.Printf("profile user.registered consumer stopped: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	publicKey, err := appjwt.LoadRSAPublicKeyFromJWKS(
		context.Background(),
		cfg.JWKSURL,
		cfg.JWTKeyID,
	)
	if err != nil {
		log.Fatalf("load jwks: %v", err)
	}

	verifier := appjwt.NewRS256Verifier(
		publicKey,
		cfg.JWTKeyID,
		cfg.OIDCIssuer,
		cfg.OIDCAudience,
	)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
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
			_ = httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "database can't be reached")
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ready",
			"service": cfg.ServiceName,
		})
	})

	r.Route("/api/v1/profiles", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(appjwt.Middleware(verifier))

			r.Get("/me", profileHandler.GetMyProfile)
			r.Patch("/me", profileHandler.UpdateMyProfile)
		})

		r.Post("/batch", profileHandler.BatchProfiles)
		r.Get("/{user_id}", profileHandler.GetProfileByUserID)
	})

	addr := ":" + cfg.HTTPPort
	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
